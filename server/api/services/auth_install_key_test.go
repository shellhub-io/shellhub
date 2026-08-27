package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/cnf/structhash"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	mockcache "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuidmock "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/mocks"
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
				storeMock.On("DeviceResolve", ctx, testifymock.Anything, store.DeviceUIDResolver, uid).Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("InstallKeyResolve", ctx, testifymock.Anything, store.InstallKeyIDResolver, badDigest).Return(nil, store.ErrNoDocuments).Once()
			},
			expected: Expected{res: nil, err: NewErrAuthInvalid(map[string]any{"install_key": "invalid"}, store.ErrNoDocuments)},
		},
		{
			description: "rejects the system legacy key presented by an agent",
			req:         requests.DeviceAuth{TenantID: tenant, Hostname: "d", Identity: &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"}, Info: &requests.DeviceInfo{}, PublicKey: "pk", InstallKey: "bad-key"},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("d", "aa:bb:cc:dd:ee:ff", "pk")
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).Return(&models.Namespace{TenantID: tenant, Name: "test"}, nil).Once()
				cacheMock.On("Get", ctx, "auth_device/"+uid, testifymock.Anything).Return(nil).Once()
				storeMock.On("DeviceResolve", ctx, testifymock.Anything, store.DeviceUIDResolver, uid).Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("InstallKeyResolve", ctx, testifymock.Anything, store.InstallKeyIDResolver, badDigest).Return(&models.InstallKey{ID: badDigest, TenantID: tenant, Type: models.InstallKeyTypeLegacy, Reusable: true}, nil).Once()
			},
			expected: Expected{res: nil, err: NewErrAuthInvalid(map[string]any{"install_key": "invalid"}, nil)},
		},
		{
			description: "attributes a tenant-only device to the legacy key without accepting it",
			req:         requests.DeviceAuth{TenantID: tenant, Hostname: "d", Identity: &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"}, Info: &requests.DeviceInfo{}, PublicKey: "pk"},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("d", "aa:bb:cc:dd:ee:ff", "pk")
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).Return(&models.Namespace{TenantID: tenant, Name: "test"}, nil).Once()
				cacheMock.On("Get", ctx, "auth_device/"+uid, testifymock.Anything).Return(nil).Once()
				storeMock.On("DeviceResolve", ctx, testifymock.Anything, store.DeviceUIDResolver, uid).Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("InstallKeyResolveSystem", ctx, scope.MustBounded(tenant)).Return(&models.InstallKey{ID: "legacydigest", TenantID: tenant, Type: models.InstallKeyTypeLegacy, Mode: models.InstallKeyModeManual}, nil).Once()
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
				storeMock.On("NamespaceIncrementDeviceCount", ctx, scope.MustBounded(tenant), models.DeviceStatusPending, int64(1)).Return(nil).Once()
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
					TenantID:  tenant,
					Status:    models.DeviceStatusPending,
				},
				err: nil,
			},
		},
	}

	service := NewService(store.Store(storeMock), privateKey, &privateKey.PublicKey, cacheMock)

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
				s.On("InstallKeyResolveSystem", ctx, scope.MustBounded(tenant)).Return(legacy, nil).Once()
			},
			expectedKey: legacy,
			expectedID:  "legacydigest",
		},
		{
			description: "attributes a paired enrollment to the pairing key, not the legacy key",
			paired:      true,
			requiredMocks: func(ctx context.Context, s *mocks.MockStore) {
				s.On("InstallKeyResolveSystemPairing", ctx, scope.MustBounded(tenant)).Return(pairing, nil).Once()
			},
			expectedKey: pairing,
			expectedID:  "pairingdigest",
		},
		{
			description: "attributes nothing when the pairing key is missing",
			paired:      true,
			requiredMocks: func(ctx context.Context, s *mocks.MockStore) {
				s.On("InstallKeyResolveSystemPairing", ctx, scope.MustBounded(tenant)).Return(nil, store.ErrNoDocuments).Once()
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

			svc := NewService(store.Store(storeMock), privateKey, &privateKey.PublicKey, cacheMock)
			key, id, err := svc.enrollmentInstallKey(ctx, scope.MustBounded(tenant), requests.DeviceAuth{TenantID: tenant}, tc.paired)

			require.NoError(tt, err)
			require.Equal(tt, tc.expectedKey, key)
			require.Equal(tt, tc.expectedID, id)
			storeMock.AssertExpectations(tt)
		})
	}
}

// TestAuthDevice_InstallKeyWithoutTenant covers the enrollment of an agent that presents an install
// key as its only credential: the namespace comes from the key, the key's mode still governs the
// resulting status, and a key that resolves to nothing is refused before any namespace is touched.
func TestAuthDevice_InstallKeyWithoutTenant(t *testing.T) {
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

	unbounded := scope.NewUnbounded(reasonInstallKeyTenant)
	keyDigest := hashInstallKey("fleet-key")
	unknownDigest := hashInstallKey("no-such-key")

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
			description: "enrolls into the key's namespace, with the key's mode still governing the status",
			req:         requests.DeviceAuth{Hostname: "fleet-a", Identity: &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:01"}, Info: &requests.DeviceInfo{}, PublicKey: "pk", InstallKey: "fleet-key"},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("fleet-a", "aa:bb:cc:dd:ee:01", "pk")
				key := &models.InstallKey{ID: keyDigest, TenantID: tenant, Type: models.InstallKeyTypeUser, Mode: models.InstallKeyModeManual, Reusable: true}
				storeMock.On("InstallKeyResolve", ctx, unbounded, store.InstallKeyIDResolver, keyDigest).Return(key, nil).Once()
				storeMock.On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, tenant).Return(&models.Namespace{TenantID: tenant, Name: "test"}, nil).Once()
				cacheMock.On("Get", ctx, "auth_device/"+uid, testifymock.Anything).Return(nil).Once()
				storeMock.On("DeviceResolve", ctx, testifymock.Anything, store.DeviceUIDResolver, uid).Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("InstallKeyResolve", ctx, scope.MustBounded(tenant), store.InstallKeyIDResolver, keyDigest).Return(key, nil).Once()
				storeMock.On("DeviceCreate", ctx, &models.Device{
					CreatedAt:       now,
					UID:             uid,
					TenantID:        tenant,
					LastSeen:        now,
					DisconnectedAt:  nil,
					Status:          models.DeviceStatusPending,
					StatusUpdatedAt: now,
					Name:            "fleet-a",
					Identity:        &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:01"},
					PublicKey:       "pk",
					RemoteAddr:      "",
					Taggable:        models.Taggable{TagIDs: []string{}},
					Position:        &models.DevicePosition{},
					Info:            &models.DeviceInfo{},
					InstallKeyID:    keyDigest,
				}).Return(uid, nil).Once()
				storeMock.On("NamespaceIncrementDeviceCount", ctx, scope.MustBounded(tenant), models.DeviceStatusPending, int64(1)).Return(nil).Once()
				storeMock.On("InstallKeyEventCreate", ctx, testifymock.Anything).Return(nil).Once()
				cacheMock.On("Set", ctx, "auth_device/"+uid, map[string]string{"device_name": "fleet-a", "namespace_name": "test"}, time.Second*30).Return(nil).Once()
			},
			expected: Expected{
				res: &models.DeviceAuthResponse{
					UID:       toUID("fleet-a", "aa:bb:cc:dd:ee:01", "pk"),
					Token:     toToken(toUID("fleet-a", "aa:bb:cc:dd:ee:01", "pk")),
					Name:      "fleet-a",
					Namespace: "test",
					TenantID:  tenant,
					Status:    models.DeviceStatusPending,
				},
				err: nil,
			},
		},
		{
			description: "refuses an unresolvable key without resolving a namespace",
			req:         requests.DeviceAuth{Hostname: "fleet-b", Identity: &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:02"}, Info: &requests.DeviceInfo{}, PublicKey: "pk", InstallKey: "no-such-key"},
			requiredMocks: func(ctx context.Context) {
				storeMock.On("InstallKeyResolve", ctx, unbounded, store.InstallKeyIDResolver, unknownDigest).Return(nil, store.ErrNoDocuments).Once()
			},
			expected: Expected{res: nil, err: NewErrAuthInvalid(map[string]any{"install_key": "invalid"}, store.ErrNoDocuments)},
		},
		{
			description: "refuses a system key presented as the only credential",
			req:         requests.DeviceAuth{Hostname: "fleet-c", Identity: &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:03"}, Info: &requests.DeviceInfo{}, PublicKey: "pk", InstallKey: "fleet-key"},
			requiredMocks: func(ctx context.Context) {
				legacy := &models.InstallKey{ID: keyDigest, TenantID: tenant, Type: models.InstallKeyTypeLegacy, Mode: models.InstallKeyModeManual}
				storeMock.On("InstallKeyResolve", ctx, unbounded, store.InstallKeyIDResolver, keyDigest).Return(legacy, nil).Once()
			},
			expected: Expected{res: nil, err: NewErrAuthInvalid(map[string]any{"install_key": "invalid"}, nil)},
		},
	}

	service := NewService(store.Store(storeMock), privateKey, &privateKey.PublicKey, cacheMock)

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

// TestInstallKeyTenant covers the contract that recovering a namespace from a key is not authorizing
// an enrollment: a revoked key still names its namespace, so a device it already enrolled keeps
// reconnecting, while enrollmentInstallKey remains the gate that refuses it a new one.
func TestInstallKeyTenant(t *testing.T) {
	const tenant = "00000000-0000-4000-0000-000000000000"

	unbounded := scope.NewUnbounded(reasonInstallKeyTenant)
	digest := hashInstallKey("fleet-key")

	type Expected struct {
		tenantID string
		err      error
	}

	cases := []struct {
		description   string
		requiredMocks func(context.Context, *mocks.MockStore)
		expected      Expected
	}{
		{
			description: "returns the namespace of a usable key",
			requiredMocks: func(ctx context.Context, storeMock *mocks.MockStore) {
				key := &models.InstallKey{ID: digest, TenantID: tenant, Type: models.InstallKeyTypeUser, Reusable: true}
				storeMock.On("InstallKeyResolve", ctx, unbounded, store.InstallKeyIDResolver, digest).Return(key, nil).Once()
			},
			expected: Expected{tenantID: tenant, err: nil},
		},
		{
			description: "returns the namespace of a revoked key, leaving the refusal to enrollment",
			requiredMocks: func(ctx context.Context, storeMock *mocks.MockStore) {
				key := &models.InstallKey{ID: digest, TenantID: tenant, Type: models.InstallKeyTypeUser, Reusable: true, Revoked: true}
				storeMock.On("InstallKeyResolve", ctx, unbounded, store.InstallKeyIDResolver, digest).Return(key, nil).Once()
			},
			expected: Expected{tenantID: tenant, err: nil},
		},
		{
			description: "refuses a system key",
			requiredMocks: func(ctx context.Context, storeMock *mocks.MockStore) {
				key := &models.InstallKey{ID: digest, TenantID: tenant, Type: models.InstallKeyTypePairing}
				storeMock.On("InstallKeyResolve", ctx, unbounded, store.InstallKeyIDResolver, digest).Return(key, nil).Once()
			},
			expected: Expected{tenantID: "", err: NewErrAuthInvalid(map[string]any{"install_key": "invalid"}, nil)},
		},
		{
			description: "refuses an unresolvable key",
			requiredMocks: func(ctx context.Context, storeMock *mocks.MockStore) {
				storeMock.On("InstallKeyResolve", ctx, unbounded, store.InstallKeyIDResolver, digest).Return(nil, store.ErrNoDocuments).Once()
			},
			expected: Expected{tenantID: "", err: NewErrAuthInvalid(map[string]any{"install_key": "invalid"}, store.ErrNoDocuments)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.TODO()
			storeMock := mocks.NewMockStore(tt)
			cacheMock := mockcache.NewMockCache(tt)
			tc.requiredMocks(ctx, storeMock)

			s := NewService(store.Store(storeMock), privateKey, &privateKey.PublicKey, cacheMock)
			tenantID, err := s.installKeyTenant(ctx, "fleet-key")

			require.Equal(tt, tc.expected.tenantID, tenantID)
			require.Equal(tt, tc.expected.err, err)
			storeMock.AssertExpectations(tt)
		})
	}
}
