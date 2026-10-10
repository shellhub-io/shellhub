package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeviceAcceptNameConflicts covers the hostname an accept would give a device being taken by
// another accepted device. A fresh device is refused outright, and so is a device whose MAC would
// merge it into a disconnected one, since the merge would still leave two devices answering to
// the name. Either accept goes through once the other device lets the name go.
func TestDeviceAcceptNameConflicts(t *testing.T) {
	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

	t.Run("a fresh device is refused the name an accepted device holds", func(t *testing.T) {
		holder := enrollDevice(t, compose, "freshtaken", "02:00:00:00:14:01")
		compose.UpdateDeviceStatus(t, holder, environment.DeviceActionAccept)
		fresh := enrollDevice(t, compose, "freshtaken", "02:00:00:00:14:02")

		requireAcceptConflict(t, compose, fresh)
		assert.Equal(t, "freshtaken", requireDevice(t, compose, holder).Name)

		renameDevice(t, compose, holder, "freshreleased")
		compose.UpdateDeviceStatus(t, fresh, environment.DeviceActionAccept)

		accepted := requireDevice(t, compose, fresh)
		assert.Equal(t, models.DeviceStatusAccepted, accepted.Status)
		assert.Equal(t, "freshtaken", accepted.Name)
	})

	t.Run("a MAC merge is refused when another accepted device holds the hostname", func(t *testing.T) {
		agent, merged := startAcceptedAgent(t, ctx, compose)
		require.NoError(t, agent.Stop(ctx, nil))
		compose.AwaitDeviceOffline(t, merged.UID)

		holder := enrollDevice(t, compose, "mergetaken", "02:00:00:00:14:03")
		compose.UpdateDeviceStatus(t, holder, environment.DeviceActionAccept)
		reinstalled := enrollDevice(t, compose, "mergetaken", merged.Identity.MAC)

		requireAcceptConflict(t, compose, reinstalled)
		assert.Equal(t, merged.Name, requireDevice(t, compose, merged.UID).Name)
		assert.Equal(t, models.DeviceStatusAccepted, requireDevice(t, compose, merged.UID).Status)

		renameDevice(t, compose, holder, "mergereleased")
		compose.UpdateDeviceStatus(t, reinstalled, environment.DeviceActionAccept)

		_, resp, err := compose.GetDevice(t.Context(), merged.UID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode(), "the merge left the disconnected device behind: %s", resp.String())

		accepted := requireDevice(t, compose, reinstalled)
		assert.Equal(t, models.DeviceStatusAccepted, accepted.Status)
		assert.Equal(t, merged.Name, accepted.Name, "the merge did not carry the disconnected device's name over")
	})
}

func requireAcceptConflict(t *testing.T, compose *environment.DockerCompose, uid string) {
	t.Helper()

	resp, err := compose.PatchDeviceStatus(t.Context(), uid, environment.DeviceActionAccept)
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, resp.StatusCode(), resp.String())
	assert.Equal(t, models.DeviceStatusPending, requireDevice(t, compose, uid).Status)
}

func renameDevice(t *testing.T, compose *environment.DockerCompose, uid, name string) {
	t.Helper()

	resp, err := compose.R(t.Context()).SetBody(map[string]string{"name": name}).Put("/api/devices/" + uid)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
}
