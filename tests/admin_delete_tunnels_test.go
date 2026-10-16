package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAdminDeletesCloseTunnels deletes, as an instance administrator, a namespace whose device is
// online, either directly or with the user who owns it, and reads the server's count of tunnels
// closed because their device was removed. The enterprise admin API deletes in the process that
// holds the tunnel, which counts the eviction before the delete answers. The community admin CLI
// runs apart from that process, so its deletion reaches the tunnel on the device's next heartbeat.
// A second device authorizes by hand before the deletion, so its authorization is cached, and again
// once the owner and the namespace are created anew: it comes back pending only if the deletion
// evicted that entry.
func TestAdminDeletesCloseTunnels(t *testing.T) {
	t.Parallel()

	const evictedTunnels = "shellhub_ssh_dialer_connections_evicted_total"

	cases := []struct {
		description string
		edition     environment.Edition
		delete      func(t *testing.T, compose *environment.DockerCompose, admin *models.UserAuthResponse)
		ownerGone   bool
		onHeartbeat bool
	}{
		{
			description: "deleting the namespace through the admin API",
			edition:     environment.EditionEnterprise,
			delete: func(t *testing.T, compose *environment.DockerCompose, admin *models.UserAuthResponse) {
				t.Helper()

				resp, err := asBearer(t, compose, admin.Token).Delete("/admin/api/namespaces/" + ShellHubNamespace)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
			},
		},
		{
			description: "deleting its owner through the admin API",
			edition:     environment.EditionEnterprise,
			delete: func(t *testing.T, compose *environment.DockerCompose, admin *models.UserAuthResponse) {
				t.Helper()

				owner := compose.AuthUser(t, ShellHubUsername, ShellHubPassword)

				resp, err := asBearer(t, compose, admin.Token).Delete("/admin/api/users/" + owner.ID)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
			},
			ownerGone: true,
		},
		{
			description: "deleting its owner through the admin CLI",
			edition:     environment.EditionCommunity,
			delete: func(t *testing.T, compose *environment.DockerCompose, _ *models.UserAuthResponse) {
				t.Helper()

				compose.DeleteUser(t, ShellHubUsername)
			},
			ownerGone:   true,
			onHeartbeat: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			compose, admin := newAdministeredTunnelEnvironment(t, t.Context(), tc.edition)

			startAcceptedAgent(t, t.Context(), compose)

			cached := newDeviceAuthRequest(t, "cached", "02:00:00:00:03:25")
			uid := authDevice(t, compose, cached).UID

			evicted, err := compose.ServerMetric(t, evictedTunnels)
			require.NoError(t, err)
			require.Zero(t, evicted)

			tc.delete(t, compose, admin)

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

			if tc.ownerGone {
				compose.NewUser(t, ShellHubUsername, ShellHubEmail, ShellHubPassword)
			}

			ownNamespace(t, t.Context(), compose, legacyNamespace)

			authDevice(t, compose, cached)

			device, resp, err := compose.GetDevice(t.Context(), uid)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
			assert.Equal(t, models.DeviceStatusPending, device.Status)
		})
	}
}

func newAdministeredTunnelEnvironment(t *testing.T, ctx context.Context, edition environment.Edition) (*environment.DockerCompose, *models.UserAuthResponse) {
	t.Helper()

	compose := environment.New(t, run).
		WithEdition(edition).
		WithEnv("SHELLHUB_TUNNEL_PING_INTERVAL", fastTunnelPingInterval.String()).
		Up(ctx)
	t.Cleanup(compose.Down)

	const adminUsername = "root"
	compose.NewUser(t, adminUsername, "root@ossystems.com.br", ShellHubPassword)
	admin := compose.AuthUser(t, adminUsername, ShellHubPassword)
	require.True(t, admin.Admin, "the first user is the instance's administrator")

	compose.NewUser(t, ShellHubUsername, ShellHubEmail, ShellHubPassword)
	ownNamespace(t, ctx, compose, legacyNamespace)

	return compose, admin
}
