package services

import (
	"context"
	"regexp"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateDeviceLoginCode(t *testing.T) {
	storeMock := new(storemock.Store)

	cases := []struct {
		description   string
		uid           string
		tenantID      string
		requiredMocks func(cacheMock *cachemock.Cache)
		expectedErr   error
	}{
		{
			description: "succeeds without a previous code",
			uid:         "uid1",
			tenantID:    "tenant1",
			requiredMocks: func(cacheMock *cachemock.Cache) {
				cacheMock.
					On("Get", mock.Anything, "login_code_device/uid1", mock.Anything).
					Return(nil).
					Once()
				cacheMock.
					On("Set", mock.Anything, mock.MatchedBy(func(key string) bool {
						return regexp.MustCompile(`^login_code/[0-9a-f]{32}$`).MatchString(key)
					}), &deviceLoginCode{UID: "uid1", TenantID: "tenant1"}, deviceLoginCodeTTL).
					Return(nil).
					Once()
				cacheMock.
					On("Set", mock.Anything, "login_code_device/uid1", mock.AnythingOfType("string"), deviceLoginCodeTTL).
					Return(nil).
					Once()
			},
			expectedErr: nil,
		},
		{
			description: "succeeds and invalidates the previous code",
			uid:         "uid1",
			tenantID:    "tenant1",
			requiredMocks: func(cacheMock *cachemock.Cache) {
				cacheMock.
					On("Get", mock.Anything, "login_code_device/uid1", mock.Anything).
					Run(func(args mock.Arguments) {
						*args.Get(2).(*string) = "previouscode"
					}).
					Return(nil).
					Once()
				cacheMock.
					On("Delete", mock.Anything, "login_code/previouscode").
					Return(nil).
					Once()
				cacheMock.
					On("Set", mock.Anything, mock.MatchedBy(func(key string) bool {
						return regexp.MustCompile(`^login_code/[0-9a-f]{32}$`).MatchString(key)
					}), &deviceLoginCode{UID: "uid1", TenantID: "tenant1"}, deviceLoginCodeTTL).
					Return(nil).
					Once()
				cacheMock.
					On("Set", mock.Anything, "login_code_device/uid1", mock.AnythingOfType("string"), deviceLoginCodeTTL).
					Return(nil).
					Once()
			},
			expectedErr: nil,
		},
		{
			description: "fails when the cache cannot store the code",
			uid:         "uid1",
			tenantID:    "tenant1",
			requiredMocks: func(cacheMock *cachemock.Cache) {
				cacheMock.
					On("Get", mock.Anything, "login_code_device/uid1", mock.Anything).
					Return(nil).
					Once()
				cacheMock.
					On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, deviceLoginCodeTTL).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expectedErr: errors.New("error", "", 0),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			cacheMock := new(cachemock.Cache)
			tc.requiredMocks(cacheMock)

			service := NewService(storeMock, privateKey, publicKey, cacheMock, clientMock)

			code, err := service.CreateDeviceLoginCode(context.TODO(), tc.uid, tc.tenantID)
			require.Equal(tt, tc.expectedErr, err)

			if tc.expectedErr == nil {
				require.Regexp(tt, `^[0-9a-f]{32}$`, code.Code)
				require.Equal(tt, int(deviceLoginCodeTTL.Seconds()), code.ExpiresIn)
			}

			cacheMock.AssertExpectations(tt)
		})
	}
}

func TestResolveDeviceLoginCode(t *testing.T) {
	type Expected struct {
		preview *models.DeviceLoginCodePreview
		err     error
	}

	device := &models.Device{
		UID:      "uid1",
		Name:     "device1",
		Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:01"},
		Info:     &models.DeviceInfo{ID: "ubuntu", PrettyName: "Ubuntu 24.04 LTS"},
		TenantID: "tenant1",
		Status:   models.DeviceStatusPending,
	}

	namespace := &models.Namespace{
		Name:     "namespace1",
		TenantID: "tenant1",
		Members:  []models.Member{{ID: "user1", Role: "owner"}},
	}

	populateCode := func(cacheMock *cachemock.Cache) {
		cacheMock.
			On("Get", mock.Anything, "login_code/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", mock.Anything).
			Run(func(args mock.Arguments) {
				*args.Get(2).(*deviceLoginCode) = deviceLoginCode{UID: "uid1", TenantID: "tenant1"}
			}).
			Return(nil).
			Once()
	}

	cases := []struct {
		description   string
		userID        string
		code          string
		requiredMocks func(cacheMock *cachemock.Cache, storeMock *storemock.Store)
		expected      Expected
	}{
		{
			description: "fails when the code is unknown or expired",
			userID:      "user1",
			code:        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			requiredMocks: func(cacheMock *cachemock.Cache, _ *storemock.Store) {
				// NOTE: A cache miss is not an error; the value is just left untouched.
				cacheMock.
					On("Get", mock.Anything, "login_code/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", mock.Anything).
					Return(nil).
					Once()
				// Not a device-bound code, so the resolve falls back to a pairing
				// lookup, which also misses.
				cacheMock.
					On("Get", mock.Anything, "pairing_code/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", mock.Anything).
					Return(nil).
					Once()
			},
			expected: Expected{
				preview: nil,
				err:     NewErrDeviceLoginCodeNotFound("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil),
			},
		},
		{
			description: "fails when the namespace does not exist",
			userID:      "user1",
			code:        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			requiredMocks: func(cacheMock *cachemock.Cache, storeMock *storemock.Store) {
				populateCode(cacheMock)
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{
				preview: nil,
				err:     NewErrDeviceLoginCodeNotFound("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", store.ErrNoDocuments),
			},
		},
		{
			description: "fails when the user is not a member of the device's namespace",
			userID:      "intruder",
			code:        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			requiredMocks: func(cacheMock *cachemock.Cache, storeMock *storemock.Store) {
				populateCode(cacheMock)
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
			},
			expected: Expected{
				preview: nil,
				err:     NewErrDeviceLoginCodeNotFound("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", nil),
			},
		},
		{
			description: "fails when the device does not exist anymore",
			userID:      "user1",
			code:        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			requiredMocks: func(cacheMock *cachemock.Cache, storeMock *storemock.Store) {
				populateCode(cacheMock)
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
				storeMock.
					On("DeviceResolve", mock.Anything, store.DeviceUIDResolver, "uid1", mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{
				preview: nil,
				err:     NewErrDeviceLoginCodeNotFound("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", store.ErrNoDocuments),
			},
		},
		{
			description: "succeeds to resolve the code",
			userID:      "user1",
			code:        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			requiredMocks: func(cacheMock *cachemock.Cache, storeMock *storemock.Store) {
				populateCode(cacheMock)
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
				storeMock.
					On("DeviceResolve", mock.Anything, store.DeviceUIDResolver, "uid1", mock.AnythingOfType("store.QueryOption")).
					Return(device, nil).
					Once()
			},
			expected: Expected{
				preview: &models.DeviceLoginCodePreview{
					Kind:      models.DeviceLoginCodeKindDevice,
					UID:       "uid1",
					Name:      "device1",
					Identity:  &models.DeviceIdentity{MAC: "00:00:00:00:00:01"},
					Info:      &models.DeviceInfo{ID: "ubuntu", PrettyName: "Ubuntu 24.04 LTS"},
					Namespace: "namespace1",
					TenantID:  "tenant1",
					Status:    models.DeviceStatusPending,
				},
				err: nil,
			},
		},
		{
			description: "succeeds to resolve a pairing code",
			userID:      "user1",
			code:        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			requiredMocks: func(cacheMock *cachemock.Cache, _ *storemock.Store) {
				cacheMock.
					On("Get", mock.Anything, "login_code/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", mock.Anything).
					Return(nil).
					Once()
				cacheMock.
					On("Get", mock.Anything, "pairing_code/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", mock.Anything).
					Run(func(args mock.Arguments) {
						*args.Get(2).(*devicePairing) = devicePairing{
							Identity:  &models.DeviceIdentity{MAC: "00:00:00:00:00:01"},
							Info:      &models.DeviceInfo{ID: "ubuntu", PrettyName: "Ubuntu 24.04 LTS"},
							PublicKey: "public-key",
							Status:    models.DeviceStatusPending,
						}
					}).
					Return(nil).
					Once()
			},
			expected: Expected{
				preview: &models.DeviceLoginCodePreview{
					Kind:     models.DeviceLoginCodeKindPairing,
					Name:     "00-00-00-00-00-01",
					Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:01"},
					Info:     &models.DeviceInfo{ID: "ubuntu", PrettyName: "Ubuntu 24.04 LTS"},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			cacheMock := new(cachemock.Cache)
			storeMock := new(storemock.Store)
			queryOptionsMock := new(storemock.QueryOptions)
			storeMock.On("Options").Return(queryOptionsMock).Maybe()
			queryOptionsMock.On("InNamespace", "tenant1").Return(nil).Maybe()

			tc.requiredMocks(cacheMock, storeMock)

			service := NewService(storeMock, privateKey, publicKey, cacheMock, clientMock)

			preview, err := service.ResolveDeviceLoginCode(context.TODO(), tc.userID, tc.code)
			require.Equal(tt, tc.expected, Expected{preview, err})

			cacheMock.AssertExpectations(tt)
			storeMock.AssertExpectations(tt)
		})
	}
}
