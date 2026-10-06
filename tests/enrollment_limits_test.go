package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testEnrollmentLimits(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("a key that has used up its enrollments blocks accepting another of its devices", func(t *testing.T) {
		key := compose.CreateProvisioningKey(t, &requests.CreateProvisioningKey{
			Name:       "manual-single-use",
			Mode:       string(models.ProvisioningKeyModeManual),
			UsageLimit: 1,
		})

		first := enroll(t, compose, newKeyedDeviceAuthRequest(t, key.Key, "single-use-first", "02:00:00:00:60:01"))
		second := enroll(t, compose, newKeyedDeviceAuthRequest(t, key.Key, "single-use-second", "02:00:00:00:60:02"))
		require.Equal(t, models.DeviceStatusPending, first.Status)
		require.Equal(t, models.DeviceStatusPending, second.Status)

		compose.UpdateDeviceStatus(t, first.UID, environment.DeviceActionAccept)
		compose.AwaitProvisioningKeyUses(t, "manual-single-use", 1)

		resp, err := compose.PatchDeviceStatus(t.Context(), second.UID, environment.DeviceActionAccept)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())

		assert.Equal(t, models.DeviceStatusPending, requireDevice(t, compose, second.UID).Status)
		compose.RequireProvisioningKeyUsesHold(t, "manual-single-use", 1)
	})

	t.Run("a namespace at its device limit keeps an automatic key's device pending until the limit is lifted", func(t *testing.T) {
		key := compose.CreateProvisioningKey(t, &requests.CreateProvisioningKey{
			Name: "automatic-limited",
			Mode: string(models.ProvisioningKeyModeAutomatic),
		})

		within := enroll(t, compose, newKeyedDeviceAuthRequest(t, key.Key, "limit-within", "02:00:00:00:60:03"))
		require.Equal(t, models.DeviceStatusAccepted, within.Status)

		namespace := new(responses.Namespace)
		resp, err := compose.R(t.Context()).SetResult(namespace).Get("/api/namespaces/" + ShellHubNamespace)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		require.NoError(t, compose.SetNamespaceMaxDevices(t.Context(), ShellHubNamespace, int(namespace.DevicesAcceptedCount)))
		t.Cleanup(func() {
			assert.NoError(t, compose.SetNamespaceMaxDevices(context.Background(), ShellHubNamespace, -1))
		})

		req := newKeyedDeviceAuthRequest(t, key.Key, "limit-over", "02:00:00:00:60:04")
		over := enroll(t, compose, req)

		assert.Equal(t, models.DeviceStatusPending, over.Status)
		compose.RequireProvisioningKeyUsesHold(t, "automatic-limited", 1)

		require.NoError(t, compose.SetNamespaceMaxDevices(t.Context(), ShellHubNamespace, -1))

		awaitStatusOnReauth(t, compose, req, over.UID, models.DeviceStatusAccepted)
		compose.AwaitProvisioningKeyUses(t, "automatic-limited", 2)
	})
}
