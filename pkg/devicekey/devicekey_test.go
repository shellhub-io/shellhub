package devicekey

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

func TestParsePublicKey(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	edPub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	pkix, err := x509.MarshalPKIXPublicKey(edPub)
	require.NoError(t, err)

	wantRSA, err := gossh.NewPublicKey(&rsaKey.PublicKey)
	require.NoError(t, err)

	wantEd, err := gossh.NewPublicKey(edPub)
	require.NoError(t, err)

	t.Run("reads the PKCS#1 key the agent enrols with", func(t *testing.T) {
		got, err := ParsePublicKey(string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&rsaKey.PublicKey)})))
		require.NoError(t, err)
		assert.Equal(t, wantRSA.Marshal(), got.Marshal())
	})

	t.Run("reads a PKIX key", func(t *testing.T) {
		got, err := ParsePublicKey(string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pkix})))
		require.NoError(t, err)
		assert.Equal(t, wantEd.Marshal(), got.Marshal())
	})

	for _, key := range []string{"", "public-key", string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: []byte("garbage")}))} {
		_, err := ParsePublicKey(key)
		require.ErrorIs(t, err, ErrInvalidPublicKey, key)
	}
}
