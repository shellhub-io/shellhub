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
	gomock "github.com/stretchr/testify/mock"
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

	// [DeviceAuthClaims.WithDefaults]
	uuidMock := &uuidmock.Uuid{}
	uuid.DefaultBackend = uuidMock
	uuidMock.
		On("Generate").
		Return("cdfd3cb0-c44e-4e54-b931-6d57713ad159").
		Once()

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

	cache := new(mockcache.Cache)
	service := NewService(store.Store(mock), privateKey, &privateKey.PublicKey, cache, clientMock, locator)

	cache.On("Get", ctx, gomock.Anything, gomock.Anything).Return(errors.New("", "", 0))
	cache.On("Set", ctx, gomock.Anything, gomock.Anything, gomock.Anything).Return(nil)

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
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
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
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
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
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
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
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
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
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
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
					On("NamespaceGetPreferred", ctx, "", "65fdd16b5f62f93184ec8a39").
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
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
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
					On("NamespaceGetPreferred", ctx, "", "65fdd16b5f62f93184ec8a39").
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
					ID:     "65fdd16b5f62f93184ec8a39",
					Name:   "john doe",
					User:   "john_doe",
					Email:  "john.doe@test.com",
					Tenant: "",
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
					Preferences: models.UserPreferences{
						PreferredNamespace: "00000000-0000-4000-0000-000000000000",
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
					On("NamespaceGetPreferred", ctx, "00000000-0000-4000-0000-000000000000", "65fdd16b5f62f93184ec8a39").
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
			description: "succeeds to authenticate with a namespace (and empty preferred namespace)",
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
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
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
					On("NamespaceGetPreferred", ctx, "", "65fdd16b5f62f93184ec8a39").
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
					Preferences: models.UserPreferences{
						PreferredNamespace: "",
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
					On("NamespaceGetPreferred", ctx, "", "65fdd16b5f62f93184ec8a39").
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
					ID:     "65fdd16b5f62f93184ec8a39",
					Name:   "john doe",
					User:   "john_doe",
					Email:  "john.doe@test.com",
					Tenant: "",
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
			description: "fails when ID is not found",
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
			description: "fails when tenant_id is empty and namespace is not found",
			req:         &requests.CreateUserToken{UserID: "000000000000000000000000", TenantID: ""},
			requiredMocks: func(ctx context.Context) {
				user := &models.User{
					ID:        "000000000000000000000000",
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
					Preferences: models.UserPreferences{
						PreferredNamespace: "00000000-0000-4000-0000-000000000000",
					},
				}

				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(user, 0, nil).
					Once()
				storeMock.
					On("NamespaceGetPreferred", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000").
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrNamespaceNotFound("", store.ErrNoDocuments),
			},
		},
		{
			description: "fails when tenant_id is not empty and namespace is not found",
			req:         &requests.CreateUserToken{UserID: "000000000000000000000000", TenantID: "00000000-0000-4000-0000-000000000000"},
			requiredMocks: func(ctx context.Context) {
				user := &models.User{
					ID:        "000000000000000000000000",
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

				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(user, 0, nil).
					Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", false).
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", store.ErrNoDocuments),
			},
		},
		{
			description: "fails when user is not member of the namespace",
			req:         &requests.CreateUserToken{UserID: "000000000000000000000000", TenantID: "00000000-0000-4000-0000-000000000000"},
			requiredMocks: func(ctx context.Context) {
				user := &models.User{
					ID:        "000000000000000000000000",
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

				ns := &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Members:  []models.Member{},
				}

				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(user, 0, nil).
					Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", false).
					Return(ns, nil).
					Once()
			},
			expected: Expected{
				res: nil,
				err: NewErrNamespaceMemberNotFound("000000000000000000000000", nil),
			},
		},
		{
			description: "succeeds when tenant id is empty",
			req:         &requests.CreateUserToken{UserID: "000000000000000000000000", TenantID: ""},
			requiredMocks: func(ctx context.Context) {
				user := &models.User{
					ID:        "000000000000000000000000",
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
					Preferences: models.UserPreferences{
						PreferredNamespace: "00000000-0000-4000-0000-000000000000",
					},
				}

				ns := &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Members: []models.Member{
						{
							ID:   "000000000000000000000000",
							Role: "owner",
						},
					},
				}

				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(user, 0, nil).
					Once()
				storeMock.
					On("NamespaceGetPreferred", ctx, "00000000-0000-4000-0000-000000000000", "000000000000000000000000").
					Return(ns, nil).
					Once()
				preferredNamespace := "00000000-0000-4000-0000-000000000000"
				storeMock.
					On("UserUpdate", ctx, user.ID, &models.UserChanges{PreferredNamespace: &preferredNamespace}).
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
			description: "succeeds when tenant id is not empty",
			req:         &requests.CreateUserToken{UserID: "000000000000000000000000", TenantID: "00000000-0000-4000-0000-000000000000"},
			requiredMocks: func(ctx context.Context) {
				user := &models.User{
					ID:        "000000000000000000000000",
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

				ns := &models.Namespace{
					TenantID: "00000000-0000-4000-0000-000000000000",
					Members: []models.Member{
						{
							ID:   "000000000000000000000000",
							Role: "owner",
						},
					},
				}

				storeMock.
					On("UserGetByID", ctx, "000000000000000000000000", false).
					Return(user, 0, nil).
					Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", false).
					Return(ns, nil).
					Once()
				preferredNamespace := "00000000-0000-4000-0000-000000000000"
				storeMock.
					On("UserUpdate", ctx, user.ID, &models.UserChanges{PreferredNamespace: &preferredNamespace}).
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
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	s := NewService(store.Store(storeMock), privateKey, &privateKey.PublicKey, cacheMock, clientMock, nil)

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
							ExpiresIn: time.Date(2000, 0o1, 0o1, 12, 0, 0, 0, time.UTC).Unix(),
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
							ExpiresIn: time.Date(3000, 0o1, 0o1, 12, 0, 0, 0, time.UTC).Unix(),
						},
						nil,
					).
					Once()
				cacheMock.
					On("Set", ctx, "api-key={00000000-0000-4000-0000-000000000000}", &models.APIKey{Name: "dev", ExpiresIn: time.Date(3000, 0o1, 0o1, 12, 0, 0, 0, time.UTC).Unix()}, 2*time.Minute).
					Return(nil).
					Once()
			},
			expected: Expected{
				apiKey: &models.APIKey{
					Name:      "dev",
					ExpiresIn: time.Date(3000, 0o1, 0o1, 12, 0, 0, 0, time.UTC).Unix(),
				},
				err: nil,
			},
		},
	}

	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	service := NewService(storeMock, privKey, &privKey.PublicKey, cacheMock, clientMock, nil)

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
