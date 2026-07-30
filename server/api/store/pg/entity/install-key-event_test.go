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

	// The agent encodes its key as PKCS#1 ("RSA PUBLIC KEY"); it must fingerprint too, to the same
	// value as the PKIX encoding of the same key.
	pkcs1 := string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
	}))
	require.Equal(t, fp, fingerprintFromPEM(pkcs1), "PKCS#1 and PKIX of the same key must match")

	// Graceful on absent/garbage input (e.g. old events with no key).
	require.Empty(t, fingerprintFromPEM(""))
	require.Empty(t, fingerprintFromPEM("not-a-pem-key"))
}
