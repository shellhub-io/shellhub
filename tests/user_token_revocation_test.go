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
	t.Parallel()

	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	cases := []struct {
		description string
		next        string
		change      func(t *testing.T, before *models.UserAuthResponse, next string)
	}{
		{
			description: "through the profile update",
			next:        "changed-by-profile",
			change: func(t *testing.T, before *models.UserAuthResponse, next string) {
				t.Helper()

				resp, err := asBearer(t, compose, before.Token).
					SetBody(map[string]string{"password": next, "current_password": ShellHubPassword}).
					Patch("/api/users")
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
			},
		},
		{
			description: "through the deprecated password route",
			next:        "changed-by-deprecated-route",
			change: func(t *testing.T, before *models.UserAuthResponse, next string) {
				t.Helper()

				resp, err := asBearer(t, compose, before.Token).
					SetBody(map[string]string{"current_password": ShellHubPassword, "new_password": next}).
					Patch("/api/users/" + before.ID + "/password")
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
			},
		},
		{
			description: "through the admin CLI",
			next:        "changed-by-admin",
			change: func(t *testing.T, _ *models.UserAuthResponse, next string) {
				t.Helper()

				compose.SetUserPassword(t, ShellHubUsername, next)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			compose.SetUserPassword(t, ShellHubUsername, ShellHubPassword)
			before := signInto(t, compose, ShellHubUsername, ShellHubPassword)
			assertTokenStatus(t, compose, before.Token, http.StatusOK)

			tc.change(t, before, tc.next)

			assertTokenStatus(t, compose, before.Token, http.StatusUnauthorized)

			after := signInto(t, compose, ShellHubUsername, tc.next)
			assertTokenStatus(t, compose, after.Token, http.StatusOK)
		})
	}
}

func TestRevokeUserTokens(t *testing.T) {
	t.Parallel()

	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	first := signInto(t, compose, ShellHubUsername, ShellHubPassword)
	second := signInto(t, compose, ShellHubUsername, ShellHubPassword)
	assertTokenStatus(t, compose, second.Token, http.StatusOK)

	resp, err := asBearer(t, compose, first.Token).Delete("/api/users/tokens")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	assertTokenStatus(t, compose, first.Token, http.StatusUnauthorized)
	assertTokenStatus(t, compose, second.Token, http.StatusUnauthorized)

	assertTokenStatus(t, compose, signInto(t, compose, ShellHubUsername, ShellHubPassword).Token, http.StatusOK)
}

func signInto(t *testing.T, compose *environment.DockerCompose, username, password string) *models.UserAuthResponse {
	t.Helper()

	return signIntoTenant(t, compose, username, password, ShellHubNamespace)
}

func signIntoTenant(t *testing.T, compose *environment.DockerCompose, username, password, tenant string) *models.UserAuthResponse {
	t.Helper()

	login := compose.AuthUser(t, username, password)

	signedInto := new(models.UserAuthResponse)
	resp, err := asBearer(t, compose, login.Token).
		SetResult(signedInto).
		Get("/api/auth/token/" + tenant)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return signedInto
}

var tokenRoutes = []string{"/api/auth/user", "/api/auth/token/" + ShellHubNamespace, "/api/namespaces"}

func assertTokenStatus(t *testing.T, compose *environment.DockerCompose, token string, status int) {
	t.Helper()

	for _, route := range tokenRoutes {
		resp, err := asBearer(t, compose, token).Get(route)
		require.NoError(t, err)
		assert.Equal(t, status, resp.StatusCode(), "GET %s: %s", route, resp.String())
	}
}
