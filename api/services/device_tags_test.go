package services

import (
	"context"
	"errors"
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
			err := s.CreateTag(ctx, tc.uid, tc.deviceName)
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
			err := s.RemoveTag(ctx, tc.uid, tc.deviceName)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestRenameTag(t *testing.T) {
	locator := &mocksGeoIp.Locator{}
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, locator)

	ctx := context.TODO()

	Err := errors.New("")

	deviceWithTags := &models.Device{
		UID:      "deviceWithTagsUID",
		Name:     "deviceWithTagsName",
		TenantID: "deviceWithTagsTenantID",
		Tags:     []string{"device3", "device4", "device5"},
	}

	namespace := &models.Namespace{
		Name:     "namespaceName",
		Owner:    "owner",
		TenantID: "namespaceTenantID",
	}

	cases := []struct {
		name              string
		tenantID          string
		currentDeviceName string
		newDeviceName     string
		requiredMocks     func()
		expected          error
	}{
		{
			name:          invalidFormat,
			tenantID:      "tenant",
			requiredMocks: func() {},
			expected:      ErrInvalidFormat,
		},
		{
			name:              "Fail when device has no tags",
			tenantID:          "namespaceTenantIDNoTag",
			currentDeviceName: "device3",
			newDeviceName:     "device1",
			requiredMocks: func() {
				mock.On("DeviceGetTags", ctx, "namespaceTenantIDNoTag").Return(nil, 0, Err)
			},
			expected: ErrNoTags,
		},
		{
			name:              "Fail when device don't have a tag",
			tenantID:          namespace.TenantID,
			currentDeviceName: "device2",
			newDeviceName:     "device1",
			requiredMocks: func() {
				mock.On("DeviceGetTags", ctx, namespace.TenantID).Return(deviceWithTags.Tags, len(deviceWithTags.Tags), nil).Once()
			},
			expected: ErrTagNameNotFound,
		},
		{
			name:              "Fail when device already have a tag",
			tenantID:          namespace.TenantID,
			currentDeviceName: "device3",
			newDeviceName:     "device5",
			requiredMocks: func() {
				mock.On("DeviceGetTags", ctx, namespace.TenantID).Return(deviceWithTags.Tags, len(deviceWithTags.Tags), nil).Once()
			},
			expected: ErrDuplicateTagName,
		},
		{
			name:              "Successful rename a tag",
			tenantID:          namespace.TenantID,
			currentDeviceName: "device3",
			newDeviceName:     "device1",
			requiredMocks: func() {
				mock.On("DeviceGetTags", ctx, namespace.TenantID).Return(deviceWithTags.Tags, len(deviceWithTags.Tags), nil).Once()
				mock.On("DeviceRenameTag", ctx, namespace.TenantID, "device3", "device1").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			err := s.RenameTag(ctx, tc.tenantID, tc.currentDeviceName, tc.newDeviceName)
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

	duplicatedTags := []string{"device1", "device1"}

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
			name: "Fails duplicated name",
			uid:  models.UID(device.UID),
			tags: duplicatedTags,
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID(device.UID)).Return(device, nil).Once()
				mock.On("DeviceUpdateTag", ctx, models.UID(device.UID), duplicatedTags).Return(ErrDuplicateTagName).Once()
			},
			expected: ErrDuplicateTagName,
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
			err := s.UpdateTag(ctx, tc.uid, tc.tags)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestGetTags(t *testing.T) {
	locator := &mocksGeoIp.Locator{}
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, locator)

	ctx := context.TODO()

	device := &models.Device{UID: "uid", Namespace: "namespace", TenantID: "tenant", Tags: []string{"device1", "device2"}}

	namespace := &models.Namespace{Name: "namespace", TenantID: "tenant"}

	type Expected struct {
		Tags  []string
		Count int
		Error error
	}

	cases := []struct {
		name          string
		uid           models.UID
		tenantID      string
		requiredMocks func()
		expected      Expected
	}{
		{
			name:     "fail find the namespace",
			tenantID: "not_found_tenant",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "not_found_tenant").Return(nil, ErrNotFound).Once()
			},
			expected: Expected{
				Tags:  nil,
				Count: 0,
				Error: ErrNamespaceNotFound,
			},
		},
		{
			name:     "successful get tags",
			tenantID: device.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("DeviceGetTags", ctx, "tenant").Return(device.Tags, len(device.Tags), nil).Once()
			},
			expected: Expected{
				Tags:  device.Tags,
				Count: len(device.Tags),
				Error: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			returnedTags, count, err := s.GetTags(ctx, tc.tenantID)
			assert.Equal(t, tc.expected, Expected{returnedTags, count, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestDeleteAllTags(t *testing.T) {
	locator := &mocksGeoIp.Locator{}
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, locator)

	ctx := context.TODO()

	device := &models.Device{UID: "uid", Namespace: "namespace", TenantID: "tenant", Tags: []string{"device1", "device2"}}

	namespace := &models.Namespace{Name: "namespace", TenantID: "tenant"}

	cases := []struct {
		name          string
		deviceName    string
		tenantID      string
		requiredMocks func()
		expected      error
	}{
		{
			name:       "fail find the namespace",
			deviceName: "device1",
			tenantID:   "not_found_tenant",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "not_found_tenant").Return(nil, ErrNotFound).Once()
			},
			expected: ErrNamespaceNotFound,
		},
		{
			name:       "successful get tags",
			deviceName: "device1",
			tenantID:   device.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("DeviceDeleteTags", ctx, "tenant", "device1").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			err := s.DeleteTags(ctx, tc.tenantID, tc.deviceName)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
