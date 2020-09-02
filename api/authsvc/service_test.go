package authsvc

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

	now := time.Now()
	uid := sha256.Sum256(structhash.Dump(authReq.DeviceAuth, 1))
	device := &models.Device{
		UID:      hex.EncodeToString(uid[:]),
		Identity: authReq.Identity,
		TenantID: authReq.TenantID,
		LastSeen: now,
	}

	user := &models.User{Username: "user", TenantID: authReq.TenantID}

	mock.On("GetDeviceByMac", ctx, device.Identity.MAC, device.TenantID).
		Return(nil, nil).Once()
	mock.On("AddDevice", ctx, *device, "").
		Return(nil).Once()
	mock.On("UpdateDeviceStatus", ctx, models.UID(device.UID), true).
		Return(nil).Once()
	mock.On("KeepAliveSession", ctx, models.UID(authReq.Sessions[0])).
		Return(nil).Once()
	mock.On("GetDevice", ctx, models.UID(device.UID)).
		Return(device, nil).Once()
	mock.On("GetUserByTenant", ctx, user.TenantID).
		Return(user, nil).Once()

	// Mock time.Now using monkey patch
	patch, err := mpatch.PatchMethod(time.Now, func() time.Time { return now })
	assert.NoError(t, err)
	defer patch.Unpatch()

	authRes, err := s.AuthDevice(ctx, authReq)
	assert.NoError(t, err)

	assert.Equal(t, device.UID, authRes.UID)
	assert.Equal(t, device.Name, authRes.Name)
	assert.Equal(t, user.Username, authRes.Namespace)
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
		TenantID: "tenant",
	}

	mock.On("GetUserByUsername", ctx, authReq.Username).
		Return(user, nil).Once()

	authRes, err := s.AuthUser(ctx, *authReq)
	assert.NoError(t, err)

	assert.Equal(t, user.Username, authRes.User)
	assert.Equal(t, user.TenantID, authRes.Tenant)
	assert.NotEmpty(t, authRes.Token)

	mock.AssertExpectations(t)
}

func TestAuthGetToken(t *testing.T) {
	mock := &mocks.Store{}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	s := NewService(store.Store(mock), privateKey, &privateKey.PublicKey)

	ctx := context.TODO()

	tenant := "tenant"

	user := &models.User{
		Username: "user",
		TenantID: "tenant",
	}

	mock.On("GetUserByTenant", ctx, tenant).
		Return(user, nil).Once()

	authRes, err := s.AuthGetToken(ctx, tenant)
	assert.NoError(t, err)

	assert.Equal(t, user.Username, authRes.User)
	assert.Equal(t, user.TenantID, authRes.Tenant)
	assert.NotEmpty(t, authRes.Token)

	mock.AssertExpectations(t)
}
