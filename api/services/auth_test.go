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
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
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
	mock := new(mocks.Store)

	ctx := context.TODO()

	type Expected struct {
		userAuthResponse *models.UserAuthResponse
		err              error
	}

	tests := []struct {
		description   string
		req           requests.UserAuth
		requiredMocks func()
		expected      Expected
		expectedErr   error
	}{
		{
			description: "Fails when username is not found",
			req: requests.UserAuth{
				Username: "user",
				Password: "passwd",
			},
			expectedErr: errors.New("error", "", 0),
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, "user").Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, NewErrAuthUnathorized(nil)},
		},
		{
			description: "Fails when email is not found",
			req: requests.UserAuth{
				Username: "user@test.com",
				Password: "passwd",
			},
			expectedErr: errors.New("error", "", 0),
			requiredMocks: func() {
				mock.On("UserGetByEmail", ctx, "user@test.com").Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, NewErrAuthUnathorized(nil)},
		},
		{
			description: "Fails when user has account but wrong password",
			req: requests.UserAuth{
				Username: "user",
				Password: "passwd",
			},
			requiredMocks: func() {
				user := &models.User{
					UserData: models.UserData{
						Username: "user",
					},

					UserPassword: models.NewUserPassword("wrongPassword"),
					ID:           "id",
					Confirmed:    true,
					LastLogin:    now,
				}

				mock.On("UserGetByUsername", ctx, "user").Return(user, nil).Once()

				namespace := &models.Namespace{
					Name:     "group1",
					Owner:    "hash1",
					TenantID: "tenant",
				}

				mock.On("NamespaceGetFirst", ctx, user.ID).Return(namespace, nil).Once()
			},
			expected: Expected{nil, NewErrAuthUnathorized(nil)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			assert.NoError(t, err)

			service := NewService(store.Store(mock), privateKey, &privateKey.PublicKey, storecache.NewNullCache(), clientMock, nil)
			authRes, err := service.AuthUser(ctx, &models.UserAuthRequest{
				Identifier: models.UserAuthIdentifier(tc.req.Username),
				Password:   tc.req.Password,
			}, true)
			assert.Equal(t, tc.expected, Expected{authRes, err})
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
