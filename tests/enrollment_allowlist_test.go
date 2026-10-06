package main

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testAllowlistEnrollment(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("the claimed MAC address decides the device", func(t *testing.T) {
		key := compose.CreateProvisioningKey(t, &requests.CreateProvisioningKey{
			Name:              "allowlist",
			Mode:              string(models.ProvisioningKeyModeAllowlist),
			AllowedIdentities: []string{"02:00:00:00:30:01"},
		})

		cases := []struct {
			description string
			hostname    string
			mac         string
			status      models.DeviceStatus
		}{
			{
				description: "a listed MAC address is accepted",
				hostname:    "allowlist-listed",
				mac:         "02:00:00:00:30:01",
				status:      models.DeviceStatusAccepted,
			},
			{
				description: "an unlisted MAC address is rejected",
				hostname:    "allowlist-unlisted",
				mac:         "02:00:00:00:30:02",
				status:      models.DeviceStatusRejected,
			},
		}

		for _, tc := range cases {
			t.Run(tc.description, func(t *testing.T) {
				device := enroll(t, compose, newKeyedDeviceAuthRequest(t, key.Key, tc.hostname, tc.mac))

				assert.Equal(t, tc.status, device.Status)
			})
		}
	})

	t.Run("a pending device is decided on its next authentication once its key turns to an allowlist", func(t *testing.T) {
		key := compose.CreateProvisioningKey(t, &requests.CreateProvisioningKey{
			Name: "allowlist-later",
			Mode: string(models.ProvisioningKeyModeManual),
		})

		listedReq := newKeyedDeviceAuthRequest(t, key.Key, "allowlist-later-listed", "02:00:00:00:30:03")
		unlistedReq := newKeyedDeviceAuthRequest(t, key.Key, "allowlist-later-unlisted", "02:00:00:00:30:04")

		listed := enroll(t, compose, listedReq)
		unlisted := enroll(t, compose, unlistedReq)
		require.Equal(t, models.DeviceStatusPending, listed.Status)
		require.Equal(t, models.DeviceStatusPending, unlisted.Status)

		compose.UpdateProvisioningKey(t, "allowlist-later", map[string]any{
			"mode":               models.ProvisioningKeyModeAllowlist,
			"allowed_identities": []string{"02:00:00:00:30:03"},
		})

		awaitStatusOnReauth(t, compose, listedReq, listed.UID, models.DeviceStatusAccepted)
		awaitStatusOnReauth(t, compose, unlistedReq, unlisted.UID, models.DeviceStatusRejected)
	})
}
