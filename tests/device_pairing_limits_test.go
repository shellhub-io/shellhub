package main

import (
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testPairingDeviceLimit(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("a namespace at its device limit refuses a pairing until the limit is lifted", func(t *testing.T) {
		acceptPairing(t, compose.R(t.Context()), startPairing(t, compose, newPairingRequest(t, "pairing-within-limit", "02:00:00:00:93:01")).Code)

		req := newPairingRequest(t, "pairing-over-limit", "02:00:00:00:93:02")
		pairing := startPairing(t, compose, req)

		compose.LimitNamespaceToItsAcceptedDevices(t, ShellHubNamespace)

		_, resp, err := postPairingAccept(compose.R(t.Context()), pairing.Code)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())

		status, resp, err := getPairingStatus(t.Context(), compose, pairing.Code)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Equal(t, models.DeviceStatusPending, status.Status)
		assert.Empty(t, status.TenantID)

		assert.NotContains(t, deviceNames(compose.ListDevices(t, models.DeviceStatusAccepted)), req.Hostname, "a namespace at its limit accepted the paired device")

		require.NoError(t, compose.SetNamespaceMaxDevices(t.Context(), ShellHubNamespace, -1))

		accepted := acceptPairing(t, compose.R(t.Context()), pairing.Code)
		assert.Equal(t, models.DeviceStatusAccepted, requireDevice(t, compose, accepted.UID).Status)
	})
}
