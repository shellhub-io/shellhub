package jwttoken_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnrollmentDecisionClaims(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	claims := jwttoken.EnrollmentDecisionClaims{DeviceUID: "uid", TenantID: "tenant", InstallKeyID: "digest"}

	t.Run("round-trips a valid token", func(t *testing.T) {
		token, err := jwttoken.EncodeEnrollmentDecisionClaims(claims, time.Hour, key)
		require.NoError(t, err)

		got, jti, err := jwttoken.DecodeEnrollmentDecisionClaims(&key.PublicKey, token)
		require.NoError(t, err)
		assert.Equal(t, claims, *got)
		assert.NotEmpty(t, jti, "the token carries a unique id for single-use tracking")
	})

	t.Run("rejects an expired token", func(t *testing.T) {
		token, err := jwttoken.EncodeEnrollmentDecisionClaims(claims, -time.Hour, key)
		require.NoError(t, err)

		_, _, err = jwttoken.DecodeEnrollmentDecisionClaims(&key.PublicKey, token)
		require.Error(t, err)
	})

	t.Run("rejects a token of another kind", func(t *testing.T) {
		deviceToken, err := jwttoken.EncodeDeviceClaims(authorizer.DeviceClaims{UID: "uid", TenantID: "tenant"}, key)
		require.NoError(t, err)

		_, _, err = jwttoken.DecodeEnrollmentDecisionClaims(&key.PublicKey, deviceToken)
		require.Error(t, err)
	})

	t.Run("rejects a token signed by another key", func(t *testing.T) {
		other, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		token, err := jwttoken.EncodeEnrollmentDecisionClaims(claims, time.Hour, key)
		require.NoError(t, err)

		_, _, err = jwttoken.DecodeEnrollmentDecisionClaims(&other.PublicKey, token)
		require.Error(t, err)
	})
}
