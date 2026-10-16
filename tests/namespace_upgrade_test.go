package main

import (
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamespaceUpgrade(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	compose := newSSHEnvironment(t, ctx, legacyNamespace)

	acceptedReq := newDeviceAuthRequest(t, "accepted-before-upgrade", "02:00:00:00:48:04")
	accepted := enroll(t, compose, acceptedReq)
	compose.UpdateDeviceStatus(t, accepted.UID, environment.DeviceActionAccept)

	pendingReq := newDeviceAuthRequest(t, "pending-before-upgrade", "02:00:00:00:48:01")
	pending := enroll(t, compose, pendingReq)
	require.Equal(t, models.DeviceStatusPending, pending.Status)

	require.True(t, compose.GetNamespace(t, ShellHubNamespace).Settings.IsLegacy(), "the fixture namespace is legacy")

	compose.UpgradeNamespace(t, ShellHubNamespace)

	t.Run("the namespace is new and seeded with the owner's access policy", func(t *testing.T) {
		assert.False(t, compose.GetNamespace(t, ShellHubNamespace).Settings.IsLegacy())

		policies := []models.AccessPolicy{}
		resp, err := compose.R(ctx).SetResult(&policies).Get("/api/access-policies")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		require.Len(t, policies, 1)
		assert.Equal(t, "Owner access", policies[0].Name)
	})

	t.Run("a second upgrade is refused", func(t *testing.T) {
		resp, err := compose.PostNamespaceUpgrade(ctx, ShellHubNamespace)
		require.NoError(t, err)
		assert.Equal(t, http.StatusConflict, resp.StatusCode(), resp.String())
	})

	t.Run("the device pending before the upgrade is still there and acceptable", func(t *testing.T) {
		requireStatusOnReauth(t, compose, pendingReq, pending.UID, models.DeviceStatusPending)

		compose.UpdateDeviceStatus(t, pending.UID, environment.DeviceActionAccept)
		assert.Equal(t, models.DeviceStatusAccepted, requireDevice(t, compose, pending.UID).Status)
	})

	t.Run("the device accepted before the upgrade re-authenticates by tenant alone", func(t *testing.T) {
		requireStatusOnReauth(t, compose, acceptedReq, accepted.UID, models.DeviceStatusAccepted)
	})

	t.Run("a new device presenting only the tenant lands pending", func(t *testing.T) {
		device := enroll(t, compose, newDeviceAuthRequest(t, "keyless-after-upgrade", "02:00:00:00:48:02"))
		assert.Equal(t, models.DeviceStatusPending, device.Status)
	})

	t.Run("the provisioning key API answers", func(t *testing.T) {
		key := compose.CreateProvisioningKey(t, &requests.CreateProvisioningKey{Name: "after-upgrade", Mode: string(models.ProvisioningKeyModeManual)})
		assert.Equal(t, models.DeviceStatusPending, enroll(t, compose, newKeyedDeviceAuthRequest(t, key.Key, "keyed-after-upgrade", "02:00:00:00:48:03")).Status)
	})
}
