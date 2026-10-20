package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/devicekey"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/storetest/pgprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceKeyVerificationE2E(t *testing.T) {
	e := setupLegacyEnrollmentE2E(t)
	ctx := context.Background()

	request := e.deviceAuthRequest

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

func (e *enrollmentE2E) deviceAuthRequest(t *testing.T, key *rsa.PrivateKey, mac string, signed bool) requests.DeviceAuth {
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
		challenge, err := e.svc.CreateDeviceAuthChallenge(context.Background())
		require.NoError(t, err)

		req.Challenge = challenge.Challenge
		req.Signature, err = devicekey.Sign(key, devicekey.Statement{Challenge: challenge.Challenge, TenantID: e.tenantID, PublicKey: publicKey})
		require.NoError(t, err)
	}

	return req
}

func (e *enrollmentE2E) requireVerifiedDeviceKey(t *testing.T, required bool) {
	t.Helper()

	req := &requests.NamespaceEdit{TenantParam: requests.TenantParam{Tenant: e.tenantID}}
	req.Settings.RequireVerifiedDeviceKey = &required

	namespace, err := e.svc.EditNamespace(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, required, namespace.Settings.RequireVerifiedDeviceKey)
}

func (e *enrollmentE2E) tenantRegistrations(t *testing.T) []models.ProvisioningKeyEvent {
	t.Helper()

	events, _, err := e.svc.ListTenantRegistrations(context.Background(), &requests.ListTenantRegistrations{
		TenantID:  e.tenantID,
		Paginator: query.Paginator{Page: 1, PerPage: 100},
	})
	require.NoError(t, err)

	return events
}

func TestNamespaceRequiresVerifiedDeviceKeyE2E(t *testing.T) {
	ctx := context.Background()

	key := func(t *testing.T) *rsa.PrivateKey {
		t.Helper()

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		return key
	}

	for _, legacy := range []bool{false, true} {
		t.Run(fmt.Sprintf("legacy %t", legacy), func(t *testing.T) {
			setup := setupEnrollmentE2E
			if legacy {
				setup = setupLegacyEnrollmentE2E
			}

			t.Run("an unsigned unknown device enrolls pending while the namespace does not require verification", func(t *testing.T) {
				e := setup(t)

				res, err := e.svc.AuthDevice(ctx, e.deviceAuthRequest(t, key(t), "bb:00:00:00:00:01", false))
				require.NoError(t, err)
				assert.Equal(t, models.DeviceStatusPending, e.status(t, res.UID))
			})

			t.Run("an unsigned unknown device is refused and recorded once while the namespace requires verification, and enrolls once it stops", func(t *testing.T) {
				e := setup(t)
				e.requireVerifiedDeviceKey(t, true)
				k := key(t)

				req := e.deviceAuthRequest(t, k, "bb:00:00:00:00:02", false)
				req.RealIP = "192.0.2.1"
				_, err := e.svc.AuthDevice(ctx, req)
				require.ErrorIs(t, err, ErrDeviceKeyVerificationRefused)
				assert.Zero(t, e.deviceCount(t), "a refused agent leaves no device behind")

				later := now.Add(time.Minute)
				e.clock.Set(later)

				req.RealIP = "192.0.2.2"
				req.Info.Version = "v0.26.1"
				_, err = e.svc.AuthDevice(ctx, req)
				require.ErrorIs(t, err, ErrDeviceKeyVerificationRefused)

				events := e.tenantRegistrations(t)
				require.Len(t, events, 1, "a retrying agent stays one line")
				assert.Equal(t, models.ProvisioningKeyEventRefusalUnverifiedDeviceKey, events[0].Refusal)
				assert.Equal(t, 2, events[0].Attempts)
				assert.Equal(t, "192.0.2.2", events[0].SourceIP)
				assert.Equal(t, "v0.26.1", events[0].Info.Version)
				assert.Equal(t, "host-bb:00:00:00:00:02", events[0].Hostname)
				assert.WithinDuration(t, later, events[0].Timestamp, time.Millisecond, "the line shows the last attempt")
				assert.Empty(t, events[0].DeviceStatus, "no device exists for a refusal")
				assert.True(t, events[0].IsCurrent)

				e.requireVerifiedDeviceKey(t, false)
				e.clock.Set(later.Add(time.Minute))

				res, err := e.svc.AuthDevice(ctx, req)
				require.NoError(t, err, "the next attempt after the requirement is lifted enrolls")
				assert.Equal(t, models.DeviceStatusPending, e.status(t, res.UID))

				events = e.tenantRegistrations(t)
				require.Len(t, events, 2)
				enrolled, refused := events[0], events[1]
				assert.Empty(t, enrolled.Refusal)
				assert.True(t, enrolled.IsCurrent)
				assert.Equal(t, 1, enrolled.Attempts)
				assert.Equal(t, models.ProvisioningKeyEventRefusalUnverifiedDeviceKey, refused.Refusal)
				assert.False(t, refused.IsCurrent, "the refusal is history once the device enrolls")
			})

			t.Run("a signed agent enrolls while the namespace requires verification, and is never accepted unsigned", func(t *testing.T) {
				e := setup(t)
				e.requireVerifiedDeviceKey(t, true)
				k := key(t)

				res, err := e.svc.AuthDevice(ctx, e.deviceAuthRequest(t, k, "bb:00:00:00:00:03", true))
				require.NoError(t, err)
				assert.Equal(t, models.DeviceStatusPending, e.status(t, res.UID))

				_, err = e.svc.AuthDevice(ctx, e.deviceAuthRequest(t, k, "bb:00:00:00:00:03", false))
				require.ErrorIs(t, err, ErrDeviceKeyVerificationRefused)

				e.requireVerifiedDeviceKey(t, false)

				_, err = e.svc.AuthDevice(ctx, e.deviceAuthRequest(t, k, "bb:00:00:00:00:03", false))
				require.ErrorIs(t, err, ErrDeviceKeyVerificationRefused, "a verified device stays verified whatever the setting")

				for _, event := range e.tenantRegistrations(t) {
					assert.Empty(t, event.Refusal, "a known device is refused without a record")
				}
			})

			t.Run("a known device that never verified is refused, and so is its token, once the namespace requires verification", func(t *testing.T) {
				e := setup(t)
				k := key(t)

				res, err := e.svc.AuthDevice(ctx, e.deviceAuthRequest(t, k, "bb:00:00:00:00:04", false))
				require.NoError(t, err)

				claims, err := jwttoken.ClaimsFromBearerToken(publicKey, res.Token)
				require.NoError(t, err)
				deviceClaims, ok := claims.(*authorizer.DeviceClaims)
				require.True(t, ok)
				require.NoError(t, e.svc.AuthDeviceToken(ctx, deviceClaims))

				e.requireVerifiedDeviceKey(t, true)

				require.ErrorIs(t, e.svc.AuthDeviceToken(ctx, deviceClaims), ErrAuthUnathorized)

				_, err = e.svc.AuthDevice(ctx, e.deviceAuthRequest(t, k, "bb:00:00:00:00:04", false))
				require.ErrorIs(t, err, ErrDeviceKeyVerificationRefused)

				for _, event := range e.tenantRegistrations(t) {
					assert.Empty(t, event.Refusal, "a known device is refused without a record")
				}
			})

			t.Run("a device the auth cache holds is refused on its next unsigned request once the namespace requires verification", func(t *testing.T) {
				e := setup(t)
				e.svc = NewService(e.st, privateKey, publicKey, newMemoryCache(), WithClock(e.clock))
				req := e.deviceAuthRequest(t, key(t), "bb:00:00:00:00:06", false)

				for range 2 {
					_, err := e.svc.AuthDevice(ctx, req)
					require.NoError(t, err)
				}

				e.requireVerifiedDeviceKey(t, true)

				_, err := e.svc.AuthDevice(ctx, req)
				require.ErrorIs(t, err, ErrDeviceKeyVerificationRefused)
			})

			t.Run("a removed device that never verified is refused and recorded once the namespace requires verification", func(t *testing.T) {
				e := setup(t)
				k := key(t)

				res, err := e.svc.AuthDevice(ctx, e.deviceAuthRequest(t, k, "bb:00:00:00:00:05", false))
				require.NoError(t, err)
				require.NoError(t, e.accept(ctx, res.UID))
				require.NoError(t, e.svc.DeleteDevice(ctx, models.UID(res.UID), e.tenantID))

				e.requireVerifiedDeviceKey(t, true)

				_, err = e.svc.AuthDevice(ctx, e.deviceAuthRequest(t, k, "bb:00:00:00:00:05", false))
				require.ErrorIs(t, err, ErrDeviceKeyVerificationRefused)

				var refused []models.ProvisioningKeyEvent
				for _, event := range e.tenantRegistrations(t) {
					if event.Refusal != "" {
						refused = append(refused, event)
					}
				}

				require.Len(t, refused, 1, "a removed device is not in the device list, so its refusal is")
				assert.Equal(t, models.ProvisioningKeyEventRefusalUnverifiedDeviceKey, refused[0].Refusal)
				assert.Equal(t, "host-bb:00:00:00:00:05", refused[0].Hostname)
				assert.Empty(t, refused[0].DeviceStatus, "a refusal carries no device status, even with the removed row still there")
			})
		})
	}
}

func TestListDevicesByKeyVerification(t *testing.T) {
	ctx := context.Background()
	e := setupEnrollmentE2E(t)

	enroll := func(t *testing.T, mac string, signed bool) string {
		t.Helper()

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		res, err := e.svc.AuthDevice(ctx, e.deviceAuthRequest(t, key, mac, signed))
		require.NoError(t, err)

		return res.UID
	}

	signed := enroll(t, "cc:00:00:00:00:01", true)
	unsigned := enroll(t, "cc:00:00:00:00:02", false)

	list := func(t *testing.T, verified bool) map[string]bool {
		t.Helper()

		devices, count, err := e.svc.ListDevices(ctx, scope.MustBounded(e.tenantID), &requests.DeviceList{
			TenantID:  e.tenantID,
			Paginator: query.Paginator{Page: 1, PerPage: 100},
			Filters: query.Filters{Data: []query.Filter{{
				Type:   query.FilterTypeProperty,
				Params: &query.FilterProperty{Name: "key_verified", Operator: "bool", Value: verified},
			}}},
		})
		require.NoError(t, err)
		require.Equal(t, int64(len(devices)), count)

		found := make(map[string]bool, len(devices))
		for _, device := range devices {
			found[device.UID] = device.KeyVerifiedAt != nil
		}

		return found
	}

	assert.Equal(t, map[string]bool{signed: true}, list(t, true))
	assert.Equal(t, map[string]bool{unsigned: false}, list(t, false))
}

func TestSetupRequiresVerifiedDeviceKeyAsConfigured(t *testing.T) {
	ctx := context.Background()

	for _, required := range []bool{true, false} {
		t.Run(fmt.Sprintf("required %t", required), func(t *testing.T) {
			provider, err := pgprovider.NewProvider(ctx)
			require.NoError(t, err)
			t.Cleanup(func() { _ = provider.Close(t) })

			st := provider.Store()
			svc := NewService(st, privateKey, publicKey, storecache.NewNullCache(), WithNamespaceRequireVerifiedDeviceKey(required), WithClock(provider.Clock()))

			_, err = svc.Setup(ctx, requests.Setup{
				Email:     "owner@example.com",
				Name:      "Owner",
				Username:  "owner",
				Password:  "secret123",
				Namespace: "first",
			})
			require.NoError(t, err)

			system, err := st.SystemGet(ctx)
			require.NoError(t, err)

			namespace, err := st.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, system.InstanceTenantID)
			require.NoError(t, err)
			assert.Equal(t, required, namespace.Settings.RequireVerifiedDeviceKey)
		})
	}
}

func TestCreateNamespaceRequiresVerifiedDeviceKeyAsConfigured(t *testing.T) {
	ctx := context.Background()

	for _, required := range []bool{true, false} {
		t.Run(fmt.Sprintf("required %t", required), func(t *testing.T) {
			e := setupEnrollmentE2E(t)

			var owner string
			require.NoError(t, e.db.NewRaw("SELECT id FROM users WHERE username = 'owner'").Scan(ctx, &owner))

			svc := NewService(e.st, privateKey, publicKey, storecache.NewNullCache(), WithNamespaceRequireVerifiedDeviceKey(required), WithClock(e.clock))

			created, err := svc.CreateNamespace(ctx, &requests.NamespaceCreate{UserID: owner, Name: "created"})
			require.NoError(t, err)

			namespace, err := e.st.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, created.TenantID)
			require.NoError(t, err)
			assert.Equal(t, required, namespace.Settings.RequireVerifiedDeviceKey)

			existing, err := e.st.NamespaceResolve(ctx, store.NamespaceTenantIDResolver, e.tenantID)
			require.NoError(t, err)
			assert.False(t, existing.Settings.RequireVerifiedDeviceKey, "a namespace that already exists keeps its own setting")
		})
	}
}
