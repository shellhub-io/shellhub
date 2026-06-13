package services

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateDevicePairing(t *testing.T) {
	req := &requests.DevicePairingCreate{
		Hostname:  "device1",
		Identity:  &requests.DeviceIdentity{MAC: "00:00:00:00:00:01"},
		Info:      &requests.DeviceInfo{ID: "ubuntu", PrettyName: "Ubuntu 24.04 LTS"},
		PublicKey: "public-key",
	}

	// notAlreadyAccepted makes the resume lookup miss, so a fresh code is minted.
	notAlreadyAccepted := func(storeMock *storemock.MockStore) {
		queryOptionsMock := new(storemock.MockQueryOptions)
		storeMock.On("Options").Return(queryOptionsMock).Maybe()
		queryOptionsMock.On("WithDeviceStatus", models.DeviceStatusAccepted).Return(nil).Maybe()
		storeMock.
			On("DeviceResolve", mock.Anything, store.DevicePublicKeyResolver, "public-key", mock.Anything).
			Return(nil, store.ErrNoDocuments).
			Once()
	}

	// noLiveCode makes the dedup-by-pubkey lookup miss, so no existing code is reused.
	noLiveCode := func(cacheMock *cachemock.MockCache) {
		cacheMock.
			On("Get", mock.Anything, mock.MatchedBy(func(key string) bool {
				return strings.HasPrefix(key, "pairing_code_pubkey/")
			}), mock.Anything).
			Return(nil).
			Once()
	}

	type Expected struct {
		resume    bool
		tenantID  string
		reuseCode string
		err       error
	}

	cases := []struct {
		description   string
		requiredMocks func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore)
		expected      Expected
	}{
		{
			description: "succeeds to create a pairing",
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				notAlreadyAccepted(storeMock)
				noLiveCode(cacheMock)
				cacheMock.
					On("Set", mock.Anything, mock.MatchedBy(func(key string) bool {
						return regexp.MustCompile(`^pairing_code/[2-9A-HJKMNP-TV-Z]{8}$`).MatchString(key)
					}), mock.MatchedBy(func(pairing *devicePairing) bool {
						return pairing.Hostname == "device1" &&
							pairing.PublicKey == "public-key" &&
							pairing.Identity.MAC == "00:00:00:00:00:01" &&
							pairing.Info.PrettyName == "Ubuntu 24.04 LTS" &&
							pairing.Status == models.DeviceStatusPending
					}), devicePairingTTL).
					Return(nil).
					Once()
				// the dedup reference mapping the public key to the new code
				cacheMock.
					On("Set", mock.Anything, mock.MatchedBy(func(key string) bool {
						return regexp.MustCompile(`^pairing_code_pubkey/[0-9a-f]{64}$`).MatchString(key)
					}), mock.AnythingOfType("string"), devicePairingTTL).
					Return(nil).
					Once()
			},
			expected: Expected{resume: false, err: nil},
		},
		{
			description: "resumes when the public key is already accepted",
			requiredMocks: func(_ *cachemock.MockCache, storeMock *storemock.MockStore) {
				queryOptionsMock := new(storemock.MockQueryOptions)
				storeMock.On("Options").Return(queryOptionsMock).Maybe()
				queryOptionsMock.On("WithDeviceStatus", models.DeviceStatusAccepted).Return(nil).Maybe()
				storeMock.
					On("DeviceResolve", mock.Anything, store.DevicePublicKeyResolver, "public-key", mock.Anything).
					Return(&models.Device{UID: "uid1", TenantID: "tenant1", Status: models.DeviceStatusAccepted}, nil).
					Once()
			},
			expected: Expected{resume: true, tenantID: "tenant1", err: nil},
		},
		{
			description: "reuses the live code when one already exists for the public key",
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				notAlreadyAccepted(storeMock)
				// the dedup lookup finds a code already mapped to this public key
				cacheMock.
					On("Get", mock.Anything, mock.MatchedBy(func(key string) bool {
						return strings.HasPrefix(key, "pairing_code_pubkey/")
					}), mock.Anything).
					Run(func(args mock.Arguments) {
						*args.Get(2).(*string) = "WXYZ2K7Q"
					}).
					Return(nil).
					Once()
				// and that code is still live
				cacheMock.
					On("Get", mock.Anything, "pairing_code/WXYZ2K7Q", mock.Anything).
					Run(func(args mock.Arguments) {
						*args.Get(2).(*devicePairing) = devicePairing{
							PublicKey: "public-key",
							Status:    models.DeviceStatusPending,
						}
					}).
					Return(nil).
					Once()
			},
			expected: Expected{resume: false, reuseCode: "WXYZ2K7Q", err: nil},
		},
		{
			description: "fails when the cache cannot store the pairing",
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				notAlreadyAccepted(storeMock)
				noLiveCode(cacheMock)
				cacheMock.
					On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, devicePairingTTL).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{resume: false, err: errors.New("error", "", 0)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			cacheMock := new(cachemock.MockCache)
			storeMock := new(storemock.MockStore)
			tc.requiredMocks(cacheMock, storeMock)

			service := NewService(storeMock, privateKey, publicKey, cacheMock, clientMock)

			pairing, err := service.CreateDevicePairing(context.TODO(), req)
			require.Equal(tt, tc.expected.err, err)

			if tc.expected.err == nil {
				switch {
				case tc.expected.resume:
					require.Equal(tt, models.DeviceStatusAccepted, pairing.Status)
					require.Equal(tt, tc.expected.tenantID, pairing.TenantID)
					require.Empty(tt, pairing.Code)
				case tc.expected.reuseCode != "":
					require.Equal(tt, tc.expected.reuseCode, pairing.Code)
					require.Equal(tt, models.DeviceStatusPending, pairing.Status)
				default:
					require.Regexp(tt, `^[2-9A-HJKMNP-TV-Z]{8}$`, pairing.Code)
					require.Equal(tt, int(devicePairingTTL.Seconds()), pairing.ExpiresIn)
					require.Equal(tt, models.DeviceStatusPending, pairing.Status)
				}
			}

			cacheMock.AssertExpectations(tt)
			storeMock.AssertExpectations(tt)
		})
	}
}

func TestGetDevicePairingStatus(t *testing.T) {
	storeMock := new(storemock.MockStore)

	type Expected struct {
		status *models.DevicePairingStatus
		err    error
	}

	cases := []struct {
		description   string
		code          string
		requiredMocks func(cacheMock *cachemock.MockCache)
		expected      Expected
	}{
		{
			description: "fails when the code is unknown or expired",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache) {
				cacheMock.
					On("Get", mock.Anything, "pairing_code/WXYZ2K7Q", mock.Anything).
					Return(nil).
					Once()
			},
			expected: Expected{
				status: nil,
				err:    NewErrDevicePairingCodeNotFound("WXYZ2K7Q", nil),
			},
		},
		{
			description: "succeeds while the pairing is pending",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache) {
				cacheMock.
					On("Get", mock.Anything, "pairing_code/WXYZ2K7Q", mock.Anything).
					Run(func(args mock.Arguments) {
						*args.Get(2).(*devicePairing) = devicePairing{
							PublicKey: "public-key",
							Status:    models.DeviceStatusPending,
						}
					}).
					Return(nil).
					Once()
			},
			expected: Expected{
				status: &models.DevicePairingStatus{Status: models.DeviceStatusPending},
				err:    nil,
			},
		},
		{
			description: "succeeds once the pairing is accepted",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache) {
				cacheMock.
					On("Get", mock.Anything, "pairing_code/WXYZ2K7Q", mock.Anything).
					Run(func(args mock.Arguments) {
						*args.Get(2).(*devicePairing) = devicePairing{
							PublicKey: "public-key",
							Status:    models.DeviceStatusAccepted,
							TenantID:  "tenant1",
						}
					}).
					Return(nil).
					Once()
			},
			expected: Expected{
				status: &models.DevicePairingStatus{Status: models.DeviceStatusAccepted, TenantID: "tenant1"},
				err:    nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			cacheMock := new(cachemock.MockCache)
			tc.requiredMocks(cacheMock)

			service := NewService(storeMock, privateKey, publicKey, cacheMock, clientMock)

			status, err := service.GetDevicePairingStatus(context.TODO(), tc.code)
			require.Equal(tt, tc.expected, Expected{status, err})

			cacheMock.AssertExpectations(tt)
		})
	}
}

func TestAcceptDevicePairing(t *testing.T) {
	namespace := &models.Namespace{
		Name:     "namespace1",
		TenantID: "tenant1",
		Members: []models.Member{
			{ID: "owner1", Role: authorizer.RoleOwner},
			{ID: "observer1", Role: authorizer.RoleObserver},
		},
	}

	populatePairing := func(cacheMock *cachemock.MockCache) {
		cacheMock.
			On("Get", mock.Anything, "pairing_code/WXYZ2K7Q", mock.Anything).
			Run(func(args mock.Arguments) {
				*args.Get(2).(*devicePairing) = devicePairing{
					Hostname:  "device1",
					Identity:  &models.DeviceIdentity{MAC: "00:00:00:00:00:01"},
					PublicKey: "public-key",
					Status:    models.DeviceStatusPending,
				}
			}).
			Return(nil).
			Once()
	}

	// NOTE: The success path goes through AuthDevice + UpdateDeviceStatus, which
	// carry their own test suites; here we cover the pairing-specific authz and
	// lookup failures, and the happy path is exercised end to end in dev.
	cases := []struct {
		description   string
		userID        string
		req           *requests.DevicePairingAccept
		requiredMocks func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore)
		expectedErr   error
	}{
		{
			description: "fails when the code is not a well-formed pairing code",
			userID:      "owner1",
			req:         &requests.DevicePairingAccept{Code: "00000000", TenantID: "tenant1"},
			// A code that fails the charset gate must be rejected before the cache
			// lookup, so no Get is expected.
			requiredMocks: func(_ *cachemock.MockCache, _ *storemock.MockStore) {},
			expectedErr:   NewErrDevicePairingCodeNotFound("00000000", nil),
		},
		{
			description: "fails when the code is unknown or expired",
			userID:      "owner1",
			req:         &requests.DevicePairingAccept{Code: "WXYZ2K7Q", TenantID: "tenant1"},
			requiredMocks: func(cacheMock *cachemock.MockCache, _ *storemock.MockStore) {
				cacheMock.
					On("Get", mock.Anything, "pairing_code/WXYZ2K7Q", mock.Anything).
					Return(nil).
					Once()
			},
			expectedErr: NewErrDevicePairingCodeNotFound("WXYZ2K7Q", nil),
		},
		{
			description: "fails when the chosen namespace does not exist",
			userID:      "owner1",
			req:         &requests.DevicePairingAccept{Code: "WXYZ2K7Q", TenantID: "tenant1"},
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				populatePairing(cacheMock)
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expectedErr: NewErrNamespaceNotFound("tenant1", store.ErrNoDocuments),
		},
		{
			description: "fails when the user is not a member of the chosen namespace",
			userID:      "intruder",
			req:         &requests.DevicePairingAccept{Code: "WXYZ2K7Q", TenantID: "tenant1"},
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				populatePairing(cacheMock)
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
			},
			expectedErr: NewErrNamespaceMemberNotFound("intruder", nil),
		},
		{
			description: "fails when the member cannot accept devices",
			userID:      "observer1",
			req:         &requests.DevicePairingAccept{Code: "WXYZ2K7Q", TenantID: "tenant1"},
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				populatePairing(cacheMock)
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
			},
			expectedErr: NewErrRoleForbidden(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			cacheMock := new(cachemock.MockCache)
			storeMock := new(storemock.MockStore)

			tc.requiredMocks(cacheMock, storeMock)

			service := NewService(storeMock, privateKey, publicKey, cacheMock, clientMock)

			accepted, err := service.AcceptDevicePairing(context.TODO(), tc.userID, tc.req)
			require.Equal(tt, tc.expectedErr, err)
			require.Nil(tt, accepted)

			cacheMock.AssertExpectations(tt)
			storeMock.AssertExpectations(tt)
		})
	}
}

func TestPrepareDevicePairing(t *testing.T) {
	namespace := &models.Namespace{
		Name:     "namespace1",
		TenantID: "tenant1",
		Members: []models.Member{
			{ID: "owner1", Role: authorizer.RoleOwner},
			{ID: "observer1", Role: authorizer.RoleObserver},
		},
	}

	cases := []struct {
		description   string
		userID        string
		tenantID      string
		requiredMocks func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore)
		expectedErr   error
	}{
		{
			description: "fails when the namespace does not exist",
			userID:      "owner1",
			tenantID:    "tenant1",
			requiredMocks: func(_ *cachemock.MockCache, storeMock *storemock.MockStore) {
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expectedErr: NewErrNamespaceNotFound("tenant1", store.ErrNoDocuments),
		},
		{
			description: "fails when the user is not a member of the namespace",
			userID:      "intruder",
			tenantID:    "tenant1",
			requiredMocks: func(_ *cachemock.MockCache, storeMock *storemock.MockStore) {
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
			},
			expectedErr: NewErrNamespaceMemberNotFound("intruder", nil),
		},
		{
			description: "fails when the member cannot accept devices",
			userID:      "observer1",
			tenantID:    "tenant1",
			requiredMocks: func(_ *cachemock.MockCache, storeMock *storemock.MockStore) {
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
			},
			expectedErr: NewErrRoleForbidden(),
		},
		{
			description: "succeeds and mints a pre-authorized code",
			userID:      "owner1",
			tenantID:    "tenant1",
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(namespace, nil).
					Once()
				cacheMock.
					On("Set", mock.Anything, mock.MatchedBy(func(key string) bool {
						return regexp.MustCompile(`^pairing_code/[2-9A-HJKMNP-TV-Z]{8}$`).MatchString(key)
					}), mock.MatchedBy(func(pairing *devicePairing) bool {
						return pairing.PreauthTenantID == "tenant1" &&
							pairing.PreauthBy == "owner1" &&
							pairing.PublicKey == "" &&
							pairing.Status == models.DeviceStatusPending
					}), devicePairingTTL).
					Return(nil).
					Once()
			},
			expectedErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			cacheMock := new(cachemock.MockCache)
			storeMock := new(storemock.MockStore)
			tc.requiredMocks(cacheMock, storeMock)

			service := NewService(storeMock, privateKey, publicKey, cacheMock, clientMock)

			pairing, err := service.PrepareDevicePairing(context.TODO(), tc.userID, tc.tenantID)
			require.Equal(tt, tc.expectedErr, err)

			if tc.expectedErr == nil {
				require.Regexp(tt, `^[2-9A-HJKMNP-TV-Z]{8}$`, pairing.Code)
				require.Equal(tt, int(devicePairingTTL.Seconds()), pairing.ExpiresIn)
				require.Equal(tt, models.DeviceStatusPending, pairing.Status)
			} else {
				require.Nil(tt, pairing)
			}

			cacheMock.AssertExpectations(tt)
			storeMock.AssertExpectations(tt)
		})
	}
}

// TestClaimDevicePairing exercises the pre-authorized claim path, which
// CreateDevicePairing takes when the request carries a code. The happy path runs
// through AuthDevice + UpdateDeviceStatus (their own suites); here we cover the
// claim-specific guards: malformed code, non-preauthorized code, single-use, the
// atomic reservation, and reservation cleanup on failure.
func TestClaimDevicePairing(t *testing.T) {
	baseReq := func(code string) *requests.DevicePairingCreate {
		return &requests.DevicePairingCreate{
			Code:      code,
			Hostname:  "device1",
			Identity:  &requests.DeviceIdentity{MAC: "00:00:00:00:00:01"},
			Info:      &requests.DeviceInfo{ID: "ubuntu", PrettyName: "Ubuntu 24.04 LTS"},
			PublicKey: "public-key",
		}
	}

	type Expected struct {
		status   models.DeviceStatus
		tenantID string
		err      error
	}

	cases := []struct {
		description   string
		req           *requests.DevicePairingCreate
		requiredMocks func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore)
		expected      Expected
	}{
		{
			description:   "fails when the code is malformed",
			req:           baseReq("00000000"), // 0 is not in the code alphabet
			requiredMocks: func(_ *cachemock.MockCache, _ *storemock.MockStore) {},
			expected:      Expected{err: NewErrDevicePairingCodeNotFound("00000000", nil)},
		},
		{
			description: "fails when the code is not pre-authorized",
			req:         baseReq("WXYZ2K7Q"),
			requiredMocks: func(cacheMock *cachemock.MockCache, _ *storemock.MockStore) {
				// An agent-minted code has a public key but no pre-authorization.
				cacheMock.
					On("Get", mock.Anything, "pairing_code/WXYZ2K7Q", mock.Anything).
					Run(func(args mock.Arguments) {
						*args.Get(2).(*devicePairing) = devicePairing{PublicKey: "someone"}
					}).
					Return(nil).
					Once()
			},
			expected: Expected{err: NewErrDevicePairingCodeNotFound("WXYZ2K7Q", nil)},
		},
		{
			description: "is idempotent for the same device",
			req:         baseReq("WXYZ2K7Q"),
			requiredMocks: func(cacheMock *cachemock.MockCache, _ *storemock.MockStore) {
				cacheMock.
					On("Get", mock.Anything, "pairing_code/WXYZ2K7Q", mock.Anything).
					Run(func(args mock.Arguments) {
						*args.Get(2).(*devicePairing) = devicePairing{
							PreauthTenantID: "tenant1",
							PublicKey:       "public-key",
							Status:          models.DeviceStatusAccepted,
							TenantID:        "tenant1",
						}
					}).
					Return(nil).
					Once()
			},
			expected: Expected{status: models.DeviceStatusAccepted, tenantID: "tenant1"},
		},
		{
			description: "rejects a different device so the code stays single-use",
			req:         baseReq("WXYZ2K7Q"),
			requiredMocks: func(cacheMock *cachemock.MockCache, _ *storemock.MockStore) {
				cacheMock.
					On("Get", mock.Anything, "pairing_code/WXYZ2K7Q", mock.Anything).
					Run(func(args mock.Arguments) {
						*args.Get(2).(*devicePairing) = devicePairing{
							PreauthTenantID: "tenant1",
							PublicKey:       "another-device-key",
						}
					}).
					Return(nil).
					Once()
			},
			expected: Expected{err: NewErrDevicePairingCodeNotFound("WXYZ2K7Q", nil)},
		},
		{
			description: "rejects when a concurrent claim already reserved the code",
			req:         baseReq("WXYZ2K7Q"),
			requiredMocks: func(cacheMock *cachemock.MockCache, _ *storemock.MockStore) {
				cacheMock.
					On("Get", mock.Anything, "pairing_code/WXYZ2K7Q", mock.Anything).
					Run(func(args mock.Arguments) {
						*args.Get(2).(*devicePairing) = devicePairing{PreauthTenantID: "tenant1"}
					}).
					Return(nil).
					Once()
				cacheMock.
					On("SetNX", mock.Anything, "pairing_claim/WXYZ2K7Q", mock.Anything, devicePairingTTL).
					Return(false, nil).
					Once()
			},
			expected: Expected{err: NewErrDevicePairingCodeNotFound("WXYZ2K7Q", nil)},
		},
		{
			description: "reserves then releases the code when the namespace is gone",
			req:         baseReq("WXYZ2K7Q"),
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				cacheMock.
					On("Get", mock.Anything, "pairing_code/WXYZ2K7Q", mock.Anything).
					Run(func(args mock.Arguments) {
						*args.Get(2).(*devicePairing) = devicePairing{PreauthTenantID: "tenant1"}
					}).
					Return(nil).
					Once()
				cacheMock.
					On("SetNX", mock.Anything, "pairing_claim/WXYZ2K7Q", mock.Anything, devicePairingTTL).
					Return(true, nil).
					Once()
				storeMock.
					On("NamespaceResolve", mock.Anything, store.NamespaceTenantIDResolver, "tenant1").
					Return(nil, store.ErrNoDocuments).
					Once()
				cacheMock.
					On("Delete", mock.Anything, "pairing_claim/WXYZ2K7Q").
					Return(nil).
					Once()
			},
			expected: Expected{err: NewErrNamespaceNotFound("tenant1", store.ErrNoDocuments)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			cacheMock := new(cachemock.MockCache)
			storeMock := new(storemock.MockStore)
			tc.requiredMocks(cacheMock, storeMock)

			service := NewService(storeMock, privateKey, publicKey, cacheMock, clientMock)

			pairing, err := service.CreateDevicePairing(context.TODO(), tc.req)
			require.Equal(tt, tc.expected.err, err)

			if tc.expected.err == nil {
				require.Equal(tt, tc.expected.status, pairing.Status)
				require.Equal(tt, tc.expected.tenantID, pairing.TenantID)
			}

			cacheMock.AssertExpectations(tt)
			storeMock.AssertExpectations(tt)
		})
	}
}
