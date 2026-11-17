package main

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSignedInPasswordLockout covers the password a signed-in user sends again before a sensitive
// action: the web-terminal step-up and a password change. A session alone must not be enough to
// guess it, so wrong ones lock the account's step-up out, while signing in with the password stays
// open to its owner.
func TestSignedInPasswordLockout(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	compose := newSSHEnvironment(t, ctx, newNamespace)

	t.Run("a key the caller does not own answers the same whatever the password", func(t *testing.T) {
		const unknown = "SHA256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

		for range 12 {
			resp, err := compose.R(ctx).
				SetBody(&requests.WebReauthVerify{Password: "wrong", Fingerprint: unknown}).
				Post("/api/web-terminal/reauth")
			require.NoError(t, err)
			require.Equal(t, http.StatusNotFound, resp.StatusCode(), resp.String())
		}

		resp, err := compose.R(ctx).
			SetBody(&requests.WebReauthVerify{Password: ShellHubPassword, Fingerprint: unknown}).
			Post("/api/web-terminal/reauth")
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode(), resp.String())
	})

	t.Run("the third wrong step-up password locks the step-up out, not the sign-in", func(t *testing.T) {
		_, fingerprint := enrollOwnerKey(t, compose)

		reauth := func(password string) (int, int64) {
			resp, err := compose.R(ctx).
				SetBody(&requests.WebReauthVerify{Password: password, Fingerprint: fingerprint}).
				Post("/api/web-terminal/reauth")
			require.NoError(t, err)

			lockout, _ := strconv.ParseInt(resp.Header().Get("X-Account-Lockout"), 10, 64)

			return resp.StatusCode(), lockout
		}

		status, _ := reauth("wrong")
		assert.Equal(t, http.StatusForbidden, status)
		status, _ = reauth("wrong")
		assert.Equal(t, http.StatusForbidden, status)
		status, lockout := reauth("wrong")
		assert.Equal(t, http.StatusTooManyRequests, status)
		assert.Positive(t, lockout)

		status, _ = reauth(ShellHubPassword)
		assert.Equal(t, http.StatusTooManyRequests, status, "a right password during the lockout was checked")

		resp, err := compose.Anonymous(ctx).
			SetBody(map[string]string{"username": ShellHubUsername, "password": ShellHubPassword}).
			Post("/api/login")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	})

	t.Run("the third wrong current password locks password changes out", func(t *testing.T) {
		member := newMember(t, compose, "lockout-member", authorizer.RoleObserver)

		change := func(current string) int {
			resp, err := compose.Anonymous(ctx).
				SetAuthToken(member.Token).
				SetBody(map[string]string{"password": "a-new-password", "current_password": current}).
				Patch("/api/users")
			require.NoError(t, err)

			return resp.StatusCode()
		}

		assert.Equal(t, http.StatusForbidden, change("wrong"))
		assert.Equal(t, http.StatusForbidden, change("wrong"))
		assert.Equal(t, http.StatusTooManyRequests, change("wrong"))
		assert.Equal(t, http.StatusTooManyRequests, change(ShellHubPassword), "a right password during the lockout was checked")
	})
}
