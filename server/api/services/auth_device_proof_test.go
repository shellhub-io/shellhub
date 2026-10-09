package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
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
	cacheMock := mockcache.NewMockCache(t)

	cacheMock.
		On("SetNX", ctx, testifymock.AnythingOfType("string"), "issued", time.Minute).
		Return(true, nil).
		Twice()

	service := NewService(mocks.NewMockStore(t), privateKey, &privateKey.PublicKey, cacheMock)

	first, err := service.CreateDeviceAuthChallenge(ctx)
	require.NoError(t, err)

	second, err := service.CreateDeviceAuthChallenge(ctx)
	require.NoError(t, err)

	assert.NotEqual(t, first.Challenge, second.Challenge, "every challenge must be fresh")
	assert.Equal(t, 60, first.ExpiresIn)
}

func TestAuthDeviceKeyProof(t *testing.T) {
	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	clock.Freeze(t, now)

	const (
		tenantID      = "00000000-0000-4000-0000-000000000000"
		otherTenantID = "00000000-0000-4000-0000-000000000001"
		hostname      = "proof-device"
		mac           = "aa:bb:cc:dd:ee:ff"
		challenge     = "issued-challenge"
	)

	deviceKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

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
		req.Challenge = challenge
		req.Signature = signature

		return req
	}

	refused := NewErrDeviceKeyProofRefused()

	t.Run("a proof signed by the device key issues the token and pins the key", func(t *testing.T) {
		ctx := context.TODO()
		storeMock := mocks.NewMockStore(t)
		cacheMock := mockcache.NewMockCache(t)

		device := &models.Device{UID: uid, Name: hostname, TenantID: tenantID, Status: models.DeviceStatusAccepted}

		cacheMock.On("CompareAndDelete", ctx, deviceAuthChallengeKey(challenge), "issued").Return(true, nil).Once()
		storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
			Return(&models.Namespace{TenantID: tenantID, Name: "test"}, nil).Once()
		cacheMock.On("Get", ctx, deviceAuthCacheKey(uid), testifymock.Anything).Return(nil).Once()
		storeMock.On("DeviceResolve", ctx, testifymock.Anything, store.DeviceUIDResolver, uid).Return(device, nil).Once()

		pinned := *device
		pinned.LastSeen = now
		pinned.KeyProvenAt = &now
		storeMock.On("DeviceUpdateUnlessRemoved", ctx, &pinned).Return(nil).Once()
		storeMock.On("DeviceHeartbeat", ctx, []string{uid}, now).Return(nil, nil).Once()
		cacheMock.On("Set", ctx, deviceAuthCacheKey(uid), map[string]string{"device_name": hostname, "namespace_name": "test", "key_proven": "true"}, 30*time.Second).
			Return(nil).Once()

		service := NewService(storeMock, privateKey, &privateKey.PublicKey, cacheMock, WithIssuer(testIssuer))

		res, err := service.AuthDevice(ctx, sign(t, deviceKey, challenge, tenantID))
		require.NoError(t, err)
		assert.Equal(t, uid, res.UID)
		assert.NotEmpty(t, res.Token)
	})

	t.Run("an unsigned request cannot obtain the token of a device that proved its key", func(t *testing.T) {
		ctx := context.TODO()
		storeMock := mocks.NewMockStore(t)
		cacheMock := mockcache.NewMockCache(t)

		storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenantID).
			Return(&models.Namespace{TenantID: tenantID, Name: "test"}, nil).Once()
		cacheMock.On("Get", ctx, deviceAuthCacheKey(uid), testifymock.Anything).Return(nil).Once()
		storeMock.On("DeviceResolve", ctx, testifymock.Anything, store.DeviceUIDResolver, uid).
			Return(&models.Device{UID: uid, Name: hostname, TenantID: tenantID, Status: models.DeviceStatusAccepted, KeyProvenAt: &now}, nil).Once()

		service := NewService(storeMock, privateKey, &privateKey.PublicKey, cacheMock, WithIssuer(testIssuer))

		res, err := service.AuthDevice(ctx, unsigned)
		assert.Nil(t, res)
		assert.Equal(t, refused, err)
	})

	t.Run("an unsigned request cannot obtain a cached token of a device that proved its key", func(t *testing.T) {
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
				(*cached)["key_proven"] = "true"
			}).
			Return(nil).Once()

		service := NewService(storeMock, privateKey, &privateKey.PublicKey, cacheMock, WithIssuer(testIssuer))

		res, err := service.AuthDevice(ctx, unsigned)
		assert.Nil(t, res)
		assert.Equal(t, refused, err)
	})

	t.Run("a token served from a cache entry that predates the proof is refused", func(t *testing.T) {
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
			Return(&models.Device{UID: uid, Name: hostname, TenantID: tenantID, Status: models.DeviceStatusAccepted, KeyProvenAt: &now}, nil).Once()

		service := NewService(storeMock, privateKey, &privateKey.PublicKey, cacheMock, WithIssuer(testIssuer))

		res, err := service.AuthDevice(ctx, unsigned)
		require.NoError(t, err)

		claims, err := jwttoken.ClaimsFromBearerToken(&privateKey.PublicKey, res.Token)
		require.NoError(t, err)

		deviceClaims, ok := claims.(*authorizer.DeviceClaims)
		require.True(t, ok)

		assert.Equal(t, NewErrAuthUnathorized(nil), service.AuthDeviceToken(ctx, deviceClaims))
	})

	t.Run("an unsigned request is refused when the instance requires a proof", func(t *testing.T) {
		service := NewService(mocks.NewMockStore(t), privateKey, &privateKey.PublicKey, mockcache.NewMockCache(t), WithIssuer(testIssuer), WithDeviceKeyProofRequired())

		res, err := service.AuthDevice(context.TODO(), unsigned)
		assert.Nil(t, res)
		assert.Equal(t, refused, err)
	})

	refusedProofs := []struct {
		description     string
		key             *rsa.PrivateKey
		signedChallenge string
		signedTenant    string
		alter           func(*requests.DeviceAuth)
		issued          bool
	}{
		{
			description:     "a challenge the server did not issue, has expired or was already used",
			key:             deviceKey,
			signedChallenge: challenge,
			signedTenant:    tenantID,
			issued:          false,
		},
		{
			description:     "a proof signed by another key",
			key:             otherKey,
			signedChallenge: challenge,
			signedTenant:    tenantID,
			issued:          true,
		},
		{
			description:     "a proof signed for another tenant",
			key:             deviceKey,
			signedChallenge: challenge,
			signedTenant:    otherTenantID,
			issued:          true,
		},
		{
			description:     "a proof signed over another challenge",
			key:             deviceKey,
			signedChallenge: "another-challenge",
			signedTenant:    tenantID,
			issued:          true,
		},
		{
			description:     "a proof signed for another provisioning key",
			key:             deviceKey,
			signedChallenge: challenge,
			signedTenant:    tenantID,
			alter:           func(req *requests.DeviceAuth) { req.ProvisioningKey = "another-provisioning-key" },
			issued:          true,
		},
		{
			description:     "an altered signature",
			key:             deviceKey,
			signedChallenge: challenge,
			signedTenant:    tenantID,
			alter:           func(req *requests.DeviceAuth) { req.Signature = req.Signature[:len(req.Signature)-8] + "AAAAAAA=" },
			issued:          true,
		},
	}

	for _, tc := range refusedProofs {
		t.Run("refuses "+tc.description, func(t *testing.T) {
			ctx := context.TODO()
			cacheMock := mockcache.NewMockCache(t)
			cacheMock.On("CompareAndDelete", ctx, deviceAuthChallengeKey(challenge), "issued").Return(tc.issued, nil).Once()

			service := NewService(mocks.NewMockStore(t), privateKey, &privateKey.PublicKey, cacheMock, WithIssuer(testIssuer))

			req := sign(t, tc.key, tc.signedChallenge, tc.signedTenant)
			if tc.alter != nil {
				tc.alter(&req)
			}

			res, err := service.AuthDevice(ctx, req)
			assert.Nil(t, res)
			assert.Equal(t, refused, err)
		})
	}

	t.Run("refuses a signature without a challenge", func(t *testing.T) {
		req := sign(t, deviceKey, challenge, tenantID)
		req.Challenge = ""

		service := NewService(mocks.NewMockStore(t), privateKey, &privateKey.PublicKey, mockcache.NewMockCache(t), WithIssuer(testIssuer))

		res, err := service.AuthDevice(context.TODO(), req)
		assert.Nil(t, res)
		assert.Equal(t, refused, err)
	})

	t.Run("reports a cache failure while consuming the challenge", func(t *testing.T) {
		ctx := context.TODO()
		cacheMock := mockcache.NewMockCache(t)
		cacheMock.On("CompareAndDelete", ctx, deviceAuthChallengeKey(challenge), "issued").Return(false, errors.New("error", "cache", 0)).Once()

		service := NewService(mocks.NewMockStore(t), privateKey, &privateKey.PublicKey, cacheMock, WithIssuer(testIssuer))

		res, err := service.AuthDevice(ctx, sign(t, deviceKey, challenge, tenantID))
		assert.Nil(t, res)
		assert.Equal(t, errors.New("error", "cache", 0), err)
	})
}

func TestAuthDeviceToken(t *testing.T) {
	const (
		tenantID = "00000000-0000-4000-0000-000000000000"
		uid      = "a3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	)

	provenAt := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		description string
		device      *models.Device
		resolveErr  error
		keyProven   bool
		options     []Option
		expected    error
	}{
		{
			description: "refuses a token issued without a proof once the device proved its key",
			device:      &models.Device{UID: uid, TenantID: tenantID, KeyProvenAt: &provenAt},
			expected:    NewErrAuthUnathorized(nil),
		},
		{
			description: "honours a token issued against a proof",
			device:      &models.Device{UID: uid, TenantID: tenantID, KeyProvenAt: &provenAt},
			keyProven:   true,
		},
		{
			description: "honours a token of a device that never proved its key",
			device:      &models.Device{UID: uid, TenantID: tenantID},
		},
		{
			description: "refuses a token of a device that never proved its key when the instance requires a proof",
			device:      &models.Device{UID: uid, TenantID: tenantID},
			options:     []Option{WithDeviceKeyProofRequired()},
			expected:    NewErrAuthUnathorized(nil),
		},
		{
			description: "leaves a token of an unknown device to the routes it reaches",
			resolveErr:  store.ErrNoDocuments,
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

			service := NewService(storeMock, privateKey, &privateKey.PublicKey, mockcache.NewMockCache(t), tc.options...)

			err := service.AuthDeviceToken(ctx, &authorizer.DeviceClaims{UID: uid, TenantID: tenantID, KeyProven: tc.keyProven})
			assert.Equal(t, tc.expected, err)
		})
	}
}
