package main

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRemovedDeviceFreesItsSlot covers a namespace at its device limit giving a slot back when an
// accepted device is removed. Removed devices do not count against the limit, so the slot is free
// at once: the namespace counts one accepted device fewer, the pending device is offered as
// acceptable again, and accepting it succeeds.
func TestRemovedDeviceFreesItsSlot(t *testing.T) {
	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	t.Run("removing an accepted device frees its slot", func(t *testing.T) {
		held, waiting := fillNamespaceLimit(t, compose, "freed", "02:00:00:00:13:01", "02:00:00:00:13:02")
		accepted := acceptedDevices(t, compose)

		compose.DeleteDevice(t, held)

		assert.Equal(t, accepted-1, acceptedDevices(t, compose))
		assert.True(t, pendingDevice(t, compose, waiting).Acceptable)
	})

	t.Run("a device is accepted into the slot a removed device freed", func(t *testing.T) {
		held, waiting := fillNamespaceLimit(t, compose, "reused", "02:00:00:00:13:03", "02:00:00:00:13:04")
		accepted := acceptedDevices(t, compose)

		compose.DeleteDevice(t, held)
		compose.UpdateDeviceStatus(t, waiting, environment.DeviceActionAccept)

		assert.Equal(t, models.DeviceStatusAccepted, requireDevice(t, compose, waiting).Status)
		assert.Equal(t, accepted, acceptedDevices(t, compose))
	})
}

func fillNamespaceLimit(t *testing.T, compose *environment.DockerCompose, prefix, heldMAC, waitingMAC string) (string, string) {
	t.Helper()

	held := enrollDevice(t, compose, prefix+"held", heldMAC)
	compose.UpdateDeviceStatus(t, held, environment.DeviceActionAccept)
	waiting := enrollDevice(t, compose, prefix+"waiting", waitingMAC)

	compose.LimitNamespaceToItsAcceptedDevices(t, ShellHubNamespace)
	require.False(t, pendingDevice(t, compose, waiting).Acceptable, "a namespace at its limit offered a pending device as acceptable")

	return held, waiting
}

func pendingDevice(t *testing.T, compose *environment.DockerCompose, uid string) models.Device {
	t.Helper()

	for _, device := range compose.ListDevices(t, models.DeviceStatusPending) {
		if device.UID == uid {
			return device
		}
	}

	require.FailNow(t, "the device is not listed as pending", uid)

	return models.Device{}
}
