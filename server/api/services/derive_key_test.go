package services

import (
	"crypto/hkdf"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"testing"

	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/require"
)

func TestService_DeriveKey(t *testing.T) {
	signingKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	otherSigningKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	replica := NewService(storemock.NewMockStore(t), signingKey, &signingKey.PublicKey, cachemock.NewMockCache(t))
	sameReplica := NewService(storemock.NewMockStore(t), signingKey, &signingKey.PublicKey, cachemock.NewMockCache(t))
	otherInstance := NewService(storemock.NewMockStore(t), otherSigningKey, &otherSigningKey.PublicKey, cachemock.NewMockCache(t))

	key := replica.DeriveKey("saml-request")
	require.Len(t, key, 32)
	require.Equal(t, key, sameReplica.DeriveKey("saml-request"))
	require.NotEqual(t, key, replica.DeriveKey("saml-link"))
	require.NotEqual(t, key, otherInstance.DeriveKey("saml-request"))

	challengeKeyBeforeDeriveKey, err := hkdf.Key(sha256.New, x509.MarshalPKCS1PrivateKey(signingKey), nil, "shellhub-device-auth-challenge-v1", sha256.Size)
	require.NoError(t, err)
	require.Equal(t, challengeKeyBeforeDeriveKey, replica.DeriveKey("device-auth-challenge"))
}
