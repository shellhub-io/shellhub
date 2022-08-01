package services

import (
	"context"
	"testing"

	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/errors"
	mocksGeoIp "github.com/shellhub-io/shellhub/pkg/geoip/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

const (
	invalidUID = "Fails to find the device invalid uid"
)

func TestCreateTag(t *testing.T) {
	locator := &mocksGeoIp.Locator{}
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, locator)

	ctx := context.TODO()

	err := errors.New("error", "", 0)

	device := &models.Device{UID: "uid", TenantID: "tenant", Tags: []string{"device1"}}

	cases := []struct {
		name          string
		uid           models.UID
		deviceName    string
		requiredMocks func()
		expected      error
	}{
		{
			name:       "Fails to find the device invalid uid",
			uid:        "invalid_uid",
			deviceName: "device1",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID("invalid_uid")).Return(nil, err).Once()
			},
			expected: NewErrDeviceNotFound(models.UID("invalid_uid"), err),
		},
		{
			name:       "Fails duplicated name",
			uid:        models.UID(device.UID),
			deviceName: "device1",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID(device.UID)).Return(device, nil).Once()
			},
			expected: NewErrTagDuplicated("device1", nil),
		},
		{
			name:       "Successful create a tag for the device",
			uid:        models.UID(device.UID),
			deviceName: "device6",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID(device.UID)).Return(device, nil).Once()
				mock.On("DeviceCreateTag", ctx, models.UID(device.UID), "device6").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			err := s.CreateDeviceTag(ctx, tc.uid, tc.deviceName)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestRemoveTag(t *testing.T) {
	locator := &mocksGeoIp.Locator{}
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, locator)

	ctx := context.TODO()

	err := errors.New("error", "", 0)

	device := &models.Device{UID: "uid", TenantID: "tenant", Tags: []string{"device1"}}

	cases := []struct {
		name          string
		uid           models.UID
		deviceName    string
		requiredMocks func()
		expected      error
	}{
		{
			name:       invalidUID,
			uid:        "invalid_uid",
			deviceName: "device1",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID("invalid_uid")).Return(nil, err).Once()
			},
			expected: NewErrDeviceNotFound(models.UID("invalid_uid"), err),
		},
		{
			name:       "fail when device does not contain the tag",
			uid:        models.UID(device.UID),
			deviceName: "device2",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID(device.UID)).Return(device, nil).Once()
			},
			expected: NewErrTagNotFound("device2", nil),
		},
		{
			name:       "fail delete a tag",
			uid:        models.UID(device.UID),
			deviceName: "device1",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID(device.UID)).Return(device, nil).Once()
				mock.On("DeviceRemoveTag", ctx, models.UID(device.UID), "device1").Return(err).Once()
			},
			expected: err,
		},
		{
			name:       "successful delete a tag",
			uid:        models.UID(device.UID),
			deviceName: "device1",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID(device.UID)).Return(device, nil).Once()
				mock.On("DeviceRemoveTag", ctx, models.UID(device.UID), "device1").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			err := s.RemoveDeviceTag(ctx, tc.uid, tc.deviceName)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestUpdateTag(t *testing.T) {
	locator := &mocksGeoIp.Locator{}
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, locator)

	ctx := context.TODO()

	err := errors.New("error", "", 0)

	device := &models.Device{UID: "uid", TenantID: "tenant"}

	tags := []string{"device1", "device2", "device3"}

	cases := []struct {
		name          string
		uid           models.UID
		tags          []string
		requiredMocks func()
		expected      error
	}{
		{
			name: invalidUID,
			uid:  "invalid_uid",
			tags: tags,
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID("invalid_uid")).Return(nil, err).Once()
			},
			expected: NewErrDeviceNotFound("invalid_uid", err),
		},
		{
			name: "Successful create tags for the device",
			uid:  models.UID(device.UID),
			tags: tags,
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID(device.UID)).Return(device, nil).Once()
				mock.On("DeviceUpdateTag", ctx, models.UID(device.UID), tags).Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			err := s.UpdateDeviceTag(ctx, tc.uid, tc.tags)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
