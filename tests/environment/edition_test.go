package environment

import (
	"encoding/base64"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEdition(t *testing.T) {
	cloudDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(cloudDir, "docker-compose.yml"), []byte("name: shellhub\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(cloudDir, "go.mod"), []byte("module github.com/shellhub-io/cloud\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(cloudDir, ".env"), []byte("SHELLHUB_BILLING=stripe\n"), 0o600))

	cloudCompose := filepath.Join(cloudDir, "docker-compose.yml")
	cloudEnv := filepath.Join(cloudDir, ".env")

	tests := []struct {
		edition      Edition
		buildEdition string
		files        []string
		envFiles     []string
		editionEnvs  map[string]string
	}{
		{
			edition:      EditionCommunity,
			buildEdition: "community",
			files: []string{
				"../docker-compose.yml",
				"../docker-compose.test.yml",
				"../docker-compose.postgres.test.yml",
			},
			envFiles:    []string{"../.env", "../versions.env"},
			editionEnvs: map[string]string{},
		},
		{
			edition:      EditionEnterprise,
			buildEdition: "enterprise",
			files: []string{
				"../docker-compose.yml",
				"../docker-compose.enterprise.yml",
				"../docker-compose.test.yml",
				"../docker-compose.postgres.test.yml",
				"../docker-compose.enterprise.test.yml",
			},
			envFiles: []string{"../.env", "../versions.env", "../.env.enterprise", cloudEnv},
			editionEnvs: map[string]string{
				"SHELLHUB_EMAIL_PROVIDER": "dummy",
				"SHELLHUB_MAXMIND_MIRROR": "",
			},
		},
		{
			edition:      EditionCloud,
			buildEdition: "enterprise",
			files: []string{
				"../docker-compose.yml",
				"../docker-compose.enterprise.yml",
				cloudCompose,
				"../docker-compose.test.yml",
				"../docker-compose.postgres.test.yml",
				"../docker-compose.enterprise.test.yml",
			},
			envFiles: []string{"../.env", "../versions.env", "../.env.enterprise", cloudEnv},
			editionEnvs: map[string]string{
				"SHELLHUB_BILLING":        "stripe",
				"SHELLHUB_EMAIL_PROVIDER": "dummy",
				"SHELLHUB_MAXMIND_MIRROR": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.edition), func(t *testing.T) {
			assert.Equal(t, tt.buildEdition, tt.edition.Build())

			files, err := tt.edition.composeFiles(cloudDir)
			require.NoError(t, err)
			assert.Equal(t, tt.files, files)
			assert.Equal(t, tt.envFiles, tt.edition.envFiles(cloudDir))

			envs, err := tt.edition.envs(cloudDir)
			require.NoError(t, err)
			assert.Equal(t, tt.buildEdition, envs["SHELLHUB_BUILD_EDITION"])
			assert.Equal(t, "production", envs["SHELLHUB_ENV"])

			assert.Equal(t, tt.editionEnvs, pick(envs, "SHELLHUB_BILLING", "COMPOSE_PROFILES", "SHELLHUB_EMAIL_PROVIDER", "SHELLHUB_MAXMIND_MIRROR"))

			secret, ok := envs["SHELLHUB_SAML_SECRET"]
			if tt.edition == EditionCommunity {
				assert.False(t, ok, "community has no SAML to sign requests for")

				return
			}

			key, err := base64.URLEncoding.DecodeString(secret)
			require.NoError(t, err, "the server decodes the SAML secret as URL-safe base64")
			assert.Len(t, key, 32, "the SAML secret encrypts the SP key with AES-256")
		})
	}

	t.Run("paid editions without the cloud source", func(t *testing.T) {
		for _, edition := range []Edition{EditionEnterprise, EditionCloud} {
			_, err := edition.composeFiles("/nonexistent/path")
			require.ErrorIs(t, err, fs.ErrNotExist, edition)
		}
	})

	t.Run("cloud without its compose file", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/shellhub-io/cloud\n"), 0o600))

		_, err := EditionCloud.composeFiles(dir)
		require.ErrorContains(t, err, "cloud edition requires")
	})

	t.Run("cloud without .env", func(t *testing.T) {
		assert.Equal(t, []string{"../.env", "../versions.env", "../.env.enterprise"}, EditionCloud.envFiles(t.TempDir()))
	})

	t.Run("parse", func(t *testing.T) {
		for _, valid := range []string{"community", "enterprise", "cloud"} {
			e, err := ParseEdition(valid)
			require.NoError(t, err)
			assert.Equal(t, Edition(valid), e)
		}

		_, err := ParseEdition("invalid")
		assert.Error(t, err)
	})
}

func pick(envs map[string]string, keys ...string) map[string]string {
	picked := map[string]string{}
	for _, key := range keys {
		if value, ok := envs[key]; ok {
			picked[key] = value
		}
	}

	return picked
}

func TestMergeEnvs(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "first.env")
	second := filepath.Join(dir, "second.env")
	require.NoError(t, os.WriteFile(first, []byte("A=first\nB=first\nC=first\nD=first\n"), 0o600))
	require.NoError(t, os.WriteFile(second, []byte("B=second\nC=second\nD=second\n"), 0o600))

	merged, err := mergeEnvs(
		[]string{first, second},
		map[string]string{"C": "layer", "D": "layer"},
		map[string]string{"D": "override"},
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"A": "first", "B": "second", "C": "layer", "D": "override"}, merged)

	_, err = mergeEnvs([]string{filepath.Join(dir, "missing.env")})
	assert.Error(t, err)
}
