package main

import (
	"maps"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigConnectorFromEnvCredential(t *testing.T) {
	base := map[string]string{
		"SHELLHUB_SERVER_ADDRESS": "http://localhost",
		"SHELLHUB_PRIVATE_KEYS":   "/tmp/keys",
	}

	withVars := func(extra map[string]string) envs.Backend {
		values := make(map[string]string, len(base)+len(extra))
		maps.Copy(values, base)
		maps.Copy(values, extra)

		return envs.Map(values)
	}

	t.Run("a tenant alone is a namespace credential", func(t *testing.T) {
		cfg, _, err := LoadConfigConnectorFromEnv(withVars(map[string]string{"SHELLHUB_TENANT_ID": "00000000-0000-4000-0000-000000000000"}))
		require.NoError(t, err)
		assert.Equal(t, "00000000-0000-4000-0000-000000000000", cfg.TenantID)
		assert.Empty(t, cfg.ProvisioningKey)
	})

	t.Run("a provisioning key alone is a namespace credential", func(t *testing.T) {
		cfg, _, err := LoadConfigConnectorFromEnv(withVars(map[string]string{"SHELLHUB_PROVISIONING_KEY": "key-1"}))
		require.NoError(t, err)
		assert.Equal(t, "key-1", cfg.ProvisioningKey)
		assert.Empty(t, cfg.TenantID)
	})

	t.Run("neither is refused", func(t *testing.T) {
		_, fields, err := LoadConfigConnectorFromEnv(withVars(nil))
		require.Error(t, err)
		assert.Contains(t, fields, "TenantID")
		assert.Contains(t, fields, "ProvisioningKey")
	})
}
