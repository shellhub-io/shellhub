package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
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
}

func testAgentLoginWithTenant(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("an agent given the tenant lands pending and login prints a code that accepts it", func(t *testing.T) {
		login := startAgentCommand(t, compose, []string{"./agent", "login"}, NewAgentContainerWithHostname("legacy-login"))
		code := awaitAgentPairingCode(t, login)

		preview := requireLoginCodePreview(t, compose, code)
		require.Equal(t, "legacy-login", preview.Name)
		require.Equal(t, models.DeviceStatusPending, preview.Status)

		compose.UpdateDeviceStatus(t, preview.UID, environment.DeviceActionAccept)

		assert.Zero(t, awaitAgentExit(t, login))
	})
}

func testEnrollmentUID(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("a device enrolled by key keeps its uid with the tenant it persisted, with or without the key", func(t *testing.T) {
		key := requireHarnessProvisioningKey(t, compose)
		keyed := newKeyedDeviceAuthRequest(t, key, "uid-by-key", "02:00:00:00:40:07")
		device := enroll(t, compose, keyed)

		withTenant := keyed
		withTenant.TenantID = ShellHubNamespace
		reauthDevice(t, compose, withTenant, device.UID)
		require.Equal(t, device.UID, authDevice(t, compose, withTenant).UID)

		tenantOnly := withTenant
		tenantOnly.ProvisioningKey = ""
		reauthDevice(t, compose, tenantOnly, device.UID)

		require.Equal(t, device.UID, authDevice(t, compose, tenantOnly).UID)
	})
}

func testKeylessRefused(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("a tenant-only enrollment is refused with 403 and enrolls nothing", func(t *testing.T) {
		req := newDeviceAuthRequest(t, "keyless-refused", "02:00:00:00:40:02")

		requireAuthForbidden(t, compose, req)
	})
}

func testProvisioningKeyRefusedInLegacy(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("a presented provisioning key is refused with 403 and enrolls nothing", func(t *testing.T) {
		req := newKeyedDeviceAuthRequest(t, unissuedProvisioningKey, "legacy-keyed", "02:00:00:00:40:05")
		req.TenantID = ShellHubNamespace

		requireAuthForbidden(t, compose, req)
	})
}

func testPairingRefusedInLegacy(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("accepting a pairing into the namespace is refused with 403 and enrolls nothing", func(t *testing.T) {
		req := newPairingRequest(t, "legacy-pairing", "02:00:00:00:40:06")
		pairing := startPairing(t, compose, req)

		_, resp, err := postPairingAccept(compose.R(t.Context()), pairing.Code)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())
		assert.NotContains(t, deviceNames(compose.ListDevices(t, models.DeviceStatusEmpty)), req.Hostname)
	})
}

func testProvisioningKeyAPIInLegacy(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("creating a key is refused", func(t *testing.T) {
		resp, err := compose.R(t.Context()).
			SetBody(&requests.CreateProvisioningKey{Name: "fleet", Mode: string(models.ProvisioningKeyModeManual)}).
			Post("/api/namespaces/provisioning-key")
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())
	})

	t.Run("editing the legacy key is refused", func(t *testing.T) {
		resp, err := compose.PatchProvisioningKey(t.Context(), string(models.ProvisioningKeyTypeLegacy), map[string]any{"disabled": true})
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())
	})

	t.Run("the system keys are still listed", func(t *testing.T) {
		assert.Equal(t, models.ProvisioningKeyTypeLegacy, compose.ProvisioningKey(t, string(models.ProvisioningKeyTypeLegacy)).Type)
		assert.Equal(t, models.ProvisioningKeyTypePairing, compose.ProvisioningKey(t, string(models.ProvisioningKeyTypePairing)).Type)
	})
}

func testSystemKeyPresented(t *testing.T, compose *environment.DockerCompose, alongsideTenantStatus int) {
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

				requireAuthAnswers(t, compose, req, alongsideTenantStatus)
			})
		})
	}
}
