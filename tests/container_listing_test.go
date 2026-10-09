package main

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContainerListing enrolls by hand one device reporting the connector platform, as a
// connector registers each container it serves, and one reporting the native platform, as an
// agent does, then reads back where the API lists each: /api/containers serves the connector's
// devices alone, and /api/devices every device but those.
func TestContainerListing(t *testing.T) {
	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	containerReq := newDeviceAuthRequest(t, "listed-container", "02:00:00:00:a1:01")
	containerReq.Info.Platform = "connector"
	container := enroll(t, compose, containerReq)
	require.Equal(t, "connector", container.Info.Platform)

	native := enroll(t, compose, newDeviceAuthRequest(t, "listed-native", "02:00:00:00:a1:02"))
	require.Equal(t, "native", native.Info.Platform)

	t.Run("pending devices are listed apart", func(t *testing.T) {
		for _, status := range []models.DeviceStatus{models.DeviceStatusEmpty, models.DeviceStatusPending} {
			requireListedApart(t, compose, status, container.UID, native.UID)
		}
	})

	t.Run("accepted devices are listed apart", func(t *testing.T) {
		compose.UpdateDeviceStatus(t, container.UID, environment.DeviceActionAccept)
		compose.UpdateDeviceStatus(t, native.UID, environment.DeviceActionAccept)

		for _, status := range []models.DeviceStatus{models.DeviceStatusEmpty, models.DeviceStatusAccepted} {
			requireListedApart(t, compose, status, container.UID, native.UID)
		}
	})
}

func requireListedApart(t *testing.T, compose *environment.DockerCompose, status models.DeviceStatus, container, native string) {
	t.Helper()

	containers := deviceUIDs(compose.ListContainers(t, status))
	assert.Contains(t, containers, container, "the container is missing from the containers listed under status %q", status)
	assert.NotContains(t, containers, native, "the native device is among the containers listed under status %q", status)

	devices := deviceUIDs(compose.ListDevices(t, status))
	assert.Contains(t, devices, native, "the native device is missing from the devices listed under status %q", status)
	assert.NotContains(t, devices, container, "the container is among the devices listed under status %q", status)
}
