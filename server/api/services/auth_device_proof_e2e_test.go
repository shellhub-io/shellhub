package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/devicekey"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceKeyProofE2E(t *testing.T) {
	e := setupEnrollmentE2E(t)
	ctx := context.Background()

	request := func(t *testing.T, key *rsa.PrivateKey, mac string, signed bool) requests.DeviceAuth {
		t.Helper()

		publicKey := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey)}))
		req := requests.DeviceAuth{
			TenantID:  e.tenantID,
			Hostname:  "host-" + mac,
			Identity:  &requests.DeviceIdentity{MAC: mac},
			Info:      &requests.DeviceInfo{ID: "debian", PrettyName: "Debian", Version: "v0.1.0", Arch: "amd64", Platform: "docker"},
			PublicKey: publicKey,
		}

		if signed {
			challenge, err := e.svc.CreateDeviceAuthChallenge(ctx)
			require.NoError(t, err)

			req.Challenge = challenge.Challenge
			req.Signature, err = devicekey.Sign(key, devicekey.Statement{Challenge: challenge.Challenge, TenantID: e.tenantID, PublicKey: publicKey})
			require.NoError(t, err)
		}

		return req
	}

	stored := func(t *testing.T, uid string) *models.Device {
		t.Helper()

		device, err := e.st.DeviceResolve(ctx, scope.MustBounded(e.tenantID), store.DeviceUIDResolver, uid)
		require.NoError(t, err)

		return device
	}

	t.Run("a legacy device is pinned by its first proof and refuses unsigned requests after it", func(t *testing.T) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		legacy, err := e.svc.AuthDevice(ctx, request(t, key, "aa:00:00:00:00:01", false))
		require.NoError(t, err)
		assert.Nil(t, stored(t, legacy.UID).KeyProvenAt, "an unsigned request proves nothing")

		stale := stored(t, legacy.UID)

		proven, err := e.svc.AuthDevice(ctx, request(t, key, "aa:00:00:00:00:01", true))
		require.NoError(t, err)
		require.Equal(t, legacy.UID, proven.UID, "the proof authenticates the same device")
		assert.NotNil(t, stored(t, proven.UID).KeyProvenAt)

		require.NoError(t, e.st.DeviceUpdate(ctx, stale))
		assert.NotNil(t, stored(t, proven.UID).KeyProvenAt, "a write from a snapshot read before the proof cannot clear it")

		_, err = e.svc.AuthDevice(ctx, request(t, key, "aa:00:00:00:00:01", false))
		require.ErrorIs(t, err, ErrDeviceKeyProofRefused)

		_, err = e.svc.AuthDevice(ctx, request(t, key, "aa:00:00:00:00:01", true))
		assert.NoError(t, err, "the device itself keeps authenticating")
	})

	t.Run("a device enrolled with a proof refuses unsigned requests from the start", func(t *testing.T) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		res, err := e.svc.AuthDevice(ctx, request(t, key, "aa:00:00:00:00:02", true))
		require.NoError(t, err)
		assert.NotNil(t, stored(t, res.UID).KeyProvenAt)

		_, err = e.svc.AuthDevice(ctx, request(t, key, "aa:00:00:00:00:02", false))
		require.ErrorIs(t, err, ErrDeviceKeyProofRefused)
	})

	deviceClaims := func(t *testing.T, token string) *authorizer.DeviceClaims {
		t.Helper()

		claims, err := jwttoken.ClaimsFromBearerToken(publicKey, token)
		require.NoError(t, err)

		device, ok := claims.(*authorizer.DeviceClaims)
		require.True(t, ok)

		return device
	}

	for _, tc := range []struct {
		name        string
		mac         string
		legacyFirst bool
	}{
		{name: "a device enrolled with a proof", mac: "aa:00:00:00:00:03"},
		{name: "a legacy device that proves its key in the same second", mac: "aa:00:00:00:00:04", legacyFirst: true},
	} {
		t.Run("only a token issued against the proof is honoured: "+tc.name, func(t *testing.T) {
			key, err := rsa.GenerateKey(rand.Reader, 2048)
			require.NoError(t, err)

			var legacyToken string
			if tc.legacyFirst {
				legacy, err := e.svc.AuthDevice(ctx, request(t, key, tc.mac, false))
				require.NoError(t, err)

				legacyToken = legacy.Token
			}

			proven, err := e.svc.AuthDevice(ctx, request(t, key, tc.mac, true))
			require.NoError(t, err)
			require.NoError(t, e.svc.AuthDeviceToken(ctx, deviceClaims(t, proven.Token)))

			if tc.legacyFirst {
				require.ErrorIs(t, e.svc.AuthDeviceToken(ctx, deviceClaims(t, legacyToken)), ErrAuthUnathorized)
			}
		})
	}
}
