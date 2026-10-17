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

func TestDeviceKeyVerificationE2E(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
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

	t.Run("a legacy device is verified by its first signature and refuses unsigned requests after it", func(t *testing.T) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		legacy, err := e.svc.AuthDevice(ctx, request(t, key, "aa:00:00:00:00:01", false))
		require.NoError(t, err)
		assert.Nil(t, stored(t, legacy.UID).KeyVerifiedAt, "an unsigned request proves nothing")

		stale := stored(t, legacy.UID)

		verified, err := e.svc.AuthDevice(ctx, request(t, key, "aa:00:00:00:00:01", true))
		require.NoError(t, err)
		require.Equal(t, legacy.UID, verified.UID, "the signature authenticates the same device")
		assert.NotNil(t, stored(t, verified.UID).KeyVerifiedAt)

		require.NoError(t, e.st.DeviceUpdate(ctx, stale))
		assert.NotNil(t, stored(t, verified.UID).KeyVerifiedAt, "a write from a snapshot read before the verification cannot clear it")

		_, err = e.svc.AuthDevice(ctx, request(t, key, "aa:00:00:00:00:01", false))
		require.ErrorIs(t, err, ErrDeviceKeyVerificationRefused)

		_, err = e.svc.AuthDevice(ctx, request(t, key, "aa:00:00:00:00:01", true))
		assert.NoError(t, err, "the device itself keeps authenticating")
	})

	t.Run("a device enrolled with a signature refuses unsigned requests from the start", func(t *testing.T) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		res, err := e.svc.AuthDevice(ctx, request(t, key, "aa:00:00:00:00:02", true))
		require.NoError(t, err)
		assert.NotNil(t, stored(t, res.UID).KeyVerifiedAt)

		_, err = e.svc.AuthDevice(ctx, request(t, key, "aa:00:00:00:00:02", false))
		require.ErrorIs(t, err, ErrDeviceKeyVerificationRefused)
	})

	t.Run("removing a verified device lets an unsigned agent enroll it again", func(t *testing.T) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		res, err := e.svc.AuthDevice(ctx, request(t, key, "aa:00:00:00:00:05", true))
		require.NoError(t, err)
		require.NoError(t, e.svc.UpdateDeviceStatus(ctx, &requests.DeviceUpdateStatus{
			TenantID: e.tenantID, UID: res.UID, Status: string(models.DeviceStatusAccepted),
		}))

		require.NoError(t, e.svc.DeleteDevice(ctx, models.UID(res.UID), e.tenantID))
		assert.Nil(t, stored(t, res.UID).KeyVerifiedAt, "removal clears the verification")

		again, err := e.svc.AuthDevice(ctx, request(t, key, "aa:00:00:00:00:05", false))
		require.NoError(t, err)
		require.Equal(t, res.UID, again.UID)
		assert.Equal(t, models.DeviceStatusPending, stored(t, again.UID).Status, "the operator accepts it again")
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
		{name: "a device enrolled with a signature", mac: "aa:00:00:00:00:03"},
		{name: "a legacy device that proves its key in the same second", mac: "aa:00:00:00:00:04", legacyFirst: true},
	} {
		t.Run("only a token issued against the signature is honoured: "+tc.name, func(t *testing.T) {
			key, err := rsa.GenerateKey(rand.Reader, 2048)
			require.NoError(t, err)

			var legacyToken string
			if tc.legacyFirst {
				legacy, err := e.svc.AuthDevice(ctx, request(t, key, tc.mac, false))
				require.NoError(t, err)

				legacyToken = legacy.Token
			}

			verified, err := e.svc.AuthDevice(ctx, request(t, key, tc.mac, true))
			require.NoError(t, err)
			require.NoError(t, e.svc.AuthDeviceToken(ctx, deviceClaims(t, verified.Token)))

			if tc.legacyFirst {
				require.ErrorIs(t, e.svc.AuthDeviceToken(ctx, deviceClaims(t, legacyToken)), ErrAuthUnathorized)
			}
		})
	}
}
