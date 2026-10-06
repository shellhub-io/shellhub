package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnrollmentPolicy enrolls devices with requests built by hand against the endpoint an agent
// enrolls through, and reads back what each provisioning key's policy made of them. The cases
// share one stack, so every device carries a hostname and a MAC address no other case uses:
// accepting a device merges it into any accepted device with the same MAC. What an agent does with
// a refusal is covered by [TestProvisioningKeyEnrollment].
func TestEnrollmentPolicy(t *testing.T) {
	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	t.Run("automatic", func(t *testing.T) { testAutomaticEnrollment(t, compose) })
	t.Run("manual", func(t *testing.T) { testManualEnrollment(t, compose) })
}

func newKeyedDeviceAuthRequest(t *testing.T, key, hostname, mac string) requests.DeviceAuth {
	t.Helper()

	req := newDeviceAuthRequest(t, hostname, mac)
	req.TenantID = ""
	req.ProvisioningKey = key

	return req
}

func enroll(t *testing.T, compose *environment.DockerCompose, req requests.DeviceAuth) models.Device {
	t.Helper()

	return requireDevice(t, compose, authDevice(t, compose, req).UID)
}

func requireDevice(t *testing.T, compose *environment.DockerCompose, uid string) models.Device {
	t.Helper()

	device, resp, err := compose.GetDevice(t.Context(), uid)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return *device
}

func postDeviceAuth(ctx context.Context, compose *environment.DockerCompose, req requests.DeviceAuth) (*resty.Response, error) {
	return compose.Anonymous(ctx).SetBody(req).Post("/api/devices/auth")
}

func requireAuthRefused(t *testing.T, compose *environment.DockerCompose, req requests.DeviceAuth) {
	t.Helper()

	resp, err := postDeviceAuth(t.Context(), compose, req)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode(), resp.String())

	for _, device := range compose.ListDevices(t, models.DeviceStatusEmpty) {
		assert.NotEqual(t, req.Hostname, device.Name, "a refused enrollment left a device behind")
	}
}

func tagNames(device models.Device) []string {
	names := make([]string, 0, len(device.Tags))
	for _, tag := range device.Tags {
		names = append(names, tag.Name)
	}

	return names
}
