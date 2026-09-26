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

const (
	unissuedProvisioningKey      = "3f2b1c44-0000-4000-8000-9a7d5e1c0b22"
	unreachableEnrollmentWebhook = "http://127.0.0.1:1/enroll"
	deviceAuthCacheTTL           = 30 * time.Second
)

// TestProvisioningKeyEnrollment enrolls a device with a provisioning key instead of a tenant id: the key
// names the namespace, and its mode decides whether the device still needs a decision. The legacy
// tenant-only path is covered by [TestSSH].
func TestProvisioningKeyEnrollment(t *testing.T) {
	tests := []struct {
		name       string
		mode       models.ProvisioningKeyMode
		status     models.DeviceStatus
		wantCharge func(t *testing.T, compose *environment.DockerCompose, name string)
	}{
		{
			name:   "an automatic key enrolls the device accepted and is charged a use",
			mode:   models.ProvisioningKeyModeAutomatic,
			status: models.DeviceStatusAccepted,
			wantCharge: func(t *testing.T, compose *environment.DockerCompose, name string) {
				t.Helper()

				compose.AwaitProvisioningKeyUses(t, name, 1)
			},
		},
		{
			name:   "a manual key leaves the device pending and is charged nothing yet",
			mode:   models.ProvisioningKeyModeManual,
			status: models.DeviceStatusPending,
			wantCharge: func(t *testing.T, compose *environment.DockerCompose, name string) {
				t.Helper()

				compose.RequireProvisioningKeyUnused(t, name)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

			key := compose.CreateProvisioningKey(t, &requests.CreateProvisioningKey{
				Name: string(tc.mode),
				Mode: string(tc.mode),
			})

			startAgent(t, ctx, compose, NewAgentContainerWithProvisioningKey(key.Key))

			compose.AwaitDeviceWithStatus(t, tc.status)
			tc.wantCharge(t, compose, string(tc.mode))
		})
	}

	t.Run("a key the namespace never issued enrolls nothing", func(t *testing.T) {
		ctx := context.Background()
		compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

		agent := startAgent(t, ctx, compose, NewAgentContainerWithProvisioningKey(unissuedProvisioningKey))

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			state, err := agent.State(ctx)
			if !assert.NoError(tt, err) {
				return
			}

			assert.False(tt, state.Running, "the agent should give up on a key the server refuses")
			assert.NotZero(tt, state.ExitCode, "the agent should report the refusal as a failure")
		}, 30*time.Second, 1*time.Second)

		devices := []models.Device{}

		resp, err := compose.R(ctx).SetResult(&devices).Get("/api/devices")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode())
		assert.Empty(t, devices)
	})

	t.Run("a pending device whose enrollment was re-evaluated is still listed", func(t *testing.T) {
		ctx := context.Background()
		compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

		key := compose.CreateProvisioningKey(t, &requests.CreateProvisioningKey{
			Name:          string(models.ProvisioningKeyModeWebhook),
			Mode:          string(models.ProvisioningKeyModeWebhook),
			WebhookURL:    unreachableEnrollmentWebhook,
			WebhookSecret: "secret",
		})

		agent := startAgent(t, ctx, compose, NewAgentContainerWithProvisioningKey(key.Key))

		compose.AwaitDeviceWithStatus(t, models.DeviceStatusPending)

		require.NoError(t, agent.Stop(ctx, nil))
		time.Sleep(deviceAuthCacheTTL)
		require.NoError(t, agent.Start(ctx))

		devices := []models.Device{}

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			resp, err := compose.R(ctx).SetResult(&devices).Get("/api/devices?status=pending")
			if !assert.NoError(tt, err) {
				return
			}

			assert.Equal(tt, 200, resp.StatusCode(), resp.String())
			if assert.Len(tt, devices, 1) {
				assert.NotNil(tt, devices[0].LastEnrollmentAttemptAt)
			}
		}, 30*time.Second, 1*time.Second)
	})
}
