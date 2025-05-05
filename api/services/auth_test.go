package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	goerrors "errors"
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
	clock.DefaultBackend = clockMock
	clockMock.On("Now").Return(now)

	uuidMock := &uuidmock.Uuid{}
	uuid.DefaultBackend = uuidMock
	uuidMock.On("Generate").Return("00000000-0000-4000-0000-000000000000")

	req := requests.DeviceAuth{
		Hostname: "hostname",
		Identity: &requests.DeviceIdentity{
			MAC: "mac",
		},
		TenantID:  "tenant",
		PublicKey: "",
		Sessions:  []string{"session"},
	}

	auth := models.DeviceAuth{
		Hostname: req.Hostname,
		Identity: &models.DeviceIdentity{
			MAC: req.Identity.MAC,
		},
		PublicKey: req.PublicKey,
		TenantID:  req.TenantID,
	}

	uid := sha256.Sum256(structhash.Dump(auth, 1))
	key := hex.EncodeToString(uid[:])

	claims := authorizer.DeviceClaims{
		UID:      key,
		TenantID: req.TenantID,
	}

	token, err := jwttoken.EncodeDeviceClaims(claims, privateKey)
	assert.NoError(t, err)

	type Expected struct {
		authRes *models.DeviceAuthResponse
		err     error
	}

	cases := []struct {
		description   string
		req           requests.DeviceAuth
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails to authenticate device due to no identity",
			req: requests.DeviceAuth{
				Hostname: "",
				Identity: nil,
			},
			requiredMocks: func(_ context.Context) {},
			expected: Expected{
				authRes: nil,
				err:     NewErrAuthDeviceNoIdentity(),
			},
		},
		{
			description: "fails to authenticate device due to no identity and hostname",
			req: requests.DeviceAuth{
				Hostname: "",
				Identity: &requests.DeviceIdentity{
					MAC: "",
				},
			},
			requiredMocks: func(_ context.Context) {},
			expected: Expected{
				authRes: nil,
				err:     NewErrAuthDeviceNoIdentityAndHostname(),
			},
		},
		{
			description: "fails to authenticate device due to namespace not found",
			req: requests.DeviceAuth{
				Hostname: "hostname",
				TenantID: "tenant",
				Identity: &requests.DeviceIdentity{
					MAC: "mac",
				},
			},
			requiredMocks: func(ctx context.Context) {
				cacheMock.On("Get", ctx, testifymock.Anything, testifymock.Anything).Return(nil).Once()

				storeMock.
					On("NamespaceGet", ctx, "tenant").
					Return(nil, goerrors.New("")).
					Once()
			},
			expected: Expected{
				authRes: nil,
				err:     NewErrNamespaceNotFound("tenant", goerrors.New("")),
			},
		},
		{
			description: "fails to authenticate device due to device creation error",
			req: requests.DeviceAuth{
				TenantID: "tenant",
				Info:     nil,
				Hostname: "hostname",
				Identity: &requests.DeviceIdentity{
					MAC: "mac",
				},
				PublicKey: "",
			},
			requiredMocks: func(ctx context.Context) {
				cacheMock.On("Get", ctx, testifymock.Anything, testifymock.Anything).Return(nil).Once()

				storeMock.
					On("NamespaceGet", ctx, "tenant").
					Return(&models.Namespace{
						Name:     "namespace-name",
						TenantID: "tenant",
					}, nil).Once()
				storeMock.
					On("DeviceCreate", ctx, models.Device{
						UID: key,
						Identity: &models.DeviceIdentity{
							MAC: "mac",
						},
						TenantID:   "tenant",
						LastSeen:   clock.Now(),
						Position:   &models.DevicePosition{},
						RemoteAddr: "127.0.0.1",
					}, req.Hostname).
					Return(goerrors.New("device creation error")).
					Once()
			},
			expected: Expected{
				authRes: nil,
				err: NewErrDeviceCreate(models.Device{
					UID: key,
					Identity: &models.DeviceIdentity{
						MAC: "mac",
					},
					TenantID:   "tenant",
					LastSeen:   clock.Now(),
					Position:   &models.DevicePosition{},
					RemoteAddr: "127.0.0.1",
				}, goerrors.New("device creation error")),
			},
		},

		{
			description: "fails to authenticate device due to device not found",
			req: requests.DeviceAuth{
				TenantID: "tenant",
				Info:     nil,
				Hostname: "hostname",
				Identity: &requests.DeviceIdentity{
					MAC: "mac",
				},
				PublicKey: "",
			},
			requiredMocks: func(ctx context.Context) {
				cacheMock.On("Get", ctx, testifymock.Anything, testifymock.Anything).Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "tenant").
					Return(&models.Namespace{Name: "namespace-name"}, nil).
					Once()
				storeMock.
					On("DeviceCreate", ctx, testifymock.Anything, req.Hostname).
					Return(nil).
					Once()
				storeMock.
					On("SessionSetLastSeen", ctx, models.UID("session")).
					Return(nil).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, testifymock.Anything, "tenant").
					Return(nil, goerrors.New("device not found")).
					Once()
			},
			expected: Expected{
				authRes: nil,
				err:     NewErrDeviceNotFound(models.UID(key), goerrors.New("device not found")),
			},
		},
		{
			description: "fails to authenticate device due to cache set error",
			req: requests.DeviceAuth{
				TenantID: "tenant",
				Info:     nil,
				Hostname: "hostname",
				Identity: &requests.DeviceIdentity{
					MAC: "mac",
				},
				PublicKey: "",
			},
			requiredMocks: func(ctx context.Context) {
				cacheMock.On("Get", ctx, testifymock.Anything, testifymock.Anything).Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "tenant").
					Return(&models.Namespace{Name: "namespace-name"}, nil).
					Once()
				storeMock.
					On("DeviceCreate", ctx, testifymock.Anything, req.Hostname).
					Return(nil).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, testifymock.Anything, "tenant").
					Return(&models.Device{
						UID:      key,
						Name:     "device-name",
						TenantID: "tenant",
					}, nil).
					Once()

				cacheMock.On("Set", ctx, testifymock.Anything, testifymock.Anything, time.Second*30).Return(goerrors.New("")).Once()
			},
			expected: Expected{
				authRes: nil,
				err:     goerrors.New(""),
			},
		},
		{
			description: "succeeds to authenticate device",
			req:         req,
			requiredMocks: func(ctx context.Context) {
				cacheMock.On("Get", ctx, testifymock.Anything, testifymock.Anything).Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "tenant").
					Return(&models.Namespace{Name: "namespace-name"}, nil).
					Once()
				storeMock.
					On("DeviceCreate", ctx, testifymock.Anything, req.Hostname).
					Return(nil).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, testifymock.Anything, "tenant").
					Return(&models.Device{
						UID:      key,
						Name:     "device-name",
						TenantID: "tenant",
					}, nil).
					Once()
				cacheMock.On("Set", ctx, testifymock.Anything, testifymock.Anything, time.Second*30).Return(nil).Once()
			},
			expected: Expected{
				authRes: &models.DeviceAuthResponse{
					UID:       key,
					Token:     token,
					Name:      "device-name",
					Namespace: "namespace-name",
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

			authRes, err := service.AuthDevice(ctx, tc.req, "127.0.0.1")
			require.Equal(tt, tc.expected.authRes, authRes)
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

				mock.On("UserGetByUsername", ctx, "john_doe").Return(user, nil).Once()
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
					On("UserGetByEmail", ctx, "john.doe@test.com").
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
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(nil, 0, store.ErrNoDocuments).
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
					On("UserGetByID", ctx, "000000000000000000000000", false).
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
						0,
						nil,
					).
					Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000").
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
					On("UserGetByID", ctx, "000000000000000000000000", false).
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
						0,
						nil,
					).
					Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000").
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
					On("UserGetByID", ctx, "000000000000000000000000", false).
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
						0,
						nil,
					).
					Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000").
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
					On("UserGetByID", ctx, "000000000000000000000000", false).
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
						0,
						nil,
					).
					Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000").
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
					On("UserGetByID", ctx, "000000000000000000000000", false).
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
						0,
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
					On("UserGetByID", ctx, "000000000000000000000000", false).
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
						0,
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
					On("APIKeyGet", ctx, hashedKey).
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
					On("APIKeyGet", ctx, hashedKey).
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
					On("APIKeyGet", ctx, hashedKey).
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
