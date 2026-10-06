package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testKeylessEnrollment(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	legacy := compose.ProvisioningKey(t, string(models.ProvisioningKeyTypeLegacy))
	require.Equal(t, models.ProvisioningKeyTypeLegacy, legacy.Type)

	t.Run("a tenant-only enrollment is attributed to the namespace's legacy key", func(t *testing.T) {
		device := enroll(t, compose, newDeviceAuthRequest(t, "keyless", "02:00:00:00:40:01"))

		assert.Equal(t, models.DeviceStatusPending, device.Status)
		assert.Equal(t, legacy.ID, device.ProvisioningKeyID)
		assert.NotNil(t, eventOf(compose.ProvisioningKeyHistory(t, legacy.ID), device.UID))
	})

	t.Run("a disabled legacy key refuses tenant-only enrollment until it is enabled again", func(t *testing.T) {
		req := newDeviceAuthRequest(t, "keyless-disabled", "02:00:00:00:40:02")

		compose.UpdateProvisioningKey(t, legacy.Name, map[string]any{"disabled": true})
		t.Cleanup(func() {
			resp, err := compose.PatchProvisioningKey(context.Background(), legacy.Name, map[string]any{"disabled": false})
			if assert.NoError(t, err) {
				assert.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
			}
		})

		requireAuthRefused(t, compose, req)

		compose.UpdateProvisioningKey(t, legacy.Name, map[string]any{"disabled": false})

		assert.Equal(t, models.DeviceStatusPending, enroll(t, compose, req).Status)
	})
}

func testSystemKeyPresented(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	cases := []struct {
		keyType   models.ProvisioningKeyType
		plaintext string
		mac       string
	}{
		{keyType: models.ProvisioningKeyTypeLegacy, plaintext: "system:" + ShellHubNamespace, mac: "02:00:00:00:40:03"},
		{keyType: models.ProvisioningKeyTypePairing, plaintext: "system:pairing:" + ShellHubNamespace, mac: "02:00:00:00:40:04"},
	}

	for _, tc := range cases {
		t.Run("the "+string(tc.keyType)+" key", func(t *testing.T) {
			digest := sha256.Sum256([]byte(tc.plaintext))
			require.Equal(t, hex.EncodeToString(digest[:]), compose.ProvisioningKey(t, string(tc.keyType)).ID,
				"the plaintext must be the one the system key is the digest of")

			t.Run("enrolls nothing in place of a tenant id", func(t *testing.T) {
				requireAuthRefused(t, compose, newKeyedDeviceAuthRequest(t, tc.plaintext, "system-"+string(tc.keyType), tc.mac))
			})

			t.Run("enrolls nothing alongside the tenant id", func(t *testing.T) {
				req := newKeyedDeviceAuthRequest(t, tc.plaintext, "system-"+string(tc.keyType)+"-tenant", tc.mac)
				req.TenantID = ShellHubNamespace

				requireAuthRefused(t, compose, req)
			})
		})
	}
}
