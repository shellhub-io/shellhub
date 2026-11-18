package environment

import (
	"crypto/tls"
	"crypto/x509"
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

func TestConfigValidateRefusesASuppliedCertificateWithACME(t *testing.T) {
	certificate, err := SelfSignedCertificate("localhost")
	require.NoError(t, err)

	err = Config{Edition: EditionCommunity, Run: &Run{}, SuppliedCertificate: &certificate, ACME: true}.validate()
	require.ErrorIs(t, err, errSuppliedCertificateAndACME)
}

func TestSelfSignedCertificateNeedsAName(t *testing.T) {
	_, err := SelfSignedCertificate()
	assert.ErrorIs(t, err, errNoCertificateName)
}
