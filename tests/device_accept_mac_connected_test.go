package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

// TestDeviceAcceptMACTwinConnected covers accepting a device whose MAC address an accepted device
// holds while its agent is still connected. The accept would merge the connected device away, so it
// is refused and both devices stay as they were. A newcomer an automatic key would accept stays
// pending instead, and is accepted on a later authentication, past the interval that throttles
// re-evaluation, once the twin's agent is gone. The cases share one stack, so each twin agent
// carries a hostname no other case uses.
func TestDeviceAcceptMACTwinConnected(t *testing.T) {
	t.Parallel()

	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	t.Run("an accept is refused on identity.mac and leaves both devices intact", func(t *testing.T) {
		_, twin := startNamedAcceptedAgent(t, compose, "connected-twin")
		newcomer := enrollDevice(t, compose, "connected-newcomer", twin.Identity.MAC)

		refusal := requireAcceptConflict(t, compose, newcomer)
		assert.Contains(t, refusal.Fields, "identity.mac")

		kept := requireDevice(t, compose, twin.UID)
		assert.Equal(t, models.DeviceStatusAccepted, kept.Status)
		assert.True(t, kept.Online, "the connected twin was merged away")
		assert.Equal(t, twin.Name, kept.Name)
	})

	t.Run("an automatic-key newcomer stays pending until the twin goes offline", func(t *testing.T) {
		key := compose.CreateProvisioningKey(t, &requests.CreateProvisioningKey{
			Name: "twin-automatic",
			Mode: string(models.ProvisioningKeyModeAutomatic),
		}).Key

		agent, twin := startNamedAcceptedAgent(t, compose, "automatic-twin")
		req := newKeyedDeviceAuthRequest(t, key, "automatic-newcomer", twin.Identity.MAC)

		newcomer := enroll(t, compose, req)
		require.Equal(t, models.DeviceStatusPending, newcomer.Status, "the automatic key merged the connected twin away")

		compose.AgeLastEnrollmentAttempt(t, newcomer.UID, models.EnrollmentReconcileInterval+time.Second)
		requireStatusOnReauth(t, compose, req, newcomer.UID, models.DeviceStatusPending)
		assert.True(t, requireDevice(t, compose, twin.UID).Online, "the connected twin was merged away")

		require.NoError(t, agent.Stop(t.Context(), nil))
		compose.AwaitDeviceOffline(t, twin.UID)

		compose.AgeLastEnrollmentAttempt(t, newcomer.UID, models.EnrollmentReconcileInterval+time.Second)
		requireStatusOnReauth(t, compose, req, newcomer.UID, models.DeviceStatusAccepted)

		_, resp, err := compose.GetDevice(t.Context(), twin.UID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode(), "the accept left the offline twin behind: %s", resp.String())
	})
}

func startNamedAcceptedAgent(t *testing.T, compose *environment.DockerCompose, hostname string) (testcontainers.Container, models.Device) {
	t.Helper()

	agent := startAgent(t, t.Context(), compose, NewAgentContainerWithHostname(hostname))

	pending := compose.AwaitDeviceNamed(t, hostname)
	compose.UpdateDeviceStatus(t, pending.UID, environment.DeviceActionAccept)

	return agent, compose.AwaitDeviceOnline(t, pending.UID)
}
