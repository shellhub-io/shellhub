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
	storeMock := new(storemock.MockStore)

	cases := []struct {
		description   string
		uid           string
		tenantID      string
		requiredMocks func(cacheMock *cachemock.MockCache)
		expectedErr   error
	}{
		{
			description: "succeeds without a previous code",
			uid:         "uid1",
			tenantID:    "tenant1",
			requiredMocks: func(cacheMock *cachemock.MockCache) {
				cacheMock.
					On("Get", mock.Anything, "login_code_device/uid1", mock.Anything).
					Return(nil).
					Once()
				cacheMock.
					On("Set", mock.Anything, mock.MatchedBy(func(key string) bool {
						return regexp.MustCompile(`^login_code/[2-9A-HJKMNP-TV-Z]{8}$`).MatchString(key)
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
			requiredMocks: func(cacheMock *cachemock.MockCache) {
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
						return regexp.MustCompile(`^login_code/[2-9A-HJKMNP-TV-Z]{8}$`).MatchString(key)
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
			requiredMocks: func(cacheMock *cachemock.MockCache) {
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
			cacheMock := new(cachemock.MockCache)
			tc.requiredMocks(cacheMock)

			service := NewService(storeMock, privateKey, publicKey, cacheMock, clientMock)

			code, err := service.CreateDeviceLoginCode(context.TODO(), tc.uid, tc.tenantID)
			require.Equal(tt, tc.expectedErr, err)

			if tc.expectedErr == nil {
				require.Regexp(tt, `^[2-9A-HJKMNP-TV-Z]{8}$`, code.Code)
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

	populateCode := func(cacheMock *cachemock.MockCache) {
		cacheMock.
			On("Get", mock.Anything, "login_code/WXYZ2K7Q", mock.Anything).
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
		requiredMocks func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore)
		expected      Expected
	}{
		{
			description: "fails when the code is unknown or expired",
			userID:      "user1",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache, _ *storemock.MockStore) {
				// NOTE: A cache miss is not an error; the value is just left untouched.
				cacheMock.
					On("Get", mock.Anything, "login_code/WXYZ2K7Q", mock.Anything).
					Return(nil).
					Once()
				// Not a device-bound code, so the resolve falls back to a pairing
				// lookup, which also misses.
				cacheMock.
					On("Get", mock.Anything, "pairing_code/WXYZ2K7Q", mock.Anything).
					Return(nil).
					Once()
			},
			expected: Expected{
				preview: nil,
				err:     NewErrDeviceLoginCodeNotFound("WXYZ2K7Q", nil),
			},
		},
		{
			description: "fails when the namespace does not exist",
			userID:      "user1",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				populateCode(cacheMock)
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{
				preview: nil,
				err:     NewErrDeviceLoginCodeNotFound("WXYZ2K7Q", store.ErrNoDocuments),
			},
		},
		{
			description: "fails when the user is not a member of the device's namespace",
			userID:      "intruder",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				populateCode(cacheMock)
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
			},
			expected: Expected{
				preview: nil,
				err:     NewErrDeviceLoginCodeNotFound("WXYZ2K7Q", nil),
			},
		},
		{
			description: "fails when the device does not exist anymore",
			userID:      "user1",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				populateCode(cacheMock)
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
				storeMock.
					On("DeviceResolve", mock.Anything, store.DeviceUIDResolver, "uid1", mock.MatchedBy(func(opts []store.QueryOption) bool { return len(opts) == 1 })).
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{
				preview: nil,
				err:     NewErrDeviceLoginCodeNotFound("WXYZ2K7Q", store.ErrNoDocuments),
			},
		},
		{
			description: "succeeds to resolve the code",
			userID:      "user1",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				populateCode(cacheMock)
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
				storeMock.
					On("DeviceResolve", mock.Anything, store.DeviceUIDResolver, "uid1", mock.MatchedBy(func(opts []store.QueryOption) bool { return len(opts) == 1 })).
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
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache, _ *storemock.MockStore) {
				cacheMock.
					On("Get", mock.Anything, "login_code/WXYZ2K7Q", mock.Anything).
					Return(nil).
					Once()
				cacheMock.
					On("Get", mock.Anything, "pairing_code/WXYZ2K7Q", mock.Anything).
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
			cacheMock := new(cachemock.MockCache)
			storeMock := new(storemock.MockStore)
			queryOptionsMock := new(storemock.MockQueryOptions)
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
