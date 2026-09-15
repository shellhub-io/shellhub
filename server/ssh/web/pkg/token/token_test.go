package token_test

import (
	"crypto/rsa"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/magickey"
	"github.com/shellhub-io/shellhub/server/ssh/web/pkg/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenCarriesTheIssuer(t *testing.T) {
	const issuer = "https://shellhub.example.com"

	created, err := token.NewToken(issuer)
	require.NoError(t, err)

	claims := new(jwt.RegisteredClaims)
	_, err = jwt.ParseWithClaims(created.Data, claims, func(*jwt.Token) (any, error) {
		public, ok := magickey.GetReference().Public().(*rsa.PublicKey)
		require.True(t, ok)

		return public, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
	require.NoError(t, err)

	assert.Equal(t, issuer, claims.Issuer)
}
