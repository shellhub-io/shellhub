package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ephemeralCleanupCron = "*/5 * * * *"
	removedCleanupCron   = "0 2 * * *"
	removedRetention     = 30 * 24 * time.Hour
	heartbeatFlushWindow = 3 * time.Second
)

// TestDeviceCleanup runs the cron jobs that delete devices for good: an ephemeral device once it has
// stayed offline past its key's timeout, and a removed device once it has been removed for 30 days.
// Each job is triggered rather than waited for, and the time it measures is moved back in the
// database. The cases share one stack, so every device carries a hostname no other case uses.
func TestDeviceCleanup(t *testing.T) {
	t.Parallel()

	compose := newConfiguredSSHEnvironment(t, t.Context(), environment.New(t, run).WithCronTrigger(), newNamespace)

	t.Run("an ephemeral device offline past its timeout is deleted and an online one is kept", func(t *testing.T) {
		key := compose.CreateProvisioningKey(t, &requests.CreateProvisioningKey{
			Name:             "ephemeral-cleanup",
			Mode:             string(models.ProvisioningKeyModeAutomatic),
			Ephemeral:        true,
			EphemeralTimeout: 1,
		}).Key

		offlineAgent := startAgent(t, t.Context(), compose, NewAgentContainerWithProvisioningKey(key), NewAgentContainerWithHostname("ephemeral-offline"))
		offline := compose.AwaitDeviceOnline(t, compose.AwaitDeviceNamed(t, "ephemeral-offline").UID)
		startAgent(t, t.Context(), compose, NewAgentContainerWithProvisioningKey(key), NewAgentContainerWithHostname("ephemeral-online"))
		online := compose.AwaitDeviceOnline(t, compose.AwaitDeviceNamed(t, "ephemeral-online").UID)
		require.True(t, offline.Ephemeral)
		require.True(t, online.Ephemeral)

		require.NoError(t, offlineAgent.Stop(t.Context(), nil))
		compose.AwaitDeviceOffline(t, offline.UID)
		time.Sleep(heartbeatFlushWindow)

		before := compose.Namespace(t, ShellHubNamespace)

		compose.AgeDeviceDisconnection(t, offline.UID, 2*time.Minute)

		compose.RunCron(t, ephemeralCleanupCron)

		awaitDeviceDeleted(t, compose, offline.UID)

		after := compose.Namespace(t, ShellHubNamespace)
		assert.Equal(t, before.DevicesAcceptedCount-1, after.DevicesAcceptedCount, "the namespace still counts the deleted device")

		kept := requireDevice(t, compose, online.UID)
		assert.Equal(t, models.DeviceStatusAccepted, kept.Status)
		assert.True(t, kept.Online, "the online ephemeral device went offline")
	})

	t.Run("a device removed more than 30 days ago is purged and a recently removed one is kept", func(t *testing.T) {
		expired := removeAcceptedDevice(t, compose, "removed-expired", "02:00:00:00:c1:01")
		recent := removeAcceptedDevice(t, compose, "removed-recent", "02:00:00:00:c1:02")

		before := compose.Namespace(t, ShellHubNamespace)

		compose.AgeDeviceRemoval(t, expired, removedRetention+time.Hour)
		compose.AgeDeviceRemoval(t, recent, removedRetention-24*time.Hour)

		compose.RunCron(t, removedCleanupCron)

		awaitDeviceDeleted(t, compose, expired)

		after := compose.Namespace(t, ShellHubNamespace)
		assert.Equal(t, before.DevicesRemovedCount-1, after.DevicesRemovedCount, "the namespace still counts the purged device")

		assert.Equal(t, models.DeviceStatusRemoved, requireDevice(t, compose, recent).Status, "a device removed inside the retention window was purged")
	})
}

func removeAcceptedDevice(t *testing.T, compose *environment.DockerCompose, hostname, mac string) string {
	t.Helper()

	uid := enrollDevice(t, compose, hostname, mac)
	compose.UpdateDeviceStatus(t, uid, environment.DeviceActionAccept)
	compose.DeleteDevice(t, uid)
	require.Equal(t, models.DeviceStatusRemoved, requireDevice(t, compose, uid).Status)

	return uid
}

func awaitDeviceDeleted(t *testing.T, compose *environment.DockerCompose, uid string) {
	t.Helper()

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		_, resp, err := compose.GetDevice(t.Context(), uid)
		if assert.NoError(tt, err) {
			assert.Equal(tt, http.StatusNotFound, resp.StatusCode(), "device %s is still there: %s", uid, resp.String())
		}
	}, 30*time.Second, time.Second)
}
