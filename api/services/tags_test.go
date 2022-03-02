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
				mock.On("TagsGet", ctx, "tenant").Return(device.Tags, len(device.Tags), nil).Once()
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
				mock.On("TagsGet", ctx, "namespaceTenantIDNoTag").Return(nil, 0, Err)
			},
			expected: ErrNoTags,
		},
		{
			name:              "Fail when device don't have a tag",
			tenantID:          namespace.TenantID,
			currentDeviceName: "device2",
			newDeviceName:     "device1",
			requiredMocks: func() {
				mock.On("TagsGet", ctx, namespace.TenantID).Return(deviceWithTags.Tags, len(deviceWithTags.Tags), nil).Once()
			},
			expected: ErrTagNameNotFound,
		},
		{
			name:              "Fail when device already have a tag",
			tenantID:          namespace.TenantID,
			currentDeviceName: "device3",
			newDeviceName:     "device5",
			requiredMocks: func() {
				mock.On("TagsGet", ctx, namespace.TenantID).Return(deviceWithTags.Tags, len(deviceWithTags.Tags), nil).Once()
			},
			expected: ErrDuplicateTagName,
		},
		{
			name:              "Successful rename a tag",
			tenantID:          namespace.TenantID,
			currentDeviceName: "device3",
			newDeviceName:     "device1",
			requiredMocks: func() {
				mock.On("TagsGet", ctx, namespace.TenantID).Return(deviceWithTags.Tags, len(deviceWithTags.Tags), nil).Once()
				mock.On("TagRename", ctx, namespace.TenantID, "device3", "device1").Return(nil).Once()
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

func TestDeleteTag(t *testing.T) {
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
				mock.On("TagDelete", ctx, "tenant", "device1").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			err := s.DeleteTag(ctx, tc.tenantID, tc.deviceName)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
