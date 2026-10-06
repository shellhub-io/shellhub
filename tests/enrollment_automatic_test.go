package main

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testAutomaticEnrollment(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	newKey := func(t *testing.T, req *requests.CreateProvisioningKey) string {
		t.Helper()

		req.Mode = string(models.ProvisioningKeyModeAutomatic)

		return compose.CreateProvisioningKey(t, req).Key
	}

	t.Run("a key past its expiry enrolls nothing more", func(t *testing.T) {
		days := 1
		key := newKey(t, &requests.CreateProvisioningKey{Name: "automatic-expiring", ExpiresIn: &days})

		before := enroll(t, compose, newKeyedDeviceAuthRequest(t, key, "automatic-before-expiry", "02:00:00:00:10:01"))
		require.Equal(t, models.DeviceStatusAccepted, before.Status)

		compose.ExpireProvisioningKey(t, "automatic-expiring")

		requireAuthRefused(t, compose, newKeyedDeviceAuthRequest(t, key, "automatic-after-expiry", "02:00:00:00:10:02"))
	})

	t.Run("a disabled key enrolls nothing until it is enabled again", func(t *testing.T) {
		key := newKey(t, &requests.CreateProvisioningKey{Name: "automatic-disabled"})
		req := newKeyedDeviceAuthRequest(t, key, "automatic-disabled", "02:00:00:00:10:03")

		compose.UpdateProvisioningKey(t, "automatic-disabled", map[string]any{"disabled": true})

		requireAuthRefused(t, compose, req)

		compose.UpdateProvisioningKey(t, "automatic-disabled", map[string]any{"disabled": false})

		assert.Equal(t, models.DeviceStatusAccepted, enroll(t, compose, req).Status)
	})

	t.Run("a revoked key enrolls nothing more", func(t *testing.T) {
		key := newKey(t, &requests.CreateProvisioningKey{Name: "automatic-revoked"})

		before := enroll(t, compose, newKeyedDeviceAuthRequest(t, key, "automatic-before-revoke", "02:00:00:00:10:04"))
		require.Equal(t, models.DeviceStatusAccepted, before.Status)

		compose.UpdateProvisioningKey(t, "automatic-revoked", map[string]any{"revoked": true})

		requireAuthRefused(t, compose, newKeyedDeviceAuthRequest(t, key, "automatic-after-revoke", "02:00:00:00:10:05"))
	})

	t.Run("the device is tagged with the key's tags", func(t *testing.T) {
		key := newKey(t, &requests.CreateProvisioningKey{Name: "automatic-tagged", Tags: []string{"fleet", "edge"}})

		device := enroll(t, compose, newKeyedDeviceAuthRequest(t, key, "automatic-tagged", "02:00:00:00:10:06"))

		assert.Equal(t, models.DeviceStatusAccepted, device.Status)
		assert.ElementsMatch(t, []string{"fleet", "edge"}, tagNames(device))
	})

	t.Run("the device is ephemeral when the key is", func(t *testing.T) {
		cases := []struct {
			description string
			name        string
			ephemeral   bool
			timeout     int
			mac         string
		}{
			{description: "an ephemeral key", name: "automatic-ephemeral", ephemeral: true, timeout: 7, mac: "02:00:00:00:10:07"},
			{description: "a lasting key", name: "automatic-lasting", ephemeral: false, timeout: 0, mac: "02:00:00:00:10:08"},
		}

		for _, tc := range cases {
			t.Run(tc.description, func(t *testing.T) {
				key := newKey(t, &requests.CreateProvisioningKey{Name: tc.name, Ephemeral: tc.ephemeral, EphemeralTimeout: tc.timeout})

				device := enroll(t, compose, newKeyedDeviceAuthRequest(t, key, tc.name, tc.mac))

				assert.Equal(t, models.DeviceStatusAccepted, device.Status)
				assert.Equal(t, tc.ephemeral, device.Ephemeral)
				assert.Equal(t, tc.timeout, device.EphemeralTimeout)
			})
		}
	})
}
