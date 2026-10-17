package environment

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloudStackLicenseAuthorization(t *testing.T) {
	issuer := newLicenseIssuer(t)
	enterprise := FullLicense()
	for _, test := range []struct {
		name    string
		edition Edition
		license *License
		want    bool
	}{
		{name: "default enterprise", edition: EditionEnterprise},
		{name: "default cloud", edition: EditionCloud, want: true},
		{name: "explicit enterprise license in cloud is not upgraded", edition: EditionCloud, license: &enterprise},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			envs, _, err := (Config{Name: "shellhub-e2e-grant", Edition: test.edition, License: test.license, Run: &Run{issuer: issuer}}).licensingEnvs()
			require.NoError(t, err)
			raw, err := os.ReadFile(envs[licenseFileEnv])
			require.NoError(t, err)
			decoded, err := base64.StdEncoding.DecodeString(string(raw))
			require.NoError(t, err)
			var license License
			require.NoError(t, json.Unmarshal(decoded[:len(decoded)-issuer.key.Size()], &license))
			require.NotNil(t, license.Features)
			assert.Equal(t, test.want, license.Features.Cloud)
		})
	}
}

func newCloudSource(t *testing.T) string {
	t.Helper()

	cloudDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(cloudDir, "go.mod"), []byte("module github.com/shellhub-io/cloud\n"), 0o600))

	return cloudDir
}

func writeIssuerKey(t *testing.T, cloudDir string) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)

	path := LicenseIssuerKeyPath(cloudDir)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), 0o600))

	return key
}

func newLicenseIssuer(t *testing.T) *LicenseIssuer {
	t.Helper()

	cloudDir := newCloudSource(t)
	writeIssuerKey(t, cloudDir)

	issuer, err := LoadLicenseIssuer(cloudDir)
	require.NoError(t, err)

	return issuer
}

func TestLoadLicenseIssuer(t *testing.T) {
	t.Run("loads the key the cloud source carries", func(t *testing.T) {
		cloudDir := newCloudSource(t)
		key := writeIssuerKey(t, cloudDir)

		issuer, err := LoadLicenseIssuer(cloudDir)
		require.NoError(t, err)
		assert.True(t, key.Equal(issuer.key))
	})

	t.Run("issues no licenses without the cloud source", func(t *testing.T) {
		issuer, err := LoadLicenseIssuer(t.TempDir())
		require.NoError(t, err)
		assert.Nil(t, issuer)
	})

	t.Run("generates no key the cloud source lacks", func(t *testing.T) {
		cloudDir := newCloudSource(t)

		_, err := LoadLicenseIssuer(cloudDir)
		require.ErrorIs(t, err, fs.ErrNotExist)
		assert.NoFileExists(t, LicenseIssuerKeyPath(cloudDir))
	})

	t.Run("fails on a file that holds no key", func(t *testing.T) {
		cloudDir := newCloudSource(t)
		path := LicenseIssuerKeyPath(cloudDir)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
		require.NoError(t, os.WriteFile(path, []byte("not a key"), 0o600))

		_, err := LoadLicenseIssuer(cloudDir)
		require.ErrorIs(t, err, errNotLicenseIssuer)
	})
}

func TestLicenseIssuerKeyPathDefaultsLikeTheLoad(t *testing.T) {
	assert.Equal(t, LicenseIssuerKeyPath(defaultCloudDir), LicenseIssuerKeyPath(""))
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
