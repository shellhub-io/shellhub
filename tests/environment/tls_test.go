package environment

import (
	"crypto/tls"
	"crypto/x509"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelfSignedCertificateVerifiesAgainstItself(t *testing.T) {
	certificate, err := SelfSignedCertificate("pebble", "localhost")
	require.NoError(t, err)

	pair, err := tls.X509KeyPair(certificate.CertificatePEM, certificate.KeyPEM)
	require.NoError(t, err)

	leaf, err := x509.ParseCertificate(pair.Certificate[0])
	require.NoError(t, err)

	roots := x509.NewCertPool()
	require.True(t, roots.AppendCertsFromPEM(certificate.CertificatePEM))

	for _, name := range []string{"pebble", "localhost"} {
		_, err := leaf.Verify(x509.VerifyOptions{DNSName: name, Roots: roots})
		assert.NoError(t, err, name)
	}
}

func TestUpRefusesASuppliedCertificateWithACME(t *testing.T) {
	t.Chdir(t.TempDir())

	certificate, err := SelfSignedCertificate("localhost")
	require.NoError(t, err)

	_, err = Up(t.Context(), Config{Edition: EditionCommunity, Name: "shellhub-e2e-a", Run: &Run{}, SuppliedCertificate: &certificate, ACME: true})
	require.ErrorIs(t, err, errSuppliedCertificateAndACME)

	_, err = os.Stat(stackArtifactsDir)
	assert.ErrorIs(t, err, fs.ErrNotExist, "a refused stack leaves no certificate or Pebble file behind")
}

func TestSelfSignedCertificateNeedsAName(t *testing.T) {
	_, err := SelfSignedCertificate()
	assert.ErrorIs(t, err, errNoCertificateName)
}
