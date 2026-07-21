package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/cnf/structhash"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	mockcache "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuidmock "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestAuthDevice_InstallKey covers the install-key enrollment decisions in AuthDevice: an invalid or
// system key is rejected without creating a device, and a tenant-only enrollment is attributed to
// the namespace's legacy key without being accepted. The full accept + usage + tag success path
// (which drives UpdateDeviceStatus) is left to end-to-end verification.
func TestAuthDevice_InstallKey(t *testing.T) {
	storeMock := mocks.NewMockStore(t)
	queryOptionsMock := mocks.NewMockQueryOptions(t)
	storeMock.On("Options").Return(queryOptionsMock).Maybe()
	cacheMock := mockcache.NewMockCache(t)
	clockMock := clockmock.NewMockClock(t)
	uuidMock := uuidmock.NewMockUUID(t)

	now := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	prevClock := clock.DefaultBackend
	prevUUID := uuid.DefaultBackend
	t.Cleanup(func() {
		clock.DefaultBackend = prevClock
		uuid.DefaultBackend = prevUUID
	})
	clock.DefaultBackend = clockMock
	clockMock.On("Now").Return(now)
	uuid.DefaultBackend = uuidMock
	uuidMock.On("Generate").Return("00000000-0000-0000-0000-000000000000")

	const tenant = "00000000-0000-4000-0000-000000000000"

	toUID := func(hostname, mac, publicKey string) string {
		auth := models.DeviceAuth{Hostname: strings.ToLower(hostname), Identity: &models.DeviceIdentity{MAC: mac}, PublicKey: publicKey, TenantID: tenant}
		uidSHA := sha256.Sum256(structhash.Dump(auth, 1))

		return hex.EncodeToString(uidSHA[:])
	}

	toToken := func(uid string) string {
		token, err := jwttoken.EncodeDeviceClaims(authorizer.DeviceClaims{UID: uid, TenantID: tenant}, privateKey)
		require.NoError(t, err)

		return token
	}

	badDigest := hashInstallKey("bad-key")

	type Expected struct {
		res *models.DeviceAuthResponse
		err error
	}

	cases := []struct {
		description   string
		req           requests.DeviceAuth
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "rejects an invalid install key without creating the device",
			req:         requests.DeviceAuth{TenantID: tenant, Hostname: "d", Identity: &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"}, Info: &requests.DeviceInfo{}, PublicKey: "pk", InstallKey: "bad-key"},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("d", "aa:bb:cc:dd:ee:ff", "pk")
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).Return(&models.Namespace{TenantID: tenant, Name: "test"}, nil).Once()
				cacheMock.On("Get", ctx, "auth_device/"+uid, testifymock.Anything).Return(nil).Once()
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, uid).Return(nil, store.ErrNoDocuments).Once()
				queryOptionsMock.On("InNamespace", tenant).Return(nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyIDResolver, badDigest, testifymock.AnythingOfType("[]store.QueryOption")).Return(nil, store.ErrNoDocuments).Once()
			},
			expected: Expected{res: nil, err: NewErrAuthInvalid(map[string]interface{}{"install_key": "invalid"}, store.ErrNoDocuments)},
		},
		{
			description: "rejects the system legacy key presented by an agent",
			req:         requests.DeviceAuth{TenantID: tenant, Hostname: "d", Identity: &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"}, Info: &requests.DeviceInfo{}, PublicKey: "pk", InstallKey: "bad-key"},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("d", "aa:bb:cc:dd:ee:ff", "pk")
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).Return(&models.Namespace{TenantID: tenant, Name: "test"}, nil).Once()
				cacheMock.On("Get", ctx, "auth_device/"+uid, testifymock.Anything).Return(nil).Once()
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, uid).Return(nil, store.ErrNoDocuments).Once()
				queryOptionsMock.On("InNamespace", tenant).Return(nil).Once()
				storeMock.On("InstallKeyResolve", ctx, store.InstallKeyIDResolver, badDigest, testifymock.AnythingOfType("[]store.QueryOption")).Return(&models.InstallKey{ID: badDigest, TenantID: tenant, Type: models.InstallKeyTypeLegacy, Reusable: true}, nil).Once()
			},
			expected: Expected{res: nil, err: NewErrAuthInvalid(map[string]interface{}{"install_key": "invalid"}, nil)},
		},
		{
			description: "attributes a tenant-only device to the legacy key without accepting it",
			req:         requests.DeviceAuth{TenantID: tenant, Hostname: "d", Identity: &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"}, Info: &requests.DeviceInfo{}, PublicKey: "pk"},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("d", "aa:bb:cc:dd:ee:ff", "pk")
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).Return(&models.Namespace{TenantID: tenant, Name: "test"}, nil).Once()
				cacheMock.On("Get", ctx, "auth_device/"+uid, testifymock.Anything).Return(nil).Once()
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, uid).Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("InstallKeyResolveSystem", ctx, tenant).Return(&models.InstallKey{ID: "legacydigest", TenantID: tenant, Type: models.InstallKeyTypeLegacy, Mode: models.InstallKeyModeManual}, nil).Once()
				storeMock.On("DeviceCreate", ctx, &models.Device{
					CreatedAt:       now,
					UID:             uid,
					TenantID:        tenant,
					LastSeen:        now,
					DisconnectedAt:  nil,
					Status:          models.DeviceStatusPending,
					StatusUpdatedAt: now,
					Name:            "d",
					Identity:        &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
					PublicKey:       "pk",
					RemoteAddr:      "",
					Taggable:        models.Taggable{TagIDs: []string{}},
					Position:        &models.DevicePosition{},
					Info:            &models.DeviceInfo{},
					InstallKeyID:    "legacydigest",
				}).Return(uid, nil).Once()
				storeMock.On("NamespaceIncrementDeviceCount", ctx, tenant, models.DeviceStatusPending, int64(1)).Return(nil).Once()
				// The legacy key is manual, so the keyless device lands pending and the enrollment is
				// recorded in the legacy key's append-only history.
				storeMock.On("InstallKeyEventCreate", ctx, testifymock.Anything).Return(nil).Once()
				cacheMock.On("Set", ctx, "auth_device/"+uid, map[string]string{"device_name": "d", "namespace_name": "test"}, time.Second*30).Return(nil).Once()
			},
			expected: Expected{
				res: &models.DeviceAuthResponse{
					UID:       toUID("d", "aa:bb:cc:dd:ee:ff", "pk"),
					Token:     toToken(toUID("d", "aa:bb:cc:dd:ee:ff", "pk")),
					Name:      "d",
					Namespace: "test",
					Status:    models.DeviceStatusPending,
				},
				err: nil,
			},
		},
	}

	service := NewService(store.Store(storeMock), privateKey, &privateKey.PublicKey, cacheMock, clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)

			res, err := service.AuthDevice(ctx, tc.req)
			require.Equal(tt, tc.expected.res, res)
			require.Equal(tt, tc.expected.err, err)
		})
	}

	storeMock.AssertExpectations(t)
}

// TestEnrollmentInstallKey covers keyless enrollment-source resolution: the pairing-code flow (paired)
// attributes to the pairing system key, a plain tenant-only enrollment attributes to the legacy key,
// and neither resolves the other's key.
func TestEnrollmentInstallKey(t *testing.T) {
	const tenant = "00000000-0000-4000-0000-000000000000"

	legacy := &models.InstallKey{ID: "legacydigest", TenantID: tenant, Type: models.InstallKeyTypeLegacy, Mode: models.InstallKeyModeManual}
	pairing := &models.InstallKey{ID: "pairingdigest", TenantID: tenant, Type: models.InstallKeyTypePairing, Mode: models.InstallKeyModeAutomatic}

	cases := []struct {
		description   string
		paired        bool
		requiredMocks func(context.Context, *mocks.MockStore)
		expectedKey   *models.InstallKey
		expectedID    string
	}{
		{
			description: "attributes a tenant-only enrollment to the legacy key",
			paired:      false,
			requiredMocks: func(ctx context.Context, s *mocks.MockStore) {
				s.On("InstallKeyResolveSystem", ctx, tenant).Return(legacy, nil).Once()
			},
			expectedKey: legacy,
			expectedID:  "legacydigest",
		},
		{
			description: "attributes a paired enrollment to the pairing key, not the legacy key",
			paired:      true,
			requiredMocks: func(ctx context.Context, s *mocks.MockStore) {
				s.On("InstallKeyResolveSystemPairing", ctx, tenant).Return(pairing, nil).Once()
			},
			expectedKey: pairing,
			expectedID:  "pairingdigest",
		},
		{
			description: "attributes nothing when the pairing key is missing",
			paired:      true,
			requiredMocks: func(ctx context.Context, s *mocks.MockStore) {
				s.On("InstallKeyResolveSystemPairing", ctx, tenant).Return(nil, store.ErrNoDocuments).Once()
			},
			expectedKey: nil,
			expectedID:  "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.TODO()
			storeMock := mocks.NewMockStore(tt)
			cacheMock := mockcache.NewMockCache(tt)
			tc.requiredMocks(ctx, storeMock)

			svc := NewService(store.Store(storeMock), privateKey, &privateKey.PublicKey, cacheMock, clientMock)
			key, id, err := svc.enrollmentInstallKey(ctx, requests.DeviceAuth{TenantID: tenant}, tc.paired)

			require.NoError(tt, err)
			require.Equal(tt, tc.expectedKey, key)
			require.Equal(tt, tc.expectedID, id)
			storeMock.AssertExpectations(tt)
		})
	}
}
