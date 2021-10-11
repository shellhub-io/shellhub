package services

import (
	"context"
	"errors"
	"testing"

	mocksGeoIp "github.com/shellhub-io/shellhub/pkg/geoip/mocks"

	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
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
	device2 := &models.Device{UID: "uid2", TenantID: "tenant2", Tags: []string{"device1", "device2", "device3", "device4", "device5"}}

	Err := errors.New("error")

	cases := []struct {
		name          string
		requiredMocks func()
		uid           models.UID
		expected      error
	}{
		{
			name: invalidUID,
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID("invalid_uid")).Return(nil, ErrDeviceNotFound).Once()
			},
			uid:      "invalid_uid",
			expected: ErrDeviceNotFound,
		},
		{
			name:          invalidFormat,
			requiredMocks: func() {},
			uid:           models.UID(device.UID),
			expected:      ErrInvalidFormat,
		},
		{
			name: "successful create a tag for the device",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID(device.UID)).Return(device, nil).Once()
				mock.On("DeviceCreateTag", ctx, models.UID(device.UID), "device6").Return(nil).Once()
			},
			uid:      models.UID(device.UID),
			expected: nil,
		},
		{
			name: "Fails duplicated name",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID(device.UID)).Return(device, nil).Once()
			},
			uid:      models.UID(device.UID),
			expected: ErrDuplicateTagName,
		},
		{
			name: "Fails max capacity reached",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID(device2.UID)).Return(device2, nil).Once()
			},
			uid:      models.UID(device2.UID),
			expected: ErrMaxTagReached,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			switch tc.name {
			case invalidUID:
				Err = s.CreateTag(ctx, "invalid_uid", "device1")
			case "Fails invalid format for name":
				Err = s.CreateTag(ctx, models.UID(device.UID), "de")
			case "successful create a tag for the device":
				Err = s.CreateTag(ctx, models.UID(device.UID), "device6")
			case "Fails duplicated name":
				Err = s.CreateTag(ctx, models.UID(device.UID), "device1")
			case "Fails max capacity reached":
				Err = s.CreateTag(ctx, models.UID(device2.UID), "device6")
			}
			assert.Equal(t, tc.expected, Err)
		})
	}

	mock.AssertExpectations(t)
}

func TestDeleteTag(t *testing.T) {
	locator := &mocksGeoIp.Locator{}
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, locator)

	ctx := context.TODO()

	device := &models.Device{UID: "uid", TenantID: "tenant", Tags: []string{"device1"}}

	Err := errors.New("error")

	cases := []struct {
		name          string
		requiredMocks func()
		uid           models.UID
		expected      error
	}{
		{
			name: invalidUID,
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID("invalid_uid")).Return(nil, ErrDeviceNotFound).Once()
			},
			uid:      "invalid_uid",
			expected: ErrDeviceNotFound,
		},
		{
			name: "successful delete a tag",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID(device.UID)).Return(device, nil).Once()
				mock.On("DeviceDeleteTag", ctx, models.UID(device.UID), "device1").Return(nil).Once()
			},
			uid:      models.UID(device.UID),
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			switch tc.name {
			case invalidUID:
				Err = s.DeleteTag(ctx, "invalid_uid", "device1")
			case "successful delete a tag":
				Err = s.DeleteTag(ctx, models.UID(device.UID), "device1")
			}
			assert.Equal(t, tc.expected, Err)
		})
	}

	mock.AssertExpectations(t)
}

func TestRenameTag(t *testing.T) {
	locator := &mocksGeoIp.Locator{}
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, locator)

	ctx := context.TODO()

	Err := errors.New("error")

	cases := []struct {
		name          string
		requiredMocks func()
		tenantID      string
		expected      error
	}{
		{
			name:          invalidFormat,
			requiredMocks: func() {},
			tenantID:      "tenant",
			expected:      ErrInvalidFormat,
		},
		{
			name: "successful rename a tag",
			requiredMocks: func() {
				mock.On("DeviceRenameTag", ctx, "tenant", "device3", "device1").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			switch tc.name {
			case invalidFormat:
				Err = s.RenameTag(ctx, "tenant", "device3", "de/@&:")
			case "successful rename a tag":
				Err = s.RenameTag(ctx, "tenant", "device3", "device1")
			}
			assert.Equal(t, tc.expected, Err)
		})
	}

	mock.AssertExpectations(t)
}

func TestListTag(t *testing.T) {
	locator := &mocksGeoIp.Locator{}
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, locator)

	ctx := context.TODO()

	device := &models.Device{UID: "uid", TenantID: "tenant", Tags: []string{"device1", "device2"}}

	type Expected struct {
		Tags  []string
		Count int
		Error error
	}

	cases := []struct {
		name          string
		requiredMocks func()
		uid           models.UID
		tenantID      string
		expected      Expected
	}{
		{
			name: "successful list tags",
			requiredMocks: func() {
				mock.On("DeviceListTag", ctx).Return(device.Tags, len(device.Tags), nil).Once()
			},
			uid:      models.UID(device.UID),
			tenantID: device.TenantID,
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
			returnedTags, count, err := s.ListTag(ctx)
			assert.Equal(t, tc.expected, Expected{returnedTags, count, err})
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

	tags := []string{"device1", "device2", "device3", "device4", "device5"}

	duplicatedTags := []string{"device1", "device1"}

	maxReachedTags := []string{"device1", "device2", "device3", "device4", "device5", "device6"}

	invalidTag := []string{"de"}

	Err := errors.New("error")

	cases := []struct {
		name          string
		requiredMocks func()
		uid           models.UID
		tags          []string
		expected      error
	}{
		{
			name: invalidUID,
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID("invalid_uid")).Return(nil, ErrDeviceNotFound).Once()
			},
			uid:      "invalid_uid",
			tags:     tags,
			expected: ErrDeviceNotFound,
		},
		{
			name:          invalidFormat,
			requiredMocks: func() {},
			uid:           models.UID(device.UID),
			expected:      ErrInvalidFormat,
		},
		{
			name: "Fails duplicated name",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID(device.UID)).Return(device, nil).Once()
				mock.On("DeviceUpdateTag", ctx, models.UID(device.UID), duplicatedTags).Return(ErrDuplicateTagName).Once()
			},
			uid:      models.UID(device.UID),
			tags:     duplicatedTags,
			expected: ErrDuplicateTagName,
		},
		{
			name: "Fails max capacity reached",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID(device.UID)).Return(device, nil).Once()
				mock.On("DeviceUpdateTag", ctx, models.UID(device.UID), maxReachedTags).Return(ErrMaxTagReached).Once()
			},
			uid:      models.UID(device.UID),
			tags:     maxReachedTags,
			expected: ErrMaxTagReached,
		},
		{
			name: "successful create tags for the device",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID(device.UID)).Return(device, nil).Once()
				mock.On("DeviceUpdateTag", ctx, models.UID(device.UID), tags).Return(nil).Once()
			},
			uid:      models.UID(device.UID),
			tags:     tags,
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			switch tc.name {
			case invalidUID:
				Err = s.UpdateTag(ctx, "invalid_uid", tags)
			case "Fails invalid format for name":
				Err = s.UpdateTag(ctx, models.UID(device.UID), invalidTag)
			case "Fails duplicated name":
				Err = s.UpdateTag(ctx, models.UID(device.UID), duplicatedTags)
			case "Fails max capacity reached":
				Err = s.UpdateTag(ctx, models.UID(device.UID), maxReachedTags)
			case "successful create tags for the device":
				Err = s.UpdateTag(ctx, models.UID(device.UID), tags)
			}
			assert.Equal(t, tc.expected, Err)
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
		requiredMocks func()
		uid           models.UID
		tenantID      string
		expected      Expected
	}{
		{
			name: "fail find the namespace",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "not_found_tenant").Return(nil, ErrNotFound).Once()
			},
			tenantID: "not_found_tenant",
			expected: Expected{
				Tags:  nil,
				Count: 0,
				Error: ErrNotFound,
			},
		},
		{
			name: "successful get tags",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("DeviceGetTags", ctx, "tenant").Return(device.Tags, len(device.Tags), nil).Once()
			},
			tenantID: device.TenantID,
			expected: Expected{
				Tags:  device.Tags,
				Count: len(device.Tags),
				Error: nil,
			},
		},
	}

	var returnedTags []string
	var count int
	var err error

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			switch tc.name {
			case "fail find the namespace":
				returnedTags, count, err = s.GetTags(ctx, "not_found_tenant")
			case "successful get tags":
				returnedTags, count, err = s.GetTags(ctx, "tenant")
			}
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
		requiredMocks func()
		uid           models.UID
		tenantID      string
		expected      error
	}{
		{
			name: "fail find the namespace",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "not_found_tenant").Return(nil, ErrNotFound).Once()
			},
			tenantID: "not_found_tenant",
			expected: ErrNotFound,
		},
		{
			name: "successful get tags",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("DeviceDeleteAllTags", ctx, "tenant", "device1").Return(nil).Once()
			},
			tenantID: device.TenantID,
			expected: nil,
		},
	}

	var err error

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			switch tc.name {
			case "fail find the namespace":
				err = s.DeleteAllTags(ctx, "not_found_tenant", "device1")
			case "successful get tags":
				err = s.DeleteAllTags(ctx, "tenant", "device1")
			}
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
