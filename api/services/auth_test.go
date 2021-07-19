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

	s := NewService(store.Store(mock), privateKey, &privateKey.PublicKey)

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
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: now,
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

	authRes, err := s.AuthDevice(ctx, authReq)
	assert.NoError(t, err)

	assert.Equal(t, device.UID, authRes.UID)
	assert.Equal(t, device.Name, authRes.Name)
	assert.Equal(t, namespace.Name, authRes.Namespace)
	assert.NotEmpty(t, authRes.Token)

	mock.AssertExpectations(t)
}

func TestAuthUser(t *testing.T) {
	mock := &mocks.Store{}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	s := NewService(store.Store(mock), privateKey, &privateKey.PublicKey)

	ctx := context.TODO()

	authReq := &models.UserAuthRequest{
		Username: "user",
		Password: "passwd",
	}

	passwd := sha256.Sum256([]byte(authReq.Password))

	user := &models.User{
		Username: "user",
		Password: hex.EncodeToString(passwd[:]),
		ID:       "id",
	}

	namespace := &models.Namespace{Name: "group1", Owner: "hash1", TenantID: "tenant"}

	mock.On("UserGetByUsername", ctx, authReq.Username).
		Return(user, nil).Once()
	mock.On("NamespaceGetFirst", ctx, user.ID).
		Return(namespace, nil).Once()

	authRes, err := s.AuthUser(ctx, *authReq)
	assert.NoError(t, err)

	assert.Equal(t, user.Username, authRes.User)
	assert.Equal(t, namespace.TenantID, authRes.Tenant)
	assert.NotEmpty(t, authRes.Token)

	mock.AssertExpectations(t)
}

func TestAuthUserInfo(t *testing.T) {
	mock := &mocks.Store{}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	s := NewService(store.Store(mock), privateKey, &privateKey.PublicKey)

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

	// error getting username
	mock.On("UserGetByUsername", ctx, "notuser").
		Return(nil, Err).Once()
	authRes, err := s.AuthUserInfo(ctx, "notuser", "xxxxx", "---------------token----------------")
	assert.Error(t, err)
	assert.Nil(t, authRes)

	// error getting namespace
	mock.On("UserGetByUsername", ctx, "user").
		Return(user, nil).Once()
	mock.On("NamespaceGet", ctx, "xxxxx").
		Return(nil, Err).Once()
	authRes, err = s.AuthUserInfo(ctx, "user", "xxxxx", "---------------token----------------")
	assert.Error(t, err)
	assert.Nil(t, authRes)

	// verify empty tenant return login auth
	mock.On("UserGetByUsername", ctx, "user").
		Return(user, nil).Once()
	mock.On("NamespaceGet", ctx, "").
		Return(nil, store.ErrNoDocuments).Once()
	authRes, err = s.AuthUserInfo(ctx, "user", "", "---------------token----------------")
	assert.Nil(t, err)
	assert.Equal(t, authRes1, authRes)

	// successful auth token login with namespace found
	mock.On("UserGetByUsername", ctx, "user").
		Return(user, nil).Once()
	mock.On("NamespaceGet", ctx, namespace.TenantID).
		Return(namespace, nil).Once()
	authRes, err = s.AuthUserInfo(ctx, "user", namespace.TenantID, "---------------token----------------")
	assert.Nil(t, err)
	assert.Equal(t, authRes2, authRes)

	mock.AssertExpectations(t)
}
