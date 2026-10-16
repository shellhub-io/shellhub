package environment

import (
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStackArtifactsCoverWhatUpWrites(t *testing.T) {
	t.Chdir(t.TempDir())

	cfg := Config{Name: "shellhub-e2e-a", Located: &Location{Country: "BR"}, Run: &Run{issuer: newLicenseIssuer(t)}}

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

func TestOnlyACloudStackReceivesStripeEvents(t *testing.T) {
	for _, edition := range []Edition{EditionCommunity, EditionEnterprise} {
		t.Run(string(edition), func(t *testing.T) {
			_, err := Up(t.Context(), Config{Edition: edition, StripeEvents: true, Run: &Run{}})
			require.ErrorIs(t, err, errStripeEventsOutsideCloud)
		})
	}
}
