package services

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	storemock "github.com/shellhub-io/shellhub/server/api/store/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var pairingNow = time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)

func cachedPairingExpiry(t *testing.T, cacheMock *cachemock.MockCache, code string, expiresAt time.Time) {
	t.Helper()

	cacheMock.
		On("Get", mock.Anything, "pairing_code_expiry/"+code, mock.Anything).
		Run(func(args mock.Arguments) {
			arg, ok := args.Get(2).(*time.Time)
			require.True(t, ok)
			*arg = expiresAt
		}).
		Return(nil).
		Once()
}

func TestCreateDevicePairing(t *testing.T) {
	req := &requests.DevicePairingCreate{
		Hostname:  "device1",
		Identity:  &requests.DeviceIdentity{MAC: "00:00:00:00:00:01"},
		Info:      &requests.DeviceInfo{ID: "ubuntu", PrettyName: "Ubuntu 24.04 LTS"},
		PublicKey: "public-key",
	}

	resolvesAcceptedByPublicKey := func(storeMock *storemock.MockStore, device *models.Device, err error) {
		queryOptionsMock := new(storemock.MockQueryOptions)
		storeMock.On("Options").Return(queryOptionsMock).Maybe()
		queryOptionsMock.On("WithDeviceStatus", models.DeviceStatusAccepted).Return(nil).Maybe()
		storeMock.
			On("DeviceResolve", mock.Anything, mock.Anything, store.DevicePublicKeyResolver, "public-key", mock.Anything).
			Return(device, err).
			Once()
	}

	notAlreadyAccepted := func(storeMock *storemock.MockStore) {
		resolvesAcceptedByPublicKey(storeMock, nil, store.ErrNoDocuments)
	}

	noLiveCode := func(cacheMock *cachemock.MockCache) {
		cacheMock.
			On("Get", mock.Anything, mock.MatchedBy(func(key string) bool {
				return strings.HasPrefix(key, "pairing_code_pubkey/")
			}), mock.Anything).
			Return(nil).
			Once()
	}

	liveCode := func(cacheMock *cachemock.MockCache, code string) {
		cacheMock.
			On("Get", mock.Anything, mock.MatchedBy(func(key string) bool {
				return strings.HasPrefix(key, "pairing_code_pubkey/")
			}), mock.Anything).
			Run(func(args mock.Arguments) {
				arg, ok := args.Get(2).(*string)
				require.True(t, ok)
				*arg = code
			}).
			Return(nil).
			Once()
	}

	cachedPairing := func(cacheMock *cachemock.MockCache, code string, pairing devicePairing) {
		cacheMock.
			On("Get", mock.Anything, "pairing_code/"+code, mock.Anything).
			Run(func(args mock.Arguments) {
				arg, ok := args.Get(2).(*devicePairing)
				require.True(t, ok)
				*arg = pairing
			}).
			Return(nil).
			Once()
	}

	acceptedPairing := devicePairing{
		PublicKey: "public-key",
		Status:    models.DeviceStatusAccepted,
		TenantID:  "tenant1",
		UID:       "uid1",
	}

	storesCodeReference := func(cacheMock *cachemock.MockCache) {
		cacheMock.
			On("Set", mock.Anything, mock.MatchedBy(func(key string) bool {
				return regexp.MustCompile(`^pairing_code_pubkey/[0-9a-f]{64}$`).MatchString(key)
			}), mock.AnythingOfType("string"), devicePairingMaxLifetime).
			Return(nil).
			Once()
	}

	storesNewPairing := func(cacheMock *cachemock.MockCache) {
		cacheMock.
			On("Set", mock.Anything, mock.MatchedBy(func(key string) bool {
				return regexp.MustCompile(`^pairing_code/[2-9A-HJKMNP-TV-Z]{8}$`).MatchString(key)
			}), mock.MatchedBy(func(pairing *devicePairing) bool {
				return pairing.Hostname == "device1" &&
					pairing.PublicKey == "public-key" &&
					pairing.Identity.MAC == "00:00:00:00:00:01" &&
					pairing.Info.PrettyName == "Ubuntu 24.04 LTS" &&
					pairing.Status == models.DeviceStatusPending &&
					pairing.TenantID == "" &&
					pairing.CreatedAt.Equal(pairingNow)
			}), devicePairingMaxLifetime).
			Return(nil).
			Once()
		cacheMock.
			On("Set", mock.Anything, mock.MatchedBy(func(key string) bool {
				return regexp.MustCompile(`^pairing_code_expiry/[2-9A-HJKMNP-TV-Z]{8}$`).MatchString(key)
			}), pairingNow.Add(devicePairingRenewal), devicePairingRenewal).
			Return(nil).
			Once()
		storesCodeReference(cacheMock)
	}

	type Expected struct {
		resume    bool
		tenantID  string
		reuseCode string
		expiresIn int
		staleCode string
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
				storesNewPairing(cacheMock)
			},
			expected: Expected{resume: false, err: nil},
		},
		{
			description: "resumes when the public key is already accepted",
			requiredMocks: func(_ *cachemock.MockCache, storeMock *storemock.MockStore) {
				resolvesAcceptedByPublicKey(storeMock, &models.Device{UID: "uid1", TenantID: "tenant1", Status: models.DeviceStatusAccepted}, nil)
			},
			expected: Expected{resume: true, tenantID: "tenant1", err: nil},
		},
		{
			description: "reuses the live code when one already exists for the public key",
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				notAlreadyAccepted(storeMock)
				liveCode(cacheMock, "WXYZ2K7Q")
				cachedPairing(cacheMock, "WXYZ2K7Q", devicePairing{PublicKey: "public-key", Status: models.DeviceStatusPending})
				cachedPairingExpiry(t, cacheMock, "WXYZ2K7Q", pairingNow.Add(4*time.Minute))
			},
			expected: Expected{resume: false, reuseCode: "WXYZ2K7Q", expiresIn: 240, err: nil},
		},
		{
			description: "mints a new code when the public key's pending code expired",
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				notAlreadyAccepted(storeMock)
				liveCode(cacheMock, "WXYZ2K7Q")
				cachedPairing(cacheMock, "WXYZ2K7Q", devicePairing{PublicKey: "public-key", Status: models.DeviceStatusPending})
				cachedPairingExpiry(t, cacheMock, "WXYZ2K7Q", time.Time{})
				storesNewPairing(cacheMock)
			},
			expected: Expected{resume: false, staleCode: "WXYZ2K7Q", err: nil},
		},
		{
			description: "mints a new code when the public key's code was already accepted",
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				notAlreadyAccepted(storeMock)
				liveCode(cacheMock, "WXYZ2K7Q")
				cachedPairing(cacheMock, "WXYZ2K7Q", acceptedPairing)
				notAlreadyAccepted(storeMock)
				storesNewPairing(cacheMock)
				cacheMock.
					On("Delete", mock.Anything, "pairing_code/WXYZ2K7Q").
					Return(nil).
					Once()
			},
			expected: Expected{resume: false, staleCode: "WXYZ2K7Q", err: nil},
		},
		{
			description: "resumes without dropping the code when its accept lands the device after the first lookup",
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				notAlreadyAccepted(storeMock)
				liveCode(cacheMock, "WXYZ2K7Q")
				cachedPairing(cacheMock, "WXYZ2K7Q", acceptedPairing)
				resolvesAcceptedByPublicKey(storeMock, &models.Device{UID: "uid1", TenantID: "tenant1", Status: models.DeviceStatusAccepted}, nil)
			},
			expected: Expected{resume: true, tenantID: "tenant1", err: nil},
		},
		{
			description: "fails without touching the cached code when the store cannot tell whether the key is accepted",
			requiredMocks: func(_ *cachemock.MockCache, storeMock *storemock.MockStore) {
				resolvesAcceptedByPublicKey(storeMock, nil, errors.New("error", "", 0))
			},
			expected: Expected{resume: false, err: errors.New("error", "", 0)},
		},
		{
			description: "fails when the cache cannot store the pairing",
			requiredMocks: func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore) {
				notAlreadyAccepted(storeMock)
				noLiveCode(cacheMock)
				cacheMock.
					On("Set", mock.Anything, mock.AnythingOfType("string"), mock.Anything, devicePairingMaxLifetime).
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

			service := NewService(storeMock, privateKey, publicKey, cacheMock, WithClock(clock.NewFixed(pairingNow)))

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
					require.Equal(tt, tc.expected.expiresIn, pairing.ExpiresIn)
					require.Equal(tt, models.DeviceStatusPending, pairing.Status)
				default:
					require.Regexp(tt, `^[2-9A-HJKMNP-TV-Z]{8}$`, pairing.Code)
					require.Equal(tt, int(devicePairingRenewal.Seconds()), pairing.ExpiresIn)
					require.Equal(tt, models.DeviceStatusPending, pairing.Status)
					require.NotEqual(tt, tc.expected.staleCode, pairing.Code)
					require.Empty(tt, pairing.TenantID)
				}
			}

			cacheMock.AssertExpectations(tt)
			storeMock.AssertExpectations(tt)
		})
	}
}

func TestGetDevicePairingStatus(t *testing.T) {
	storeMock := new(storemock.MockStore)

	cachedPendingPairing := func(cacheMock *cachemock.MockCache, createdAt time.Time) {
		cacheMock.
			On("Get", mock.Anything, "pairing_code/WXYZ2K7Q", mock.Anything).
			Run(func(args mock.Arguments) {
				arg, ok := args.Get(2).(*devicePairing)
				require.True(t, ok)
				*arg = devicePairing{
					PublicKey: "public-key",
					Status:    models.DeviceStatusPending,
					CreatedAt: createdAt,
				}
			}).
			Return(nil).
			Once()
	}

	renewsUntil := func(cacheMock *cachemock.MockCache, ttl time.Duration) {
		cacheMock.
			On("Set", mock.Anything, "pairing_code_expiry/WXYZ2K7Q", pairingNow.Add(ttl), ttl).
			Return(nil).
			Once()
	}

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
			description: "fails when the pending code stopped being renewed",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache) {
				cachedPendingPairing(cacheMock, pairingNow.Add(-15*time.Minute))
				cachedPairingExpiry(t, cacheMock, "WXYZ2K7Q", time.Time{})
			},
			expected: Expected{
				status: nil,
				err:    NewErrDevicePairingCodeNotFound("WXYZ2K7Q", nil),
			},
		},
		{
			description: "renews the pending code for another full TTL",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache) {
				cachedPendingPairing(cacheMock, pairingNow.Add(-20*time.Minute))
				cachedPairingExpiry(t, cacheMock, "WXYZ2K7Q", pairingNow.Add(2*time.Minute))
				renewsUntil(cacheMock, devicePairingRenewal)
			},
			expected: Expected{
				status: &models.DevicePairingStatus{Status: models.DeviceStatusPending, ExpiresIn: 600},
				err:    nil,
			},
		},
		{
			description: "renews the pending code no further than an hour after its creation",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache) {
				cachedPendingPairing(cacheMock, pairingNow.Add(-55*time.Minute))
				cachedPairingExpiry(t, cacheMock, "WXYZ2K7Q", pairingNow.Add(3*time.Minute))
				renewsUntil(cacheMock, 5*time.Minute)
			},
			expected: Expected{
				status: &models.DevicePairingStatus{Status: models.DeviceStatusPending, ExpiresIn: 300},
				err:    nil,
			},
		},
		{
			description: "stops renewing once the code reached its hour",
			code:        "WXYZ2K7Q",
			requiredMocks: func(cacheMock *cachemock.MockCache) {
				cachedPendingPairing(cacheMock, pairingNow.Add(-58*time.Minute))
				cachedPairingExpiry(t, cacheMock, "WXYZ2K7Q", pairingNow.Add(2*time.Minute))
			},
			expected: Expected{
				status: &models.DevicePairingStatus{Status: models.DeviceStatusPending, ExpiresIn: 120},
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
						arg, ok := args.Get(2).(*devicePairing)
						require.True(t, ok)
						*arg = devicePairing{
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

			service := NewService(storeMock, privateKey, publicKey, cacheMock, WithClock(clock.NewFixed(pairingNow)))

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

	populatePairingWithStatus := func(cacheMock *cachemock.MockCache, status models.DeviceStatus) {
		cacheMock.
			On("Get", mock.Anything, "pairing_code/WXYZ2K7Q", mock.Anything).
			Run(func(args mock.Arguments) {
				arg, ok := args.Get(2).(*devicePairing)
				require.True(t, ok)
				*arg = devicePairing{
					Hostname:  "device1",
					Identity:  &models.DeviceIdentity{MAC: "00:00:00:00:00:01"},
					PublicKey: "public-key",
					Status:    status,
				}
			}).
			Return(nil).
			Once()
	}

	populatePairing := func(cacheMock *cachemock.MockCache) {
		populatePairingWithStatus(cacheMock, models.DeviceStatusPending)
		cachedPairingExpiry(t, cacheMock, "WXYZ2K7Q", pairingNow.Add(5*time.Minute))
	}

	cases := []struct {
		description   string
		userID        string
		req           *requests.DevicePairingAccept
		requiredMocks func(cacheMock *cachemock.MockCache, storeMock *storemock.MockStore)
		expectedErr   error
	}{
		{
			description:   "fails when the code is not a well-formed pairing code",
			userID:        "owner1",
			req:           &requests.DevicePairingAccept{Code: "00000000", TenantID: "tenant1"},
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
			description: "fails when the code was already accepted",
			userID:      "owner1",
			req:         &requests.DevicePairingAccept{Code: "WXYZ2K7Q", TenantID: "tenant1"},
			requiredMocks: func(cacheMock *cachemock.MockCache, _ *storemock.MockStore) {
				populatePairingWithStatus(cacheMock, models.DeviceStatusAccepted)
			},
			expectedErr: NewErrDevicePairingCodeNotFound("WXYZ2K7Q", nil),
		},
		{
			description: "fails when the pending code stopped being renewed",
			userID:      "owner1",
			req:         &requests.DevicePairingAccept{Code: "WXYZ2K7Q", TenantID: "tenant1"},
			requiredMocks: func(cacheMock *cachemock.MockCache, _ *storemock.MockStore) {
				populatePairingWithStatus(cacheMock, models.DeviceStatusPending)
				cachedPairingExpiry(t, cacheMock, "WXYZ2K7Q", time.Time{})
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

			service := NewService(storeMock, privateKey, publicKey, cacheMock, WithClock(clock.NewFixed(pairingNow)))

			accepted, err := service.AcceptDevicePairing(context.TODO(), tc.userID, tc.req)
			require.Equal(tt, tc.expectedErr, err)
			require.Nil(tt, accepted)

			cacheMock.AssertExpectations(tt)
			storeMock.AssertExpectations(tt)
		})
	}
}
