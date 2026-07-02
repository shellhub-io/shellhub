package agentd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantFilePath(t *testing.T) {
	assert.Equal(t, "/host/etc/shellhub.key.tenant", TenantFilePath("/host/etc/shellhub.key"))
}

func TestPersistAndReadTenant(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shellhub.key.tenant")

	tenant, err := ReadPersistedTenant(path)
	require.NoError(t, err)
	assert.Empty(t, tenant, "missing file is not an error and yields an empty tenant")

	require.NoError(t, PersistTenant(path, "00000000-0000-4000-0000-000000000000"))

	tenant, err = ReadPersistedTenant(path)
	require.NoError(t, err)
	assert.Equal(t, "00000000-0000-4000-0000-000000000000", tenant)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	// Overwrite is atomic and replaces the previous tenant.
	require.NoError(t, PersistTenant(path, "11111111-1111-4111-1111-111111111111"))

	tenant, err = ReadPersistedTenant(path)
	require.NoError(t, err)
	assert.Equal(t, "11111111-1111-4111-1111-111111111111", tenant)
}
