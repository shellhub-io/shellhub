package environment

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLicenseEnvs(t *testing.T) {
	tests := []struct {
		name     string
		override string
		shell    string
		absolute bool
		want     string
		wantErr  error
	}{
		{
			name:     "resolves an override path against the override's directory",
			override: "SHELLHUB_LICENSE_FILE=license.dat\n",
			want:     "license.dat",
		},
		{
			name:     "prefers an absolute shell path over the override file",
			override: "SHELLHUB_LICENSE_FILE=missing.dat\n",
			shell:    "license.dat",
			absolute: true,
			want:     "license.dat",
		},
		{
			name:     "resolves a relative shell path against the override's directory",
			override: "SHELLHUB_LICENSE_FILE=missing.dat\n",
			shell:    "license.dat",
			want:     "license.dat",
		},
		{
			name:    "fails when neither sets a license",
			wantErr: errNoLicense,
		},
		{
			name:     "fails on a license file that does not exist",
			override: "SHELLHUB_LICENSE_FILE=missing.dat\n",
			wantErr:  os.ErrNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, "license.dat"), []byte("license"), 0o600))
			overridePath := filepath.Join(dir, ".env.override")
			require.NoError(t, os.WriteFile(overridePath, []byte(tt.override), 0o600))

			shell := tt.shell
			if tt.absolute {
				shell = filepath.Join(dir, shell)
			}
			t.Setenv(licenseFileEnv, shell)

			envs, err := licenseEnvs(overridePath)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, map[string]string{licenseFileEnv: filepath.Join(dir, tt.want)}, envs)
		})
	}
}

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
