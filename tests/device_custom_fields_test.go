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

const maxDeviceCustomFields = 20

// TestDeviceCustomFieldLimit covers the cap of twenty custom fields a device holds. A device at the
// cap refuses a new key and keeps its fields, while overwriting a key it holds still works, and
// removing a key makes room for another.
func TestDeviceCustomFieldLimit(t *testing.T) {
	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	uid := enrollDevice(t, compose, "customfields", "02:00:00:00:15:01")
	compose.UpdateDeviceStatus(t, uid, environment.DeviceActionAccept)

	for i := range maxDeviceCustomFields {
		setCustomField(t, compose, uid, fmt.Sprintf("field%02d", i), "value", http.StatusOK)
	}

	t.Run("a device at the cap refuses a new key", func(t *testing.T) {
		setCustomField(t, compose, uid, "onetoomany", "value", http.StatusForbidden)

		fields := requireDevice(t, compose, uid).CustomFields
		assert.Len(t, fields, maxDeviceCustomFields)
		assert.NotContains(t, fields, "onetoomany")
	})

	t.Run("a device at the cap still overwrites a key it holds", func(t *testing.T) {
		setCustomField(t, compose, uid, "field00", "overwritten", http.StatusOK)

		fields := requireDevice(t, compose, uid).CustomFields
		assert.Len(t, fields, maxDeviceCustomFields)
		assert.Equal(t, "overwritten", fields["field00"])
	})

	t.Run("removing a key makes room for another", func(t *testing.T) {
		resp, err := compose.R(t.Context()).Delete("/api/devices/" + uid + "/custom_fields/field01")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		setCustomField(t, compose, uid, "replacement", "value", http.StatusOK)

		fields := requireDevice(t, compose, uid).CustomFields
		assert.Len(t, fields, maxDeviceCustomFields)
		assert.NotContains(t, fields, "field01")
		assert.Equal(t, "value", fields["replacement"])
	})
}

func setCustomField(t *testing.T, compose *environment.DockerCompose, uid, key, value string, status int) {
	t.Helper()

	resp, err := compose.R(t.Context()).
		SetBody(map[string]string{"value": value}).
		Put("/api/devices/" + uid + "/custom_fields/" + key)
	require.NoError(t, err)
	require.Equal(t, status, resp.StatusCode(), "setting %q: %s", key, resp.String())
}
