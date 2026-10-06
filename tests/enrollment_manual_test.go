package main

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/require"
)

func testManualEnrollment(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	cases := []struct {
		description string
		name        string
		mac         string
		decision    environment.DeviceStatusAction
		status      models.DeviceStatus
		wantCharge  func(t *testing.T, name string)
	}{
		{
			description: "accepting the pending device charges the key a use",
			name:        "manual-accepted",
			mac:         "02:00:00:00:20:01",
			decision:    environment.DeviceActionAccept,
			status:      models.DeviceStatusAccepted,
			wantCharge: func(t *testing.T, name string) {
				t.Helper()

				compose.AwaitProvisioningKeyUses(t, name, 1)
			},
		},
		{
			description: "rejecting the pending device charges the key nothing",
			name:        "manual-rejected",
			mac:         "02:00:00:00:20:02",
			decision:    environment.DeviceActionReject,
			status:      models.DeviceStatusRejected,
			wantCharge: func(t *testing.T, name string) {
				t.Helper()

				compose.RequireProvisioningKeyUsesHold(t, name, 0)
				compose.RequireProvisioningKeyUnused(t, name)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			key := compose.CreateProvisioningKey(t, &requests.CreateProvisioningKey{
				Name: tc.name,
				Mode: string(models.ProvisioningKeyModeManual),
			})

			device := enroll(t, compose, newKeyedDeviceAuthRequest(t, key.Key, tc.name, tc.mac))
			require.Equal(t, models.DeviceStatusPending, device.Status)
			compose.RequireProvisioningKeyUnused(t, tc.name)

			compose.UpdateDeviceStatus(t, device.UID, tc.decision)

			require.Equal(t, tc.status, requireDevice(t, compose, device.UID).Status)
			tc.wantCharge(t, tc.name)
		})
	}
}
