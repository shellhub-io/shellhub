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
		models.Device{UID: "uid"},
	}

	filters := []models.Filter{
		models.Filter{
			Type:   "property",
			Params: &models.PropertyParams{Name: "hostname", Operator: "eq"}},
	}

	filterJSON, err := json.Marshal(filters)
	assert.NoError(t, err)

	encodedFilter := base64.StdEncoding.EncodeToString(filterJSON)

	query := paginator.Query{Page: 1, PerPage: 10}

	mock.On("ListDevices", ctx, query, filters, "accepted").
		Return(devices, len(devices), nil).Once()

	returnedDevices, count, err := s.ListDevices(ctx, query, encodedFilter, "accepted")
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
	mock.On("GetDevice", ctx, models.UID(device.UID)).
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

	device := &models.Device{UID: "uid", TenantID: "tenant"}

	mock.On("DeleteDevice", ctx, models.UID(device.UID)).
		Return(nil).Once()
	mock.On("GetDeviceByUID", ctx, models.UID(device.UID), device.TenantID).
		Return(device, nil).Once()

	err := s.DeleteDevice(ctx, models.UID(device.UID), device.TenantID)
	assert.NoError(t, err)

	mock.AssertExpectations(t)
}

func TestRenameDevice(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	device := &models.Device{UID: "uid", Name: "name", TenantID: "tenant"}
	renamedDevice := &models.Device{UID: "uid", Name: "rename", TenantID: "tenant"}

	mock.On("GetDeviceByUID", ctx, models.UID(device.UID), device.TenantID).
		Return(device, nil).Once()
	mock.On("GetDeviceByName", ctx, renamedDevice.Name, device.TenantID).
		Return(nil, nil).Once()
	mock.On("RenameDevice", ctx, models.UID(device.UID), renamedDevice.Name).
		Return(nil).Once()

	err := s.RenameDevice(ctx, models.UID(device.UID), renamedDevice.Name, device.TenantID)
	assert.NoError(t, err)

	mock.AssertExpectations(t)
}

func TestLookupDevice(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	device := &models.Device{UID: "uid", Name: "name", TenantID: "tenant"}

	mock.On("LookupDevice", ctx, device.TenantID, device.UID).
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

	mock.On("UpdateDeviceStatus", ctx, models.UID("uid"), true).
		Return(nil).Once()

	err := s.UpdateDeviceStatus(ctx, "uid", true)
	assert.NoError(t, err)

	mock.AssertExpectations(t)
}

func TestUpdatePendingStatus(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	mock.On("UpdatePendingStatus", ctx, models.UID("uid"), "accepted").
		Return(nil).Once()

	err := s.UpdatePendingStatus(ctx, "uid", "accepted")
	assert.NoError(t, err)

	mock.AssertExpectations(t)
}
