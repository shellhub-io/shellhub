package entity

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFingerprintFromPEM(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	pemKey := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))

	fp := fingerprintFromPEM(pemKey)
	require.True(t, strings.HasPrefix(fp, "SHA256:"), "expected an SSH SHA256 fingerprint, got %q", fp)
	require.Equal(t, fp, fingerprintFromPEM(pemKey), "same key must yield the same fingerprint")

	pkcs1 := string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
	}))
	require.Equal(t, fp, fingerprintFromPEM(pkcs1), "PKCS#1 and PKIX of the same key must match")

	require.Empty(t, fingerprintFromPEM(""))
	require.Empty(t, fingerprintFromPEM("not-a-pem-key"))
}
