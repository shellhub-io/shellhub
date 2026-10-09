package environment

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadLicenseIssuer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "issuer", "license-issuer.pem")

	first, err := LoadLicenseIssuer(path)
	require.NoError(t, err)

	again, err := LoadLicenseIssuer(path)
	require.NoError(t, err)

	assert.True(t, first.key.Equal(again.key), "a second load generated another key")

	t.Run("fails on a file that holds no key", func(t *testing.T) {
		garbage := filepath.Join(t.TempDir(), "license-issuer.pem")
		require.NoError(t, os.WriteFile(garbage, []byte("not a key"), 0o600))

		_, err := LoadLicenseIssuer(garbage)
		require.ErrorIs(t, err, errNotLicenseIssuer)
	})
}

func TestLicensingEnvsWithoutIssuer(t *testing.T) {
	run := &Run{}

	_, _, err := Config{Name: "shellhub-e2e-a", Run: run}.licensingEnvs()
	require.ErrorIs(t, err, errRunIssuesNoLicense)

	envs, written, err := Config{Name: "shellhub-e2e-a", Run: run, Unlicensed: true}.licensingEnvs()
	require.NoError(t, err)
	assert.Empty(t, written)
	assert.Equal(t, map[string]string{licenseFileEnv: ""}, envs)
}
