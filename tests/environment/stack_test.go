package environment

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStackArtifactsCoverWhatUpWrites(t *testing.T) {
	t.Chdir(t.TempDir())

	issuer, err := LoadLicenseIssuer(filepath.Join(t.TempDir(), "issuer.pem"))
	require.NoError(t, err)

	cfg := Config{Name: "shellhub-e2e-a", LocatedCountry: "BR", Run: &Run{issuer: issuer}}

	_, written, err := cfg.licensingEnvs()
	require.NoError(t, err)

	_, dir, err := cfg.geoIPEnvs()
	require.NoError(t, err)

	written = append(written, dir)

	artifacts, err := stackArtifacts(cfg.Name)
	require.NoError(t, err)
	require.NoError(t, removeArtifacts(artifacts))

	for _, path := range written {
		_, err := os.Stat(path)
		assert.ErrorIs(t, err, fs.ErrNotExist, path)
	}
}
