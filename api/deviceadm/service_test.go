package deviceadm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestListDevices(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	devices := []models.Device{
		{UID: "uid"},
	}

	filters := []models.Filter{
		{
			Type:   "property",
			Params: &models.PropertyParams{Name: "hostname", Operator: "eq"}},
	}

	filterJSON, err := json.Marshal(filters)
	assert.NoError(t, err)

	encodedFilter := base64.StdEncoding.EncodeToString(filterJSON)

	query := paginator.Query{Page: 1, PerPage: 10}

	mock.On("DeviceList", ctx, query, filters, "accepted", "name", "asc").
		Return(devices, len(devices), nil).Once()

	returnedDevices, count, err := s.ListDevices(ctx, query, encodedFilter, "accepted", "name", "asc")
	assert.NoError(t, err)
	assert.Equal(t, devices, returnedDevices)
	assert.Equal(t, count, len(devices))

	mock.AssertExpectations(t)
}

func TestGetDevice(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	device := &models.Device{UID: "uid"}
	mock.On("DeviceGet", ctx, models.UID(device.UID)).
		Return(device, nil).Once()

	returnedDevice, err := s.GetDevice(ctx, models.UID(device.UID))
	assert.NoError(t, err)
	assert.Equal(t, device, returnedDevice)

	mock.AssertExpectations(t)
}

func TestDeleteDevice(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	user := &models.User{Name: "name", Email: "", Username: "username", ID: "id"}
	namespace := &models.Namespace{Name: "group1", Owner: "id", TenantID: "tenant"}
	device := &models.Device{UID: "uid", TenantID: "tenant"}

	mock.On("UserGetByUsername", ctx, user.Username).
		Return(user, nil).Once()
	mock.On("NamespaceGet", ctx, device.TenantID).
		Return(namespace, nil).Once()
	mock.On("DeviceGetByUID", ctx, models.UID(device.UID), device.TenantID).
		Return(device, nil).Once()
	mock.On("DeviceDelete", ctx, models.UID(device.UID)).
		Return(nil).Once()

	err := s.DeleteDevice(ctx, models.UID(device.UID), device.TenantID, user.Username)
	assert.NoError(t, err)

	// Tests to user doesnt owner namespace
	userDoesnotOwner := &models.User{Name: "name1", Email: "email1", Username: "username1", ID: "id1"}

	mock.On("UserGetByUsername", ctx, userDoesnotOwner.Username).
		Return(userDoesnotOwner, nil).Once()
	mock.On("NamespaceGet", ctx, device.TenantID).
		Return(namespace, nil).Once()

	err = s.DeleteDevice(ctx, models.UID(device.UID), device.TenantID, userDoesnotOwner.Username)
	assert.EqualError(t, err, "unauthorized")

	mock.AssertExpectations(t)
}

func TestRenameDevice(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	user := &models.User{Name: "name", Email: "email", Username: "username", ID: "id"}
	namespace := &models.Namespace{Name: "group1", Owner: "id", TenantID: "tenant"}
	device := &models.Device{UID: "uid", Name: "name", TenantID: "tenant"}
	renamedDevice := &models.Device{UID: "uid", Name: "rename", TenantID: "tenant"}

	mock.On("UserGetByUsername", ctx, user.Username).
		Return(user, nil).Once()
	mock.On("NamespaceGet", ctx, device.TenantID).
		Return(namespace, nil).Once()
	mock.On("DeviceGetByUID", ctx, models.UID(device.UID), device.TenantID).
		Return(device, nil).Once()
	mock.On("DeviceGetByName", ctx, renamedDevice.Name, device.TenantID).
		Return(nil, nil).Once()
	mock.On("DeviceRename", ctx, models.UID(device.UID), renamedDevice.Name).
		Return(nil).Once()

	err := s.RenameDevice(ctx, models.UID(device.UID), renamedDevice.Name, device.TenantID, user.Username)
	assert.NoError(t, err)

	// Tests to user doesnt owner namespace
	userDoesnotOwner := &models.User{Name: "name1", Email: "email1", Username: "username1", ID: "id1"}

	mock.On("UserGetByUsername", ctx, userDoesnotOwner.Username).
		Return(userDoesnotOwner, nil).Once()
	mock.On("NamespaceGet", ctx, device.TenantID).
		Return(namespace, nil).Once()

	err = s.RenameDevice(ctx, models.UID(device.UID), renamedDevice.Name, device.TenantID, userDoesnotOwner.Username)
	assert.EqualError(t, err, "unauthorized")

	mock.AssertExpectations(t)
}

func TestLookupDevice(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	device := &models.Device{UID: "uid", Name: "name", TenantID: "tenant"}

	mock.On("DeviceLookup", ctx, device.TenantID, device.UID).
		Return(device, nil).Once()

	returnedDevice, err := s.LookupDevice(ctx, device.TenantID, device.UID)
	assert.NoError(t, err)
	assert.Equal(t, device, returnedDevice)

	mock.AssertExpectations(t)
}

func TestUpdateDeviceStatus(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	mock.On("DeviceSetOnline", ctx, models.UID("uid"), true).
		Return(nil).Once()

	err := s.UpdateDeviceStatus(ctx, "uid", true)
	assert.NoError(t, err)

	mock.AssertExpectations(t)
}

func TestUpdatePendingStatus(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	user := &models.User{Name: "name", Email: "", Username: "username", ID: "id"}
	namespace := &models.Namespace{Name: "group1", Owner: "id", TenantID: "tenant", MaxDevices: -1}
	identity := &models.DeviceIdentity{MAC: "mac"}
	device := &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: identity}
	oldDevice := &models.Device{UID: "old_uid", Name: "name", TenantID: "tenant", Identity: identity}
	ctx := context.TODO()

	mock.On("UserGetByUsername", ctx, user.Username).
		Return(user, nil).Once()
	mock.On("NamespaceGet", ctx, device.TenantID).
		Return(namespace, nil).Once()
	mock.On("DeviceGetByUID", ctx, models.UID(device.UID), device.TenantID).
		Return(device, nil).Once()
	mock.On("DeviceGetByMac", ctx, "mac", device.TenantID, "accepted").
		Return(oldDevice, nil).Once()
	mock.On("SessionUpdateDeviceUID", ctx, models.UID(oldDevice.UID), models.UID(device.UID)).
		Return(nil).Once()
	mock.On("DeviceDelete", ctx, models.UID(oldDevice.UID)).
		Return(nil).Once()
	mock.On("DeviceRename", ctx, models.UID(device.UID), oldDevice.Name).
		Return(nil).Once()
	mock.On("DeviceUpdateStatus", ctx, models.UID(device.UID), "accepted").
		Return(nil).Once()

	err := s.UpdatePendingStatus(ctx, models.UID("uid"), "accepted", "tenant", user.Username)
	assert.NoError(t, err)

	// Tests to user doesnt owner namespace
	userDoesnotOwner := &models.User{Name: "name1", Email: "email1", Username: "username1", ID: "id1"}

	mock.On("UserGetByUsername", ctx, userDoesnotOwner.Username).
		Return(userDoesnotOwner, nil).Once()
	mock.On("NamespaceGet", ctx, device.TenantID).
		Return(namespace, nil).Once()

	err = s.UpdatePendingStatus(ctx, models.UID("uid"), "accepted", "tenant", userDoesnotOwner.Username)
	assert.EqualError(t, err, "unauthorized")

	// Would exceed limit
	mock.On("UserGetByUsername", ctx, user.Username).
		Return(user, nil).Once()
	mock.On("NamespaceGet", ctx, device.TenantID).
		Return(namespace, nil).Once()
	mock.On("DeviceGetByUID", ctx, models.UID(device.UID), device.TenantID).
		Return(device, nil).Once()
	mock.On("DeviceGetByMac", ctx, "mac", device.TenantID, "accepted").
		Return(nil, nil).Once()
	mock.On("NamespaceGet", ctx, device.TenantID).
		Return(namespace, nil).Once()
	mock.On("DeviceUpdateStatus", ctx, models.UID(device.UID), "accepted").
		Return(nil).Once()

	err = s.UpdatePendingStatus(ctx, models.UID("uid"), "accepted", "tenant", user.Username)
	assert.NoError(t, err)

	mock.AssertExpectations(t)
}
