package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeviceLifecycle drives a device's status and name through the API, with devices enrolled by
// hand so no agent container is needed.
func TestDeviceLifecycle(t *testing.T) {
	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	t.Run("status transitions", func(t *testing.T) {
		cases := []struct {
			description    string
			setup          []environment.DeviceStatusAction
			transition     environment.DeviceStatusAction
			expectedCode   int
			expectedStatus models.DeviceStatus
		}{
			{
				description:    "a pending device can be accepted",
				transition:     environment.DeviceActionAccept,
				expectedCode:   http.StatusOK,
				expectedStatus: models.DeviceStatusAccepted,
			},
			{
				description:    "a pending device can be rejected",
				transition:     environment.DeviceActionReject,
				expectedCode:   http.StatusOK,
				expectedStatus: models.DeviceStatusRejected,
			},
			{
				description:    "a rejected device can be accepted",
				setup:          []environment.DeviceStatusAction{environment.DeviceActionReject},
				transition:     environment.DeviceActionAccept,
				expectedCode:   http.StatusOK,
				expectedStatus: models.DeviceStatusAccepted,
			},
			{
				description:    "an accepted device cannot be rejected",
				setup:          []environment.DeviceStatusAction{environment.DeviceActionAccept},
				transition:     environment.DeviceActionReject,
				expectedCode:   http.StatusBadRequest,
				expectedStatus: models.DeviceStatusAccepted,
			},
			{
				description:    "an accepted device cannot return to pending",
				setup:          []environment.DeviceStatusAction{environment.DeviceActionAccept},
				transition:     environment.DeviceActionPending,
				expectedCode:   http.StatusBadRequest,
				expectedStatus: models.DeviceStatusAccepted,
			},
		}

		for i, tc := range cases {
			t.Run(tc.description, func(t *testing.T) {
				uid := enrollDevice(t, compose, fmt.Sprintf("status-%d", i), fmt.Sprintf("02:00:00:00:07:%02x", i))
				for _, transition := range tc.setup {
					compose.UpdateDeviceStatus(t, uid, transition)
				}

				resp, err := compose.PatchDeviceStatus(t.Context(), uid, tc.transition)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedCode, resp.StatusCode(), resp.String())

				device, resp, err := compose.GetDevice(t.Context(), uid)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
				assert.Equal(t, tc.expectedStatus, device.Status)
			})
		}
	})

	t.Run("an accepted device is kept as removed", func(t *testing.T) {
		uid := enrollDevice(t, compose, "removal-accepted", "02:00:00:00:08:ff")
		compose.UpdateDeviceStatus(t, uid, environment.DeviceActionAccept)

		compose.DeleteDevice(t, uid)

		device, resp, err := compose.GetDevice(t.Context(), uid)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Equal(t, models.DeviceStatusRemoved, device.Status)
		assert.Contains(t, deviceUIDs(compose.ListDevices(t, models.DeviceStatusRemoved)), uid)
	})

	t.Run("a non-accepted device is hard-deleted", func(t *testing.T) {
		cases := []struct {
			description string
			setup       []environment.DeviceStatusAction
			status      models.DeviceStatus
		}{
			{
				description: "a pending device",
				status:      models.DeviceStatusPending,
			},
			{
				description: "a rejected device",
				setup:       []environment.DeviceStatusAction{environment.DeviceActionReject},
				status:      models.DeviceStatusRejected,
			},
		}

		for i, tc := range cases {
			t.Run(tc.description, func(t *testing.T) {
				uid := enrollDevice(t, compose, fmt.Sprintf("removal-%d", i), fmt.Sprintf("02:00:00:00:08:%02x", i))
				for _, transition := range tc.setup {
					compose.UpdateDeviceStatus(t, uid, transition)
				}
				require.Contains(t, deviceUIDs(compose.ListDevices(t, tc.status)), uid)

				compose.DeleteDevice(t, uid)

				_, resp, err := compose.GetDevice(t.Context(), uid)
				require.NoError(t, err)
				assert.Equal(t, http.StatusNotFound, resp.StatusCode(), resp.String())

				for _, status := range []models.DeviceStatus{
					models.DeviceStatusEmpty,
					models.DeviceStatusAccepted,
					models.DeviceStatusPending,
					models.DeviceStatusRejected,
					models.DeviceStatusRemoved,
					models.DeviceStatusUnused,
				} {
					assert.NotContains(t, deviceUIDs(compose.ListDevices(t, status)), uid, "listed under status %q", status)
				}
			})
		}
	})

	t.Run("renaming", func(t *testing.T) {
		taken := enrollDevice(t, compose, "taken", "02:00:00:00:09:ff")
		compose.UpdateDeviceStatus(t, taken, environment.DeviceActionAccept)

		cases := []struct {
			description  string
			method       string
			hostname     string
			name         string
			expectedCode int
			expectedName string
		}{
			{
				description:  "through PUT to a name another accepted device holds fails",
				method:       http.MethodPut,
				hostname:     "put-duplicate",
				name:         "taken",
				expectedCode: http.StatusConflict,
				expectedName: "put-duplicate",
			},
			{
				description:  "through PUT to that name in another case fails",
				method:       http.MethodPut,
				hostname:     "put-duplicate-in-another-case",
				name:         "TAKEN",
				expectedCode: http.StatusConflict,
				expectedName: "put-duplicate-in-another-case",
			},
			{
				description:  "through PUT stores the new name in lowercase",
				method:       http.MethodPut,
				hostname:     "put-mixed-case",
				name:         "Put-Renamed",
				expectedCode: http.StatusOK,
				expectedName: "put-renamed",
			},
			{
				description:  "through PATCH to a name another accepted device holds fails",
				method:       http.MethodPatch,
				hostname:     "patch-duplicate",
				name:         "taken",
				expectedCode: http.StatusConflict,
				expectedName: "patch-duplicate",
			},
			{
				description:  "through PATCH to that name in another case fails",
				method:       http.MethodPatch,
				hostname:     "patch-duplicate-in-another-case",
				name:         "TAKEN",
				expectedCode: http.StatusConflict,
				expectedName: "patch-duplicate-in-another-case",
			},
			{
				description:  "through PATCH stores the new name in lowercase",
				method:       http.MethodPatch,
				hostname:     "patch-mixed-case",
				name:         "Patch-Renamed",
				expectedCode: http.StatusOK,
				expectedName: "patch-renamed",
			},
		}

		for i, tc := range cases {
			t.Run(tc.description, func(t *testing.T) {
				uid := enrollDevice(t, compose, tc.hostname, fmt.Sprintf("02:00:00:00:09:%02x", i))
				compose.UpdateDeviceStatus(t, uid, environment.DeviceActionAccept)

				resp, err := compose.R(t.Context()).
					SetBody(map[string]string{"name": tc.name}).
					Execute(tc.method, "/api/devices/"+uid)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedCode, resp.StatusCode(), resp.String())

				device, resp, err := compose.GetDevice(t.Context(), uid)
				require.NoError(t, err)
				require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
				assert.Equal(t, tc.expectedName, device.Name)
			})
		}
	})
}

func enrollDevice(t *testing.T, compose *environment.DockerCompose, hostname, mac string) string {
	t.Helper()

	res := authDevice(t, compose, newDeviceAuthRequest(t, hostname, mac))
	require.Equal(t, models.DeviceStatusPending, res.Status)

	return res.UID
}

func deviceUIDs(devices []models.Device) []string {
	uids := make([]string, 0, len(devices))
	for _, device := range devices {
		uids = append(uids, device.UID)
	}

	return uids
}
