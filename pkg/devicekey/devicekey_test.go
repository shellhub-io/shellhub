package devicekey

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
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

func TestProof(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	other, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	encode := func(key *rsa.PrivateKey) string {
		return string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey)}))
	}

	const (
		challenge = "challenge"
		tenant    = "00000000-0000-4000-0000-000000000000"
	)

	publicKey := encode(key)

	statement := Statement{Challenge: challenge, TenantID: tenant, PublicKey: publicKey}

	signature, err := Sign(key, statement)
	require.NoError(t, err)

	t.Run("the device key's proof verifies", func(t *testing.T) {
		require.NoError(t, Verify(statement, signature))
	})

	t.Run("a proof does not verify for another challenge, tenant, provisioning key or key", func(t *testing.T) {
		for _, altered := range []Statement{
			{Challenge: "another-challenge", TenantID: tenant, PublicKey: publicKey},
			{Challenge: challenge, TenantID: "00000000-0000-4000-0000-000000000001", PublicKey: publicKey},
			{Challenge: challenge, TenantID: tenant, ProvisioningKey: "provisioning-key", PublicKey: publicKey},
			{Challenge: challenge, TenantID: tenant, PublicKey: encode(other)},
		} {
			require.ErrorIs(t, Verify(altered, signature), ErrInvalidProof)
		}
	})

	t.Run("a key cannot sign for another device's public key", func(t *testing.T) {
		forged, err := Sign(other, statement)
		require.NoError(t, err)

		require.ErrorIs(t, Verify(statement, forged), ErrInvalidProof)
	})

	t.Run("a SHA-1 signature is refused", func(t *testing.T) {
		signer, err := gossh.NewSignerFromKey(key)
		require.NoError(t, err)

		algorithmSigner, ok := signer.(gossh.AlgorithmSigner)
		require.True(t, ok)

		sig, err := algorithmSigner.SignWithAlgorithm(rand.Reader, statement.message(), gossh.KeyAlgoRSA)
		require.NoError(t, err)

		require.ErrorIs(t, Verify(statement, base64.StdEncoding.EncodeToString(gossh.Marshal(sig))), ErrInvalidProof)
	})

	t.Run("a malformed signature is refused", func(t *testing.T) {
		require.ErrorIs(t, Verify(statement, "not base64"), ErrInvalidProof)
		require.ErrorIs(t, Verify(statement, base64.StdEncoding.EncodeToString([]byte("garbage"))), ErrInvalidProof)
	})
}
