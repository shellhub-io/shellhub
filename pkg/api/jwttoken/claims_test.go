package jwttoken_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodedClaimsCarryTheIssuer(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	const issuer = "https://shellhub.example.com"

	issuerOf := func(t *testing.T, raw string) string {
		t.Helper()

		claims := new(jwt.RegisteredClaims)
		_, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) {
			return &key.PublicKey, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
		require.NoError(t, err)

		return claims.Issuer
	}

	t.Run("user claims", func(t *testing.T) {
		token, err := jwttoken.EncodeUserClaims(authorizer.UserClaims{ID: "000000000000000000000000"}, issuer, key)
		require.NoError(t, err)

		assert.Equal(t, issuer, issuerOf(t, token))
	})

	t.Run("device claims", func(t *testing.T) {
		token, err := jwttoken.EncodeDeviceClaims(authorizer.DeviceClaims{UID: "uid"}, issuer, key)
		require.NoError(t, err)

		assert.Equal(t, issuer, issuerOf(t, token))
	})

	t.Run("enrollment decision claims", func(t *testing.T) {
		claims := jwttoken.EnrollmentDecisionClaims{DeviceUID: "uid", TenantID: "tenant", InstallKeyID: "key"}
		token, err := jwttoken.EncodeEnrollmentDecisionClaims(claims, time.Hour, issuer, key)
		require.NoError(t, err)

		assert.Equal(t, issuer, issuerOf(t, token))
	})
}
