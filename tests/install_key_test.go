package main

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const unissuedInstallKey = "3f2b1c44-0000-4000-8000-9a7d5e1c0b22"

// TestInstallKeyEnrollment enrolls a device with an install key instead of a tenant id: the key
// names the namespace, and its mode decides whether the device still needs a decision. The legacy
// tenant-only path is covered by [TestSSH].
func TestInstallKeyEnrollment(t *testing.T) {
	tests := []struct {
		name   string
		mode   models.InstallKeyMode
		status models.DeviceStatus
		uses   int
	}{
		{
			name:   "an automatic key enrolls the device accepted and is charged a use",
			mode:   models.InstallKeyModeAutomatic,
			status: models.DeviceStatusAccepted,
			uses:   1,
		},
		{
			name:   "a manual key leaves the device pending and is charged nothing yet",
			mode:   models.InstallKeyModeManual,
			status: models.DeviceStatusPending,
			uses:   0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

			key := compose.CreateInstallKey(t, &requests.CreateInstallKey{
				Name: string(tc.mode),
				Mode: string(tc.mode),
			})

			startAgent(t, ctx, compose, NewAgentContainerWithInstallKey(key.Key))

			compose.AwaitDeviceWithStatus(t, tc.status)
			compose.AwaitInstallKeyUses(t, string(tc.mode), tc.uses)
		})
	}

	t.Run("a key the namespace never issued enrolls nothing", func(t *testing.T) {
		ctx := context.Background()
		compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

		agent := startAgent(t, ctx, compose, NewAgentContainerWithInstallKey(unissuedInstallKey))

		environment.AwaitLogContains(t, agent, `error="failed to authorize device: bad request"`)

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			state, err := agent.State(ctx)
			if !assert.NoError(tt, err) {
				return
			}

			assert.False(tt, state.Running, "the agent should give up on a key the server refuses")
		}, 30*time.Second, 1*time.Second)

		devices := []models.Device{}

		resp, err := compose.R(ctx).SetResult(&devices).Get("/api/devices")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode())
		assert.Empty(t, devices)
	})
}
