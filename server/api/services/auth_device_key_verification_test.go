package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"testing"
	"time"

	"github.com/cnf/structhash"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	mockcache "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/devicekey"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateDeviceAuthChallenge(t *testing.T) {
	ctx := context.TODO()

	service := NewService(mocks.NewMockStore(t), privateKey, &privateKey.PublicKey, mockcache.NewMockCache(t))

	first, err := service.CreateDeviceAuthChallenge(ctx)
	require.NoError(t, err)

	second, err := service.CreateDeviceAuthChallenge(ctx)
	require.NoError(t, err)

	assert.NotEqual(t, first.Challenge, second.Challenge, "every challenge must be fresh")
	assert.Equal(t, 60, first.ExpiresIn)
}

func TestAuthDeviceKeyVerification(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	const (
		tenantID      = "00000000-0000-4000-0000-000000000000"
		otherTenantID = "00000000-0000-4000-0000-000000000001"
		hostname      = "verified-device"
		mac           = "aa:bb:cc:dd:ee:ff"
	)

	deviceKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	otherServerKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	issue := func(t *testing.T, serverKey *rsa.PrivateKey, at time.Time) string {
		t.Helper()

		issued, err := NewService(mocks.NewMockStore(t), serverKey, &serverKey.PublicKey, mockcache.NewMockCache(t), WithClock(clock.NewFixed(at))).
			CreateDeviceAuthChallenge(context.TODO())
		require.NoError(t, err)

		return issued.Challenge
	}

	tamper := func(t *testing.T, challenge string) string {
		t.Helper()

		raw, err := base64.RawURLEncoding.DecodeString(challenge)
		require.NoError(t, err)
		raw[len(raw)-1] ^= 0x01

		return base64.RawURLEncoding.EncodeToString(raw)
	}

	challenge := issue(t, privateKey, now)
	otherChallenge := issue(t, privateKey, now)

	publicKey := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&deviceKey.PublicKey)}))

	uidSHA := sha256.Sum256(structhash.Dump(models.DeviceAuth{
		Hostname:  hostname,
		Identity:  &models.DeviceIdentity{MAC: mac},
		PublicKey: publicKey,
		TenantID:  tenantID,
	}, 1))
	uid := hex.EncodeToString(uidSHA[:])

	unsigned := requests.DeviceAuth{
		TenantID:  tenantID,
		Hostname:  hostname,
		Identity:  &requests.DeviceIdentity{MAC: mac},
		PublicKey: publicKey,
	}

	sign := func(t *testing.T, key *rsa.PrivateKey, signed, tenantID string) requests.DeviceAuth {
		t.Helper()

		signature, err := devicekey.Sign(key, devicekey.Statement{Challenge: signed, TenantID: tenantID, PublicKey: publicKey})
		require.NoError(t, err)

		req := unsigned
		req.Challenge = signed
		req.Signature = signature

		return req
	}

	refused := NewErrDeviceKeyVerificationRefused()

	t.Run("a signature by the device key issues the token and records the verification", func(t *testing.T) {
		ctx := context.TODO()
		storeMock := mocks.NewMockStore(t)
		cacheMock := mockcache.NewMockCache(t)

		device := &models.Device{UID: uid, Name: hostname, TenantID: tenantID, Status: models.DeviceStatusAccepted}

		storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
			Return(&models.Namespace{TenantID: tenantID, Name: "test"}, nil).Once()
		cacheMock.On("Get", ctx, deviceAuthCacheKey(uid), testifymock.Anything).Return(nil).Once()
		storeMock.On("DeviceResolve", ctx, testifymock.Anything, store.DeviceUIDResolver, uid).Return(device, nil).Once()

		pinned := *device
		pinned.KeyVerifiedAt = &now
		storeMock.On("DeviceUpdateUnlessRemoved", ctx, &pinned).Return(nil).Once()
		cacheMock.On("Set", ctx, deviceAuthCacheKey(uid), map[string]string{"device_name": hostname, "namespace_name": "test", "key_verified": "true"}, 30*time.Second).
			Return(nil).Once()

		service := NewService(storeMock, privateKey, &privateKey.PublicKey, cacheMock, WithIssuer(testIssuer), WithClock(clock.NewFixed(now)))

		res, err := service.AuthDevice(ctx, sign(t, deviceKey, challenge, tenantID))
		require.NoError(t, err)
		assert.Equal(t, uid, res.UID)

		claims, err := jwttoken.ClaimsFromBearerToken(&privateKey.PublicKey, res.Token)
		require.NoError(t, err)

		deviceClaims, ok := claims.(*authorizer.DeviceClaims)
		require.True(t, ok)
		assert.True(t, deviceClaims.KeyVerified, "a signed request earns a token that says so")
	})

	t.Run("a signature skips a cache entry from before the verification, so the verification is recorded", func(t *testing.T) {
		ctx := context.TODO()
		storeMock := mocks.NewMockStore(t)
		cacheMock := mockcache.NewMockCache(t)

		device := &models.Device{UID: uid, Name: hostname, TenantID: tenantID, Status: models.DeviceStatusAccepted}

		storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
			Return(&models.Namespace{TenantID: tenantID, Name: "test"}, nil).Once()
		cacheMock.On("Get", ctx, deviceAuthCacheKey(uid), testifymock.Anything).
			Run(func(args testifymock.Arguments) {
				cached, ok := args.Get(2).(*map[string]string)
				require.True(t, ok)
				(*cached)["device_name"] = hostname
				(*cached)["namespace_name"] = "test"
			}).
			Return(nil).Once()
		storeMock.On("DeviceResolve", ctx, testifymock.Anything, store.DeviceUIDResolver, uid).Return(device, nil).Once()

		verified := *device
		verified.KeyVerifiedAt = &now
		storeMock.On("DeviceUpdateUnlessRemoved", ctx, &verified).Return(nil).Once()
		cacheMock.On("Set", ctx, deviceAuthCacheKey(uid), map[string]string{"device_name": hostname, "namespace_name": "test", "key_verified": "true"}, 30*time.Second).
			Return(nil).Once()

		service := NewService(storeMock, privateKey, &privateKey.PublicKey, cacheMock, WithIssuer(testIssuer), WithClock(clock.NewFixed(now)))

		res, err := service.AuthDevice(ctx, sign(t, deviceKey, challenge, tenantID))
		require.NoError(t, err)
		assert.Equal(t, uid, res.UID)
	})

	t.Run("an unsigned request cannot obtain the token of a device that verified its key", func(t *testing.T) {
		ctx := context.TODO()
		storeMock := mocks.NewMockStore(t)
		cacheMock := mockcache.NewMockCache(t)

		storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
			Return(&models.Namespace{TenantID: tenantID, Name: "test"}, nil).Once()
		cacheMock.On("Get", ctx, deviceAuthCacheKey(uid), testifymock.Anything).Return(nil).Once()
		storeMock.On("DeviceResolve", ctx, testifymock.Anything, store.DeviceUIDResolver, uid).
			Return(&models.Device{UID: uid, Name: hostname, TenantID: tenantID, Status: models.DeviceStatusAccepted, KeyVerifiedAt: &now}, nil).Once()

		service := NewService(storeMock, privateKey, &privateKey.PublicKey, cacheMock, WithIssuer(testIssuer), WithClock(clock.NewFixed(now)))

		res, err := service.AuthDevice(ctx, unsigned)
		assert.Nil(t, res)
		assert.Equal(t, refused, err)
	})

	t.Run("an unsigned request cannot obtain a cached token of a device that verified its key", func(t *testing.T) {
		ctx := context.TODO()
		storeMock := mocks.NewMockStore(t)
		cacheMock := mockcache.NewMockCache(t)

		storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
			Return(&models.Namespace{TenantID: tenantID, Name: "test"}, nil).Once()
		cacheMock.On("Get", ctx, deviceAuthCacheKey(uid), testifymock.Anything).
			Run(func(args testifymock.Arguments) {
				cached, ok := args.Get(2).(*map[string]string)
				require.True(t, ok)
				(*cached)["device_name"] = hostname
				(*cached)["namespace_name"] = "test"
				(*cached)["key_verified"] = "true"
			}).
			Return(nil).Once()

		service := NewService(storeMock, privateKey, &privateKey.PublicKey, cacheMock, WithIssuer(testIssuer), WithClock(clock.NewFixed(now)))

		res, err := service.AuthDevice(ctx, unsigned)
		assert.Nil(t, res)
		assert.Equal(t, refused, err)
	})

	t.Run("a token served from a cache entry that predates the verification is refused", func(t *testing.T) {
		ctx := context.TODO()
		storeMock := mocks.NewMockStore(t)
		cacheMock := mockcache.NewMockCache(t)

		storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
			Return(&models.Namespace{TenantID: tenantID, Name: "test"}, nil).Once()
		cacheMock.On("Get", ctx, deviceAuthCacheKey(uid), testifymock.Anything).
			Run(func(args testifymock.Arguments) {
				cached, ok := args.Get(2).(*map[string]string)
				require.True(t, ok)
				(*cached)["device_name"] = hostname
				(*cached)["namespace_name"] = "test"
			}).
			Return(nil).Once()
		storeMock.On("DeviceResolve", ctx, testifymock.Anything, store.DeviceUIDResolver, uid).
			Return(&models.Device{UID: uid, Name: hostname, TenantID: tenantID, Status: models.DeviceStatusAccepted, KeyVerifiedAt: &now}, nil).Once()

		service := NewService(storeMock, privateKey, &privateKey.PublicKey, cacheMock, WithIssuer(testIssuer), WithClock(clock.NewFixed(now)))

		res, err := service.AuthDevice(ctx, unsigned)
		require.NoError(t, err)

		claims, err := jwttoken.ClaimsFromBearerToken(&privateKey.PublicKey, res.Token)
		require.NoError(t, err)

		deviceClaims, ok := claims.(*authorizer.DeviceClaims)
		require.True(t, ok)

		assert.Equal(t, NewErrAuthUnathorized(nil), service.AuthDeviceToken(ctx, deviceClaims))
	})

	t.Run("the same signed request is honoured twice within the window", func(t *testing.T) {
		ctx := context.TODO()
		storeMock := mocks.NewMockStore(t)
		cacheMock := mockcache.NewMockCache(t)

		storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
			Return(&models.Namespace{TenantID: tenantID, Name: "test"}, nil).Twice()
		cacheMock.On("Get", ctx, deviceAuthCacheKey(uid), testifymock.Anything).
			Run(func(args testifymock.Arguments) {
				cached, ok := args.Get(2).(*map[string]string)
				require.True(t, ok)
				(*cached)["device_name"] = hostname
				(*cached)["namespace_name"] = "test"
				(*cached)["key_verified"] = "true"
			}).
			Return(nil).Twice()

		service := NewService(storeMock, privateKey, &privateKey.PublicKey, cacheMock, WithIssuer(testIssuer), WithClock(clock.NewFixed(now)))
		req := sign(t, deviceKey, challenge, tenantID)

		for range 2 {
			res, err := service.AuthDevice(ctx, req)
			require.NoError(t, err)
			assert.Equal(t, uid, res.UID)
		}
	})

	refusedSignatures := []struct {
		description string
		key         *rsa.PrivateKey
		challenge   func(t *testing.T) string
		tenant      string
		alter       func(*requests.DeviceAuth)
	}{
		{
			description: "a challenge with a tampered tag",
			key:         deviceKey,
			challenge: func(t *testing.T) string {
				t.Helper()

				return tamper(t, challenge)
			},
			tenant: tenantID,
		},
		{
			description: "a challenge of the wrong length",
			key:         deviceKey,
			challenge:   func(*testing.T) string { return "issued-challenge" },
			tenant:      tenantID,
		},
		{
			description: "a challenge that is not base64",
			key:         deviceKey,
			challenge:   func(*testing.T) string { return "not base64 at all!" },
			tenant:      tenantID,
		},
		{
			description: "a challenge issued by a server with another signing key",
			key:         deviceKey,
			challenge: func(t *testing.T) string {
				t.Helper()

				return issue(t, otherServerKey, now)
			},
			tenant: tenantID,
		},
		{
			description: "a signature by another key",
			key:         otherKey,
			challenge:   func(*testing.T) string { return challenge },
			tenant:      tenantID,
		},
		{
			description: "a signature for another tenant",
			key:         deviceKey,
			challenge:   func(*testing.T) string { return challenge },
			tenant:      otherTenantID,
		},
		{
			description: "a signature over another challenge",
			key:         deviceKey,
			challenge:   func(*testing.T) string { return otherChallenge },
			tenant:      tenantID,
			alter:       func(req *requests.DeviceAuth) { req.Challenge = challenge },
		},
		{
			description: "a signature for another provisioning key",
			key:         deviceKey,
			challenge:   func(*testing.T) string { return challenge },
			tenant:      tenantID,
			alter:       func(req *requests.DeviceAuth) { req.ProvisioningKey = "another-provisioning-key" },
		},
		{
			description: "a challenge without a signature",
			key:         deviceKey,
			challenge:   func(*testing.T) string { return challenge },
			tenant:      tenantID,
			alter:       func(req *requests.DeviceAuth) { req.Signature = "" },
		},
		{
			description: "a signature without a challenge",
			key:         deviceKey,
			challenge:   func(*testing.T) string { return challenge },
			tenant:      tenantID,
			alter:       func(req *requests.DeviceAuth) { req.Challenge = "" },
		},
		{
			description: "an altered signature",
			key:         deviceKey,
			challenge:   func(*testing.T) string { return challenge },
			tenant:      tenantID,
			alter:       func(req *requests.DeviceAuth) { req.Signature = req.Signature[:len(req.Signature)-8] + "AAAAAAA=" },
		},
	}

	for _, tc := range refusedSignatures {
		t.Run("refuses "+tc.description, func(t *testing.T) {
			service := NewService(mocks.NewMockStore(t), privateKey, &privateKey.PublicKey, mockcache.NewMockCache(t), WithIssuer(testIssuer), WithClock(clock.NewFixed(now)))

			req := sign(t, tc.key, tc.challenge(t), tc.tenant)
			if tc.alter != nil {
				tc.alter(&req)
			}

			res, err := service.AuthDevice(context.TODO(), req)
			assert.Nil(t, res)
			assert.Equal(t, refused, err)
		})
	}

	t.Run("refuses a challenge older than the window", func(t *testing.T) {
		service := NewService(mocks.NewMockStore(t), privateKey, &privateKey.PublicKey, mockcache.NewMockCache(t), WithIssuer(testIssuer), WithClock(clock.NewFixed(now.Add(61*time.Second))))

		res, err := service.AuthDevice(context.TODO(), sign(t, deviceKey, challenge, tenantID))
		assert.Nil(t, res)
		assert.Equal(t, refused, err)
	})

	t.Run("refuses a challenge stamped further ahead than the clock skew", func(t *testing.T) {
		ahead := issue(t, privateKey, now.Add(6*time.Second))

		service := NewService(mocks.NewMockStore(t), privateKey, &privateKey.PublicKey, mockcache.NewMockCache(t), WithIssuer(testIssuer), WithClock(clock.NewFixed(now)))

		res, err := service.AuthDevice(context.TODO(), sign(t, deviceKey, ahead, tenantID))
		assert.Nil(t, res)
		assert.Equal(t, refused, err)
	})
}

func TestAuthDeviceToken(t *testing.T) {
	const (
		tenantID = "00000000-0000-4000-0000-000000000000"
		uid      = "a3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	)

	verifiedAt := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		description string
		device      *models.Device
		resolveErr  error
		keyVerified bool
		namespace   *models.Namespace
		resolveNs   error
		expected    error
	}{
		{
			description: "refuses a token issued without a signature once the device verified its key",
			device:      &models.Device{UID: uid, TenantID: tenantID, KeyVerifiedAt: &verifiedAt},
			expected:    NewErrAuthUnathorized(nil),
		},
		{
			description: "honours a token issued against a signature",
			device:      &models.Device{UID: uid, TenantID: tenantID, KeyVerifiedAt: &verifiedAt},
			keyVerified: true,
		},
		{
			description: "honours a token of a device that never verified its key",
			device:      &models.Device{UID: uid, TenantID: tenantID},
			namespace:   &models.Namespace{TenantID: tenantID, Settings: &models.NamespaceSettings{}},
		},
		{
			description: "refuses a token of a device that never verified its key when its namespace requires verification",
			device:      &models.Device{UID: uid, TenantID: tenantID},
			namespace:   &models.Namespace{TenantID: tenantID, Settings: &models.NamespaceSettings{RequireVerifiedDeviceKey: true}},
			expected:    NewErrAuthUnathorized(nil),
		},
		{
			description: "refuses a token of a device that never verified its key once its namespace is gone",
			device:      &models.Device{UID: uid, TenantID: tenantID},
			resolveNs:   store.ErrNoDocuments,
			expected:    NewErrAuthUnathorized(nil),
		},
		{
			description: "reports a store failure reading the namespace of a device that never verified its key",
			device:      &models.Device{UID: uid, TenantID: tenantID},
			resolveNs:   errors.New("error", "store", 0),
			expected:    errors.New("error", "store", 0),
		},
		{
			description: "refuses a token of a device the store no longer has",
			resolveErr:  store.ErrNoDocuments,
			expected:    NewErrAuthUnathorized(nil),
		},
		{
			description: "reports a store failure",
			resolveErr:  errors.New("error", "store", 0),
			expected:    errors.New("error", "store", 0),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()
			storeMock := mocks.NewMockStore(t)
			storeMock.On("DeviceResolve", ctx, testifymock.Anything, store.DeviceUIDResolver, uid).Return(tc.device, tc.resolveErr).Once()
			if tc.namespace != nil || tc.resolveNs != nil {
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).Return(tc.namespace, tc.resolveNs).Once()
			}

			service := NewService(storeMock, privateKey, &privateKey.PublicKey, mockcache.NewMockCache(t))

			err := service.AuthDeviceToken(ctx, &authorizer.DeviceClaims{UID: uid, TenantID: tenantID, KeyVerified: tc.keyVerified})
			assert.Equal(t, tc.expected, err)
		})
	}

	t.Run("refuses claims without a tenant", func(t *testing.T) {
		service := NewService(mocks.NewMockStore(t), privateKey, &privateKey.PublicKey, mockcache.NewMockCache(t), WithIssuer(testIssuer))

		err := service.AuthDeviceToken(context.TODO(), &authorizer.DeviceClaims{UID: uid})
		assert.ErrorIs(t, err, ErrAuthUnathorized)
	})
}
