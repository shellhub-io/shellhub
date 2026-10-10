package main

import (
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeviceAcceptanceLimit accepts a device enrolled by hand through the route a member accepts
// a pending device with, in a namespace whose device limit the test sets. Pairing and automatic
// provisioning keys reach the same limit through their own routes, in [TestDevicePairing] and
// [TestEnrollmentPolicy].
func TestDeviceAcceptanceLimit(t *testing.T) {
	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	t.Run("a namespace at its device limit refuses accepting a device until the limit is lifted", func(t *testing.T) {
		within := enrollDevice(t, compose, "accept-within-limit", "02:00:00:00:a2:01")
		compose.UpdateDeviceStatus(t, within, environment.DeviceActionAccept)

		over := enrollDevice(t, compose, "accept-over-limit", "02:00:00:00:a2:02")

		compose.LimitNamespaceToItsAcceptedDevices(t, ShellHubNamespace)

		resp, err := compose.PatchDeviceStatus(t.Context(), over, environment.DeviceActionAccept)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())

		assert.Equal(t, models.DeviceStatusPending, requireDevice(t, compose, over).Status)

		require.NoError(t, compose.SetNamespaceMaxDevices(t.Context(), ShellHubNamespace, -1))

		compose.UpdateDeviceStatus(t, over, environment.DeviceActionAccept)
		assert.Equal(t, models.DeviceStatusAccepted, requireDevice(t, compose, over).Status)
	})
}
