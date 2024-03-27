package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/cnf/structhash"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	cachemock "github.com/shellhub-io/shellhub/pkg/cache/mocks"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
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
		RemoteAddr: "0.0.0.0",
	}

	clockMock.On("Now").Return(now).Twice()
	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "tenant"}

	mock.On("DeviceCreate", ctx, *device, "").
		Return(nil).Once()
	mock.On("SessionSetLastSeen", ctx, models.UID(authReq.Sessions[0])).
		Return(nil).Once()
	mock.On("DeviceGetByUID", ctx, models.UID(device.UID), device.TenantID).
		Return(device, nil).Once()
	mock.On("NamespaceGet", ctx, namespace.TenantID).
		Return(namespace, nil).Once()

	// Mock time.Now using monkey patch
	patch, err := mpatch.PatchMethod(time.Now, func() time.Time { return now })
	assert.NoError(t, err)
	defer patch.Unpatch() //nolint:errcheck

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	service := NewService(store.Store(mock), privateKey, &privateKey.PublicKey, storecache.NewNullCache(), clientMock, nil)

	authRes, err := service.AuthDevice(ctx, authReq, "0.0.0.0")
	assert.NoError(t, err)

	assert.Equal(t, device.UID, authRes.UID)
	assert.Equal(t, device.Name, authRes.Name)
	assert.Equal(t, namespace.Name, authRes.Namespace)
	assert.NotEmpty(t, authRes.Token)
	assert.Equal(t, device.RemoteAddr, "0.0.0.0")

	mock.AssertExpectations(t)
}

func TestAuthUser(t *testing.T) {
	storeMock := new(mocks.Store)
	cacheMock := new(cachemock.Cache)

	type Actual struct {
		res     *models.UserAuthResponse
		lockout int64
		err     error
	}

	tests := []struct {
		description string
		req         *requests.UserAuth
		source      string
		mocks       func(context.Context)
		expected    Actual
	}{
		{
			description: "fails when username is not found",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "secret",
			},
			mocks: func(ctx context.Context) {
				storeMock.
					On("UserGetByUsername", ctx, "john_doe").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Actual{
				res:     nil,
				lockout: int64(0),
				err:     NewErrAuthUnathorized(nil),
			},
		},
		{
			description: "fails when email is not found",
			req: &requests.UserAuth{
				Identifier: "john.doe@test.com",
				Password:   "secret",
			},
			mocks: func(ctx context.Context) {
				storeMock.
					On("UserGetByEmail", ctx, "john.doe@test.com").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Actual{
				res:     nil,
				lockout: int64(0),
				err:     NewErrAuthUnathorized(nil),
			},
		},
		{
			description: "fails when user is not confimed",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "secret",
			},
			mocks: func(ctx context.Context) {
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Confirmed: false,
					LastLogin: now,
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
					},
					Password: models.UserPassword{
						Hash: "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b",
					},
				}

				storeMock.On("UserGetByUsername", ctx, "john_doe").Return(user, nil).Once()
			},
			expected: Actual{
				res:     nil,
				lockout: int64(0),
				err:     NewErrUserNotConfirmed(nil),
			},
		},
		{
			description: "fails when an account lockout occurs",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "wrong_password",
			},
			source: "127.0.0.1",
			mocks: func(ctx context.Context) {
				user := &models.User{
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
				}

				storeMock.
					On("UserGetByUsername", ctx, "john_doe").
					Return(user, nil).
					Once()
				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(1711510689), 3, nil).
					Once()
			},
			expected: Actual{
				res:     nil,
				lockout: int64(1711510689),
				err:     NewErrAuthUnathorized(nil),
			},
		},
		{
			description: "fails when input password is wrong",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "wrong_password",
			},
			source: "127.0.0.1",
			mocks: func(ctx context.Context) {
				user := &models.User{
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
				}

				storeMock.
					On("UserGetByUsername", ctx, "john_doe").
					Return(user, nil).
					Once()
				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()
				passwordMock.
					On("Compare", "wrong_password", "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(false).
					Once()
				cacheMock.
					On("StoreLoginAttempt", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()
			},
			expected: Actual{
				res:     nil,
				lockout: int64(0),
				err:     NewErrAuthUnathorized(nil),
			},
		},
		{
			description: "fails when can not retrieve MFA status",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "secret",
			},
			source: "127.0.0.1",
			mocks: func(ctx context.Context) {
				user := &models.User{
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
				}

				storeMock.
					On("UserGetByUsername", ctx, "john_doe").
					Return(user, nil).
					Once()
				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()
				passwordMock.
					On("Compare", "secret", "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(true).
					Once()
				cacheMock.
					On("ResetLoginAttempts", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(nil).
					Once()
				storeMock.
					On("GetStatusMFA", ctx, "65fdd16b5f62f93184ec8a39").
					Return(false, errors.New("error", "", 0)).
					Once()
			},
			expected: Actual{
				res:     nil,
				lockout: int64(0),
				err:     errors.New("error", "", 0),
			},
		},
		{
			description: "fails when can not update the last_login field",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "secret",
			},
			source: "127.0.0.1",
			mocks: func(ctx context.Context) {
				user := &models.User{
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
				}

				storeMock.
					On("UserGetByUsername", ctx, "john_doe").
					Return(user, nil).
					Once()
				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()
				passwordMock.
					On("Compare", "secret", "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(true).
					Once()
				cacheMock.
					On("ResetLoginAttempts", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(nil).
					Once()
				storeMock.
					On("GetStatusMFA", ctx, "65fdd16b5f62f93184ec8a39").
					Return(true, nil).
					Once()
				storeMock.
					On("NamespaceGetFirst", ctx, "65fdd16b5f62f93184ec8a39").
					Return(nil, nil).
					Once()

				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)

				storeMock.
					On("UserUpdateData", ctx, user.ID, *user).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: Actual{
				res:     nil,
				lockout: int64(0),
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
			description: "succeeds to authenticate with MFA",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "secret",
			},
			source: "127.0.0.1",
			mocks: func(ctx context.Context) {
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Confirmed: true,
					LastLogin: now,
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
						Name:     "john doe",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
				}

				storeMock.
					On("UserGetByUsername", ctx, "john_doe").
					Return(user, nil).
					Once()
				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()
				passwordMock.
					On("Compare", "secret", "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(true).
					Once()
				cacheMock.
					On("ResetLoginAttempts", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(nil).
					Once()
				storeMock.
					On("GetStatusMFA", ctx, "65fdd16b5f62f93184ec8a39").
					Return(true, nil).
					Once()
				storeMock.
					On("NamespaceGetFirst", ctx, "65fdd16b5f62f93184ec8a39").
					Return(nil, nil).
					Once()

				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)

				storeMock.
					On("UserUpdateData", ctx, user.ID, *user).
					Return(nil).
					Once()
				cacheMock.
					On("Set", ctx, "token_65fdd16b5f62f93184ec8a39", testifymock.Anything, time.Hour*72).
					Return(nil).
					Once()
			},
			expected: Actual{
				res: &models.UserAuthResponse{
					ID:     "65fdd16b5f62f93184ec8a39",
					Name:   "john doe",
					User:   "john_doe",
					Email:  "john.doe@test.com",
					Tenant: "",
					Role:   "",
					Token:  "must ignore",
					MFA: models.MFA{
						Enable:   true,
						Validate: false,
					},
				},
				lockout: int64(0),
				err:     nil,
			},
		},
		{
			description: "succeeds to authenticate without a namespace",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "secret",
			},
			source: "127.0.0.1",
			mocks: func(ctx context.Context) {
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Confirmed: true,
					LastLogin: now,
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
						Name:     "john doe",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
				}

				storeMock.
					On("UserGetByUsername", ctx, "john_doe").
					Return(user, nil).
					Once()
				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()
				passwordMock.
					On("Compare", "secret", "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(true).
					Once()
				cacheMock.
					On("ResetLoginAttempts", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(nil).
					Once()
				storeMock.
					On("GetStatusMFA", ctx, "65fdd16b5f62f93184ec8a39").
					Return(false, nil).
					Once()
				storeMock.
					On("NamespaceGetFirst", ctx, "65fdd16b5f62f93184ec8a39").
					Return(nil, nil).
					Once()

				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)

				storeMock.
					On("UserUpdateData", ctx, user.ID, *user).
					Return(nil).
					Once()
				cacheMock.
					On("Set", ctx, "token_65fdd16b5f62f93184ec8a39", testifymock.Anything, time.Hour*72).
					Return(nil).
					Once()
			},
			expected: Actual{
				res: &models.UserAuthResponse{
					ID:     "65fdd16b5f62f93184ec8a39",
					Name:   "john doe",
					User:   "john_doe",
					Email:  "john.doe@test.com",
					Tenant: "",
					Role:   "",
					Token:  "must ignore",
					MFA: models.MFA{
						Enable:   false,
						Validate: false,
					},
				},
				lockout: int64(0),
				err:     nil,
			},
		},
		{
			description: "succeeds to authenticate with a namespace",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "secret",
			},
			source: "127.0.0.1",
			mocks: func(ctx context.Context) {
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Confirmed: true,
					LastLogin: now,
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
						Name:     "john doe",
					},
					Password: models.UserPassword{
						Hash: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi",
					},
				}

				storeMock.
					On("UserGetByUsername", ctx, "john_doe").
					Return(user, nil).
					Once()
				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()
				passwordMock.
					On("Compare", "secret", "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi").
					Return(true).
					Once()
				cacheMock.
					On("ResetLoginAttempts", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(nil).
					Once()
				storeMock.
					On("GetStatusMFA", ctx, "65fdd16b5f62f93184ec8a39").
					Return(false, nil).
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

				storeMock.
					On("NamespaceGetFirst", ctx, "65fdd16b5f62f93184ec8a39").
					Return(ns, nil).
					Once()

				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)

				storeMock.
					On("UserUpdateData", ctx, user.ID, *user).
					Return(nil).
					Once()
				cacheMock.
					On("Set", ctx, "token_00000000-0000-4000-0000-00000000000065fdd16b5f62f93184ec8a39", testifymock.Anything, time.Hour*72).
					Return(nil).
					Once()
			},
			expected: Actual{
				res: &models.UserAuthResponse{
					ID:     "65fdd16b5f62f93184ec8a39",
					Name:   "john doe",
					User:   "john_doe",
					Email:  "john.doe@test.com",
					Tenant: "00000000-0000-4000-0000-000000000000",
					Role:   "owner",
					Token:  "must ignore",
					MFA: models.MFA{
						Enable:   false,
						Validate: false,
					},
				},
				lockout: int64(0),
				err:     nil,
			},
		},
		{
			description: "succeeds to authenticate and update non-bcypt hashes",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "secret",
			},
			source: "127.0.0.1",
			mocks: func(ctx context.Context) {
				user := &models.User{
					ID:        "65fdd16b5f62f93184ec8a39",
					Confirmed: true,
					LastLogin: now,
					UserData: models.UserData{
						Username: "john_doe",
						Email:    "john.doe@test.com",
						Name:     "john doe",
					},
					Password: models.UserPassword{
						Hash: "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b",
					},
				}

				storeMock.
					On("UserGetByUsername", ctx, "john_doe").
					Return(user, nil).
					Once()
				cacheMock.
					On("HasAccountLockout", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(int64(0), 0, nil).
					Once()
				passwordMock.
					On("Compare", "secret", "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b").
					Return(true).
					Once()
				cacheMock.
					On("ResetLoginAttempts", ctx, "127.0.0.1", "65fdd16b5f62f93184ec8a39").
					Return(nil).
					Once()
				storeMock.
					On("GetStatusMFA", ctx, "65fdd16b5f62f93184ec8a39").
					Return(false, nil).
					Once()

				storeMock.
					On("NamespaceGetFirst", ctx, "65fdd16b5f62f93184ec8a39").
					Return(nil, nil).
					Once()

				clockMock := new(clockmock.Clock)
				clock.DefaultBackend = clockMock
				clockMock.On("Now").Return(now)

				storeMock.
					On("UserUpdateData", ctx, user.ID, *user).
					Return(nil).
					Once()
				cacheMock.
					On("Set", ctx, "token_65fdd16b5f62f93184ec8a39", testifymock.Anything, time.Hour*72).
					Return(nil).
					Once()

				passwordMock.
					On("Hash", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi", nil).
					Once()

				storeMock.
					On("UserUpdatePassword", ctx, "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi", "65fdd16b5f62f93184ec8a39").
					Return(nil).
					Once()
			},
			expected: Actual{
				res: &models.UserAuthResponse{
					ID:     "65fdd16b5f62f93184ec8a39",
					Name:   "john doe",
					User:   "john_doe",
					Email:  "john.doe@test.com",
					Tenant: "",
					Role:   "",
					Token:  "must ignore",
					MFA: models.MFA{
						Enable:   false,
						Validate: false,
					},
				},
				lockout: int64(0),
				err:     nil,
			},
		},
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	service := NewService(store.Store(storeMock), privateKey, &privateKey.PublicKey, storecache.Cache(cacheMock), clientMock, nil)

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.TODO()

			tc.mocks(ctx)
			res, lockout, err := service.AuthUser(ctx, tc.req, tc.source)
			// Since the resulting token is not crucial for the assertion and
			// difficult to mock, it is safe to ignore this field.
			if res != nil {
				res.Token = "must ignore"
			}

			assert.Equal(t, tc.expected, Actual{res, lockout, err})
		})
	}

	storeMock.AssertExpectations(t)
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
				mock.On("NamespaceGet", ctx, "xxxxxx").Return(namespace, nil).Once()
				mock.On("GetStatusMFA", ctx, "id").Return(false, nil).Once()
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
				mock.On("GetStatusMFA", ctx, "id").Return(false, nil).Once()

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

			authRes, err := service.AuthGetToken(ctx, tc.userID, false)
			assert.NotNil(t, authRes)
			assert.Equal(t, tc.expected.err, err)

			mock.AssertExpectations(t)
		})
	}
}
