package main

import (
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordChangeRevokesTokens(t *testing.T) {
	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	cases := []struct {
		description string
		next        string
		change      func(t *testing.T, before *models.UserAuthResponse, current, next string)
	}{
		{
			description: "through the profile update",
			next:        "changed-by-profile",
			change: func(t *testing.T, before *models.UserAuthResponse, current, next string) {
				t.Helper()

				resp, err := asBearer(t, compose, before.Token).
					SetBody(map[string]string{"password": next, "current_password": current}).
					Patch("/api/users")
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
			},
		},
		{
			description: "through the deprecated password route",
			next:        "changed-by-deprecated-route",
			change: func(t *testing.T, before *models.UserAuthResponse, current, next string) {
				t.Helper()

				resp, err := asBearer(t, compose, before.Token).
					SetBody(map[string]string{"current_password": current, "new_password": next}).
					Patch("/api/users/" + before.ID + "/password")
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
			},
		},
		{
			description: "through the admin CLI",
			next:        "changed-by-admin",
			change: func(t *testing.T, _ *models.UserAuthResponse, _, next string) {
				t.Helper()

				compose.SetUserPassword(t, ShellHubUsername, next)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			compose.SetUserPassword(t, ShellHubUsername, ShellHubPassword)
			before := signInto(t, compose, ShellHubPassword)

			tc.change(t, before, ShellHubPassword, tc.next)

			assertTokenRefused(t, compose, before.Token)

			after := signInto(t, compose, tc.next)
			assertTokenHonoured(t, compose, after.Token)
		})
	}
}

func TestRevokeUserTokens(t *testing.T) {
	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	first := signInto(t, compose, ShellHubPassword)
	second := signInto(t, compose, ShellHubPassword)

	resp, err := asBearer(t, compose, first.Token).Delete("/api/users/tokens")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	assertTokenRefused(t, compose, first.Token)
	assertTokenRefused(t, compose, second.Token)

	assertTokenHonoured(t, compose, signInto(t, compose, ShellHubPassword).Token)
}

func signInto(t *testing.T, compose *environment.DockerCompose, password string) *models.UserAuthResponse {
	t.Helper()

	login := compose.AuthUser(t, ShellHubUsername, password)

	signedInto := new(models.UserAuthResponse)
	resp, err := asBearer(t, compose, login.Token).
		SetResult(signedInto).
		Get("/api/auth/token/" + ShellHubNamespace)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return signedInto
}

var tokenRoutes = []string{"/api/auth/user", "/api/auth/token/" + ShellHubNamespace, "/api/namespaces"}

func assertTokenRefused(t *testing.T, compose *environment.DockerCompose, token string) {
	t.Helper()

	for _, route := range tokenRoutes {
		resp, err := asBearer(t, compose, token).Get(route)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode(), "GET %s", route)
	}
}

func assertTokenHonoured(t *testing.T, compose *environment.DockerCompose, token string) {
	t.Helper()

	for _, route := range tokenRoutes {
		resp, err := asBearer(t, compose, token).Get(route)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode(), "GET %s: %s", route, resp.String())
	}
}
