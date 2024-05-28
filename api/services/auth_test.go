package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"testing"
	"time"

	"github.com/cnf/structhash"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	mockcache "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	mocksGeoIp "github.com/shellhub-io/shellhub/pkg/geoip/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	uuidmock "github.com/shellhub-io/shellhub/pkg/uuid/mocks"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/undefinedlabs/go-mpatch"
)

func TestAuthDevice(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	authReq := requests.DeviceAuth{
		TenantID: "tenant",
		Identity: &requests.DeviceIdentity{
			MAC: "mac",
		},
		Sessions: []string{"session"},
	}

	auth := models.DeviceAuth{
		Hostname: authReq.Hostname,
		Identity: &models.DeviceIdentity{
			MAC: authReq.Identity.MAC,
		},
		PublicKey: authReq.PublicKey,
		TenantID:  authReq.TenantID,
	}
	uid := sha256.Sum256(structhash.Dump(auth, 1))
	device := &models.Device{
		UID: hex.EncodeToString(uid[:]),
		Identity: &models.DeviceIdentity{
			MAC: authReq.Identity.MAC,
		},
		TenantID:   authReq.TenantID,
		LastSeen:   now,
		RemoteAddr: "127.0.0.1",
		Position: &models.DevicePosition{
			Latitude:  0,
			Longitude: 0,
		},
	}

	clockMock.On("Now").Return(now).Twice()
	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "tenant"}

	mock.On("DeviceCreate", ctx, *device, "").
		Return(nil).Once()
	mock.On("SessionSetLastSeen", ctx, models.UID(authReq.Sessions[0])).
		Return(nil).Once()
	mock.On("DeviceGetByUID", ctx, models.UID(device.UID), device.TenantID).
		Return(device, nil).Once()
	mock.On("NamespaceGet", ctx, namespace.TenantID, false).
		Return(namespace, nil).Once()

	// Mock time.Now using monkey patch
	patch, err := mpatch.PatchMethod(time.Now, func() time.Time { return now })
	assert.NoError(t, err)
	defer patch.Unpatch() //nolint:errcheck

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	locator := &mocksGeoIp.Locator{}
	locator.On("GetPosition", net.ParseIP("127.0.0.1")).
		Return(geoip.Position{
			Latitude:  0,
			Longitude: 0,
		}, nil).Once()

	service := NewService(store.Store(mock), privateKey, &privateKey.PublicKey, storecache.NewNullCache(), clientMock, locator)

	authRes, err := service.AuthDevice(ctx, authReq, "127.0.0.1")
	assert.NoError(t, err)

	assert.Equal(t, device.UID, authRes.UID)
	assert.Equal(t, device.Name, authRes.Name)
	assert.Equal(t, namespace.Name, authRes.Namespace)
	assert.NotEmpty(t, authRes.Token)
	assert.Equal(t, device.RemoteAddr, "127.0.0.1")

	mock.AssertExpectations(t)
}

func TestAuthUser(t *testing.T) {
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
		req           *requests.UserAuth
		sourceIP      string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when username is not found",
			sourceIP:    "127.0.0.1",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				mock.
					On("UserGetByUsername", ctx, "john_doe").
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
			req: &requests.UserAuth{
				Identifier: "john.doe@test.com",
				Password:   "secret",
			},
			requiredMocks: func() {
				mock.
					On("UserGetByEmail", ctx, "john.doe@test.com").
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
			description: "fails when user is not confimed",
			sourceIP:    "127.0.0.1",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Confirmed: false,
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
				}

				mock.On("UserGetByUsername", ctx, "john_doe").Return(user, nil).Once()
			},
			expected: Expected{
				res:      nil,
				lockout:  0,
				mfaToken: "",
				err:      NewErrUserNotConfirmed(nil),
			},
		},
		{
			description: "fails when an account lockout occurs",
			sourceIP:    "127.0.0.1",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "wrong_password",
			},
			requiredMocks: func() {
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Confirmed: true,
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
				}

				mock.
					On("UserGetByUsername", ctx, "john_doe").
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
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "wrong_password",
			},
			requiredMocks: func() {
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Confirmed: true,
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
				}

				mock.
					On("UserGetByUsername", ctx, "john_doe").
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
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Confirmed: true,
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
				}

				mock.
					On("UserGetByUsername", ctx, "john_doe").
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
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Confirmed: true,
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
				}

				mock.
					On("UserGetByUsername", ctx, "john_doe").
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
					On("NamespaceGetFirst", ctx, "65fdd16b5f62f93184ec8a39").
					Return(nil, nil).
					Once()

				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)

				mock.
					On("UserUpdate", ctx, user.ID, &models.UserChanges{LastLogin: now}).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				res:      nil,
				lockout:  0,
				mfaToken: "",
				err: NewErrUserUpdate(&models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Confirmed: true,
					LastLogin: now,
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
				}, errors.New("error", "", 0)),
			},
		},
		{
			description: "succeeds to authenticate without a namespace",
			sourceIP:    "127.0.0.1",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Confirmed: true,
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
				}

				mock.
					On("UserGetByUsername", ctx, "john_doe").
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
					On("NamespaceGetFirst", ctx, "65fdd16b5f62f93184ec8a39").
					Return(nil, nil).
					Once()

				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)

				mock.
					On("UserUpdate", ctx, user.ID, &models.UserChanges{LastLogin: now}).
					Return(nil).
					Once()
				cacheMock.
					On("Set", ctx, "token_65fdd16b5f62f93184ec8a39", testifymock.Anything, time.Hour*72).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.UserAuthResponse{
					ID:     "65fdd16b5f62f93184ec8a39",
					Name:   "john doe",
					User:   "john_doe",
					Email:  "john.doe@test.com",
					Tenant: "",
					Role:   "",
					Token:  "must ignore",
				},
				lockout:  0,
				mfaToken: "",
				err:      nil,
			},
		},
		{
			description: "succeeds to authenticate with a namespace",
			sourceIP:    "127.0.0.1",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Confirmed: true,
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
				}

				mock.
					On("UserGetByUsername", ctx, "john_doe").
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
					On("NamespaceGetFirst", ctx, "65fdd16b5f62f93184ec8a39").
					Return(ns, nil).
					Once()

				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)

				mock.
					On("UserUpdate", ctx, user.ID, &models.UserChanges{LastLogin: now}).
					Return(nil).
					Once()
				cacheMock.
					On("Set", ctx, "token_00000000-0000-4000-0000-00000000000065fdd16b5f62f93184ec8a39", testifymock.Anything, time.Hour*72).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.UserAuthResponse{
					ID:     "65fdd16b5f62f93184ec8a39",
					Name:   "john doe",
					User:   "john_doe",
					Email:  "john.doe@test.com",
					Tenant: "00000000-0000-4000-0000-000000000000",
					Role:   "owner",
					Token:  "must ignore",
				},
				lockout:  0,
				mfaToken: "",
				err:      nil,
			},
		},
		{
			description: "succeeds to authenticate and update non-bcypt hashes",
			sourceIP:    "127.0.0.1",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "secret",
			},
			requiredMocks: func() {
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Confirmed: true,
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
				}

				mock.
					On("UserGetByUsername", ctx, "john_doe").
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
					On("NamespaceGetFirst", ctx, "65fdd16b5f62f93184ec8a39").
					Return(nil, nil).
					Once()

				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)

				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()
				hashMock.
					On("Do", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi", nil).
					Once()
				mock.
					On("UserUpdate", ctx, user.ID, &models.UserChanges{LastLogin: now, Password: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi"}).
					Return(nil).
					Once()
				cacheMock.
					On("Set", ctx, "token_65fdd16b5f62f93184ec8a39", testifymock.Anything, time.Hour*72).
					Return(nil).
					Once()
			},
			expected: Expected{
				res: &models.UserAuthResponse{
					ID:     "65fdd16b5f62f93184ec8a39",
					Name:   "john doe",
					User:   "john_doe",
					Email:  "john.doe@test.com",
					Tenant: "",
					Role:   "",
					Token:  "must ignore",
				},
				lockout:  0,
				mfaToken: "",
				err:      nil,
			},
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	service := NewService(store.Store(mock), privateKey, &privateKey.PublicKey, cacheMock, clientMock, nil)

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			res, lockout, mfaToken, err := service.AuthUser(ctx, tc.req, tc.sourceIP)
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

func TestAuthUserInfo(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	type Expected struct {
		userAuthResponse *models.UserAuthResponse
		err              error
	}

	tests := []struct {
		description   string
		username      string
		tenantID      string
		requiredMocks func()
		expected      Expected
		expectedErr   error
	}{
		{
			description: "Fails to find the user",
			username:    "notuser",
			expectedErr: errors.New("error", "", 0),
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, "notuser").Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, NewErrUserNotFound("notuser", errors.New("error", "", 0))},
		},
		{
			description: "Successful auth login",
			username:    "user",
			tenantID:    "xxxxxx",
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "id",
					TenantID: "xxxxxx",
					Members: []models.Member{
						{
							ID:   "id",
							Role: "owner",
						},
					},
				}

				mock.On("UserGetByUsername", ctx, "user").Return(&models.User{
					UserData: models.UserData{
						Username: "user",
						Name:     "user",
						Email:    "email@email.com",
					},
					ID: "id",
				}, nil).Once()
				mock.On("NamespaceGet", ctx, "xxxxxx", false).Return(namespace, nil).Once()
			},
			expected: Expected{
				userAuthResponse: &models.UserAuthResponse{
					Name:   "user",
					Token:  "---------------token----------------",
					User:   "user",
					Tenant: "xxxxxx",
					ID:     "id",
					Role:   "owner",
					Email:  "email@email.com",
				},
				err: nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			mock.ExpectedCalls = nil
			tc.requiredMocks()

			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			assert.NoError(t, err)

			service := NewService(store.Store(mock), privateKey, &privateKey.PublicKey, storecache.NewNullCache(), clientMock, nil)

			authRes, err := service.AuthUserInfo(ctx, tc.username, tc.tenantID, "---------------token----------------")
			assert.Equal(t, tc.expected.userAuthResponse, authRes)
			assert.Equal(t, tc.expected.err, err)

			mock.AssertExpectations(t)
		})
	}
}

func TestAuthGetToken(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	type Expected struct {
		userAuthResponse *models.UserAuthResponse
		err              error
	}

	tests := []struct {
		description   string
		userID        string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "success when try to get a token",
			userID:      "user",
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "id",
					TenantID: "xxxxxx",
					Members: []models.Member{
						{
							ID:   "memberID",
							Role: "owner",
						},
					},
				}

				mock.On("UserGetByID", ctx, "user", false).Return(&models.User{
					UserData: models.UserData{
						Username: "user",
						Name:     "user",
						Email:    "email@email.com",
					},
					ID: "id",
				}, 1, nil).Once()
				mock.On("NamespaceGetFirst", ctx, "id").Return(namespace, nil).Once()

				clockMock.On("Now").Return(now).Twice()
			},
			expected: Expected{
				userAuthResponse: &models.UserAuthResponse{},
				err:              nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			mock.ExpectedCalls = nil
			tc.requiredMocks()

			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			assert.NoError(t, err)

			service := NewService(mock, privateKey, &privateKey.PublicKey, storecache.NewNullCache(), clientMock, nil)

			authRes, err := service.AuthGetToken(ctx, tc.userID)
			assert.NotNil(t, authRes)
			assert.Equal(t, tc.expected.err, err)

			mock.AssertExpectations(t)
		})
	}
}
