package environment

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEdition(t *testing.T) {
	cloudDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(cloudDir, "docker-compose.yml"), []byte("name: shellhub\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(cloudDir, ".env"), []byte("SHELLHUB_BILLING=stripe\n"), 0o600))

	cloudCompose := filepath.Join(cloudDir, "docker-compose.yml")

	tests := []struct {
		edition      Edition
		buildEdition string
		files        []string
		dummyBilling bool
	}{
		{
			edition:      EditionCommunity,
			buildEdition: "community",
			files: []string{
				"../docker-compose.yml",
				"../docker-compose.test.yml",
				"../docker-compose.postgres.test.yml",
			},
			dummyBilling: false,
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
			dummyBilling: true,
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
			dummyBilling: true,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.edition), func(t *testing.T) {
			assert.Equal(t, tt.buildEdition, tt.edition.Build())

			files, err := tt.edition.composeFiles(cloudDir)
			require.NoError(t, err)
			assert.Equal(t, tt.files, files)

			envs, err := tt.edition.envs(cloudDir)
			require.NoError(t, err)
			assert.Equal(t, tt.buildEdition, envs["SHELLHUB_BUILD_EDITION"])
			assert.Equal(t, "production", envs["SHELLHUB_ENV"])

			if tt.dummyBilling {
				assert.Equal(t, "dummy", envs["SHELLHUB_BILLING"])
				assert.Equal(t, "dummy", envs["SHELLHUB_EMAIL_PROVIDER"])
			} else {
				_, hasBilling := envs["SHELLHUB_BILLING"]
				assert.False(t, hasBilling)
			}
		})
	}

	t.Run("cloud without dir", func(t *testing.T) {
		_, err := EditionCloud.composeFiles("/nonexistent/path")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cloud edition requires")
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
