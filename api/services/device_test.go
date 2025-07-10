package services

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/envs"
	envsmocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListDevices(t *testing.T) {
	storeMock := new(storemock.Store)

	type Expected struct {
		devices []models.Device
		count   int
		err     error
	}

	cases := []struct {
		description   string
		req           *requests.DeviceList
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails to list devices",
			req: &requests.DeviceList{
				TenantID:     "",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableIfNotAccepted).
					Return([]models.Device{}, 0, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     errors.New("error", "", 0),
			},
		},
		{
			description: "succeeds to list devices",
			req: &requests.DeviceList{
				TenantID:     "",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableIfNotAccepted).
					Return([]models.Device{}, 0, nil).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     nil,
			},
		},
	}

	service := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)

			devices, count, err := service.ListDevices(ctx, tc.req)
			require.Equal(tt, tc.expected, Expected{devices, count, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestListDevices_status_removed(t *testing.T) {
	storeMock := new(storemock.Store)

	type Expected struct {
		devices []models.Device
		count   int
		err     error
	}

	cases := []struct {
		description   string
		req           *requests.DeviceList
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails when could not list the removed devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusRemoved,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusRemoved, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableFromRemoved).
					Return([]models.Device{}, 0, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     errors.New("error", "", 0),
			},
		},
		{
			description: "succeeds to list the removed devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusRemoved,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusRemoved, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableFromRemoved).
					Return([]models.Device{{Name: "dev"}}, 1, nil).
					Once()
			},
			expected: Expected{
				devices: []models.Device{
					{
						Name: "dev",
					},
				},
				count: 1,
				err:   nil,
			},
		},
	}

	service := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)

			devices, count, err := service.ListDevices(ctx, tc.req)
			require.Equal(tt, tc.expected, Expected{devices, count, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestListDevices_tenant_not_empty(t *testing.T) {
	storeMock := new(storemock.Store)

	type Expected struct {
		devices []models.Device
		count   int
		err     error
	}

	cases := []struct {
		description   string
		req           *requests.DeviceList
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails when the namespace does not exists",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				devices: nil,
				count:   0,
				err:     NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", errors.New("error", "", 0)),
			},
		},
		{
			description: "[cloud] fails when the namespace reached the device limit and cannot list the devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", MaxDevices: 3, DevicesAcceptedCount: 2, DevicesRemovedCount: 1}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableFromRemoved).
					Return([]models.Device{}, 0, errors.New("error", "layer", 0)).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     errors.New("error", "layer", 0),
			},
		},
		{
			description: "[cloud] succeeds when the namespace reached the device limit and list the devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", MaxDevices: 3, DevicesAcceptedCount: 2, DevicesRemovedCount: 1}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableFromRemoved).
					Return([]models.Device{}, 0, nil).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     nil,
			},
		},
		{
			description: "[cloud] fails when the namespace do not reached the device limit and cannot list the devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", MaxDevices: 3, DevicesAcceptedCount: 2, DevicesRemovedCount: 0}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableIfNotAccepted).
					Return([]models.Device{}, 0, errors.New("error", "layer", 0)).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     errors.New("error", "layer", 0),
			},
		},
		{
			description: "[cloud] succeeds when the namespace do not reached the device limit and list the devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", MaxDevices: 3, DevicesAcceptedCount: 2, DevicesRemovedCount: 0}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableIfNotAccepted).
					Return([]models.Device{}, 0, nil).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     nil,
			},
		},
		{
			description: "[enterprise|community] fails when the namespace reached the device limit and cannot list the devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", MaxDevices: 3, DevicesAcceptedCount: 3}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("true").
					Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableAsFalse).
					Return([]models.Device{}, 0, errors.New("error", "layer", 0)).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     errors.New("error", "layer", 0),
			},
		},
		{
			description: "[enterprise|community] succeeds when the namespace reached the device limit and list the devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", MaxDevices: 3, DevicesAcceptedCount: 3}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("true").
					Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableAsFalse).
					Return([]models.Device{}, 0, nil).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     nil,
			},
		},
		{
			description: "[enterprise|community] fails when the namespace do not reached the device limit and cannot list the devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", MaxDevices: 3, DevicesAcceptedCount: 2}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("true").
					Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableIfNotAccepted).
					Return([]models.Device{}, 0, errors.New("error", "layer", 0)).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     errors.New("error", "layer", 0),
			},
		},
		{
			description: "[enterprise|community] succeeds when the namespace do not reached the device limit and list the devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-4000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", MaxDevices: 3, DevicesAcceptedCount: 2}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("true").
					Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableIfNotAccepted).
					Return([]models.Device{}, 0, nil).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     nil,
			},
		},
	}

	service := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)

			devices, count, err := service.ListDevices(ctx, tc.req)
			require.Equal(tt, tc.expected, Expected{devices, count, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestGetDevice(t *testing.T) {
	mock := new(storemock.Store)

	ctx := context.TODO()

	type Expected struct {
		device *models.Device
		err    error
	}

	cases := []struct {
		description   string
		requiredMocks func()
		uid           models.UID
		expected      Expected
	}{
		{
			description: "fails when the store get device fails",
			requiredMocks: func() {
				mock.On("DeviceResolve", ctx, store.DeviceUIDResolver, "_uid").Return(nil, errors.New("error", "", 0)).Once()
			},
			uid: models.UID("_uid"),
			expected: Expected{
				nil,
				NewErrDeviceNotFound(models.UID("_uid"), errors.New("error", "", 0)),
			},
		},
		{
			description: "succeeds",
			requiredMocks: func() {
				device := &models.Device{UID: "uid"}
				mock.On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid").Return(device, nil).Once()
			},
			uid: models.UID("uid"),
			expected: Expected{
				&models.Device{UID: "uid"},
				nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

			returnedDevice, err := service.GetDevice(ctx, tc.uid)
			assert.Equal(t, tc.expected, Expected{returnedDevice, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestResolveDevice(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	ctx := context.TODO()

	type Expected struct {
		device *models.Device
		err    error
	}

	cases := []struct {
		description   string
		requiredMocks func()
		req           *requests.ResolveDevice
		expected      Expected
	}{
		{
			description: "fails when namespace does not exists",
			req:         &requests.ResolveDevice{TenantID: "00000000-0000-0000-0000-000000000000", UID: "uid", Hostname: ""},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				nil,
				NewErrNamespaceNotFound("00000000-0000-0000-0000-000000000000", errors.New("error", "", 0)),
			},
		},
		{
			description: "fails when cannot retrieve a device with the specified UID",
			req:         &requests.ResolveDevice{TenantID: "00000000-0000-0000-0000-000000000000", UID: "uid", Hostname: ""},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(&models.Namespace{Name: "namespace", TenantID: "00000000-0000-0000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid", mock.AnythingOfType("store.QueryOption")).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				nil,
				NewErrDeviceNotFound(models.UID(""), errors.New("error", "", 0)),
			},
		},
		{
			description: "succeeds to fetch a device using UID",
			req:         &requests.ResolveDevice{TenantID: "00000000-0000-0000-0000-000000000000", UID: "uid", Hostname: ""},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(&models.Namespace{Name: "namespace", TenantID: "00000000-0000-0000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid", mock.AnythingOfType("store.QueryOption")).
					Return(&models.Device{UID: "uid"}, nil).
					Once()
			},
			expected: Expected{
				&models.Device{UID: "uid"},
				nil,
			},
		},
		{
			description: "fails when cannot retrieve a device with the specified hostname",
			req:         &requests.ResolveDevice{TenantID: "00000000-0000-0000-0000-000000000000", UID: "", Hostname: "hostname"},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(&models.Namespace{Name: "namespace", TenantID: "00000000-0000-0000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceHostnameResolver, "hostname", mock.AnythingOfType("store.QueryOption")).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				nil,
				NewErrDeviceNotFound(models.UID(""), errors.New("error", "", 0)),
			},
		},
		{
			description: "succeeds to fetch a device using hostname",
			req:         &requests.ResolveDevice{TenantID: "00000000-0000-0000-0000-000000000000", UID: "", Hostname: "hostname"},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(&models.Namespace{Name: "namespace", TenantID: "00000000-0000-0000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceHostnameResolver, "hostname", mock.AnythingOfType("store.QueryOption")).
					Return(&models.Device{UID: "uid"}, nil).
					Once()
			},
			expected: Expected{
				&models.Device{UID: "uid"},
				nil,
			},
		},
		{
			description: "succeeds to fetch a device using uid when both are provided",
			req:         &requests.ResolveDevice{TenantID: "00000000-0000-0000-0000-000000000000", UID: "uid", Hostname: "hostname"},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(&models.Namespace{Name: "namespace", TenantID: "00000000-0000-0000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid", mock.AnythingOfType("store.QueryOption")).
					Return(&models.Device{UID: "uid"}, nil).
					Once()
			},
			expected: Expected{
				&models.Device{UID: "uid"},
				nil,
			},
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			device, err := s.ResolveDevice(ctx, tc.req)
			assert.Equal(t, tc.expected, Expected{device, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestDeleteDevice(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	ctx := context.TODO()

	cases := []struct {
		description   string
		requiredMocks func()
		uid           models.UID
		tenant        string
		expected      error
	}{
		{
			description: "fails when the store device get by uid fails",
			uid:         models.UID("_uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				queryOptionsMock.
					On("InNamespace", "tenant").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "_uid", mock.AnythingOfType("store.QueryOption")).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceNotFound(models.UID("_uid"), errors.New("error", "", 0)),
		},
		{
			description: "fails when the store namespace get fails",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				queryOptionsMock.
					On("InNamespace", "tenant").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:       "uid",
							TenantID:  "tenant",
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrNamespaceNotFound("tenant", errors.New("error", "", 0)),
		},
		{
			description: "fails when the store device delete fails",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				queryOptionsMock.
					On("InNamespace", "tenant").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:       "uid",
							TenantID:  "tenant",
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(
						&models.Namespace{
							Name:     "group1",
							Owner:    "id",
							TenantID: "tenant",
							Members: []models.Member{
								{
									ID:   "id",
									Role: authorizer.RoleOwner,
								},
								{
									ID:   "id2",
									Role: authorizer.RoleObserver,
								},
							},
							MaxDevices: 3,
						},
						nil,
					).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").Return("false").
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("uid")).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "succeeds",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				queryOptionsMock.
					On("InNamespace", "tenant").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:       "uid",
							Status:    models.DeviceStatusAccepted,
							TenantID:  "tenant",
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(
						&models.Namespace{
							Name:     "group1",
							Owner:    "id",
							TenantID: "tenant",
							Members: []models.Member{
								{
									ID:   "id",
									Role: authorizer.RoleOwner,
								},
								{
									ID:   "id2",
									Role: authorizer.RoleObserver,
								},
							},
							MaxDevices: 3,
						},
						nil,
					).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").Return("false").
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("uid")).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "tenant", models.DeviceStatusAccepted, int64(-1)).
					Return(nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "[with_billing] fails when cannot update the device",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				queryOptionsMock.
					On("InNamespace", "tenant").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:       "uid",
							TenantID:  "tenant",
							CreatedAt: time.Time{},
							Status:    models.DeviceStatusAccepted,
						},
						nil,
					).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(
						&models.Namespace{
							Name:     "group1",
							Owner:    "id",
							TenantID: "tenant",
							Members: []models.Member{
								{
									ID:   "id",
									Role: authorizer.RoleOwner,
								},
								{
									ID:   "id2",
									Role: authorizer.RoleObserver,
								},
							},
							MaxDevices: 3,
						},
						nil,
					).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").Return("true").
					Once()
				envMock.
					On("Get", "SHELLHUB_BILLING").Return("true").
					Once()
				storeMock.
					On(
						"DeviceUpdate",
						ctx,
						"tenant",
						"uid",
						&models.DeviceChanges{Status: models.DeviceStatusRemoved},
					).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "[with_billing] succeeds",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				queryOptionsMock.
					On("InNamespace", "tenant").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:       "uid",
							Status:    models.DeviceStatusAccepted,
							TenantID:  "tenant",
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(
						&models.Namespace{
							Name:     "group1",
							Owner:    "id",
							TenantID: "tenant",
							Members: []models.Member{
								{
									ID:   "id",
									Role: authorizer.RoleOwner,
								},
								{
									ID:   "id2",
									Role: authorizer.RoleObserver,
								},
							},
							MaxDevices: 3,
						},
						nil,
					).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").Return("true").
					Once()
				envMock.
					On("Get", "SHELLHUB_BILLING").Return("true").
					Once()
				storeMock.
					On(
						"DeviceUpdate",
						ctx,
						"tenant",
						"uid",
						&models.DeviceChanges{Status: models.DeviceStatusRemoved},
					).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "tenant", models.DeviceStatusRemoved, int64(1)).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "tenant", models.DeviceStatusAccepted, int64(-1)).
					Return(nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "[with_billing] succeeds but device status isn't accepted",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				queryOptionsMock.
					On("InNamespace", "tenant").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:       "uid",
							Status:    models.DeviceStatusPending,
							TenantID:  "tenant",
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "tenant").
					Return(
						&models.Namespace{
							Name:     "group1",
							Owner:    "id",
							TenantID: "tenant",
							Members: []models.Member{
								{
									ID:   "id",
									Role: authorizer.RoleOwner,
								},
								{
									ID:   "id2",
									Role: authorizer.RoleObserver,
								},
							},
							MaxDevices: 3,
						},
						nil,
					).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").Return("true").
					Once()
				envMock.
					On("Get", "SHELLHUB_BILLING").Return("true").
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("uid")).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "tenant", models.DeviceStatusPending, int64(-1)).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			err := service.DeleteDevice(ctx, tc.uid, tc.tenant)
			assert.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestRenameDevice(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	ctx := context.TODO()

	cases := []struct {
		description   string
		requiredMocks func(device *models.Device)
		uid           models.UID
		device        *models.Device
		deviceNewName string
		tenant        string
		expected      error
	}{
		{
			description: "fails when store device get fails",
			tenant:      "tenant",
			uid:         models.UID("uid"),
			device:      &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device) {
				queryOptionsMock.On("InNamespace", "tenant").Return(nil).Once()
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid", mock.AnythingOfType("store.QueryOption")).Return(device, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound(models.UID("uid"), errors.New("error", "", 0)),
		},
		{
			description:   "returns nil if the name is the same",
			tenant:        "tenant",
			deviceNewName: "name",
			uid:           models.UID("uid"),
			device:        &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device) {
				queryOptionsMock.On("InNamespace", "tenant").Return(nil).Once()
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid", mock.AnythingOfType("store.QueryOption")).Return(device, nil).Once()
			},
			expected: nil,
		},
		{
			description:   "fails when store get by device name fails",
			tenant:        "tenant",
			deviceNewName: "newname",
			uid:           models.UID("uid"),
			device:        &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device) {
				device2 := &models.Device{
					UID:      "uid2",
					Name:     "newname",
					TenantID: "tenant2",
				}

				queryOptionsMock.On("InNamespace", "tenant").Return(nil).Once()
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid", mock.AnythingOfType("store.QueryOption")).Return(device, nil).Once()
				queryOptionsMock.On("InNamespace", "tenant").Return(nil).Once()
				queryOptionsMock.On("WithDeviceStatus", models.DeviceStatusAccepted).Return(nil).Once()
				storeMock.On("DeviceResolve", ctx, store.DeviceHostnameResolver, "newname", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).Return(device2, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound(models.UID("uid"), errors.New("error", "", 0)),
		},
		{
			description:   "fails when the name already exists",
			tenant:        "tenant",
			deviceNewName: "newname",
			uid:           models.UID("uid"),
			device:        &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device) {
				device2 := &models.Device{
					UID:      "uid2",
					Name:     "newname",
					TenantID: "tenant2",
				}

				queryOptionsMock.On("InNamespace", "tenant").Return(nil).Once()
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid", mock.AnythingOfType("store.QueryOption")).Return(device, nil).Once()
				queryOptionsMock.On("InNamespace", "tenant").Return(nil).Once()
				queryOptionsMock.On("WithDeviceStatus", models.DeviceStatusAccepted).Return(nil).Once()
				storeMock.On("DeviceResolve", ctx, store.DeviceHostnameResolver, "newname", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).Return(device2, nil).Once()
			},
			expected: NewErrDeviceDuplicated("newname", nil),
		},
		{
			description:   "fails when the store device rename fails",
			tenant:        "tenant",
			deviceNewName: "newname",
			uid:           models.UID("uid"),
			device:        &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device) {
				queryOptionsMock.On("InNamespace", "tenant").Return(nil).Once()
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid", mock.AnythingOfType("store.QueryOption")).Return(device, nil).Once()
				queryOptionsMock.On("InNamespace", "tenant").Return(nil).Once()
				queryOptionsMock.On("WithDeviceStatus", models.DeviceStatusAccepted).Return(nil).Once()
				storeMock.On("DeviceResolve", ctx, store.DeviceHostnameResolver, "newname", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("DeviceRename", ctx, models.UID("uid"), "newname").Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description:   "succeeds",
			tenant:        "tenant",
			deviceNewName: "newname",
			uid:           models.UID("uid"),
			device:        &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device) {
				queryOptionsMock.On("InNamespace", "tenant").Return(nil).Once()
				storeMock.On("DeviceResolve", ctx, store.DeviceUIDResolver, "uid", mock.AnythingOfType("store.QueryOption")).Return(device, nil).Once()
				queryOptionsMock.On("InNamespace", "tenant").Return(nil).Once()
				queryOptionsMock.On("WithDeviceStatus", models.DeviceStatusAccepted).Return(nil).Once()
				storeMock.On("DeviceResolve", ctx, store.DeviceHostnameResolver, "newname", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).Return(nil, store.ErrNoDocuments).Once()
				storeMock.On("DeviceRename", ctx, models.UID("uid"), "newname").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks(tc.device)

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			err := service.RenameDevice(ctx, tc.uid, tc.deviceNewName, tc.tenant)
			assert.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestLookupDevice(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	ctx := context.TODO()

	type Expected struct {
		device *models.Device
		err    error
	}

	cases := []struct {
		description   string
		namespace     string
		device        *models.Device
		requiredMocks func(device *models.Device, namespace string)
		expected      Expected
	}{
		{
			description: "fails when namespace does not exists",
			namespace:   "namespace",
			device:      &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(_ *models.Device, namespace string) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceNameResolver, namespace).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				nil,
				NewErrNamespaceNotFound("namespace", errors.New("error", "", 0)),
			},
		},
		{
			description: "fails when device is not found",
			namespace:   "namespace",
			device:      &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device, namespace string) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceNameResolver, namespace).
					Return(&models.Namespace{TenantID: "00000000-0000-0000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				storeMock.
					On(
						"DeviceResolve",
						ctx,
						store.DeviceHostnameResolver,
						"name",
						mock.AnythingOfType("store.QueryOption"),
						mock.AnythingOfType("store.QueryOption"),
					).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				nil,
				NewErrDeviceNotFound(models.UID("name"), errors.New("error", "", 0)),
			},
		},
		{
			description: "succeeds to lookup for device",
			namespace:   "namespace",
			device:      &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device, namespace string) {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceNameResolver, namespace).
					Return(&models.Namespace{TenantID: "00000000-0000-0000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				storeMock.
					On(
						"DeviceResolve",
						ctx,
						store.DeviceHostnameResolver,
						"name",
						mock.AnythingOfType("store.QueryOption"),
						mock.AnythingOfType("store.QueryOption"),
					).
					Return(device, nil).
					Once()
			},
			expected: Expected{
				&models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
				nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks(tc.device, tc.namespace)

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			returnedDevice, err := service.LookupDevice(ctx, tc.namespace, tc.device.Name)
			assert.Equal(t, tc.expected, Expected{returnedDevice, err})
		})
	}
	storeMock.AssertExpectations(t)
}

func TestOfflineDevice(t *testing.T) {
	storeMock := new(storemock.Store)

	cases := []struct {
		name     string
		uid      models.UID
		mocks    func(context.Context)
		expected error
	}{
		{
			name: "fails when operation does not succeeds",
			uid:  models.UID("uid"),
			mocks: func(ctx context.Context) {
				storeMock.
					On("DeviceUpdate", ctx, "", "uid", &models.DeviceChanges{DisconnectedAt: &now}).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			name: "fails when connected_device does not exist",
			uid:  models.UID("uid"),
			mocks: func(ctx context.Context) {
				storeMock.
					On("DeviceUpdate", ctx, "", "uid", &models.DeviceChanges{DisconnectedAt: &now}).
					Return(store.ErrNoDocuments).
					Once()
			},
			expected: NewErrDeviceNotFound(models.UID("uid"), store.ErrNoDocuments),
		},
		{
			name: "succeeds",
			uid:  models.UID("uid"),
			mocks: func(ctx context.Context) {
				storeMock.
					On("DeviceUpdate", ctx, "", "uid", &models.DeviceChanges{DisconnectedAt: &now}).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.mocks(ctx)
			assert.Equal(t, tc.expected, s.OfflineDevice(ctx, tc.uid))
		})
	}

	storeMock.AssertExpectations(t)
}

func TestUpdateDeviceStatus(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	envMock := new(envsmocks.Backend)
	envs.DefaultBackend = envMock

	ctx := context.Background()
	cases := []struct {
		description   string
		req           *requests.DeviceUpdateStatus
		requiredMocks func()
		expectedError error
	}{
		{
			description: "failure - namespace not found",
			req: &requests.DeviceUpdateStatus{
				TenantID: "invalid-tenant",
				UID:      "new-device",
				Status:   "accepted",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "invalid-tenant").
					Return(nil, errors.New("namespace not found", "", 0)).
					Once()
			},
			expectedError: NewErrNamespaceNotFound("invalid-tenant", errors.New("namespace not found", "", 0)),
		},
		{
			description: "failure - device not found",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "nonexistent-device",
				Status:   "accepted",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-0000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "nonexistent-device", mock.AnythingOfType("store.QueryOption")).
					Return(nil, errors.New("device not found", "", 0)).
					Once()
			},
			expectedError: NewErrDeviceNotFound(models.UID("nonexistent-device"), errors.New("device not found", "", 0)),
		},
		{
			description: "failure - device already accepted",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "accepted-device",
				Status:   "accepted",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-0000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "accepted-device", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "accepted-device",
							Name:     "test-device",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusAccepted,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
			},
			expectedError: NewErrDeviceStatusAccepted(nil),
		},
		{
			description: "success - same status",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "new-device",
				Status:   "pending",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-0000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "new-device", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "accepted-device",
							Name:     "test-device",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusPending,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
			},
			expectedError: nil,
		},
		{
			description: "success (rejected) - status change to pending",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "device-to-pending",
				Status:   "pending",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-0000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "device-to-pending", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "device-to-pending",
							Name:     "test-device",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusRejected,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceUpdate", ctx, "00000000-0000-0000-0000-000000000000", "device-to-pending", mock.MatchedBy(func(changes *models.DeviceChanges) bool {
						return changes.Status == models.DeviceStatusPending
					})).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-0000-0000-000000000000", models.DeviceStatusRejected, int64(-1)).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-0000-0000-000000000000", models.DeviceStatusPending, int64(1)).
					Return(nil).
					Once()
			},
			expectedError: nil,
		},
		{
			description: "success (pending) - status change to rejected",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "device-to-reject",
				Status:   "rejected",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-0000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "device-to-reject", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "device-to-reject",
							Name:     "test-device",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusPending,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceUpdate", ctx, "00000000-0000-0000-0000-000000000000", "device-to-reject", mock.MatchedBy(func(changes *models.DeviceChanges) bool {
						return changes.Status == models.DeviceStatusRejected
					})).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-0000-0000-000000000000", models.DeviceStatusPending, int64(-1)).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-0000-0000-000000000000", models.DeviceStatusRejected, int64(1)).
					Return(nil).
					Once()
			},
			expectedError: nil,
		},
		{
			description: "failure (accepted) (same MAC) - hostname conflict",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "conflicting-device",
				Status:   "accepted",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-0000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "conflicting-device", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "conflicting-device",
							Name:     "device-name",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusPending,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceMACResolver, "aa:bb:cc:dd:ee:ff", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "old-device",
							Name:     "device-name",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusAccepted,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceHostnameResolver, "device-name", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "old-device",
							Name:     "device-name",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusAccepted,
							Identity: &models.DeviceIdentity{MAC: "00:11:22:33:44:55"},
						},
						nil,
					).
					Once()
			},
			expectedError: NewErrDeviceDuplicated("device-name", nil),
		},
		{
			description: "success (accepted) (same MAC) - device merge",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "new-device",
				Status:   "accepted",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-0000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "new-device", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "new-device",
							Name:     "device-name",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusPending,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceMACResolver, "aa:bb:cc:dd:ee:ff", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "old-device",
							Name:     "device-name",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusAccepted,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceHostnameResolver, "device-name", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "old-device",
							Name:     "device-name",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusAccepted,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				// Merge operations
				storeMock.
					On("SessionUpdateDeviceUID", ctx, models.UID("old-device"), models.UID("new-device")).
					Return(nil).
					Once()
				storeMock.
					On("DeviceUpdate", ctx, "00000000-0000-0000-0000-000000000000", "new-device", mock.MatchedBy(func(changes *models.DeviceChanges) bool {
						return changes.Name == "device-name"
					})).
					Return(nil).
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("old-device")).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted, int64(-1)).
					Return(nil).
					Once()
				// Final status update
				storeMock.
					On("DeviceUpdate", ctx, "00000000-0000-0000-0000-000000000000", "new-device", mock.MatchedBy(func(changes *models.DeviceChanges) bool {
						return changes.Status == models.DeviceStatusAccepted
					})).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-0000-0000-000000000000", models.DeviceStatusPending, int64(-1)).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted, int64(1)).
					Return(nil).
					Once()
			},
			expectedError: nil,
		},
		{
			description: "failure (accepted) (different MAC) - hostname conflict",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "conflicting-device",
				Status:   "accepted",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-0000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "conflicting-device", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "conflicting-device",
							Name:     "duplicate-name",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusPending,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceMACResolver, "aa:bb:cc:dd:ee:ff", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceHostnameResolver, "duplicate-name", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "existing-device",
							Name:     "duplicate-name",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusAccepted,
							Identity: &models.DeviceIdentity{MAC: "11:22:33:44:55:66"},
						},
						nil,
					).
					Once()
			},
			expectedError: NewErrDeviceDuplicated("duplicate-name", nil),
		},
		{
			description: "failure (accepted) (different MAC) - device limit reached [community]",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "limit-device",
				Status:   "accepted",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Namespace{
							TenantID:             "00000000-0000-0000-0000-000000000000",
							MaxDevices:           3,
							DevicesAcceptedCount: 3,
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "limit-device", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "limit-device",
							Name:     "test-device",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusPending,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceMACResolver, "aa:bb:cc:dd:ee:ff", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceHostnameResolver, "test-device", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
			},
			expectedError: NewErrDeviceMaxDevicesReached(3),
		},
		{
			description: "success (accepted) (different MAC) - device acceptance [community]",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "pending-device",
				Status:   "accepted",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-0000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "pending-device", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "pending-device",
							Name:     "test-device",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusPending,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceMACResolver, "aa:bb:cc:dd:ee:ff", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceHostnameResolver, "test-device", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				storeMock.
					On("DeviceUpdate", ctx, "00000000-0000-0000-0000-000000000000", "pending-device", mock.MatchedBy(func(changes *models.DeviceChanges) bool {
						return changes.Status == models.DeviceStatusAccepted
					})).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-0000-0000-000000000000", models.DeviceStatusPending, int64(-1)).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted, int64(1)).
					Return(nil).
					Once()
			},
			expectedError: nil,
		},
		{
			description: "failure (accepted) (different MAC) - device limit reached [enterprise]",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "limit-device",
				Status:   "accepted",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Namespace{
							TenantID:             "00000000-0000-0000-0000-000000000000",
							MaxDevices:           3,
							DevicesAcceptedCount: 3,
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "limit-device", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "limit-device",
							Name:     "test-device",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusPending,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceMACResolver, "aa:bb:cc:dd:ee:ff", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceHostnameResolver, "test-device", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
			},
			expectedError: NewErrDeviceMaxDevicesReached(3),
		},
		{
			description: "success (accepted) (different MAC) - device acceptance [enterprise]",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "pending-device",
				Status:   "accepted",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(&models.Namespace{TenantID: "00000000-0000-0000-0000-000000000000"}, nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "pending-device", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "pending-device",
							Name:     "test-device",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusPending,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceMACResolver, "aa:bb:cc:dd:ee:ff", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceHostnameResolver, "test-device", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				storeMock.
					On("DeviceUpdate", ctx, "00000000-0000-0000-0000-000000000000", "pending-device", mock.MatchedBy(func(changes *models.DeviceChanges) bool {
						return changes.Status == models.DeviceStatusAccepted
					})).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-0000-0000-000000000000", models.DeviceStatusPending, int64(-1)).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted, int64(1)).
					Return(nil).
					Once()
			},
			expectedError: nil,
		},
		{
			description: "failure (accepted) (different MAC) (billing inactive) (removed device) - billing evaluate [cloud]",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "removed-device",
				Status:   "accepted",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
							Billing:  &models.Billing{Active: false},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "removed-device", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "removed-device",
							Name:     "test-device",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusRemoved,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceMACResolver, "aa:bb:cc:dd:ee:ff", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceHostnameResolver, "test-device", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				clientMock.
					On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").
					Return(&models.BillingEvaluation{CanAccept: false}, 0, errors.New("error", "store", 0)).
					Once()
			},
			expectedError: NewErrBillingEvaluate(errors.New("evaluate error", "service", 4)),
		},
		{
			description: "failure (accepted) (different MAC) (billing inactive) (removed device) - can't accept [cloud]",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "removed-device",
				Status:   "accepted",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
							Billing:  &models.Billing{Active: false},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "removed-device", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "removed-device",
							Name:     "test-device",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusRemoved,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceMACResolver, "aa:bb:cc:dd:ee:ff", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceHostnameResolver, "test-device", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				clientMock.
					On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").
					Return(&models.BillingEvaluation{CanAccept: false}, 0, nil).
					Once()
			},
			expectedError: ErrDeviceLimit,
		},
		{
			description: "success (accepted) (different MAC) (billing inactive) (removed device) - [cloud]",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "removed-device",
				Status:   "accepted",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
							Billing:  &models.Billing{Active: false},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "removed-device", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "removed-device",
							Name:     "test-device",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusRemoved,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceMACResolver, "aa:bb:cc:dd:ee:ff", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceHostnameResolver, "test-device", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				clientMock.
					On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").
					Return(&models.BillingEvaluation{CanAccept: true}, 0, nil).
					Once()
				storeMock.
					On("DeviceUpdate", ctx, "00000000-0000-0000-0000-000000000000", "removed-device", mock.MatchedBy(func(changes *models.DeviceChanges) bool {
						return changes.Status == models.DeviceStatusAccepted
					})).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-0000-0000-000000000000", models.DeviceStatusRemoved, int64(-1)).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted, int64(1)).
					Return(nil).
					Once()
			},
			expectedError: nil,
		},
		{
			description: "failure (accepted) (different MAC) (billing active) - billing report error [cloud]",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "billing-error-device",
				Status:   "accepted",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
							Billing:  &models.Billing{Active: true},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "billing-error-device", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "billing-error-device",
							Name:     "test-device",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusPending,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceMACResolver, "aa:bb:cc:dd:ee:ff", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceHostnameResolver, "test-device", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				clientMock.
					On("BillingReport", "00000000-0000-0000-0000-000000000000", ReportDeviceAccept).
					Return(0, errors.New("billing error", "", 0)).
					Once()
			},
			expectedError: NewErrBillingReportNamespaceDelete(errors.New("billing error", "", 0)),
		},
		{
			description: "failure (accepted) (different MAC) (billing active) - payment required [cloud]",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "payment-required-device",
				Status:   "accepted",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
							Billing:  &models.Billing{Active: true},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "payment-required-device", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "payment-required-device",
							Name:     "test-device",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusPending,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceMACResolver, "aa:bb:cc:dd:ee:ff", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceHostnameResolver, "test-device", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				clientMock.
					On("BillingReport", "00000000-0000-0000-0000-000000000000", ReportDeviceAccept).
					Return(402, nil).
					Once()
			},
			expectedError: NewErrBillingReportNamespaceDelete(ErrPaymentRequired),
		},
		{
			description: "success (accepted) (different MAC) (billing active) - device acceptance [cloud]",
			req: &requests.DeviceUpdateStatus{
				TenantID: "00000000-0000-0000-0000-000000000000",
				UID:      "cloud-device",
				Status:   "accepted",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceResolve", ctx, store.NamespaceTenantIDResolver, "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
							Billing:  &models.Billing{Active: true},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "cloud-device", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:      "cloud-device",
							Name:     "test-device",
							TenantID: "00000000-0000-0000-0000-000000000000",
							Status:   models.DeviceStatusPending,
							Identity: &models.DeviceIdentity{MAC: "aa:bb:cc:dd:ee:ff"},
						},
						nil,
					).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceMACResolver, "aa:bb:cc:dd:ee:ff", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				queryOptionsMock.
					On("WithDeviceStatus", models.DeviceStatusAccepted).
					Return(nil).
					Once()
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceHostnameResolver, "test-device", mock.AnythingOfType("store.QueryOption"), mock.AnythingOfType("store.QueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				clientMock.
					On("BillingReport", "00000000-0000-0000-0000-000000000000", ReportDeviceAccept).
					Return(200, nil).
					Once()
				storeMock.
					On("DeviceUpdate", ctx, "00000000-0000-0000-0000-000000000000", "cloud-device", mock.MatchedBy(func(changes *models.DeviceChanges) bool {
						return changes.Status == models.DeviceStatusAccepted
					})).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-0000-0000-000000000000", models.DeviceStatusPending, int64(-1)).
					Return(nil).
					Once()
				storeMock.
					On("NamespaceIncrementDeviceCount", ctx, "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted, int64(1)).
					Return(nil).
					Once()
			},
			expectedError: nil,
		},
	}

	service := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	storeMock.
		On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
		Return(func(ctx context.Context, cb store.TransactionCb) error { return cb(ctx) }).
		Times(len(cases))

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			err := service.UpdateDeviceStatus(ctx, tc.req)
			require.Equal(t, tc.expectedError, err)
		})
	}

	storeMock.AssertExpectations(t)
	envMock.AssertExpectations(t)
}

func TestDeviceUpdate(t *testing.T) {
	now := time.Now()
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	cases := []struct {
		description   string
		req           *requests.DeviceUpdate
		requiredMocks func(ctx context.Context)
		expected      error
	}{
		{
			description: "fails when could not get the device by UID",
			req: &requests.DeviceUpdate{
				UID:      "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
				TenantID: "00000000-0000-0000-0000-000000000000",
				Name:     "",
			},
			requiredMocks: func(ctx context.Context) {
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e", mock.AnythingOfType("store.QueryOption")).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceNotFound(models.UID("d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e"), errors.New("error", "", 0)),
		},
		{
			description: "fails when already exists a device with same name",
			req: &requests.DeviceUpdate{
				UID:      "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
				TenantID: "00000000-0000-0000-0000-000000000000",
				Name:     "name",
			},
			requiredMocks: func(ctx context.Context) {
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:            "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
							DisconnectedAt: &now,
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceConflicts", ctx, &models.DeviceConflicts{Name: "name"}).
					Return([]string{"name"}, true, nil).
					Once()
			},
			expected: NewErrDeviceDuplicated("name", nil),
		},
		{
			description: "success when updating the device name to same name",
			req: &requests.DeviceUpdate{
				UID:      "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
				TenantID: "00000000-0000-0000-0000-000000000000",
				Name:     "name",
			},
			requiredMocks: func(ctx context.Context) {
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:            "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
							Name:           "name",
							DisconnectedAt: &now,
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceConflicts", ctx, &models.DeviceConflicts{Name: ""}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("DeviceUpdate", ctx, "00000000-0000-0000-0000-000000000000", "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e", &models.DeviceChanges{DisconnectedAt: &now}).
					Return(nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "success when update device",
			req: &requests.DeviceUpdate{
				UID:      "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
				TenantID: "00000000-0000-0000-0000-000000000000",
				Name:     "name",
			},
			requiredMocks: func(ctx context.Context) {
				queryOptionsMock.
					On("InNamespace", "00000000-0000-0000-0000-000000000000").
					Return(nil).
					Once()
				storeMock.
					On("DeviceResolve", ctx, store.DeviceUIDResolver, "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e", mock.AnythingOfType("store.QueryOption")).
					Return(
						&models.Device{
							UID:            "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
							DisconnectedAt: &now,
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceConflicts", ctx, &models.DeviceConflicts{Name: "name"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("DeviceUpdate", ctx, "00000000-0000-0000-0000-000000000000", "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e", &models.DeviceChanges{Name: "name", DisconnectedAt: &now}).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	service := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, test := range cases {
		t.Run(test.description, func(t *testing.T) {
			ctx := context.Background()
			test.requiredMocks(ctx)

			err := service.UpdateDevice(ctx, test.req)
			assert.Equal(t, test.expected, err)
		})
	}
}
