package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnterpriseSignInPolicy runs an enterprise instance whose administrator turned local sign-in
// off and left SAML as the only way in. A recovery code is refused with the status the password
// sign-in answers, a token issued by local sign-in is no longer renewed, and neither issues a token.
func TestEnterpriseSignInPolicy(t *testing.T) {
	t.Parallel()

	compose := newEnterpriseEnvironment(t, context.Background(), environment.New(t, run))

	user := newMFAUser(t, compose, "policy-recover")
	mfaToken := passwordStep(t, compose, user)

	enableSAMLPostBinding(t, compose)
	requireAnswer(t, asBearer(t, compose, user.token), http.MethodGet, "/api/auth/user", http.StatusOK)
	setLocalSignIn(t, compose, false)

	passwordSignIn, err := compose.Anonymous(t.Context()).
		SetBody(map[string]string{"username": user.username, "password": ShellHubPassword}).
		Post("/api/login")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotImplemented, passwordSignIn.StatusCode(), passwordSignIn.String())

	t.Run("a recovery code does not sign in while local sign-in is off", func(t *testing.T) {
		auth := new(models.UserAuthResponse)
		resp, err := compose.Anonymous(t.Context()).
			SetBody(map[string]string{"token": mfaToken, "recovery_code": user.recoveryCodes[0]}).
			SetResult(auth).
			Post("/api/user/mfa/recover")
		require.NoError(t, err)
		assert.Equal(t, passwordSignIn.StatusCode(), resp.StatusCode(), resp.String())
		assert.Empty(t, auth.Token)
	})

	t.Run("a token issued by local sign-in is not renewed while local sign-in is off", func(t *testing.T) {
		auth := new(models.UserAuthResponse)
		resp, err := asBearer(t, compose, user.token).SetResult(auth).Get("/api/auth/user")
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode(), resp.String())
		assert.Empty(t, auth.Token)
	})
}

func setLocalSignIn(t *testing.T, compose *environment.DockerCompose, enabled bool) {
	t.Helper()

	resp, err := compose.R(t.Context()).
		SetBody(map[string]bool{"enable": enabled}).
		Put("/admin/api/authentication/local")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
}
