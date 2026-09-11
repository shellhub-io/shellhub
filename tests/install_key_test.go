package main

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

// NewAgentContainerWithInstallKey drops the tenant id, so the device proves the key alone
// resolved the namespace.
func NewAgentContainerWithInstallKey(key string) NewAgentContainerOption {
	return func(envs map[string]string) {
		delete(envs, "SHELLHUB_TENANT_ID")
		envs["SHELLHUB_INSTALL_KEY"] = key
	}
}

func createInstallKey(t *testing.T, ctx context.Context, compose *environment.DockerCompose, req *requests.CreateInstallKey) *responses.CreateInstallKey {
	t.Helper()

	key := new(responses.CreateInstallKey)

	resp, err := compose.R(ctx).
		SetBody(req).
		SetResult(key).
		Post("/api/namespaces/install-key")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode())
	require.NotEmpty(t, key.Key)

	return key
}

func awaitDevicesWithStatus(t *testing.T, ctx context.Context, compose *environment.DockerCompose, status string, count int) []models.Device {
	t.Helper()

	devices := []models.Device{}

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		resp, err := compose.R(ctx).SetResult(&devices).Get("/api/devices?status=" + status)
		assert.NoError(tt, err)
		assert.Equal(tt, 200, resp.StatusCode())
		assert.Len(tt, devices, count)
	}, 30*time.Second, 1*time.Second)

	return devices
}

func startAgentWithInstallKey(t *testing.T, ctx context.Context, compose *environment.DockerCompose, key string) testcontainers.Container {
	t.Helper()

	agent, err := NewAgentContainer(ctx, compose.Env("SHELLHUB_HTTP_PORT"), NewAgentContainerWithInstallKey(key))
	require.NoError(t, err)

	require.NoError(t, agent.Start(ctx))

	t.Cleanup(func() {
		_ = agent.Stop(context.Background(), nil)
		_ = agent.Terminate(context.Background())
	})

	return agent
}

// TestInstallKeyEnrollment enrolls a device with an install key instead of a tenant id: the key
// names the namespace, and its mode decides whether the device still needs a decision. The legacy
// tenant-only path is covered by [TestSSH].
func TestInstallKeyEnrollment(t *testing.T) {
	t.Run("an automatic key enrolls the device already accepted", func(t *testing.T) {
		ctx := context.Background()
		compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

		key := createInstallKey(t, ctx, compose, &requests.CreateInstallKey{
			Name: "automatic",
			Mode: "automatic",
		})

		startAgentWithInstallKey(t, ctx, compose, key.Key)

		devices := awaitDevicesWithStatus(t, ctx, compose, "accepted", 1)
		assert.Equal(t, models.DeviceStatusAccepted, devices[0].Status)
	})

	t.Run("a manual key leaves the device waiting for a decision", func(t *testing.T) {
		ctx := context.Background()
		compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

		key := createInstallKey(t, ctx, compose, &requests.CreateInstallKey{
			Name: "manual",
			Mode: "manual",
		})

		startAgentWithInstallKey(t, ctx, compose, key.Key)

		devices := awaitDevicesWithStatus(t, ctx, compose, "pending", 1)
		assert.Equal(t, models.DeviceStatusPending, devices[0].Status)
	})

	t.Run("an automatic enrollment charges one use of the key", func(t *testing.T) {
		ctx := context.Background()
		compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

		key := createInstallKey(t, ctx, compose, &requests.CreateInstallKey{
			Name: "counted",
			Mode: "automatic",
		})

		startAgentWithInstallKey(t, ctx, compose, key.Key)
		awaitDevicesWithStatus(t, ctx, compose, "accepted", 1)

		keys := []models.InstallKey{}

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			resp, err := compose.R(ctx).SetResult(&keys).Get("/api/namespaces/install-key")
			assert.NoError(tt, err)
			assert.Equal(tt, 200, resp.StatusCode())

			found := false

			for _, k := range keys {
				if k.Name != "counted" {
					continue
				}

				found = true

				assert.Equal(tt, 1, k.UsedTimes)
				assert.NotNil(tt, k.LastUsedAt)
			}

			assert.True(tt, found, "the key was not listed")
		}, 30*time.Second, 1*time.Second)
	})

	t.Run("a key the namespace never issued enrolls nothing", func(t *testing.T) {
		ctx := context.Background()
		compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

		agent := startAgentWithInstallKey(t, ctx, compose, "3f2b1c44-0000-4000-8000-9a7d5e1c0b22")

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			state, err := agent.State(ctx)
			assert.NoError(tt, err)
			assert.False(tt, state.Running, "the agent should give up on a key the server refuses")
		}, 30*time.Second, 1*time.Second)

		devices := []models.Device{}

		resp, err := compose.R(ctx).SetResult(&devices).Get("/api/devices")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode())
		assert.Empty(t, devices)
	})
}
