package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
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
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuidmock "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthDevice(t *testing.T) {
	storeMock := new(mocks.Store)
	cacheMock := new(mockcache.Cache)
	clockMock := new(clockmock.Clock)
	uuidMock := new(uuidmock.Uuid)

	clock.DefaultBackend = clockMock
	clockMock.On("Now").Return(now)
	uuid.DefaultBackend = uuidMock
	uuidMock.On("Generate").Return("00000000-0000-0000-0000-000000000000")

	toUID := func(tenantID, hostname, mac, publicKey string) string {
		auth := models.DeviceAuth{
			Hostname:  strings.ToLower(hostname),
			Identity:  &models.DeviceIdentity{MAC: mac},
			PublicKey: publicKey,
			TenantID:  tenantID,
		}

		uidSHA := sha256.Sum256(structhash.Dump(auth, 1))

		return hex.EncodeToString(uidSHA[:])
	}

	toToken := func(tenantID, uid string) string {
		token, err := jwttoken.EncodeDeviceClaims(authorizer.DeviceClaims{UID: uid, TenantID: tenantID}, privateKey)
		require.NoError(t, err)

		return token
	}

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
			description: "fails when tenant does not exist",
			req: requests.DeviceAuth{
				TenantID: "00000000-0000-4000-0000-000000000000",
				RealIP:   "127.0.0.1",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(nil, errors.New("error", "store", 0)).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", errors.New("error", "store", 0)),
			},
		},
		{
			description: "fails to authenticate device due to no identity",
			req: requests.DeviceAuth{
				TenantID: "00000000-0000-4000-0000-000000000000",
				Hostname: "",
				Identity: nil,
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", Name: "test"}, nil).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrAuthDeviceNoIdentity(),
			},
		},
		{
			description: "fails to authenticate device due to no identity and hostname",
			req: requests.DeviceAuth{
				TenantID: "00000000-0000-4000-0000-000000000000",
				Hostname: "",
				Identity: &requests.DeviceIdentity{MAC: ""},
				RealIP:   "127.0.0.1",
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", Name: "test"}, nil).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrAuthDeviceNoIdentityAndHostname(),
			},
		},
		{
			description: "fails to resolve the device without ErrNoDocuments error",
			req: requests.DeviceAuth{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Hostname:  "hostname",
				Identity:  &requests.DeviceIdentity{MAC: ""},
				Info:      nil,
				PublicKey: "",
			},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("00000000-0000-4000-0000-000000000000", "hostname", "", "")

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", Name: "test"}, nil).
					Once()
				cacheMock.
					On("Get", ctx, "auth_device/"+uid, testifymock.Anything).
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, uid).
					Return(nil, errors.New("error", "store", 0)).
					Once()
			},
			expected: Expected{
				res: nil,
				err: errors.New("error", "store", 0),
			},
		},
		{
			description: "[device exists] fails when cannot update device with info",
			req: requests.DeviceAuth{
				TenantID: "00000000-0000-4000-0000-000000000000",
				Hostname: "hostname",
				Identity: &requests.DeviceIdentity{MAC: ""},
				Info: &requests.DeviceInfo{
					ID:         "test",
					PrettyName: "Test",
					Version:    "v0.20.0",
					Arch:       "arch64",
					Platform:   "native",
				},
				PublicKey: "",
				RealIP:    "127.0.0.1",
			},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("00000000-0000-4000-0000-000000000000", "hostname", "", "")

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", Name: "test"}, nil).
					Once()
				cacheMock.
					On("Get", ctx, "auth_device/"+uid, testifymock.Anything).
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, uid).
					Return(&models.Device{UID: uid, Name: "hostname"}, nil).
					Once()
				storeMock.
					On("DeviceUpdate", ctx, "00000000-0000-4000-0000-000000000000", uid, &models.DeviceChanges{Info: &models.DeviceInfo{
						ID:         "test",
						PrettyName: "Test",
						Version:    "v0.20.0",
						Arch:       "arch64",
						Platform:   "native",
					}, LastSeen: now, DisconnectedAt: nil}).
					Return(errors.New("error", "store", 0)).
					Once()
			},
			expected: Expected{
				res: nil,
				err: errors.New("error", "store", 0),
			},
		},
		{
			description: "[device exists] fails when cannot set device as online",
			req: requests.DeviceAuth{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Hostname:  "hostname",
				Identity:  &requests.DeviceIdentity{MAC: ""},
				Info:      nil,
				PublicKey: "",
				RealIP:    "127.0.0.1",
			},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("00000000-0000-4000-0000-000000000000", "hostname", "", "")

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", Name: "test"}, nil).
					Once()
				cacheMock.
					On("Get", ctx, "auth_device/"+uid, testifymock.Anything).
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, uid).
					Return(&models.Device{UID: uid, Name: "hostname"}, nil).
					Once()
				storeMock.
					On("DeviceUpdate", ctx, "00000000-0000-4000-0000-000000000000", uid, &models.DeviceChanges{LastSeen: now, DisconnectedAt: nil}).
					Return(errors.New("error", "store", 0)).
					Once()
			},
			expected: Expected{
				res: nil,
				err: errors.New("error", "store", 0),
			},
		},
		{
			description: "[device exists] [without session] succeeds to authenticate device",
			req: requests.DeviceAuth{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Hostname:  "hostname",
				Identity:  &requests.DeviceIdentity{MAC: ""},
				Info:      nil,
				PublicKey: "",
				Sessions:  []string{},
				RealIP:    "127.0.0.1",
			},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("00000000-0000-4000-0000-000000000000", "hostname", "", "")

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", Name: "test"}, nil).
					Once()
				cacheMock.
					On("Get", ctx, "auth_device/"+uid, testifymock.Anything).
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, uid).
					Return(&models.Device{UID: uid, Name: "hostname"}, nil).
					Once()
				storeMock.
					On("DeviceUpdate", ctx, "00000000-0000-4000-0000-000000000000", uid, &models.DeviceChanges{LastSeen: now, DisconnectedAt: nil}).
					Return(nil).
					Once()
				cacheMock.
					On("Set", ctx, "auth_device/"+uid, map[string]string{"device_name": "hostname", "namespace_name": "test"}, time.Second*30).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.DeviceAuthResponse{
					UID:       toUID("00000000-0000-4000-0000-000000000000", "hostname", "", ""),
					Token:     toToken("00000000-0000-4000-0000-000000000000", toUID("00000000-0000-4000-0000-000000000000", "hostname", "", "")),
					Name:      "hostname",
					Namespace: "test",
				},
				err: nil,
			},
		},
		{
			description: "[device exists] succeeds to authenticate device with sessions",
			req: requests.DeviceAuth{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Hostname:  "hostname",
				Identity:  &requests.DeviceIdentity{MAC: ""},
				Info:      nil,
				PublicKey: "",
				Sessions:  []string{"session_1", "session_2"},
				RealIP:    "127.0.0.1",
			},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("00000000-0000-4000-0000-000000000000", "hostname", "", "")

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", Name: "test"}, nil).
					Once()
				cacheMock.
					On("Get", ctx, "auth_device/"+uid, testifymock.Anything).
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, uid).
					Return(&models.Device{UID: uid, Name: "hostname"}, nil).
					Once()
				storeMock.
					On("DeviceUpdate", ctx, "00000000-0000-4000-0000-000000000000", uid, &models.DeviceChanges{LastSeen: now, DisconnectedAt: nil}).
					Return(nil).
					Once()
				storeMock.
					On("SessionSetLastSeen", ctx, models.UID("session_1")).
					Return(nil).
					Once()
				storeMock.
					On("SessionSetLastSeen", ctx, models.UID("session_2")).
					Return(nil).
					Once()
				cacheMock.
					On("Set", ctx, "auth_device/"+uid, map[string]string{"device_name": "hostname", "namespace_name": "test"}, time.Second*30).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.DeviceAuthResponse{
					UID:       toUID("00000000-0000-4000-0000-000000000000", "hostname", "", ""),
					Token:     toToken("00000000-0000-4000-0000-000000000000", toUID("00000000-0000-4000-0000-000000000000", "hostname", "", "")),
					Name:      "hostname",
					Namespace: "test",
				},
				err: nil,
			},
		},
		{
			description: "[device exists] succeeds to authenticate device with info",
			req: requests.DeviceAuth{
				TenantID: "00000000-0000-4000-0000-000000000000",
				Hostname: "hostname",
				Identity: &requests.DeviceIdentity{MAC: ""},
				Info: &requests.DeviceInfo{
					ID:         "test",
					PrettyName: "Test",
					Version:    "v0.20.0",
					Arch:       "arch64",
					Platform:   "native",
				},
				PublicKey: "",
				Sessions:  []string{},
				RealIP:    "127.0.0.1",
			},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("00000000-0000-4000-0000-000000000000", "hostname", "", "")

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", Name: "test"}, nil).
					Once()
				cacheMock.
					On("Get", ctx, "auth_device/"+uid, testifymock.Anything).
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, uid).
					Return(&models.Device{UID: uid, Name: "hostname"}, nil).
					Once()
				storeMock.
					On("DeviceUpdate", ctx, "00000000-0000-4000-0000-000000000000", uid, &models.DeviceChanges{Info: &models.DeviceInfo{
						ID:         "test",
						PrettyName: "Test",
						Version:    "v0.20.0",
						Arch:       "arch64",
						Platform:   "native",
					}, LastSeen: now, DisconnectedAt: nil}).
					Return(nil).
					Once()
				cacheMock.
					On("Set", ctx, "auth_device/"+uid, map[string]string{"device_name": "hostname", "namespace_name": "test"}, time.Second*30).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.DeviceAuthResponse{
					UID:       toUID("00000000-0000-4000-0000-000000000000", "hostname", "", ""),
					Token:     toToken("00000000-0000-4000-0000-000000000000", toUID("00000000-0000-4000-0000-000000000000", "hostname", "", "")),
					Name:      "hostname",
					Namespace: "test",
				},
				err: nil,
			},
		},
		{
			description: "[device creation] fails when device creation fails",
			req: requests.DeviceAuth{
				TenantID: "00000000-0000-4000-0000-000000000000",
				Hostname: "new-device",
				Identity: &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
				Info: &requests.DeviceInfo{
					ID:         "device-id",
					PrettyName: "My Device",
					Version:    "1.0.0",
					Arch:       "x86_64",
					Platform:   "linux",
				},
				PublicKey: "public-key",
				Sessions:  []string{},
				RealIP:    "127.0.0.1",
			},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("00000000-0000-4000-0000-000000000000", "new-device", "aa:bb:cc:dd:ee:ff", "public-key")

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", Name: "test"}, nil).
					Once()
				cacheMock.
					On("Get", ctx, "auth_device/"+uid, testifymock.Anything).
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, uid).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On(
						"DeviceCreate",
						ctx,
						&models.Device{
							CreatedAt:       now,
							UID:             uid,
							TenantID:        "00000000-0000-4000-0000-000000000000",
							LastSeen:        now,
							DisconnectedAt:  nil,
							Status:          models.DeviceStatusPending,
							StatusUpdatedAt: now,
							Name:            "new-device",
							Identity:        &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
							PublicKey:       "public-key",
							RemoteAddr:      "127.0.0.1",
							Tags:            []string{},
							Position:        &models.DevicePosition{Longitude: 0., Latitude: 0.},
							Info: &models.DeviceInfo{
								ID:         "device-id",
								PrettyName: "My Device",
								Version:    "1.0.0",
								Arch:       "x86_64",
								Platform:   "linux",
							},
						},
					).
					Return("", errors.New("database error", "store", 0)).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrDeviceCreate(models.Device{}, errors.New("database error", "store", 0)),
			},
		},
		{
			description: "[device creation] fails when namespace increment fails",
			req: requests.DeviceAuth{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Hostname:  "new-device",
				Identity:  &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
				PublicKey: "public-key",
				Info: &requests.DeviceInfo{
					ID:         "device-id",
					PrettyName: "My Device",
					Version:    "1.0.0",
					Arch:       "x86_64",
					Platform:   "linux",
				},
				Sessions: []string{},
				RealIP:   "127.0.0.1",
			},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("00000000-0000-4000-0000-000000000000", "new-device", "aa:bb:cc:dd:ee:ff", "public-key")

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", Name: "test"}, nil).
					Once()
				cacheMock.
					On("Get", ctx, "auth_device/"+uid, testifymock.Anything).
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, uid).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On(
						"DeviceCreate",
						ctx,
						&models.Device{
							CreatedAt:       now,
							UID:             uid,
							TenantID:        "00000000-0000-4000-0000-000000000000",
							LastSeen:        now,
							DisconnectedAt:  nil,
							Status:          models.DeviceStatusPending,
							StatusUpdatedAt: now,
							Name:            "new-device",
							Identity:        &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
							PublicKey:       "public-key",
							RemoteAddr:      "127.0.0.1",
							Tags:            []string{},
							Position:        &models.DevicePosition{Longitude: 0., Latitude: 0.},
							Info: &models.DeviceInfo{
								ID:         "device-id",
								PrettyName: "My Device",
								Version:    "1.0.0",
								Arch:       "x86_64",
								Platform:   "linux",
							},
						},
					).
					Return(uid, nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-4000-0000-000000000000", models.DeviceStatusPending, int64(1)).
					Return(errors.New("increment error", "store", 0)).
					Once()
			},
			expected: Expected{
				res: nil,
				err: errors.New("increment error", "store", 0),
			},
		},
		{
			description: "[device creation] succeeds to create and authenticate new device",
			req: requests.DeviceAuth{
				TenantID: "00000000-0000-4000-0000-000000000000",
				Hostname: "new-device",
				Identity: &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
				Info: &requests.DeviceInfo{
					ID:         "device-id",
					PrettyName: "My Device",
					Version:    "1.0.0",
					Arch:       "x86_64",
					Platform:   "linux",
				},
				PublicKey: "public-key",
				Sessions:  []string{"session_1", "session_2"},
				RealIP:    "127.0.0.1",
			},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("00000000-0000-4000-0000-000000000000", "new-device", "aa:bb:cc:dd:ee:ff", "public-key")

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", Name: "test"}, nil).
					Once()
				cacheMock.
					On("Get", ctx, "auth_device/"+uid, testifymock.Anything).
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, uid).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On(
						"DeviceCreate",
						ctx,
						&models.Device{
							CreatedAt:       now,
							UID:             uid,
							TenantID:        "00000000-0000-4000-0000-000000000000",
							LastSeen:        now,
							DisconnectedAt:  nil,
							Status:          models.DeviceStatusPending,
							StatusUpdatedAt: now,
							Name:            "new-device",
							Identity:        &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
							PublicKey:       "public-key",
							RemoteAddr:      "127.0.0.1",
							Tags:            []string{},
							Position:        &models.DevicePosition{Longitude: 0., Latitude: 0.},
							Info: &models.DeviceInfo{
								ID:         "device-id",
								PrettyName: "My Device",
								Version:    "1.0.0",
								Arch:       "x86_64",
								Platform:   "linux",
							},
						},
					).
					Return(uid, nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-4000-0000-000000000000", models.DeviceStatusPending, int64(1)).
					Return(nil).
					Once()
				storeMock.
					On("SessionSetLastSeen", ctx, models.UID("session_1")).
					Return(nil).
					Once()
				storeMock.
					On("SessionSetLastSeen", ctx, models.UID("session_2")).
					Return(nil).
					Once()
				cacheMock.
					On("Set", ctx, "auth_device/"+uid, map[string]string{"device_name": "new-device", "namespace_name": "test"}, time.Second*30).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.DeviceAuthResponse{
					UID:       toUID("00000000-0000-4000-0000-000000000000", "new-device", "aa:bb:cc:dd:ee:ff", "public-key"),
					Token:     toToken("00000000-0000-4000-0000-000000000000", toUID("00000000-0000-4000-0000-000000000000", "new-device", "aa:bb:cc:dd:ee:ff", "public-key")),
					Name:      "new-device",
					Namespace: "test",
				},
				err: nil,
			},
		},
		{
			description: "[device creation] succeeds to create and authenticate new device with sessions",
			req: requests.DeviceAuth{
				TenantID: "00000000-0000-4000-0000-000000000000",
				Hostname: "new-device",
				Identity: &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
				Info: &requests.DeviceInfo{
					ID:         "device-id",
					PrettyName: "My Device",
					Version:    "1.0.0",
					Arch:       "x86_64",
					Platform:   "linux",
				},
				PublicKey: "public-key",
				Sessions:  []string{},
				RealIP:    "127.0.0.1",
			},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("00000000-0000-4000-0000-000000000000", "new-device", "aa:bb:cc:dd:ee:ff", "public-key")

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", Name: "test"}, nil).
					Once()
				cacheMock.
					On("Get", ctx, "auth_device/"+uid, testifymock.Anything).
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, uid).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On(
						"DeviceCreate",
						ctx,
						&models.Device{
							CreatedAt:       now,
							UID:             uid,
							TenantID:        "00000000-0000-4000-0000-000000000000",
							LastSeen:        now,
							DisconnectedAt:  nil,
							Status:          models.DeviceStatusPending,
							StatusUpdatedAt: now,
							Name:            "new-device",
							Identity:        &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
							PublicKey:       "public-key",
							RemoteAddr:      "127.0.0.1",
							Tags:            []string{},
							Position:        &models.DevicePosition{Longitude: 0., Latitude: 0.},
							Info: &models.DeviceInfo{
								ID:         "device-id",
								PrettyName: "My Device",
								Version:    "1.0.0",
								Arch:       "x86_64",
								Platform:   "linux",
							},
						},
					).
					Return(uid, nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-4000-0000-000000000000", models.DeviceStatusPending, int64(1)).
					Return(nil).
					Once()
				cacheMock.
					On("Set", ctx, "auth_device/"+uid, map[string]string{"device_name": "new-device", "namespace_name": "test"}, time.Second*30).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.DeviceAuthResponse{
					UID:       toUID("00000000-0000-4000-0000-000000000000", "new-device", "aa:bb:cc:dd:ee:ff", "public-key"),
					Token:     toToken("00000000-0000-4000-0000-000000000000", toUID("00000000-0000-4000-0000-000000000000", "new-device", "aa:bb:cc:dd:ee:ff", "public-key")),
					Name:      "new-device",
					Namespace: "test",
				},
				err: nil,
			},
		},
		{
			description: "[device creation] succeeds when hostname is derived from MAC",
			req: requests.DeviceAuth{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Hostname:  "",
				Identity:  &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
				PublicKey: "public-key",
				Info: &requests.DeviceInfo{
					ID:         "device-id",
					PrettyName: "My Device",
					Version:    "1.0.0",
					Arch:       "x86_64",
					Platform:   "linux",
				},
				Sessions: []string{},
				RealIP:   "127.0.0.1",
			},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("00000000-0000-4000-0000-000000000000", "aa-bb-cc-dd-ee-ff", "aa:bb:cc:dd:ee:ff", "public-key")

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", Name: "test"}, nil).
					Once()
				cacheMock.
					On("Get", ctx, "auth_device/"+uid, testifymock.Anything).
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, uid).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On(
						"DeviceCreate",
						ctx,
						&models.Device{
							CreatedAt:       now,
							UID:             uid,
							TenantID:        "00000000-0000-4000-0000-000000000000",
							LastSeen:        now,
							DisconnectedAt:  nil,
							Status:          models.DeviceStatusPending,
							StatusUpdatedAt: now,
							Name:            "aa-bb-cc-dd-ee-ff",
							Identity:        &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
							PublicKey:       "public-key",
							RemoteAddr:      "127.0.0.1",
							Tags:            []string{},
							Position:        &models.DevicePosition{Longitude: 0., Latitude: 0.},
							Info: &models.DeviceInfo{
								ID:         "device-id",
								PrettyName: "My Device",
								Version:    "1.0.0",
								Arch:       "x86_64",
								Platform:   "linux",
							},
						},
					).
					Return(uid, nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-4000-0000-000000000000", models.DeviceStatusPending, int64(1)).
					Return(nil).
					Once()
				cacheMock.
					On("Set", ctx, "auth_device/"+uid, map[string]string{"device_name": "aa-bb-cc-dd-ee-ff", "namespace_name": "test"}, time.Second*30).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.DeviceAuthResponse{
					UID:       toUID("00000000-0000-4000-0000-000000000000", "aa-bb-cc-dd-ee-ff", "aa:bb:cc:dd:ee:ff", "public-key"),
					Token:     toToken("00000000-0000-4000-0000-000000000000", toUID("00000000-0000-4000-0000-000000000000", "aa-bb-cc-dd-ee-ff", "aa:bb:cc:dd:ee:ff", "public-key")),
					Name:      "aa-bb-cc-dd-ee-ff",
					Namespace: "test",
				},
				err: nil,
			},
		},
		{
			description: "[device creation] succeeds to create and authenticate new device with null information",
			req: requests.DeviceAuth{
				TenantID:  "00000000-0000-4000-0000-000000000000",
				Hostname:  "new-device",
				Identity:  &requests.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
				Info:      nil,
				PublicKey: "public-key",
				Sessions:  []string{"session_1", "session_2"},
				RealIP:    "127.0.0.1",
			},
			requiredMocks: func(ctx context.Context) {
				uid := toUID("00000000-0000-4000-0000-000000000000", "new-device", "aa:bb:cc:dd:ee:ff", "public-key")

				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", Name: "test"}, nil).
					Once()
				cacheMock.
					On("Get", ctx, "auth_device/"+uid, testifymock.Anything).
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, uid).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On(
						"DeviceCreate",
						ctx,
						&models.Device{
							CreatedAt:       now,
							UID:             uid,
							TenantID:        "00000000-0000-4000-0000-000000000000",
							LastSeen:        now,
							DisconnectedAt:  nil,
							Status:          models.DeviceStatusPending,
							StatusUpdatedAt: now,
							Name:            "new-device",
							Identity:        &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
							PublicKey:       "public-key",
							RemoteAddr:      "127.0.0.1",
							Tags:            []string{},
							Position:        &models.DevicePosition{Longitude: 0., Latitude: 0.},
							Info:            nil,
						},
					).
					Return(uid, nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-4000-0000-000000000000", models.DeviceStatusPending, int64(1)).
					Return(nil).
					Once()
				storeMock.
					On("SessionSetLastSeen", ctx, models.UID("session_1")).
					Return(nil).
					Once()
				storeMock.
					On("SessionSetLastSeen", ctx, models.UID("session_2")).
					Return(nil).
					Once()
				cacheMock.
					On("Set", ctx, "auth_device/"+uid, map[string]string{"device_name": "new-device", "namespace_name": "test"}, time.Second*30).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.DeviceAuthResponse{
					UID:       toUID("00000000-0000-4000-0000-000000000000", "new-device", "aa:bb:cc:dd:ee:ff", "public-key"),
					Token:     toToken("00000000-0000-4000-0000-000000000000", toUID("00000000-0000-4000-0000-000000000000", "new-device", "aa:bb:cc:dd:ee:ff", "public-key")),
					Name:      "new-device",
					Namespace: "test",
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

			authRes, err := service.AuthDevice(ctx, tc.req)
			require.Equal(tt, tc.expected.res, authRes)
			require.Equal(tt, tc.expected.err, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_AuthLocalUser(t *testing.T) {
	mock := new(mocks.Store)
	cacheMock := new(mockcache.Cache)

	ctx := context.TODO()

	type Expected struct {
		res      *models.UserAuthResponse
		lockout  int64
		mfaToken string
		err      error
	}

	tests := []struct {
		description   string
		req           *requests.AuthLocalUser
		sourceIP      string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when could not retrieve the system",
			sourceIP:    "127.0.0.1",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				mock.
					On("SystemGet", ctx).
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{
				res:      nil,
				lockout:  0,
				mfaToken: "",
				err:      NewErrAuthMethodNotAllowed(models.UserAuthMethodLocal.String()),
			},
		},
		{
			description: "fails when local authentication is not allowed",
			sourceIP:    "127.0.0.1",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				mock.
					On("SystemGet", ctx).
					Return(
						&models.System{
							Authentication: &models.SystemAuthentication{
								Local: &models.SystemAuthenticationLocal{
									Enabled: false,
								},
							},
						},
						nil,
					).
					Once()
			},
			expected: Expected{
				res:      nil,
				lockout:  0,
				mfaToken: "",
				err:      NewErrAuthMethodNotAllowed(models.UserAuthMethodLocal.String()),
			},
		},
		{
			description: "fails when username is not found",
			sourceIP:    "127.0.0.1",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				mock.
					On("SystemGet", ctx).
					Return(
						&models.System{
							Authentication: &models.SystemAuthentication{
								Local: &models.SystemAuthenticationLocal{
									Enabled: true,
								},
							},
						},
						nil,
					).
					Once()
				mock.
					On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{
				res:      nil,
				lockout:  0,
				mfaToken: "",
				err:      NewErrAuthUnathorized(nil),
			},
		},
		{
			description: "fails when email is not found",
			sourceIP:    "127.0.0.1",
			req: &requests.AuthLocalUser{
				Identifier: "john.doe@test.com",
				Password:   "secret",
			},
			requiredMocks: func() {
				mock.
					On("SystemGet", ctx).
					Return(
						&models.System{
							Authentication: &models.SystemAuthentication{
								Local: &models.SystemAuthenticationLocal{
									Enabled: true,
								},
							},
						},
						nil,
					).
					Once()
				mock.
					On("UserResolve", ctx, store.UserEmailResolver, "john.doe@test.com").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{
				res:      nil,
				lockout:  0,
				mfaToken: "",
				err:      NewErrAuthUnathorized(nil),
			},
		},
		{
			description: "fails when user does not have local as authentication method",
			sourceIP:    "127.0.0.1",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				mock.
					On("SystemGet", ctx).
					Return(
						&models.System{
							Authentication: &models.SystemAuthentication{
								Local: &models.SystemAuthenticationLocal{
									Enabled: true,
								},
							},
						},
						nil,
					).
					Once()
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Origin:    models.UserOriginLocal,
					Status:    models.UserStatusNotConfirmed,
					LastLogin: now,
					MFA: models.UserMFA{
						Enabled: false,
					},
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
					},
					Password: models.UserPassword{
						Hash: "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b",
					},
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
						AuthMethods:        []models.UserAuthMethod{},
					},
				}

				mock.
					On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").
					Return(user, nil).
					Once()
			},
			expected: Expected{
				res:      nil,
				lockout:  0,
				mfaToken: "",
				err:      NewErrAuthUnathorized(nil),
			},
		},
		{
			description: "fails when user is not confirmed",
			sourceIP:    "127.0.0.1",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				mock.
					On("SystemGet", ctx).
					Return(
						&models.System{
							Authentication: &models.SystemAuthentication{
								Local: &models.SystemAuthenticationLocal{
									Enabled: true,
								},
							},
						},
						nil,
					).
					Once()
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Origin:    models.UserOriginLocal,
					Status:    models.UserStatusNotConfirmed,
					LastLogin: now,
					MFA: models.UserMFA{
						Enabled: false,
					},
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
					},
					Password: models.UserPassword{
						Hash: "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b",
					},
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
						AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
				}

				mock.
					On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").
					Return(user, nil).
					Once()
			},
			expected: Expected{
				res:      nil,
				lockout:  0,
				mfaToken: "",
				err:      NewErrUserNotConfirmed(nil),
			},
		},
		{
			description: "fails when user is 'invited'",
			sourceIP:    "127.0.0.1",
			req: &requests.AuthLocalUser{
				Identifier: "john.doe@test.com",
				Password:   "secret",
			},
			requiredMocks: func() {
				mock.
					On("SystemGet", ctx).
					Return(
						&models.System{
							Authentication: &models.SystemAuthentication{
								Local: &models.SystemAuthenticationLocal{
									Enabled: true,
								},
							},
						},
						nil,
					).
					Once()
				mock.
					On("UserResolve", ctx, store.UserEmailResolver, "john.doe@test.com").
					Return(
						&models.User{
							ID:        "65fdd16b5f62f93184ec8a39",
							Origin:    models.UserOriginLocal,
							Status:    models.UserStatusInvited,
							LastLogin: now,
							MFA: models.UserMFA{
								Enabled: false,
							},
							UserData: models.UserData{
								Username: "john_doe",
								Email:    "john.doe@test.com",
							},
							Password: models.UserPassword{
								Hash: "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b",
							},
							Preferences: models.UserPreferences{
								PreferredNamespace: "",
								AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
							},
						},
						nil,
					).
					Once()
			},
			expected: Expected{
				res:      nil,
				lockout:  0,
				mfaToken: "",
				err:      NewErrAuthUnathorized(nil),
			},
		},
		{
			description: "fails when an account lockout occurs",
			sourceIP:    "127.0.0.1",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "wrong_password",
			},
			requiredMocks: func() {
				mock.
					On("SystemGet", ctx).
					Return(
						&models.System{
							Authentication: &models.SystemAuthentication{
								Local: &models.SystemAuthenticationLocal{
									Enabled: true,
								},
							},
						},
						nil,
					).
					Once()
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Origin:    models.UserOriginLocal,
					Status:    models.UserStatusConfirmed,
					LastLogin: now,
					MFA: models.UserMFA{
						Enabled: false,
					},
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
						AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
				}

				mock.
					On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").
					Return(user, nil).
					Once()
				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(1711510689), 3, nil).
					Once()
			},
			expected: Expected{
				res:      nil,
				lockout:  1711510689,
				mfaToken: "",
				err:      NewErrAuthUnathorized(nil),
			},
		},
		{
			description: "fails when input password is wrong",
			sourceIP:    "127.0.0.1",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "wrong_password",
			},
			requiredMocks: func() {
				mock.
					On("SystemGet", ctx).
					Return(
						&models.System{
							Authentication: &models.SystemAuthentication{
								Local: &models.SystemAuthenticationLocal{
									Enabled: true,
								},
							},
						},
						nil,
					).
					Once()
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Origin:    models.UserOriginLocal,
					Status:    models.UserStatusConfirmed,
					LastLogin: now,
					MFA: models.UserMFA{
						Enabled: false,
					},
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
						AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
				}

				mock.
					On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").
					Return(user, nil).
					Once()
				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()
				hashMock.
					On("CompareWith", "wrong_password", "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(false).
					Once()
				cacheMock.
					On("StoreLoginAttempt", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(1711510689), 3, nil).
					Once()
			},
			expected: Expected{
				res:      nil,
				lockout:  1711510689,
				mfaToken: "",
				err:      NewErrAuthUnathorized(nil),
			},
		},
		{
			description: "fails when user has MFA enable",
			sourceIP:    "127.0.0.1",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				mock.
					On("SystemGet", ctx).
					Return(
						&models.System{
							Authentication: &models.SystemAuthentication{
								Local: &models.SystemAuthenticationLocal{
									Enabled: true,
								},
							},
						},
						nil,
					).
					Once()
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Origin:    models.UserOriginLocal,
					Status:    models.UserStatusConfirmed,
					LastLogin: now,
					MFA: models.UserMFA{
						Enabled: true,
					},
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
						AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
				}

				mock.
					On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").
					Return(user, nil).
					Once()
				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()
				hashMock.
					On("CompareWith", "secret", "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(true).
					Once()
				cacheMock.
					On("ResetLoginAttempts", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(nil).
					Once()
				uuidMock := &uuidmock.Uuid{}
				uuid.DefaultBackend = uuidMock
				uuidMock.
					On("Generate").
					Return("00000000-0000-4000-0000-000000000000")
				cacheMock.
					On("Set", ctx, "mfa-token={00000000-0000-4000-0000-000000000000}", "65fdd16b5f62f93184ec8a39", 30*time.Minute).
					Return(nil).
					Once()
			},
			expected: Expected{
				res:      nil,
				lockout:  0,
				mfaToken: "00000000-0000-4000-0000-000000000000",
				err:      nil,
			},
		},
		{
			description: "fails when can not update the last_login field",
			sourceIP:    "127.0.0.1",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				mock.
					On("SystemGet", ctx).
					Return(
						&models.System{
							Authentication: &models.SystemAuthentication{
								Local: &models.SystemAuthenticationLocal{
									Enabled: true,
								},
							},
						},
						nil,
					).
					Once()
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Origin:    models.UserOriginLocal,
					Status:    models.UserStatusConfirmed,
					LastLogin: now,
					MFA: models.UserMFA{
						Enabled: false,
					},
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
						AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
				}

				mock.
					On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").
					Return(user, nil).
					Once()
				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()
				hashMock.
					On("CompareWith", "secret", "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(true).
					Once()
				cacheMock.
					On("ResetLoginAttempts", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(nil).
					Once()
				mock.
					On("NamespaceGetPreferred", ctx, "65fdd16b5f62f93184ec8a39").
					Return(nil, errors.New("error", "layer", 0)).
					Once()

				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)

				cacheMock.
					On("Set", ctx, "token_65fdd16b5f62f93184ec8a39", testifymock.Anything, time.Hour*72).
					Return(nil).
					Once()

				preferredNamespace := ""
				mock.
					On("UserUpdate", ctx, user.ID, &models.UserChanges{LastLogin: now, PreferredNamespace: &preferredNamespace}).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				res:      nil,
				lockout:  0,
				mfaToken: "",
				err: NewErrUserUpdate(
					&models.User{
						ID:        "65fdd16b5f62f93184ec8a39",
						Origin:    models.UserOriginLocal,
						Status:    models.UserStatusConfirmed,
						LastLogin: now,
						UserData: models.UserData{
							Username: "john_doe",
							Email:    "john.doe@test.com",
						},
						Password: models.UserPassword{
							Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
						},
						Preferences: models.UserPreferences{
							AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
						},
					},
					errors.New("error", "", 0),
				),
			},
		},
		{
			description: "succeeds to authenticate without a namespace",
			sourceIP:    "127.0.0.1",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				mock.
					On("SystemGet", ctx).
					Return(
						&models.System{
							Authentication: &models.SystemAuthentication{
								Local: &models.SystemAuthenticationLocal{
									Enabled: true,
								},
							},
						},
						nil,
					).
					Once()
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Origin:    models.UserOriginLocal,
					Status:    models.UserStatusConfirmed,
					LastLogin: now,
					MFA: models.UserMFA{
						Enabled: false,
					},
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
						Name:     "john doe",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
						AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
				}

				mock.
					On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").
					Return(user, nil).
					Once()
				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()
				hashMock.
					On("CompareWith", "secret", "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(true).
					Once()
				cacheMock.
					On("ResetLoginAttempts", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(nil).
					Once()
				mock.
					On("NamespaceGetPreferred", ctx, "65fdd16b5f62f93184ec8a39").
					Return(nil, errors.New("error", "layer", 0)).
					Once()

				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)

				cacheMock.
					On("Set", ctx, "token_65fdd16b5f62f93184ec8a39", testifymock.Anything, time.Hour*72).
					Return(nil).
					Once()

				preferredNamespace := ""
				mock.
					On("UserUpdate", ctx, user.ID, &models.UserChanges{LastLogin: now, PreferredNamespace: &preferredNamespace}).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.UserAuthResponse{
					ID:          "65fdd16b5f62f93184ec8a39",
					Origin:      models.UserOriginLocal.String(),
					AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
					Name:        "john doe",
					User:        "john_doe",
					Email:       "john.doe@test.com",
					Tenant:      "",
					Token:       "must ignore",
				},
				lockout:  0,
				mfaToken: "",
				err:      nil,
			},
		},
		{
			description: "succeeds to authenticate with a namespace",
			sourceIP:    "127.0.0.1",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				mock.
					On("SystemGet", ctx).
					Return(
						&models.System{
							Authentication: &models.SystemAuthentication{
								Local: &models.SystemAuthenticationLocal{
									Enabled: true,
								},
							},
						},
						nil,
					).
					Once()
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Origin:    models.UserOriginLocal,
					Status:    models.UserStatusConfirmed,
					LastLogin: now,
					MFA: models.UserMFA{
						Enabled: false,
					},
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
						Name:     "john doe",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
					Preferences: models.UserPreferences{
						PreferredNamespace: "00000000-0000-4000-0000-000000000000",
						AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
				}

				mock.
					On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").
					Return(user, nil).
					Once()
				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()
				hashMock.
					On("CompareWith", "secret", "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(true).
					Once()
				cacheMock.
					On("ResetLoginAttempts", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(nil).
					Once()

				ns := &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Members: []models.Member{
						{
							ID:   "65fdd16b5f62f93184ec8a39",
							Role: "owner",
						},
					},
				}

				mock.
					On("NamespaceGetPreferred", ctx, "65fdd16b5f62f93184ec8a39").
					Return(ns, nil).
					Once()

				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)

				cacheMock.
					On("Set", ctx, "token_00000000-0000-4000-0000-00000000000065fdd16b5f62f93184ec8a39", testifymock.Anything, time.Hour*72).
					Return(nil).
					Once()

				preferredNamespace := "00000000-0000-4000-0000-000000000000"
				mock.
					On("UserUpdate", ctx, user.ID, &models.UserChanges{LastLogin: now, PreferredNamespace: &preferredNamespace}).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.UserAuthResponse{
					ID:          "65fdd16b5f62f93184ec8a39",
					Origin:      models.UserOriginLocal.String(),
					AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
					Name:        "john doe",
					User:        "john_doe",
					Email:       "john.doe@test.com",
					Tenant:      "00000000-0000-4000-0000-000000000000",
					Role:        "owner",
					Token:       "must ignore",
				},
				lockout:  0,
				mfaToken: "",
				err:      nil,
			},
		},
		{
			description: "succeeds to authenticate with a namespace (and member status 'pending')",
			sourceIP:    "127.0.0.1",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				mock.
					On("SystemGet", ctx).
					Return(
						&models.System{
							Authentication: &models.SystemAuthentication{
								Local: &models.SystemAuthenticationLocal{
									Enabled: true,
								},
							},
						},
						nil,
					).
					Once()
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Origin:    models.UserOriginLocal,
					Status:    models.UserStatusConfirmed,
					LastLogin: now,
					MFA: models.UserMFA{
						Enabled: false,
					},
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
						Name:     "john doe",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
					Preferences: models.UserPreferences{
						PreferredNamespace: "00000000-0000-4000-0000-000000000000",
						AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
				}

				mock.
					On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").
					Return(user, nil).
					Once()
				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()
				hashMock.
					On("CompareWith", "secret", "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(true).
					Once()
				cacheMock.
					On("ResetLoginAttempts", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(nil).
					Once()

				ns := &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Members: []models.Member{
						{
							ID:     "65fdd16b5f62f93184ec8a39",
							Role:   "owner",
							Status: models.MemberStatusPending,
						},
					},
				}

				mock.
					On("NamespaceGetPreferred", ctx, "65fdd16b5f62f93184ec8a39").
					Return(ns, nil).
					Once()

				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)

				cacheMock.
					On("Set", ctx, "token_65fdd16b5f62f93184ec8a39", testifymock.Anything, time.Hour*72).
					Return(nil).
					Once()

				preferredNamespace := ""
				mock.
					On("UserUpdate", ctx, user.ID, &models.UserChanges{LastLogin: now, PreferredNamespace: &preferredNamespace}).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.UserAuthResponse{
					ID:          "65fdd16b5f62f93184ec8a39",
					Origin:      models.UserOriginLocal.String(),
					AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
					Name:        "john doe",
					User:        "john_doe",
					Email:       "john.doe@test.com",
					Tenant:      "",
					Role:        "",
					Token:       "must ignore",
				},
				lockout:  0,
				mfaToken: "",
				err:      nil,
			},
		},
		{
			description: "succeeds to authenticate with a namespace (and empty preferred namespace)",
			sourceIP:    "127.0.0.1",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				mock.
					On("SystemGet", ctx).
					Return(
						&models.System{
							Authentication: &models.SystemAuthentication{
								Local: &models.SystemAuthenticationLocal{
									Enabled: true,
								},
							},
						},
						nil,
					).
					Once()
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Origin:    models.UserOriginLocal,
					Status:    models.UserStatusConfirmed,
					LastLogin: now,
					MFA: models.UserMFA{
						Enabled: false,
					},
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
						Name:     "john doe",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
						AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
				}

				mock.
					On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").
					Return(user, nil).
					Once()
				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()
				hashMock.
					On("CompareWith", "secret", "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(true).
					Once()
				cacheMock.
					On("ResetLoginAttempts", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(nil).
					Once()

				ns := &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Members: []models.Member{
						{
							ID:   "65fdd16b5f62f93184ec8a39",
							Role: "owner",
						},
					},
				}

				mock.
					On("NamespaceGetPreferred", ctx, "65fdd16b5f62f93184ec8a39").
					Return(ns, nil).
					Once()

				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)

				cacheMock.
					On("Set", ctx, "token_00000000-0000-4000-0000-00000000000065fdd16b5f62f93184ec8a39", testifymock.Anything, time.Hour*72).
					Return(nil).
					Once()

				preferredNamespace := "00000000-0000-4000-0000-000000000000"
				mock.
					On("UserUpdate", ctx, user.ID, &models.UserChanges{LastLogin: now, PreferredNamespace: &preferredNamespace}).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.UserAuthResponse{
					ID:          "65fdd16b5f62f93184ec8a39",
					Origin:      models.UserOriginLocal.String(),
					AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
					Name:        "john doe",
					User:        "john_doe",
					Email:       "john.doe@test.com",
					Tenant:      "00000000-0000-4000-0000-000000000000",
					Role:        "owner",
					Token:       "must ignore",
				},
				lockout:  0,
				mfaToken: "",
				err:      nil,
			},
		},
		{
			description: "succeeds to authenticate and update non-bcypt hashes",
			sourceIP:    "127.0.0.1",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				mock.
					On("SystemGet", ctx).
					Return(
						&models.System{
							Authentication: &models.SystemAuthentication{
								Local: &models.SystemAuthenticationLocal{
									Enabled: true,
								},
							},
						},
						nil,
					).
					Once()
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Origin:    models.UserOriginLocal,
					Status:    models.UserStatusConfirmed,
					LastLogin: now,
					MFA: models.UserMFA{
						Enabled: false,
					},
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
						Name:     "john doe",
					},
					Password: models.UserPassword{
						Hash: "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b",
					},
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
						AuthMethods:        []models.UserAuthMethod{models.UserAuthMethodLocal},
					},
				}

				mock.
					On("UserResolve", ctx, store.UserUsernameResolver, "john_doe").
					Return(user, nil).
					Once()
				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()
				hashMock.
					On("CompareWith", "secret", "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b").
					Return(true).
					Once()
				cacheMock.
					On("ResetLoginAttempts", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(nil).
					Once()
				mock.
					On("NamespaceGetPreferred", ctx, "65fdd16b5f62f93184ec8a39").
					Return(nil, errors.New("error", "layer", 0)).
					Once()

				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)

				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()

				cacheMock.
					On("Set", ctx, "token_65fdd16b5f62f93184ec8a39", testifymock.Anything, time.Hour*72).
					Return(nil).
					Once()
				hashMock.
					On("Do", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi", nil).
					Once()

				preferredNamespace := ""
				mock.
					On("UserUpdate", ctx, user.ID, &models.UserChanges{LastLogin: now, PreferredNamespace: &preferredNamespace, Password: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi"}).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.UserAuthResponse{
					ID:          "65fdd16b5f62f93184ec8a39",
					Origin:      models.UserOriginLocal.String(),
					AuthMethods: []models.UserAuthMethod{models.UserAuthMethodLocal},
					Name:        "john doe",
					User:        "john_doe",
					Email:       "john.doe@test.com",
					Tenant:      "",
					Token:       "must ignore",
				},
				lockout:  0,
				mfaToken: "",
				err:      nil,
			},
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	service := NewService(store.Store(mock), privateKey, &privateKey.PublicKey, cacheMock, clientMock)

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			res, lockout, mfaToken, err := service.AuthLocalUser(ctx, tc.req, tc.sourceIP)
			// Since the resulting token is not crucial for the assertion and
			// difficult to mock, it is safe to ignore this field.
			if res != nil {
				res.Token = "must ignore"
			}

			assert.Equal(t, tc.expected, Expected{res, lockout, mfaToken, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestCreateUserToken(t *testing.T) {
	storeMock := new(mocks.Store)
	cacheMock := new(mockcache.Cache)

	type Expected struct {
		res *models.UserAuthResponse
		err error
	}

	tests := []struct {
		description   string
		req           *requests.CreateUserToken
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails when user is not found",
			req:         &requests.CreateUserToken{UserID: "000000000000000000000000", TenantID: "00000000-0000-4000-0000-000000000000"},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrUserNotFound("000000000000000000000000", store.ErrNoDocuments),
			},
		},
		{
			description: "[with-tenant] fails when namespace is not found",
			req:         &requests.CreateUserToken{UserID: "000000000000000000000000", TenantID: "00000000-0000-4000-0000-000000000000"},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(
						&models.User{
							ID:        "000000000000000000000000",
							Status:    models.UserStatusConfirmed,
							LastLogin: now,
							MFA: models.UserMFA{
								Enabled: false,
							},
							UserData: models.UserData{
								Username: "john_doe",
								Email:    "john.doe@test.com",
								Name:     "john doe",
							},
							Password: models.UserPassword{
								Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
							},
							Preferences: models.UserPreferences{
								PreferredNamespace: "",
							},
						},
						nil,
					).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", store.ErrNoDocuments),
			},
		},
		{
			description: "[with-tenant] fails when user is not a member of the namespace",
			req:         &requests.CreateUserToken{UserID: "000000000000000000000000", TenantID: "00000000-0000-4000-0000-000000000000"},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(
						&models.User{
							ID:        "000000000000000000000000",
							Status:    models.UserStatusConfirmed,
							LastLogin: now,
							MFA: models.UserMFA{
								Enabled: false,
							},
							UserData: models.UserData{
								Username: "john_doe",
								Email:    "john.doe@test.com",
								Name:     "john doe",
							},
							Password: models.UserPassword{
								Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
							},
							Preferences: models.UserPreferences{
								PreferredNamespace: "",
							},
						},
						nil,
					).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-4000-0000-000000000000",
							Members:  []models.Member{},
						},
						nil,
					).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrNamespaceMemberNotFound("000000000000000000000000", nil),
			},
		},
		{
			description: "[with-tenant] fails when user membership is pending",
			req:         &requests.CreateUserToken{UserID: "000000000000000000000000", TenantID: "00000000-0000-4000-0000-000000000000"},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(
						&models.User{
							ID:        "000000000000000000000000",
							Status:    models.UserStatusConfirmed,
							LastLogin: now,
							MFA: models.UserMFA{
								Enabled: false,
							},
							UserData: models.UserData{
								Username: "john_doe",
								Email:    "john.doe@test.com",
								Name:     "john doe",
							},
							Password: models.UserPassword{
								Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
							},
							Preferences: models.UserPreferences{
								PreferredNamespace: "",
							},
						},
						nil,
					).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-4000-0000-000000000000",
							Members: []models.Member{
								{
									ID:     "000000000000000000000000",
									Role:   "administrator",
									Status: models.MemberStatusPending,
								},
							},
						},
						nil,
					).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrNamespaceMemberNotFound("000000000000000000000000", nil),
			},
		},
		{
			description: "[with-tenant] succeeds",
			req:         &requests.CreateUserToken{UserID: "000000000000000000000000", TenantID: "00000000-0000-4000-0000-000000000000"},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(
						&models.User{
							ID:        "000000000000000000000000",
							Status:    models.UserStatusConfirmed,
							LastLogin: now,
							MFA: models.UserMFA{
								Enabled: false,
							},
							UserData: models.UserData{
								Username: "john_doe",
								Email:    "john.doe@test.com",
								Name:     "john doe",
							},
							Password: models.UserPassword{
								Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
							},
							Preferences: models.UserPreferences{
								PreferredNamespace: "",
							},
						},
						nil,
					).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-4000-0000-000000000000",
							Members: []models.Member{
								{
									ID:     "000000000000000000000000",
									Role:   "owner",
									Status: models.MemberStatusAccepted,
								},
							},
						},
						nil,
					).
					Once()
				preferredNamespace := "00000000-0000-4000-0000-000000000000"
				storeMock.
					On("UserUpdate", ctx, "000000000000000000000000", &models.UserChanges{PreferredNamespace: &preferredNamespace}).
					Return(nil).
					Once()
				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)
				cacheMock.
					On("Set", ctx, "token_00000000-0000-4000-0000-000000000000000000000000000000000000", testifymock.Anything, time.Hour*72).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.UserAuthResponse{
					ID:     "000000000000000000000000",
					Name:   "john doe",
					User:   "john_doe",
					Email:  "john.doe@test.com",
					Tenant: "00000000-0000-4000-0000-000000000000",
					Role:   "owner",
					Token:  "must ignore",
				},
				err: nil,
			},
		},
		{
			description: "[without-tenant] succeeds when user has a preferred namespace",
			req:         &requests.CreateUserToken{UserID: "000000000000000000000000", TenantID: ""},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(
						&models.User{
							ID:        "000000000000000000000000",
							Status:    models.UserStatusConfirmed,
							LastLogin: now,
							MFA: models.UserMFA{
								Enabled: false,
							},
							UserData: models.UserData{
								Username: "john_doe",
								Email:    "john.doe@test.com",
								Name:     "john doe",
							},
							Password: models.UserPassword{
								Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
							},
							Preferences: models.UserPreferences{
								PreferredNamespace: "00000000-0000-4000-0000-000000000000",
							},
						},
						nil,
					).
					Once()
				storeMock.
					On("NamespaceGetPreferred", ctx, "000000000000000000000000").
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-4000-0000-000000000000",
							Members: []models.Member{
								{
									ID:     "000000000000000000000000",
									Role:   "owner",
									Status: models.MemberStatusAccepted,
								},
							},
						},
						nil,
					).
					Once()
				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)
				cacheMock.
					On("Set", ctx, "token_00000000-0000-4000-0000-000000000000000000000000000000000000", testifymock.Anything, time.Hour*72).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.UserAuthResponse{
					ID:     "000000000000000000000000",
					Name:   "john doe",
					User:   "john_doe",
					Email:  "john.doe@test.com",
					Tenant: "00000000-0000-4000-0000-000000000000",
					Role:   "owner",
					Token:  "must ignore",
				},
				err: nil,
			},
		},
		{
			description: "[without-tenant] succeeds when user doesn't has a preferred namespace",
			req:         &requests.CreateUserToken{UserID: "000000000000000000000000", TenantID: ""},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("UserResolve", ctx, store.UserIDResolver, "000000000000000000000000").
					Return(
						&models.User{
							ID:        "000000000000000000000000",
							Status:    models.UserStatusConfirmed,
							LastLogin: now,
							MFA: models.UserMFA{
								Enabled: false,
							},
							UserData: models.UserData{
								Username: "john_doe",
								Email:    "john.doe@test.com",
								Name:     "john doe",
							},
							Password: models.UserPassword{
								Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
							},
							Preferences: models.UserPreferences{
								PreferredNamespace: "",
							},
						},
						nil,
					).
					Once()
				storeMock.
					On("NamespaceGetPreferred", ctx, "000000000000000000000000").
					Return(nil, store.ErrNoDocuments).
					Once()
				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)
				cacheMock.
					On("Set", ctx, "token_000000000000000000000000", testifymock.Anything, time.Hour*72).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.UserAuthResponse{
					ID:     "000000000000000000000000",
					Name:   "john doe",
					User:   "john_doe",
					Email:  "john.doe@test.com",
					Tenant: "",
					Role:   "",
					Token:  "must ignore",
				},
				err: nil,
			},
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	s := NewService(store.Store(storeMock), privateKey, &privateKey.PublicKey, cacheMock, clientMock)

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)

			res, err := s.CreateUserToken(ctx, tc.req)
			// Since the resulting token is not crucial for the assertion and
			// difficult to mock, it is safe to ignore this field.
			if res != nil {
				res.Token = "must ignore"
			}

			assert.Equal(t, tc.expected, Expected{res, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestAuthAPIKey(t *testing.T) {
	type Expected struct {
		apiKey *models.APIKey
		err    error
	}

	storeMock := new(mocks.Store)
	cacheMock := new(mockcache.Cache)

	tests := []struct {
		description   string
		key           string
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails when could not get the api key from store",
			key:         "00000000-0000-4000-0000-000000000000",
			requiredMocks: func(ctx context.Context) {
				cacheMock.
					On("Get", ctx, "api-key={00000000-0000-4000-0000-000000000000}", testifymock.Anything).
					Return(nil).
					Once()
				keySum := sha256.Sum256([]byte("00000000-0000-4000-0000-000000000000"))
				hashedKey := hex.EncodeToString(keySum[:])
				storeMock.
					On("APIKeyResolve", ctx, store.APIKeyIDResolver, hashedKey).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				apiKey: nil,
				err:    NewErrAPIKeyNotFound("", errors.New("error", "", 0)),
			},
		},
		{
			description: "fails when the api key is not valid",
			key:         "00000000-0000-4000-0000-000000000000",
			requiredMocks: func(ctx context.Context) {
				cacheMock.
					On("Get", ctx, "api-key={00000000-0000-4000-0000-000000000000}", testifymock.Anything).
					Return(nil).
					Once()
				keySum := sha256.Sum256([]byte("00000000-0000-4000-0000-000000000000"))
				hashedKey := hex.EncodeToString(keySum[:])
				storeMock.
					On("APIKeyResolve", ctx, store.APIKeyIDResolver, hashedKey).
					Return(
						&models.APIKey{
							Name:      "dev",
							ExpiresIn: time.Date(2000, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC).Unix(),
						},
						nil,
					).
					Once()
			},
			expected: Expected{
				apiKey: nil,
				err:    NewErrAPIKeyInvalid("dev"),
			},
		},
		{
			description: "succeeds",
			key:         "00000000-0000-4000-0000-000000000000",
			requiredMocks: func(ctx context.Context) {
				cacheMock.
					On("Get", ctx, "api-key={00000000-0000-4000-0000-000000000000}", testifymock.Anything).
					Return(nil).
					Once()
				keySum := sha256.Sum256([]byte("00000000-0000-4000-0000-000000000000"))
				hashedKey := hex.EncodeToString(keySum[:])
				storeMock.
					On("APIKeyResolve", ctx, store.APIKeyIDResolver, hashedKey).
					Return(
						&models.APIKey{
							Name:      "dev",
							ExpiresIn: time.Date(3000, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC).Unix(),
						},
						nil,
					).
					Once()
				cacheMock.
					On("Set", ctx, "api-key={00000000-0000-4000-0000-000000000000}", &models.APIKey{Name: "dev", ExpiresIn: time.Date(3000, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC).Unix()}, 2*time.Minute).
					Return(nil).
					Once()
			},
			expected: Expected{
				apiKey: &models.APIKey{
					Name:      "dev",
					ExpiresIn: time.Date(3000, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC).Unix(),
				},
				err: nil,
			},
		},
	}

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	service := NewService(storeMock, privKey, &privKey.PublicKey, cacheMock, clientMock)

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()
			tc.requiredMocks(ctx)
			apiKey, err := service.AuthAPIKey(ctx, tc.key)
			require.Equal(t, tc.expected, Expected{apiKey, err})
		})
	}

	storeMock.AssertExpectations(t)
}
