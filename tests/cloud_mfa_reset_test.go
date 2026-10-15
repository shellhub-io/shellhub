package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha1" //nolint:depguard,gosec // RFC 6238 TOTP, which the server checks, is HMAC-SHA1
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	mfaSecret            = "JBSWY3DPEHPK3PXP" //nolint:gosec // the TOTP secret the test registers for its own user
	mfaResetCodeLifetime = 24 * time.Hour
)

var mfaResetCodeLine = regexp.MustCompile(`following code\.(?:\\n|\s)*([A-Z2-7]{5})\b`)

// TestCloudMFAReset runs a cloud instance, the only edition that sends email, for the reset of a
// user's MFA by the two codes it emails to their main and recovery addresses. Each code lasts 24
// hours in the cache, and the reset refuses the pair once either code has expired.
func TestCloudMFAReset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := environment.New(t, run).WithEdition(environment.EditionCloud).Up(ctx)
	t.Cleanup(compose.Down)

	t.Run("the emailed codes reset MFA", func(t *testing.T) {
		user := newMFAUser(t, compose, "mfa-reset")
		codes := requestMFAReset(t, compose, user)

		auth := new(models.UserAuthResponse)
		resp, err := compose.Anonymous(t.Context()).SetBody(codes).SetResult(auth).Put("/api/user/mfa/reset/" + user.id)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.False(t, auth.MFA, "the reset left MFA on")
	})

	t.Run("the emailed codes last 24 hours and the pair is refused once one of them expires", func(t *testing.T) {
		user := newMFAUser(t, compose, "mfa-expired")
		codes := requestMFAReset(t, compose, user)

		for _, address := range []string{"main", "recovery"} {
			key := mfaResetCodeCacheKey(address, user.id)
			left := compose.CacheEntryTTL(t, key)
			assert.Greater(t, left, mfaResetCodeLifetime-5*time.Minute, key)
			assert.LessOrEqual(t, left, mfaResetCodeLifetime, key)
		}

		compose.ExpireCacheEntryIn(t, mfaResetCodeCacheKey("recovery", user.id), 0)

		resp, err := compose.Anonymous(t.Context()).SetBody(codes).Put("/api/user/mfa/reset/" + user.id)
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())

		auth := new(models.UserAuthResponse)
		resp, err = asBearer(t, compose, user.token).SetResult(auth).Get("/api/auth/user")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.True(t, auth.MFA, "expired codes turned MFA off")
	})
}

type mfaUser struct {
	id            string
	token         string
	email         string
	recoveryEmail string
	username      string
}

func newMFAUser(t *testing.T, compose *environment.DockerCompose, username string) mfaUser {
	t.Helper()

	user := mfaUser{
		username:      username,
		email:         username + stripeTestEmailDomain,
		recoveryEmail: username + "-recovery" + stripeTestEmailDomain,
	}

	compose.NewUser(t, username, user.email, ShellHubPassword)
	compose.NewNamespace(t, username, username, uuid.Generate(), "")

	login := compose.AuthUser(t, username, ShellHubPassword)
	user.id, user.token = login.ID, login.Token

	resp, err := asBearer(t, compose, user.token).
		SetBody(map[string]string{"recovery_email": user.recoveryEmail}).
		Patch("/api/users")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(mfaSecret)
	require.NoError(t, err)

	recoveryCodes := make([]string, 6)
	for i := range recoveryCodes {
		recoveryCodes[i] = strings.ToUpper(strings.ReplaceAll(uuid.Generate(), "-", "")[:16])
	}

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		resp, err := asBearer(t, compose, user.token).
			SetBody(map[string]any{"secret": mfaSecret, "code": totp(key), "recovery_codes": recoveryCodes}).
			Put("/api/user/mfa/enable")
		assert.NoError(tt, err)
		assert.Equal(tt, http.StatusOK, resp.StatusCode(), resp.String())
	}, time.Minute, 5*time.Second)

	return user
}

func requestMFAReset(t *testing.T, compose *environment.DockerCompose, user mfaUser) map[string]string {
	t.Helper()

	mark := compose.ServerLogMark(t)

	resp, err := compose.Anonymous(t.Context()).
		SetBody(map[string]string{"identifier": user.username}).
		Post("/api/user/mfa/reset")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return map[string]string{
		"main_email_code":     emailedMFAResetCode(t, compose, mark, user.email),
		"recovery_email_code": emailedMFAResetCode(t, compose, mark, user.recoveryEmail),
	}
}

func emailedMFAResetCode(t *testing.T, compose *environment.DockerCompose, mark int, to string) string {
	t.Helper()

	line := compose.AwaitServerLogLine(t, mark, "[mail/dummy]", "to="+to)

	match := mfaResetCodeLine.FindStringSubmatch(line)
	require.NotNil(t, match, "no reset code in the email to %s: %s", to, line)

	return match[1]
}

func mfaResetCodeCacheKey(address, userID string) string {
	return "reset-mfa={" + address + ":" + userID + "}"
}

func totp(key []byte) string {
	counter := make([]byte, 8)
	binary.BigEndian.PutUint64(counter, uint64(wallClock.Now().Unix()/30)) //nolint:gosec // a Unix time after 1970 is positive

	mac := hmac.New(sha1.New, key)
	mac.Write(counter)
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0xf
	code := binary.BigEndian.Uint32(sum[offset:]) & 0x7fffffff

	return fmt.Sprintf("%06d", code%1_000_000)
}
