package main

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testReRegistration(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("a removed device re-registers through its key with the key's current tags and ephemeral setting", func(t *testing.T) {
		key := compose.CreateProvisioningKey(t, &requests.CreateProvisioningKey{
			Name: "reregistered",
			Mode: string(models.ProvisioningKeyModeAutomatic),
			Tags: []string{"enrolled"},
		})
		req := newKeyedDeviceAuthRequest(t, key.Key, "reregistered", "02:00:00:00:50:01")

		enrolled := enroll(t, compose, req)
		require.Equal(t, models.DeviceStatusAccepted, enrolled.Status)
		require.False(t, enrolled.Ephemeral)

		compose.DeleteDevice(t, enrolled.UID)
		require.Equal(t, models.DeviceStatusRemoved, requireDevice(t, compose, enrolled.UID).Status)

		compose.UpdateProvisioningKey(t, "reregistered", map[string]any{
			"tags":              []string{"reenrolled"},
			"ephemeral":         true,
			"ephemeral_timeout": 4,
		})

		reregistered := enroll(t, compose, req)

		assert.Equal(t, enrolled.UID, reregistered.UID)
		assert.Equal(t, models.DeviceStatusAccepted, reregistered.Status)
		assert.Equal(t, key.ID, reregistered.ProvisioningKeyID)
		assert.ElementsMatch(t, []string{"enrolled", "reenrolled"}, tagNames(reregistered.Tags),
			"re-registration adds the key's current tags to those the removed device kept")
		assert.True(t, reregistered.Ephemeral)
		assert.Equal(t, 4, reregistered.EphemeralTimeout)
		compose.AwaitProvisioningKeyUses(t, "reregistered", 2)

		events := compose.ProvisioningKeyHistory(t, key.ID)
		require.Len(t, events, 2)
		assert.True(t, events[0].ReRegistration, "the newest event records the re-registration")
		assert.True(t, events[0].Ephemeral)
		assert.False(t, events[1].ReRegistration)
	})
}
