package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha1" //nolint:depguard,gosec // RFC 6238 TOTP, which the server checks, is HMAC-SHA1
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mfaResetCodeLifetime = 24 * time.Hour

// TestCloudMFARecovery runs a cloud instance, the only edition that sends email, for a user who
// lost their authenticator. A recovery code stands in for the TOTP code after the password, never
// for the password. The reset of a user's MFA by the two codes it emails to their main and
// recovery addresses signs no one in: the user signs in with the password afterwards. Each code
// lasts 24 hours in the cache, and the reset refuses the pair once either code has expired.
func TestCloudMFARecovery(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := environment.New(t, run).WithEdition(environment.EditionCloud).Up(ctx)
	t.Cleanup(compose.Down)

	t.Run("a recovery code completes the password step", func(t *testing.T) {
		user := newMFAUser(t, compose, "mfa-recover")
		mfaToken := passwordStep(t, compose, user)

		auth := new(models.UserAuthResponse)
		resp, err := compose.Anonymous(t.Context()).
			SetBody(map[string]string{"token": mfaToken, "recovery_code": user.recoveryCodes[0]}).
			SetResult(auth).
			Post("/api/user/mfa/recover")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Equal(t, user.id, auth.ID)
		assert.NotEmpty(t, auth.Token)
	})

	t.Run("a recovery code without the password step signs no one in", func(t *testing.T) {
		user := newMFAUser(t, compose, "mfa-recover-alone")

		auth := new(models.UserAuthResponse)
		resp, err := compose.Anonymous(t.Context()).
			SetBody(map[string]string{"identifier": user.username, "recovery_code": user.recoveryCodes[0]}).
			SetResult(auth).
			Post("/api/user/mfa/recover")
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode(), resp.String())
		assert.Empty(t, auth.Token)

		resp, err = compose.Anonymous(t.Context()).
			SetBody(map[string]string{"token": uuid.Generate(), "recovery_code": user.recoveryCodes[0]}).
			SetResult(auth).
			Post("/api/user/mfa/recover")
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode(), resp.String())
		assert.Empty(t, auth.Token)
	})

	t.Run("the emailed codes reset MFA and the user signs in with the password", func(t *testing.T) {
		user := newMFAUser(t, compose, "mfa-reset")
		codes := requestMFAReset(t, compose, user)

		resp, err := compose.Anonymous(t.Context()).SetBody(codes).Put("/api/user/mfa/reset/" + user.id)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Empty(t, resp.Body(), "the reset signed the user in")

		requireAnswer(t, asBearer(t, compose, user.token), http.MethodGet, "/api/auth/user", http.StatusUnauthorized)

		signedIn := compose.AuthUser(t, user.username, ShellHubPassword)
		assert.NotEmpty(t, signedIn.Token)
		assert.False(t, signedIn.MFA, "the reset left MFA on")
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
	recoveryCodes []string
}

func newMFAUser(t *testing.T, compose *environment.DockerCompose, username string) mfaUser {
	t.Helper()

	user := mfaUser{
		username:      username,
		email:         username + stripeTestEmailDomain,
		recoveryEmail: username + "-recovery" + stripeTestEmailDomain,
	}

	compose.NewUser(t, username, user.email, ShellHubPassword)
	compose.NewNamespace(t, username, username, uuid.Generate())

	login := compose.AuthUser(t, username, ShellHubPassword)
	user.id, user.token = login.ID, login.Token

	resp, err := asBearer(t, compose, user.token).
		SetBody(map[string]string{"recovery_email": user.recoveryEmail}).
		Patch("/api/users")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	generated := struct {
		Secret        string   `json:"secret"`
		RecoveryCodes []string `json:"recovery_codes"`
	}{}
	resp, err = asBearer(t, compose, user.token).SetResult(&generated).Get("/api/user/mfa/generate")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	user.recoveryCodes = generated.RecoveryCodes

	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(generated.Secret)
	require.NoError(t, err)

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		resp, err := asBearer(t, compose, user.token).
			SetBody(map[string]string{"code": totp(key)}).
			Put("/api/user/mfa/enable")
		assert.NoError(tt, err)
		assert.Equal(tt, http.StatusOK, resp.StatusCode(), resp.String())
	}, time.Minute, 5*time.Second)

	return user
}

func passwordStep(t *testing.T, compose *environment.DockerCompose, user mfaUser) string {
	t.Helper()

	resp, err := compose.Anonymous(t.Context()).
		SetBody(map[string]string{"username": user.username, "password": ShellHubPassword}).
		Post("/api/login")
	require.NoError(t, err)

	mfaToken := resp.Header().Get("X-MFA-Token")
	require.NotEmpty(t, mfaToken, resp.String())

	return mfaToken
}

func requestMFAReset(t *testing.T, compose *environment.DockerCompose, user mfaUser) map[string]string {
	t.Helper()

	resp, err := compose.Anonymous(t.Context()).
		SetBody(map[string]string{"identifier": user.username}).
		Post("/api/user/mfa/reset")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return map[string]string{
		"main_email_code":     emailedMFAResetCode(t, compose, user.email),
		"recovery_email_code": emailedMFAResetCode(t, compose, user.recoveryEmail),
	}
}

func emailedMFAResetCode(t *testing.T, compose *environment.DockerCompose, to string) string {
	t.Helper()

	code, ok := compose.AwaitEmail(t, to, "/reset-mfa").Labelled("MFA reset code")
	require.True(t, ok, "no reset code in the email to %s", to)

	return code
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
