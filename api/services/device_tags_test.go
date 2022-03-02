package services

import (
	"context"
	"testing"

	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	mocksGeoIp "github.com/shellhub-io/shellhub/pkg/geoip/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

const (
	invalidUID    = "Fails to find the device invalid uid"
	invalidFormat = "Fails invalid format for name"
)

func TestCreateTag(t *testing.T) {
	locator := &mocksGeoIp.Locator{}
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, locator)

	ctx := context.TODO()

	device := &models.Device{UID: "uid", TenantID: "tenant", Tags: []string{"device1"}}

	device2 := &models.Device{UID: "uid2", TenantID: "tenant2", Tags: []string{"device1", "device2", "device3"}}

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
				mock.On("DeviceGet", ctx, models.UID("invalid_uid")).Return(nil, ErrDeviceNotFound).Once()
			},
			expected: ErrDeviceNotFound,
		},
		{
			name:          "Fails invalid format for name",
			uid:           models.UID(device.UID),
			deviceName:    "de",
			requiredMocks: func() {},
			expected:      ErrInvalidFormat,
		},
		{
			name:       "Fails duplicated name",
			uid:        models.UID(device.UID),
			deviceName: "device1",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID(device.UID)).Return(device, nil).Once()
			},
			expected: ErrDuplicateTagName,
		},
		{
			name:       "Fails max capacity reached",
			uid:        models.UID(device2.UID),
			deviceName: "device6",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID(device2.UID)).Return(device2, nil).Once()
			},
			expected: ErrMaxTagReached,
		},
		{
			name:       "successful create a tag for the device",
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
				mock.On("DeviceGet", ctx, models.UID("invalid_uid")).Return(nil, ErrDeviceNotFound).Once()
			},
			expected: ErrDeviceNotFound,
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

	device := &models.Device{UID: "uid", TenantID: "tenant"}

	tags := []string{"device1", "device2", "device3"}

	maxReachedTags := []string{"device1", "device2", "device3", "device4"}

	invalidTag := []string{"de"}

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
				mock.On("DeviceGet", ctx, models.UID("invalid_uid")).Return(nil, ErrDeviceNotFound).Once()
			},
			expected: ErrDeviceNotFound,
		},
		{
			name:          invalidFormat,
			uid:           models.UID(device.UID),
			tags:          invalidTag,
			requiredMocks: func() {},
			expected:      ErrInvalidFormat,
		},
		{
			name: "Fails max capacity reached",
			uid:  models.UID(device.UID),
			tags: maxReachedTags,
			requiredMocks: func() {
			},
			expected: ErrMaxTagReached,
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
