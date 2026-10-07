package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	leaverUsername = "leaver"
	leaverEmail    = "leaver@ossystems.com.br"
	leaverPassword = "password"

	leftBehindNamespaceName = "otherspace"
	leftBehindNamespace     = "00000000-0000-4000-0000-000000000001"

	neverLeftNamespaceName = "neverleftspace"
	neverLeftNamespace     = "00000000-0000-4000-0000-000000000002"
)

// TestLeaveNamespace covers both shapes the route answers. Leaving the namespace the caller's
// token is scoped to mints a replacement token, because the one they hold names a namespace they
// no longer belong to. Leaving any other namespace leaves that token valid, so there is nothing
// to send back and the route says so with its own status rather than an empty 200.
func TestLeaveNamespace(t *testing.T) {
	ctx := context.Background()

	compose := environment.New(t, run).Up(ctx)
	t.Cleanup(compose.Down)

	compose.NewUser(t, ShellHubUsername, ShellHubEmail, ShellHubPassword)
	compose.NewNamespace(t, ShellHubUsername, ShellHubNamespaceName, ShellHubNamespace, "")
	compose.NewNamespace(t, ShellHubUsername, leftBehindNamespaceName, leftBehindNamespace, "")
	compose.NewNamespace(t, ShellHubUsername, neverLeftNamespaceName, neverLeftNamespace, "")
	compose.NewUser(t, leaverUsername, leaverEmail, leaverPassword)
	compose.NewMember(t, leaverUsername, ShellHubNamespaceName, string(authorizer.RoleObserver))
	compose.NewMember(t, leaverUsername, leftBehindNamespaceName, string(authorizer.RoleObserver))
	compose.NewMember(t, leaverUsername, neverLeftNamespaceName, string(authorizer.RoleObserver))

	auth := compose.AuthUser(t, leaverUsername, leaverPassword)

	signedInto := models.UserAuthResponse{}

	resp, err := compose.Anonymous(ctx).
		SetAuthToken(auth.Token).
		SetResult(&signedInto).
		Get("/api/auth/token/" + ShellHubNamespace)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, signedInto.Tenant)
	require.Equal(t, ShellHubNamespace, *signedInto.Tenant)

	t.Run("leaving another namespace answers no content", func(t *testing.T) {
		resp, err := compose.Anonymous(ctx).
			SetAuthToken(signedInto.Token).
			Delete("/api/namespaces/" + leftBehindNamespace + "/members")

		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode())
		require.Empty(t, resp.Body())
	})

	t.Run("leaving the namespace the token names answers a replacement token", func(t *testing.T) {
		replacement := models.UserAuthResponse{}

		resp, err := compose.Anonymous(ctx).
			SetAuthToken(signedInto.Token).
			SetResult(&replacement).
			Delete("/api/namespaces/" + ShellHubNamespace + "/members")

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.NotEmpty(t, replacement.Token)
		require.NotEqual(t, signedInto.Token, replacement.Token)
		require.NotNil(t, replacement.Tenant)
		require.Equal(t, neverLeftNamespace, *replacement.Tenant)
	})
}

// TestDeleteNamespaceClosesItsTunnels deletes a namespace whose device is online and creates it
// again under the same tenant. The agent answers a deleted namespace with a 404 it never treats as
// a removal, and pings only every 8 to 12 minutes, so it re-enrolls within a heartbeat only when
// the server closed the tunnel itself. The admin CLI runs apart from the server that holds the
// tunnel.
func TestDeleteNamespaceClosesItsTunnels(t *testing.T) {
	cases := []struct {
		description     string
		deleteNamespace func(t *testing.T, compose *environment.DockerCompose)
	}{
		{
			description: "through the API",
			deleteNamespace: func(t *testing.T, compose *environment.DockerCompose) {
				t.Helper()

				resp, err := compose.R(t.Context()).Delete("/api/namespaces/" + ShellHubNamespace)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
			},
		},
		{
			description: "through the admin CLI",
			deleteNamespace: func(t *testing.T, compose *environment.DockerCompose) {
				t.Helper()

				compose.DeleteNamespace(t, ShellHubNamespaceName)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

			_, device := startAcceptedAgent(t, t.Context(), compose)

			tc.deleteNamespace(t, compose)

			compose.NewNamespace(t, ShellHubUsername, ShellHubNamespaceName, ShellHubNamespace, models.SSHAccessModeLegacy)

			require.EventuallyWithT(t, func(tt *assert.CollectT) {
				current, resp, err := compose.GetDevice(t.Context(), device.UID)
				if !assert.NoError(tt, err) {
					return
				}

				if !assert.Equal(tt, http.StatusOK, resp.StatusCode(), resp.String()) {
					return
				}

				assert.Equal(tt, models.DeviceStatusPending, current.Status)
			}, tunnelPingTimeout, time.Second)
		})
	}
}

// TestDeleteNamespaceRefusesItsAPIKeys uses a key once, so the server caches it, then deletes the
// key's namespace. Deleting the namespace evicts nothing from that cache.
func TestDeleteNamespaceRefusesItsAPIKeys(t *testing.T) {
	compose := environment.New(t, run).Up(t.Context())
	t.Cleanup(compose.Down)

	compose.NewUser(t, ShellHubUsername, ShellHubEmail, ShellHubPassword)
	compose.NewNamespace(t, ShellHubUsername, ShellHubNamespaceName, ShellHubNamespace, "")
	compose.JWT(compose.AuthUser(t, ShellHubUsername, ShellHubPassword).Token)

	key := compose.CreateAPIKey(t, &requests.CreateAPIKey{
		Name:      "automation",
		ExpiresAt: -1,
		OptRole:   authorizer.RoleAdministrator,
	})

	resp, err := compose.Anonymous(t.Context()).SetHeader("X-API-Key", key.Key).Get("/api/devices")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	resp, err = compose.R(t.Context()).Delete("/api/namespaces/" + ShellHubNamespace)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	resp, err = compose.Anonymous(t.Context()).SetHeader("X-API-Key", key.Key).Get("/api/devices")
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode(), resp.String())
}
