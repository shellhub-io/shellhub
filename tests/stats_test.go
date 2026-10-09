package main

import (
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatsDeviceCounts(t *testing.T) {
	compose := environment.New(t, run).Up(t.Context())
	t.Cleanup(compose.Down)

	compose.NewUser(t, ShellHubUsername, ShellHubEmail, ShellHubPassword)
	compose.NewNamespace(t, ShellHubUsername, ShellHubNamespaceName, ShellHubNamespace, "")

	compose.JWT(compose.AuthUser(t, ShellHubUsername, ShellHubPassword).Token)

	stats := func(t *testing.T) models.Stats {
		t.Helper()

		var current models.Stats
		resp, err := compose.R(t.Context()).SetResult(&current).Get("/api/stats")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		return current
	}

	t.Run("a namespace with no device has nothing registered", func(t *testing.T) {
		current := stats(t)
		assert.Zero(t, current.RegisteredDevices)
		assert.Zero(t, current.PendingDevices)
	})

	uid := enrollDevice(t, compose, "first", "02:00:00:00:0a:01")

	t.Run("the first device counts as pending until it is accepted", func(t *testing.T) {
		current := stats(t)
		assert.Zero(t, current.RegisteredDevices)
		assert.Equal(t, int64(1), current.PendingDevices)
	})

	t.Run("the first accepted device counts as registered", func(t *testing.T) {
		compose.UpdateDeviceStatus(t, uid, environment.DeviceActionAccept)

		current := stats(t)
		assert.Equal(t, int64(1), current.RegisteredDevices)
		assert.Zero(t, current.PendingDevices)
	})
}
