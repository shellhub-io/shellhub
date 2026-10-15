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
	t.Parallel()

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

	signedInto := signInto(t, compose, leaverUsername, leaverPassword)
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

// TestDeleteNamespaceClosesItsTunnels deletes a namespace whose device is online and reads the
// server's count of tunnels closed because their device was removed. The API deletes in the
// process that holds the tunnel, which counts the eviction before the delete answers. The admin CLI
// runs apart from that process, so its deletion reaches the tunnel on the device's next heartbeat.
// A second device authorizes by hand before the deletion, so its authorization is cached, and again
// once the namespace is created anew: it comes back pending only if the deletion evicted that entry.
func TestDeleteNamespaceClosesItsTunnels(t *testing.T) {
	t.Parallel()

	const evictedTunnels = "shellhub_ssh_dialer_connections_evicted_total"

	cases := []struct {
		description     string
		deleteNamespace func(t *testing.T, compose *environment.DockerCompose)
		onHeartbeat     bool
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
			onHeartbeat: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			compose := newFastTunnelPingEnvironment(t, t.Context())

			startAcceptedAgent(t, t.Context(), compose)

			cached := newDeviceAuthRequest(t, "cached", "02:00:00:00:03:24")
			uid := authDevice(t, compose, cached).UID

			evicted, err := compose.ServerMetric(t, evictedTunnels)
			require.NoError(t, err)
			require.Zero(t, evicted)

			tc.deleteNamespace(t, compose)

			if tc.onHeartbeat {
				require.EventuallyWithT(t, func(tt *assert.CollectT) {
					evicted, err := compose.ServerMetric(t, evictedTunnels)
					if !assert.NoError(tt, err) {
						return
					}

					assert.InDelta(tt, 1, evicted, 0)
				}, tunnelBeatTimeout, time.Second)
			} else {
				evicted, err := compose.ServerMetric(t, evictedTunnels)
				require.NoError(t, err)
				assert.InDelta(t, 1, evicted, 0)
			}

			compose.NewNamespace(t, ShellHubUsername, ShellHubNamespaceName, ShellHubNamespace, models.SSHAccessModeLegacy)

			authDevice(t, compose, cached)

			device, resp, err := compose.GetDevice(t.Context(), uid)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
			assert.Equal(t, models.DeviceStatusPending, device.Status)
		})
	}
}

// TestDeleteNamespaceRefusesItsAPIKeys uses a key once, so the server caches it, then deletes the
// key's namespace. Deleting the namespace evicts nothing from that cache.
func TestDeleteNamespaceRefusesItsAPIKeys(t *testing.T) {
	t.Parallel()

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
