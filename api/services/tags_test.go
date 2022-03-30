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

func TestGetTags(t *testing.T) {
	locator := &mocksGeoIp.Locator{}
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, locator)

	ctx := context.TODO()

	Err := errors.New("", "", 0)

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
			name:     "fail when namespace is not found",
			tenantID: "not_found_tenant",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "not_found_tenant").Return(nil, Err).Once()
			},
			expected: Expected{
				Tags:  nil,
				Count: 0,
				Error: NewErrNamespaceNotFound("not_found_tenant", Err),
			},
		},
		{
			name:     "fail when store function to get tags fails",
			tenantID: device.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("TagsGet", ctx, "tenant").Return(nil, 0, Err).Once()
			},
			expected: Expected{
				Tags:  nil,
				Count: 0,
				Error: Err,
			},
		},
		{
			name:     "success to get tags",
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

			tags, count, err := s.GetTags(ctx, tc.tenantID)
			assert.Equal(t, tc.expected, Expected{tags, count, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestRenameTag(t *testing.T) {
	locator := &mocksGeoIp.Locator{}
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, locator)

	ctx := context.TODO()

	Err := errors.New("", "", 0)

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
		name          string
		tenantID      string
		currentTag    string
		newTag        string
		requiredMocks func()
		expected      error
	}{
		{
			name:          "fail when tag is invalid",
			tenantID:      "tenant",
			currentTag:    "currentTag",
			newTag:        "invalid_tag",
			requiredMocks: func() {},
			expected:      NewErrTagInvalid("invalid_tag", nil),
		},
		{
			name:       "fail when device has no tags",
			tenantID:   "namespaceTenantIDNoTag",
			currentTag: "device3",
			newTag:     "device1",
			requiredMocks: func() {
				mock.On("TagsGet", ctx, "namespaceTenantIDNoTag").Return(nil, 0, Err)
			},
			expected: NewErrTagEmpty("namespaceTenantIDNoTag", Err),
		},
		{
			name:       "fail when device don't have the tag",
			tenantID:   namespace.TenantID,
			currentTag: "device2",
			newTag:     "device1",
			requiredMocks: func() {
				mock.On("TagsGet", ctx, namespace.TenantID).Return(deviceWithTags.Tags, len(deviceWithTags.Tags), nil).Once()
			},
			expected: NewErrTagNotFound("device2", nil),
		},
		{
			name:       "fail when device already have the tag",
			tenantID:   namespace.TenantID,
			currentTag: "device3",
			newTag:     "device5",
			requiredMocks: func() {
				mock.On("TagsGet", ctx, namespace.TenantID).Return(deviceWithTags.Tags, len(deviceWithTags.Tags), nil).Once()
			},
			expected: NewErrTagDuplicated("device5", nil),
		},
		{
			name:       "fail when the store function to rename the tag fails",
			tenantID:   namespace.TenantID,
			currentTag: "device3",
			newTag:     "device1",
			requiredMocks: func() {
				mock.On("TagsGet", ctx, namespace.TenantID).Return(deviceWithTags.Tags, len(deviceWithTags.Tags), nil).Once()
				mock.On("TagRename", ctx, namespace.TenantID, "device3", "device1").Return(nil).Once()
			},
			expected: nil,
		},
		{
			name:       "success to rename the tag",
			tenantID:   namespace.TenantID,
			currentTag: "device3",
			newTag:     "device1",
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

			err := s.RenameTag(ctx, tc.tenantID, tc.currentTag, tc.newTag)
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

	Err := errors.New("", "", 0)

	device := &models.Device{UID: "uid", Namespace: "namespace", TenantID: "tenant", Tags: []string{"device1", "device2"}}

	namespace := &models.Namespace{Name: "namespace", TenantID: "tenant"}

	cases := []struct {
		name          string
		tag           string
		tenant        string
		requiredMocks func()
		expected      error
	}{
		{
			name:   "fail when tag is invalid",
			tag:    "invalid_tag",
			tenant: device.TenantID,
			requiredMocks: func() {
			},
			expected: NewErrTagInvalid("invalid_tag", nil),
		},
		{
			name:   "fail when could not find the namespace",
			tag:    "device1",
			tenant: "not_found_tenant",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "not_found_tenant").Return(nil, Err).Once()
			},
			expected: NewErrNamespaceNotFound("not_found_tenant", Err),
		},
		{
			name:   "fail when tags are empty",
			tag:    "device1",
			tenant: device.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("TagsGet", ctx, "tenant").Return(nil, 0, Err).Once()
			},
			expected: NewErrTagEmpty("tenant", Err),
		},
		{
			name:   "fail when tag does not exist",
			tag:    "device3",
			tenant: device.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("TagsGet", ctx, "tenant").Return(device.Tags, len(device.Tags), nil).Once()
			},
			expected: NewErrTagNotFound("device3", nil),
		},
		{
			name:   "fail when the store function to delete the tag fails",
			tag:    "device1",
			tenant: device.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("TagsGet", ctx, "tenant").Return(device.Tags, len(device.Tags), nil).Once()
				mock.On("TagDelete", ctx, "tenant", "device1").Return(Err).Once()
			},
			expected: Err,
		},
		{
			name:   "success to delete tags",
			tag:    "device1",
			tenant: device.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("TagsGet", ctx, "tenant").Return(device.Tags, len(device.Tags), nil).Once()
				mock.On("TagDelete", ctx, "tenant", "device1").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()

			err := s.DeleteTag(ctx, tc.tenant, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
