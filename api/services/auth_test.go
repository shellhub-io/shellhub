package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/cnf/structhash"
	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/undefinedlabs/go-mpatch"
)

func TestAuthDevice(t *testing.T) {
	mock := &mocks.Store{}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	s := NewService(store.Store(mock), privateKey, &privateKey.PublicKey, storecache.NewNullCache())

	ctx := context.TODO()

	authReq := &models.DeviceAuthRequest{
		DeviceAuth: &models.DeviceAuth{
			TenantID: "tenant",
			Identity: &models.DeviceIdentity{
				MAC: "mac",
			},
		},
		Sessions: []string{"session"},
	}

	now := clock.Now()
	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))
	device := &models.Device{
		UID:        hex.EncodeToString(uid[:]),
		Identity:   authReq.Identity,
		TenantID:   authReq.TenantID,
		LastSeen:   now,
		RemoteAddr: "0.0.0.0",
	}

	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "tenant"}

	mock.On("DeviceCreate", ctx, *device, "").
		Return(nil).Once()
	mock.On("DeviceSetOnline", ctx, models.UID(device.UID), true).
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

	authRes, err := s.AuthDevice(ctx, authReq, "0.0.0.0")
	assert.NoError(t, err)

	assert.Equal(t, device.UID, authRes.UID)
	assert.Equal(t, device.Name, authRes.Name)
	assert.Equal(t, namespace.Name, authRes.Namespace)
	assert.NotEmpty(t, authRes.Token)
	assert.Equal(t, device.RemoteAddr, "0.0.0.0")

	mock.AssertExpectations(t)
}

func TestAuthUser(t *testing.T) {
	mock := &mocks.Store{}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	s := NewService(store.Store(mock), privateKey, &privateKey.PublicKey, storecache.NewNullCache())

	ctx := context.TODO()

	authReq := &models.UserAuthRequest{
		Username: "user",
		Password: "passwd",
	}

	wrongPasswd := sha256.Sum256([]byte("wrongPassword"))
	passwd := sha256.Sum256([]byte(authReq.Password))

	userWithWrongPassword := &models.User{
		Username:      "user",
		Password:      hex.EncodeToString(wrongPasswd[:]),
		ID:            "id",
		Authenticated: true,
		LastLogin:     clock.Now(),
	}

	userAuthenticated := &models.User{
		Username:      "user",
		Password:      hex.EncodeToString(passwd[:]),
		ID:            "id",
		Authenticated: true,
		LastLogin:     clock.Now(),
	}

	userNotActivatedAccount := &models.User{
		Username:      "user",
		Password:      hex.EncodeToString(passwd[:]),
		ID:            "id",
		Authenticated: false,
		LastLogin:     clock.Now(),
	}

	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "tenant"}

	mock.On("UserGetByUsername", ctx, authReq.Username).Return(userAuthenticated, nil).Once()
	mock.On("NamespaceGetFirst", ctx, userAuthenticated.ID).Return(namespace, nil).Once()
	mock.On("UserUpdateData", ctx, userAuthenticated, userAuthenticated.ID).Return(nil).Once()

	authRes, err := s.AuthUser(ctx, *authReq)
	assert.NoError(t, err)

	Err := errors.New("error")

	type Expected struct {
		userAuthResponse *models.UserAuthResponse
		err              error
	}

	tests := []struct {
		description   string
		args          models.UserAuthRequest
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "Fails when user has no account",
			args:        *authReq,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, authReq.Username).Return(nil, Err).Once()
				mock.On("UserGetByEmail", ctx, authReq.Username).Return(nil, Err).Once()
			},
			expected: Expected{nil, Err},
		},
		{
			description: "Fails when user has account but wrong password",
			args:        *authReq,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, authReq.Username).Return(userWithWrongPassword, nil).Once()
				mock.On("NamespaceGetFirst", ctx, userWithWrongPassword.ID).Return(namespace, nil).Once()
			},
			expected: Expected{nil, ErrUnauthorized},
		},
		{
			description: "Fails when user has account but not activated",
			args:        *authReq,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, authReq.Username).Return(userNotActivatedAccount, nil).Once()
			},
			expected: Expected{nil, ErrForbidden},
		},
		{
			description: "Successful authentication",
			args:        *authReq,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, authReq.Username).Return(userAuthenticated, nil).Once()
				mock.On("NamespaceGetFirst", ctx, userAuthenticated.ID).Return(namespace, nil).Once()
				mock.On("UserUpdateData", ctx, userAuthenticated, userAuthenticated.ID).Return(nil).Once()
			},
			expected: Expected{authRes, nil},
		},
	}

	for _, test := range tests {
		t.Log("PASS:  ", test.description)
		test.requiredMocks()
		authRes, err := s.AuthUser(ctx, test.args)
		assert.Equal(t, test.expected, Expected{authRes, err})
	}

	mock.AssertExpectations(t)
}

func TestAuthUserInfo(t *testing.T) {
	mock := &mocks.Store{}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	s := NewService(store.Store(mock), privateKey, &privateKey.PublicKey, storecache.NewNullCache())

	ctx := context.TODO()

	authRes1 := &models.UserAuthResponse{
		Name:   "user",
		Token:  "---------------token----------------",
		User:   "user",
		Tenant: "",
		ID:     "id",
		Email:  "email@email.com",
	}

	authRes2 := &models.UserAuthResponse{
		Name:   "user",
		Token:  "---------------token----------------",
		User:   "user",
		Tenant: "xxxxxx",
		ID:     "id",
		Email:  "email@email.com",
	}

	user := &models.User{
		Username: "user",
		Name:     "user",
		ID:       "id",
		Email:    "email@email.com",
	}

	namespace := &models.Namespace{
		Name:     "namespace",
		Owner:    "id",
		TenantID: "xxxxxx",
	}

	Err := errors.New("error")

	type Expected struct {
		userAuthResponse *models.UserAuthResponse
		err              error
	}

	tests := []struct {
		description   string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "Fails to find the user",
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, "notuser").Return(nil, Err).Once()
			},
			expected: Expected{nil, Err},
		},
		{
			description: "Fails to find the namespace",
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, "user").Return(user, nil).Once()
				mock.On("NamespaceGet", ctx, "xxxxx").Return(nil, Err).Once()
			},
			expected: Expected{nil, Err},
		},
		{
			description: "Fails empty tenant return login auth",
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, "user").Return(user, nil).Once()
				mock.On("NamespaceGet", ctx, "").Return(nil, store.ErrNoDocuments).Once()
			},
			expected: Expected{authRes1, nil},
		},
		{
			description: "Successful auth login",
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, "user").Return(user, nil).Once()
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
			},
			expected: Expected{authRes2, nil},
		},
	}

	authRes := new(models.UserAuthResponse)

	for _, test := range tests {
		t.Log("PASS:  ", test.description)
		test.requiredMocks()
		switch test.description {
		case "Fails to find the user":
			authRes, err = s.AuthUserInfo(ctx, "notuser", "xxxxx", "---------------token----------------")
		case "Fails to find the namespace":
			authRes, err = s.AuthUserInfo(ctx, "user", "xxxxx", "---------------token----------------")
		case "Fails empty tenant return login auth":
			authRes, err = s.AuthUserInfo(ctx, "user", "", "---------------token----------------")
		case "Successful auth login":
			authRes, err = s.AuthUserInfo(ctx, "user", namespace.TenantID, "---------------token----------------")
		}
		assert.Equal(t, test.expected, Expected{authRes, err})
	}

	mock.AssertExpectations(t)
}
