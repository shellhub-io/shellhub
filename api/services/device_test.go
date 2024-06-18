package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/auth"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestListDevices_cloud(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	type Expected struct {
		devices []models.Device
		count   int
		err     error
	}

	cases := []struct {
		description   string
		tenant        string
		sorter        query.Sorter
		pagination    query.Paginator
		filter        query.Filters
		status        models.DeviceStatus
		requiredMocks func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter)
		expected      Expected
	}{
		{
			description: "fail when namespace does not exist",
			tenant:      "00000000-0000-4000-0000-000000000000",
			sorter:      query.Sorter{By: "name", Order: query.OrderAsc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusAccepted,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				mock.On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{
				devices: nil,
				count:   0,
				err:     NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", errors.New("error", "", 0)),
			},
		},
		{
			description: "fail to list devices when status is removed",
			tenant:      "00000000-0000-4000-0000-000000000000",
			sorter:      query.Sorter{By: "name", Order: query.OrderAsc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusRemoved,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				namespace := &models.Namespace{
					TenantID:     "00000000-0000-4000-0000-000000000000",
					MaxDevices:   3,
					DevicesCount: 3,
				}

				mock.On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).Return(namespace, nil).Once()
				mock.On("DeviceRemovedList", ctx, "00000000-0000-4000-0000-000000000000", paginator, filters, sorter).
					Return(nil, 0, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				devices: nil,
				count:   0,
				err:     errors.New("error", "", 0),
			},
		},
		{
			description: "fail to list devices when could not list how many removed devices exist",
			tenant:      "00000000-0000-4000-0000-000000000000",
			sorter:      query.Sorter{By: "name", Order: query.OrderAsc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusPending,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Once()

				namespace := &models.Namespace{
					TenantID:     "00000000-0000-4000-0000-000000000000",
					MaxDevices:   3,
					DevicesCount: 3,
				}

				mock.On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).Return(namespace, nil).Once()
				mock.On("DeviceRemovedCount", ctx, "00000000-0000-4000-0000-000000000000").
					Return(int64(0), errors.New("error", "", 0)).Once()
			},
			expected: Expected{
				devices: nil,
				count:   0,
				err:     NewErrDeviceRemovedCount(errors.New("error", "", 0)),
			},
		},
		{
			description: "fail to list the devices when the device number has reached its limit",
			tenant:      "00000000-0000-4000-0000-000000000000",
			sorter:      query.Sorter{By: "name", Order: query.OrderAsc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusPending,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Once()

				namespace := &models.Namespace{
					TenantID:     "00000000-0000-4000-0000-000000000000",
					MaxDevices:   3,
					DevicesCount: 3,
				}

				mock.On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).Return(namespace, nil).Once()
				mock.On("DeviceRemovedCount", ctx, "00000000-0000-4000-0000-000000000000").
					Return(int64(0), nil).Once()
				mock.On("DeviceList", ctx, status, paginator, filters, sorter, store.DeviceAcceptableFromRemoved).
					Return(nil, 0, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				devices: nil,
				count:   0,
				err:     errors.New("error", "", 0),
			},
		},
		{
			description: "fail to list the devices when the device number is under its limit",
			tenant:      "00000000-0000-4000-0000-000000000000",
			sorter:      query.Sorter{By: "name", Order: query.OrderAsc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusPending,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Once()

				namespace := &models.Namespace{
					TenantID:     "00000000-0000-4000-0000-000000000000",
					MaxDevices:   3,
					DevicesCount: 2,
				}

				mock.On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).Return(namespace, nil).Once()
				mock.On("DeviceRemovedCount", ctx, "00000000-0000-4000-0000-000000000000").
					Return(int64(0), nil).Once()
				mock.On("DeviceList", ctx, status, paginator, filters, sorter, store.DeviceAcceptableIfNotAccepted).
					Return(nil, 0, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				devices: nil,
				count:   0,
				err:     errors.New("error", "", 0),
			},
		},
		{
			description: "success to list devices when status is pending",
			tenant:      "00000000-0000-4000-0000-000000000000",
			sorter:      query.Sorter{By: "name", Order: query.OrderAsc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusPending,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Once()

				namespace := &models.Namespace{
					TenantID:     "00000000-0000-4000-0000-000000000000",
					MaxDevices:   3,
					DevicesCount: 2,
				}

				mock.On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).Return(namespace, nil).Once()
				mock.On("DeviceRemovedCount", ctx, "00000000-0000-4000-0000-000000000000").
					Return(int64(0), nil).Once()
				mock.On("DeviceList", ctx, status, paginator, filters, sorter, store.DeviceAcceptableIfNotAccepted).
					Return([]models.Device{
						{
							Acceptable: true,
						},
						{
							Acceptable: true,
						},
					}, 2, nil).
					Once()
			},
			expected: Expected{
				devices: []models.Device{
					{
						Acceptable: true,
					},
					{
						Acceptable: true,
					},
				},
				count: 2,
				err:   nil,
			},
		},
		{
			description: "success to list devices when status is accepted",
			tenant:      "00000000-0000-4000-0000-000000000000",
			sorter:      query.Sorter{By: "name", Order: query.OrderAsc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusAccepted,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Once()

				namespace := &models.Namespace{
					TenantID:     "00000000-0000-4000-0000-000000000000",
					MaxDevices:   3,
					DevicesCount: 2,
				}

				mock.On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).Return(namespace, nil).Once()
				mock.On("DeviceRemovedCount", ctx, "00000000-0000-4000-0000-000000000000").
					Return(int64(0), nil).Once()
				mock.On("DeviceList", ctx, status, paginator, filters, sorter, store.DeviceAcceptableIfNotAccepted).
					Return([]models.Device{
						{
							Acceptable: false,
						},
						{
							Acceptable: false,
						},
					}, 2, nil).
					Once()
			},
			expected: Expected{
				devices: []models.Device{
					{
						Acceptable: false,
					},
					{
						Acceptable: false,
					},
				},
				count: 2,
				err:   nil,
			},
		},
		{
			description: "success to list devices when status is empty",
			tenant:      "00000000-0000-4000-0000-000000000000",
			sorter:      query.Sorter{By: "name", Order: query.OrderAsc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusEmpty,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Once()

				namespace := &models.Namespace{
					TenantID:     "00000000-0000-4000-0000-000000000000",
					MaxDevices:   3,
					DevicesCount: 2,
				}

				mock.On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", true).Return(namespace, nil).Once()
				mock.On("DeviceRemovedCount", ctx, "00000000-0000-4000-0000-000000000000").
					Return(int64(0), nil).Once()
				mock.On("DeviceList", ctx, status, paginator, filters, sorter, store.DeviceAcceptableIfNotAccepted).
					Return([]models.Device{
						{
							Acceptable: true,
						},
						{
							Acceptable: false,
						},
					}, 2, nil).
					Once()
			},
			expected: Expected{
				devices: []models.Device{
					{
						Acceptable: true,
					},
					{
						Acceptable: false,
					},
				},
				count: 2,
				err:   nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(*testing.T) {
			tc.requiredMocks(tc.status, tc.pagination, tc.filter, tc.sorter)

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			devices, count, err := service.ListDevices(ctx, tc.tenant, tc.status, tc.pagination, tc.filter, tc.sorter)

			assert.Equal(t, tc.expected.devices, devices)
			assert.Equal(t, tc.expected.count, count)
			assert.Equal(t, tc.expected.err, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestListDevices_enterprise(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	type Expected struct {
		devices []models.Device
		count   int
		err     error
	}

	cases := []struct {
		description   string
		tenant        string
		sorter        query.Sorter
		pagination    query.Paginator
		filter        query.Filters
		status        models.DeviceStatus
		requiredMocks func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter)
		expected      Expected
	}{
		{
			description: "fails when the store device list fails when status is pending",
			tenant:      "tenant",
			sorter:      query.Sorter{By: "name", Order: query.OrderAsc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusPending,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("true").Twice()

				namespace := &models.Namespace{
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 3,
				}

				mock.On("NamespaceGet", ctx, "tenant", true).Return(namespace, nil).Once()
				mock.On("DeviceList", ctx, status, paginator, filters, sorter, store.DeviceAcceptableAsFalse).
					Return(nil, 0, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				nil,
				0,
				errors.New("error", "", 0),
			},
		},
		{
			description: "fails when the store device list fails when status is accepted",
			tenant:      "tenant",
			sorter:      query.Sorter{By: "name", Order: query.OrderDesc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusAccepted,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("true").Twice()

				namespace := &models.Namespace{
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 2,
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("DeviceList", ctx, status, paginator, filters, sorter, store.DeviceAcceptableIfNotAccepted).
					Return(nil, 0, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				nil,
				0,
				errors.New("error", "", 0),
			},
		},
		{
			description: "succeeds when status is pending",
			tenant:      "tenant",
			sorter:      query.Sorter{By: "name", Order: query.OrderAsc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusPending,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("true").Twice()

				namespace := &models.Namespace{
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 3,
				}

				devices := []models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("DeviceList", ctx, status, paginator, filters, sorter, store.DeviceAcceptableAsFalse).
					Return(devices, len(devices), nil).
					Once()
			},
			expected: Expected{
				[]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				},
				len([]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}),
				nil,
			},
		},
		{
			description: "succeeds when status is accepted",
			tenant:      "tenant",
			sorter:      query.Sorter{By: "name", Order: query.OrderDesc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusAccepted,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("true").Twice()

				namespace := &models.Namespace{
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 2,
				}

				devices := []models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("DeviceList", ctx, status, paginator, filters, sorter, store.DeviceAcceptableIfNotAccepted).
					Return(devices, len(devices), nil).
					Once()
			},
			expected: Expected{
				[]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				},
				len([]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}),
				nil,
			},
		},
		{
			description: "fails when status is removed",
			tenant:      "tenant",
			sorter:      query.Sorter{By: "name", Order: query.OrderDesc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusRemoved,
			requiredMocks: func(_ models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				namespace := &models.Namespace{
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 3,
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("DeviceRemovedList", ctx, "tenant", paginator, filters, sorter).
					Return(nil, 0, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				nil,
				0,
				errors.New("error", "", 0),
			},
		},
		{
			description: "succeeds when status is removed",
			tenant:      "tenant",
			sorter:      query.Sorter{By: "name", Order: query.OrderDesc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusRemoved,
			requiredMocks: func(_ models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				namespace := &models.Namespace{
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 3,
				}

				devices := []models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}

				removedDevices := []models.DeviceRemoved{
					{Device: &devices[0]},
					{Device: &devices[1]},
					{Device: &devices[2]},
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("DeviceRemovedList", ctx, "tenant", paginator, filters, sorter).
					Return(removedDevices, len(removedDevices), nil).
					Once()
			},
			expected: Expected{
				[]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				},
				len([]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}),
				nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(*testing.T) {
			tc.requiredMocks(tc.status, tc.pagination, tc.filter, tc.sorter)

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			returnedDevices, count, err := service.ListDevices(ctx, tc.tenant, tc.status, tc.pagination, tc.filter, tc.sorter)
			assert.Equal(t, tc.expected, Expected{returnedDevices, count, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestListDevices_community(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	type Expected struct {
		devices []models.Device
		count   int
		err     error
	}

	cases := []struct {
		description   string
		tenant        string
		sorter        query.Sorter
		pagination    query.Paginator
		filter        query.Filters
		status        models.DeviceStatus
		requiredMocks func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter)
		expected      Expected
	}{
		{
			description: "fails when the store device list fails when status is pending",
			tenant:      "tenant",
			sorter:      query.Sorter{By: "name", Order: query.OrderAsc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusPending,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				namespace := &models.Namespace{
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 3,
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("DeviceList", ctx, status, paginator, filters, sorter, store.DeviceAcceptableAsFalse).
					Return(nil, 0, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				nil,
				0,
				errors.New("error", "", 0),
			},
		},
		{
			description: "fails when the store device list fails when status is accepted",
			tenant:      "tenant",
			sorter:      query.Sorter{By: "name", Order: query.OrderDesc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusAccepted,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				namespace := &models.Namespace{
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 2,
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("DeviceList", ctx, status, paginator, filters, sorter, store.DeviceAcceptableIfNotAccepted).
					Return(nil, 0, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				nil,
				0,
				errors.New("error", "", 0),
			},
		},
		{
			description: "succeeds when status is pending and the namespace has not reached its limit",
			tenant:      "tenant",
			sorter:      query.Sorter{By: "name", Order: query.OrderAsc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusPending,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				namespace := &models.Namespace{
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 2,
				}

				devices := []models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("DeviceList", ctx, status, paginator, filters, sorter, store.DeviceAcceptableIfNotAccepted).
					Return(devices, len(devices), nil).
					Once()
			},
			expected: Expected{
				[]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				},
				len([]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}),
				nil,
			},
		},
		{
			description: "succeeds when status is pending and the namespace has reached its limit",
			tenant:      "tenant",
			sorter:      query.Sorter{By: "name", Order: query.OrderAsc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusPending,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				namespace := &models.Namespace{
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 3,
				}

				devices := []models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("DeviceList", ctx, status, paginator, filters, sorter, store.DeviceAcceptableAsFalse).
					Return(devices, len(devices), nil).
					Once()
			},
			expected: Expected{
				[]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				},
				len([]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}),
				nil,
			},
		},
		{
			description: "succeeds when status is pending and namespace has no limit",
			tenant:      "tenant",
			sorter:      query.Sorter{By: "name", Order: query.OrderAsc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusPending,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				namespace := &models.Namespace{
					TenantID:     "tenant",
					MaxDevices:   -1,
					DevicesCount: 3,
				}

				devices := []models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("DeviceList", ctx, status, paginator, filters, sorter, store.DeviceAcceptableIfNotAccepted).
					Return(devices, len(devices), nil).
					Once()
			},
			expected: Expected{
				[]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				},
				len([]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}),
				nil,
			},
		},
		{
			description: "succeeds when status is accepted and the namespace has reached its limit",
			tenant:      "tenant",
			sorter:      query.Sorter{By: "name", Order: query.OrderDesc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusAccepted,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				namespace := &models.Namespace{
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 3,
				}

				devices := []models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("DeviceList", ctx, status, paginator, filters, sorter, store.DeviceAcceptableAsFalse).
					Return(devices, len(devices), nil).
					Once()
			},
			expected: Expected{
				[]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				},
				len([]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}),
				nil,
			},
		},
		{
			description: "succeeds when status is accepted and the namespace has no limit",
			tenant:      "tenant",
			sorter:      query.Sorter{By: "name", Order: query.OrderDesc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusAccepted,
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				namespace := &models.Namespace{
					TenantID:     "tenant",
					MaxDevices:   -1,
					DevicesCount: 2,
				}

				devices := []models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("DeviceList", ctx, status, paginator, filters, sorter, store.DeviceAcceptableIfNotAccepted).
					Return(devices, len(devices), nil).
					Once()
			},
			expected: Expected{
				[]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				},
				len([]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}),
				nil,
			},
		},
		{
			description: "fails when status is removed",
			tenant:      "tenant",
			sorter:      query.Sorter{By: "name", Order: query.OrderDesc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusRemoved,
			requiredMocks: func(_ models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				namespace := &models.Namespace{
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 3,
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("DeviceRemovedList", ctx, "tenant", paginator, filters, sorter).
					Return(nil, 0, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				nil,
				0,
				errors.New("error", "", 0),
			},
		},
		{
			description: "succeeds when status is removed",
			tenant:      "tenant",
			sorter:      query.Sorter{By: "name", Order: query.OrderDesc},
			pagination:  query.Paginator{Page: 1, PerPage: 10},
			filter: query.Filters{
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "hostname",
							Operator: "eq",
						},
					},
				},
			},
			status: models.DeviceStatusRemoved,
			requiredMocks: func(_ models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				namespace := &models.Namespace{
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 3,
				}

				devices := []models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}

				removedDevices := []models.DeviceRemoved{
					{Device: &devices[0]},
					{Device: &devices[1]},
					{Device: &devices[2]},
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID, true).Return(namespace, nil).Once()
				mock.On("DeviceRemovedList", ctx, "tenant", paginator, filters, sorter).
					Return(removedDevices, len(removedDevices), nil).
					Once()
			},
			expected: Expected{
				[]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				},
				len([]models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}),
				nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(*testing.T) {
			tc.requiredMocks(tc.status, tc.pagination, tc.filter, tc.sorter)

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			returnedDevices, count, err := service.ListDevices(ctx, tc.tenant, tc.status, tc.pagination, tc.filter, tc.sorter)
			assert.Equal(t, tc.expected, Expected{returnedDevices, count, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestGetDevice(t *testing.T) {
	mock := new(mocks.Store)

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
				mock.On("DeviceGet", ctx, models.UID("_uid")).
					Return(nil, errors.New("error", "", 0)).Once()
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

				mock.On("DeviceGet", ctx, models.UID("uid")).
					Return(device, nil).Once()
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

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

			returnedDevice, err := service.GetDevice(ctx, tc.uid)
			assert.Equal(t, tc.expected, Expected{returnedDevice, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestDeleteDevice(t *testing.T) {
	mock := new(mocks.Store)

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
				mock.On("DeviceGetByUID", ctx, models.UID("_uid"), "tenant").
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound(models.UID("_uid"), errors.New("error", "", 0)),
		},
		{
			description: "fails when the store namespace get fails",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				device := &models.Device{
					UID:       "uid",
					TenantID:  "tenant",
					CreatedAt: time.Time{},
				}

				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), "tenant").
					Return(device, nil).Once()
				mock.On("NamespaceGet", ctx, "tenant", false).
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrNamespaceNotFound("tenant", errors.New("error", "", 0)),
		},
		{
			description: "fails when device removed insert return error",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "group1",
					Owner:    "id",
					TenantID: "tenant",
					Members: []models.Member{
						{
							ID:   "id",
							Role: auth.RoleOwner,
						},
						{
							ID:   "id2",
							Role: auth.RoleObserver,
						},
					},
					MaxDevices: 3,
				}

				device := &models.Device{
					UID:       "uid",
					TenantID:  "tenant",
					CreatedAt: time.Time{},
				}

				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), "tenant").
					Return(device, nil).Once()
				mock.On("NamespaceGet", ctx, "tenant", false).
					Return(namespace, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Once()
				envMock.On("Get", "SHELLHUB_BILLING").Return("true").Once()
				mock.On("DeviceRemovedInsert", ctx, "tenant", device).
					Return(errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceRemovedInsert(errors.New("error", "", 0)),
		},
		{
			description: "fails when the store device delete fails",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "group1",
					Owner:    "id",
					TenantID: "tenant",
					Members: []models.Member{
						{
							ID:   "id",
							Role: auth.RoleOwner,
						},
						{
							ID:   "id2",
							Role: auth.RoleObserver,
						},
					},
					MaxDevices: 3,
				}

				device := &models.Device{
					UID:       "uid",
					TenantID:  "tenant",
					CreatedAt: time.Time{},
				}

				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), "tenant").
					Return(device, nil).Once()
				mock.On("NamespaceGet", ctx, "tenant", false).
					Return(namespace, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				mock.On("DeviceDelete", ctx, models.UID(device.UID)).
					Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "succeeds",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				device := &models.Device{
					UID:       "uid",
					TenantID:  "tenant",
					CreatedAt: time.Time{},
				}

				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), "tenant").
					Return(device, nil).Once()
				mock.On("NamespaceGet", ctx, "tenant", false).
					Return(&models.Namespace{TenantID: "tenant"}, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				mock.On("DeviceDelete", ctx, models.UID(device.UID)).
					Return(nil).Once()
			},
			expected: nil,
		},
		/*{
			description: "fails to report usage",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				device := &models.Device{
					UID:       "uid",
					TenantID:  "tenant",
					CreatedAt: time.Time{},
				}

				namespaceBilling := &models.Namespace{
					Name:       "namespace1",
					MaxDevices: -1,
					Billing: &models.Billing{
						Active: true,
					},
				}
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), "tenant").
					Return(device, nil).Once()
				mock.On("NamespaceGet", ctx, "tenant").
					Return(namespaceBilling, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				clockMock.On("Now").Return(now).Twice()
				envMock.On("Get", "SHELLHUB_BILLING").Return(strconv.FormatBool(true)).Once()
				clientMock.On("ReportUsage", &models.UsageRecord{
					Device:    device,
					Namespace: namespaceBilling,
					Timestamp: now.Unix(),
				}).Return(500, nil).Once()
			},
			expected: ErrReport,
		},
		{
			description: "reports usage with success",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				device := &models.Device{
					UID:       "uid",
					TenantID:  "tenant",
					CreatedAt: time.Time{},
				}

				namespaceBilling := &models.Namespace{
					Name:    "namespace1",
					Members: []models.Member{{ID: "id", Role: auth.RoleOwner}, {ID: "id2", Role: auth.RoleObserver}},
					Billing: &models.Billing{
						Active: true,
					},
				}

				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), "tenant").
					Return(device, nil).Once()
				mock.On("NamespaceGet", ctx, "tenant").
					Return(namespaceBilling, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				clockMock.On("Now").Return(now).Twice()
				envMock.On("Get", "SHELLHUB_BILLING").Return(strconv.FormatBool(true)).Once()
				clientMock.On("ReportUsage", &models.UsageRecord{
					Device:    device,
					Namespace: namespaceBilling,
					Timestamp: now.Unix(),
				}).Return(200, nil).Once()
				mock.On("DeviceDelete", ctx, models.UID(device.UID)).
					Return(nil).Once()
			},
			expected: nil,
		},*/
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			err := service.DeleteDevice(ctx, tc.uid, tc.tenant)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestRenameDevice(t *testing.T) {
	mock := new(mocks.Store)

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
				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").Return(device, errors.New("error", "", 0)).Once()
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
				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").Return(device, nil).Once()
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

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").Return(device, nil).Once()
				mock.On("DeviceGetByName", ctx, "newname", "tenant", models.DeviceStatusAccepted).Return(device2, errors.New("error", "", 0)).Once()
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

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").Return(device, nil).Once()
				mock.On("DeviceGetByName", ctx, "newname", "tenant", models.DeviceStatusAccepted).Return(device2, nil).Once()
			},
			expected: NewErrDeviceDuplicated("newname", nil),
		},
		{
			description:   "fails when the store device rename fails",
			tenant:        "tenant",
			deviceNewName: "anewname",
			uid:           models.UID("uid"),
			device:        &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device) {
				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").Return(device, nil).Once()
				mock.On("DeviceGetByName", ctx, "anewname", "tenant", models.DeviceStatusAccepted).Return(nil, store.ErrNoDocuments).Once()
				mock.On("DeviceRename", ctx, models.UID("uid"), "anewname").Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description:   "succeeds",
			tenant:        "tenant",
			deviceNewName: "anewname",
			uid:           models.UID("uid"),
			device:        &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device) {
				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").Return(device, nil).Once()
				mock.On("DeviceGetByName", ctx, "anewname", "tenant", models.DeviceStatusAccepted).Return(nil, store.ErrNoDocuments).Once()
				mock.On("DeviceRename", ctx, models.UID("uid"), "anewname").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks(tc.device)

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			err := service.RenameDevice(ctx, tc.uid, tc.deviceNewName, tc.tenant)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestLookupDevice(t *testing.T) {
	mock := new(mocks.Store)

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
			description: "fails when store device lookup fails",
			namespace:   "namespace",
			device:      &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device, namespace string) {
				mock.On("DeviceLookup", ctx, namespace, device.Name).Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{
				nil,
				NewErrDeviceLookupNotFound("namespace", "name", errors.New("error", "", 0)),
			},
		},
		{
			description: "fails when the device is not found",
			namespace:   "namespace",
			device:      &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device, namespace string) {
				mock.On("DeviceLookup", ctx, namespace, device.Name).
					Return(nil, store.ErrNoDocuments).Once()
			},
			expected: Expected{
				nil,
				NewErrDeviceLookupNotFound("namespace", "name", store.ErrNoDocuments),
			},
		},
		{
			description: "succeeds",
			namespace:   "namespace",
			device:      &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device, namespace string) {
				mock.On("DeviceLookup", ctx, namespace, device.Name).
					Return(device, nil).Once()
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

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			returnedDevice, err := service.LookupDevice(ctx, tc.namespace, tc.device.Name)
			assert.Equal(t, tc.expected, Expected{returnedDevice, err})
		})
	}
	mock.AssertExpectations(t)
}

func TestOfflineDevice(t *testing.T) {
	storeMock := new(mocks.Store)

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
					On("DeviceSetOffline", ctx, "uid").
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
					On("DeviceSetOffline", ctx, "uid").
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
					On("DeviceSetOffline", ctx, "uid").
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.mocks(ctx)
			assert.Equal(t, tc.expected, s.OfflineDevice(ctx, tc.uid))
		})
	}

	storeMock.AssertExpectations(t)
}

func TestUpdateDeviceStatus_same_mac(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	cases := []struct {
		description   string
		uid           models.UID
		status        models.DeviceStatus
		tenant        string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when could not get the namespace",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrNamespaceNotFound("00000000-0000-0000-0000-000000000000", errors.New("error", "", 0)),
		},
		{
			description: "fails when could not get the devcie",
			uid:         models.UID("uid"),
			tenant:      "00000000-0000-0000-0000-000000000000",
			status:      "accepted",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound("uid", errors.New("error", "", 0)),
		},
		{
			description: "fails when device already accepted",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "accepted",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()
			},
			expected: NewErrDeviceStatusAccepted(nil),
		},
		{
			description: "fails when could not get the device by MAC",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound(models.UID("uid"), errors.New("error", "", 0)),
		},
		{
			description: "fails when already exist a device with same name and a different mac",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(&models.Device{
						UID:      "uid",
						Name:     "name",
						Identity: &models.DeviceIdentity{MAC: "mac"},
					}, nil).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(&models.Device{
						UID:      "notsameuid",
						Name:     "name",
						Identity: &models.DeviceIdentity{MAC: "anothermac"},
					}, nil).Once()
			},
			expected: NewErrDeviceDuplicated("name", nil),
		},
		{
			description: "fails to update device UID",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(&models.Device{
						UID:      "notsameuid",
						Name:     "name",
						Identity: &models.DeviceIdentity{MAC: "mac"},
					}, nil).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(&models.Device{
						UID:      "notsameuid",
						Name:     "name",
						Identity: &models.DeviceIdentity{MAC: "mac"},
					}, nil).Once()

				mock.On("SessionUpdateDeviceUID", ctx, models.UID("notsameuid"), models.UID("uid")).
					Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "fails to update device to the old name",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(&models.Device{
						UID:      "notsameuid",
						Name:     "name",
						Identity: &models.DeviceIdentity{MAC: "mac"},
					}, nil).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(&models.Device{
						UID:      "notsameuid",
						Name:     "name",
						Identity: &models.DeviceIdentity{MAC: "mac"},
					}, nil).Once()

				mock.On("SessionUpdateDeviceUID", ctx, models.UID("notsameuid"), models.UID("uid")).
					Return(nil).Once()

				mock.On("DeviceRename", ctx, models.UID("uid"), "name").
					Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "fails to delete device with the same mac",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(&models.Device{
						UID:      "notsameuid",
						Name:     "name",
						Identity: &models.DeviceIdentity{MAC: "mac"},
					}, nil).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(&models.Device{
						UID:      "notsameuid",
						Name:     "name",
						Identity: &models.DeviceIdentity{MAC: "mac"},
					}, nil).Once()

				mock.On("SessionUpdateDeviceUID", ctx, models.UID("notsameuid"), models.UID("uid")).
					Return(nil).Once()

				mock.On("DeviceRename", ctx, models.UID("uid"), "name").
					Return(nil).Once()

				mock.On("DeviceDelete", ctx, models.UID("notsameuid")).
					Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "fails to update device status",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(&models.Device{
						UID:      "notsameuid",
						Name:     "name",
						Identity: &models.DeviceIdentity{MAC: "mac"},
					}, nil).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(&models.Device{
						UID:      "notsameuid",
						Name:     "name",
						Identity: &models.DeviceIdentity{MAC: "mac"},
					}, nil).Once()

				mock.On("SessionUpdateDeviceUID", ctx, models.UID("notsameuid"), models.UID("uid")).
					Return(nil).Once()

				mock.On("DeviceRename", ctx, models.UID("uid"), "name").
					Return(nil).Once()

				mock.On("DeviceDelete", ctx, models.UID("notsameuid")).
					Return(nil).Once()

				mock.On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "success to update device status",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(&models.Device{
						UID:      "notsameuid",
						Name:     "name",
						Identity: &models.DeviceIdentity{MAC: "mac"},
					}, nil).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(&models.Device{
						UID:      "notsameuid",
						Name:     "name",
						Identity: &models.DeviceIdentity{MAC: "mac"},
					}, nil).Once()

				mock.On("SessionUpdateDeviceUID", ctx, models.UID("notsameuid"), models.UID("uid")).
					Return(nil).Once()

				mock.On("DeviceRename", ctx, models.UID("uid"), "name").
					Return(nil).Once()

				mock.On("DeviceDelete", ctx, models.UID("notsameuid")).
					Return(nil).Once()

				mock.On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			err := service.UpdateDeviceStatus(ctx, tc.tenant, tc.uid, tc.status)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestUpdateDeviceStatus_community_and_enterprise(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	cases := []struct {
		description   string
		uid           models.UID
		status        models.DeviceStatus
		tenant        string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when could not get the namespace",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrNamespaceNotFound("00000000-0000-0000-0000-000000000000", errors.New("error", "", 0)),
		},
		{
			description: "fails when could not get the devcie",
			uid:         models.UID("uid"),
			tenant:      "00000000-0000-0000-0000-000000000000",
			status:      "accepted",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound("uid", errors.New("error", "", 0)),
		},
		{
			description: "fails when device already accepted",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "accepted",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()
			},
			expected: NewErrDeviceStatusAccepted(nil),
		},
		{
			description: "fails when could not get the device by MAC",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound(models.UID("uid"), errors.New("error", "", 0)),
		},
		{
			description: "fails when already exist a device with same name",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(&models.Device{
						UID:  "fb2de504e98d3ccab342b53d83395cd7fda297c71e8da550c31478bae0dbb8c5",
						Name: "name",
					}, nil).Once()
			},
			expected: NewErrDeviceDuplicated("name", nil),
		},
		{
			description: "fails namespace has reached the limit of devices in community instance",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID:     "00000000-0000-0000-0000-000000000000",
						MaxDevices:   3,
						DevicesCount: 3,
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()
			},
			expected: NewErrDeviceMaxDevicesReached(3),
		},
		{
			description: "fails namespace has reached the limit of devices in enterprise instance",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID:     "00000000-0000-0000-0000-000000000000",
						MaxDevices:   3,
						DevicesCount: 3,
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("true").Twice()
			},
			expected: NewErrDeviceMaxDevicesReached(3),
		},
		{
			description: "fails when could not update device status on database",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				mock.On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "success to update device status",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				mock.On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			err := service.UpdateDeviceStatus(ctx, tc.tenant, tc.uid, tc.status)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestUpdateDeviceStatus_cloud_subscription_active(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	cases := []struct {
		description   string
		uid           models.UID
		status        models.DeviceStatus
		tenant        string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when could not get the namespace",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrNamespaceNotFound("00000000-0000-0000-0000-000000000000", errors.New("error", "", 0)),
		},
		{
			description: "fails when could not get the devcie",
			uid:         models.UID("uid"),
			tenant:      "00000000-0000-0000-0000-000000000000",
			status:      "accepted",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound("uid", errors.New("error", "", 0)),
		},
		{
			description: "fails when device already accepted",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "accepted",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()
			},
			expected: NewErrDeviceStatusAccepted(nil),
		},
		{
			description: "fails when could not get the device by MAC",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound(models.UID("uid"), errors.New("error", "", 0)),
		},
		{
			description: "fails when namespace has a subscription active and could not report the device accepted",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
						Billing: &models.Billing{
							Active: true,
						},
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				clientMock.On("BillingReport", "00000000-0000-0000-0000-000000000000", "device_accept").Return(0, errors.New("error", "", 0)).Once()
			},
			expected: NewErrBillingReportNamespaceDelete(errors.New("error", "", 0)),
		},
		{
			description: "fails when namespace has a subscription active and report block the action",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
						Billing: &models.Billing{
							Active: true,
						},
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				clientMock.On("BillingReport", "00000000-0000-0000-0000-000000000000", "device_accept").Return(402, nil).Once()
			},
			expected: NewErrBillingReportNamespaceDelete(ErrPaymentRequired),
		},
		{
			description: "fails when could not update device status on database",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
						Billing: &models.Billing{
							Active: true,
						},
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				clientMock.On("BillingReport", "00000000-0000-0000-0000-000000000000", "device_accept").Return(200, nil).Once()

				mock.On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "success to update device status",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
						Billing: &models.Billing{
							Active: true,
						},
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				clientMock.On("BillingReport", "00000000-0000-0000-0000-000000000000", "device_accept").Return(200, nil).Once()

				mock.On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			err := service.UpdateDeviceStatus(ctx, tc.tenant, tc.uid, tc.status)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestUpdateDeviceStatus_cloud_subscription_inactive(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	cases := []struct {
		description   string
		uid           models.UID
		status        models.DeviceStatus
		tenant        string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when could not get the namespace",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrNamespaceNotFound("00000000-0000-0000-0000-000000000000", errors.New("error", "", 0)),
		},
		{
			description: "fails when could not get the devcie",
			uid:         models.UID("uid"),
			tenant:      "00000000-0000-0000-0000-000000000000",
			status:      "accepted",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound("uid", errors.New("error", "", 0)),
		},
		{
			description: "fails when device already accepted",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "accepted",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()
			},
			expected: NewErrDeviceStatusAccepted(nil),
		},
		{
			description: "fails when could not get the device by MAC",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound(models.UID("uid"), errors.New("error", "", 0)),
		},
		{
			description: "fails when could not check if device was removed recently",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
						Billing: &models.Billing{
							Active: false,
						},
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				mock.On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceRemovedGet(errors.New("error", "", 0)),
		},
		{
			description: "fails when could not count how many devices were removed recently",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
						Billing: &models.Billing{
							Active: false,
						},
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				mock.On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil, nil).Once()

				mock.On("DeviceRemovedCount", ctx, "00000000-0000-0000-0000-000000000000").
					Return(int64(0), errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceRemovedCount(errors.New("error", "", 0)),
		},
		{
			description: "fails when namespace has reached the limit counting with removed devices",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID:     "00000000-0000-0000-0000-000000000000",
						MaxDevices:   3,
						DevicesCount: 1,
						Billing: &models.Billing{
							Active: false,
						},
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				mock.On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil, nil).Once()

				mock.On("DeviceRemovedCount", ctx, "00000000-0000-0000-0000-000000000000").
					Return(int64(2), nil).Once()
			},
			expected: NewErrDeviceRemovedFull(3, nil),
		},
		{
			description: "fails when could not evaluate the namespace capabilities when accepted device is not removed",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID:     "00000000-0000-0000-0000-000000000000",
						MaxDevices:   3,
						DevicesCount: 1,
						Billing: &models.Billing{
							Active: false,
						},
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				mock.On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil, nil).Once()

				mock.On("DeviceRemovedCount", ctx, "00000000-0000-0000-0000-000000000000").
					Return(int64(1), nil).Once()

				clientMock.On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").Return(nil, 0, errors.New("error", "", 0)).Once()
			},
			expected: NewErrBillingEvaluate(ErrEvaluate),
		},
		{
			description: "fails when namespace cannot accept more devices",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID:     "00000000-0000-0000-0000-000000000000",
						MaxDevices:   3,
						DevicesCount: 1,
						Billing: &models.Billing{
							Active: false,
						},
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				mock.On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil, nil).Once()

				mock.On("DeviceRemovedCount", ctx, "00000000-0000-0000-0000-000000000000").
					Return(int64(1), nil).Once()

				clientMock.On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").Return(&models.BillingEvaluation{
					CanAccept: false,
				}, 0, nil).Once()
			},
			expected: ErrDeviceLimit,
		},
		{
			description: "fails to update the device status when device is not on removed list",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID:     "00000000-0000-0000-0000-000000000000",
						MaxDevices:   3,
						DevicesCount: 1,
						Billing: &models.Billing{
							Active: false,
						},
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				mock.On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil, nil).Once()

				mock.On("DeviceRemovedCount", ctx, "00000000-0000-0000-0000-000000000000").
					Return(int64(1), nil).Once()

				clientMock.On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").Return(&models.BillingEvaluation{
					CanAccept: true,
				}, 0, nil).Once()

				mock.On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "success to update the device status when device is not on removed list",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID:     "00000000-0000-0000-0000-000000000000",
						MaxDevices:   3,
						DevicesCount: 1,
						Billing: &models.Billing{
							Active: false,
						},
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				mock.On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil, nil).Once()

				mock.On("DeviceRemovedCount", ctx, "00000000-0000-0000-0000-000000000000").
					Return(int64(1), nil).Once()

				clientMock.On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").Return(&models.BillingEvaluation{
					CanAccept: true,
				}, 0, nil).Once()

				mock.On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(nil).Once()
			},
			expected: nil,
		},
		{
			description: "fail when could not remove the device from removed device list when device is on removed list",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID:     "00000000-0000-0000-0000-000000000000",
						MaxDevices:   3,
						DevicesCount: 1,
						Billing: &models.Billing{
							Active: false,
						},
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				mock.On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(&models.DeviceRemoved{}, nil).Once()

				mock.On("DeviceRemovedDelete", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceRemovedDelete(errors.New("error", "", 0)),
		},
		{
			description: "fail when could not evaluate the namespace when device is on removed list",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID:     "00000000-0000-0000-0000-000000000000",
						MaxDevices:   3,
						DevicesCount: 1,
						Billing: &models.Billing{
							Active: false,
						},
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				mock.On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(&models.DeviceRemoved{}, nil).Once()

				mock.On("DeviceRemovedDelete", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil).Once()

				clientMock.On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").Return(nil, 0, errors.New("error", "", 0)).Once()
			},
			expected: NewErrBillingEvaluate(ErrEvaluate),
		},
		{
			description: "fails when namespace evaluation block device acceptance when device is on removed list",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID:     "00000000-0000-0000-0000-000000000000",
						MaxDevices:   3,
						DevicesCount: 1,
						Billing: &models.Billing{
							Active: false,
						},
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				mock.On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(&models.DeviceRemoved{}, nil).Once()

				mock.On("DeviceRemovedDelete", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil).Once()

				clientMock.On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").Return(&models.BillingEvaluation{
					CanAccept: false,
				}, 0, nil).Once()
			},
			expected: ErrDeviceLimit,
		},
		{
			description: "fails to update device status when device is on removed list",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID:     "00000000-0000-0000-0000-000000000000",
						MaxDevices:   3,
						DevicesCount: 1,
						Billing: &models.Billing{
							Active: false,
						},
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				mock.On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(&models.DeviceRemoved{}, nil).Once()

				mock.On("DeviceRemovedDelete", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil).Once()

				clientMock.On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").Return(&models.BillingEvaluation{
					CanAccept: true,
				}, 0, nil).Once()

				mock.On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "success to update device status when device is on removed list",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID:     "00000000-0000-0000-0000-000000000000",
						MaxDevices:   3,
						DevicesCount: 1,
						Billing: &models.Billing{
							Active: false,
						},
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Twice()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()

				mock.On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(&models.DeviceRemoved{}, nil).Once()

				mock.On("DeviceRemovedDelete", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil).Once()

				clientMock.On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").Return(&models.BillingEvaluation{
					CanAccept: true,
				}, 0, nil).Once()

				mock.On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			err := service.UpdateDeviceStatus(ctx, tc.tenant, tc.uid, tc.status)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestDeviceUpdate(t *testing.T) {
	mock := new(mocks.Store)
	service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	toPointer := func(s string) *string {
		return &s
	}

	other := toPointer("other")

	tests := []struct {
		description   string
		uid           string
		tenant        string
		name          *string
		publicKey     *bool
		requiredMocks func(ctx context.Context)
		expected      error
	}{
		{
			description: "fails when could not get the device by UID",
			uid:         "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
			tenant:      "00000000-0000-0000-0000-000000000000",
			name:        nil,
			publicKey:   nil,
			requiredMocks: func(ctx context.Context) {
				mock.On("DeviceGetByUID", ctx, models.UID("d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e"), "00000000-0000-0000-0000-000000000000").
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound(models.UID("d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e"), errors.New("error", "", 0)),
		},
		{
			description: "success when updating the device name to same name",
			uid:         "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
			tenant:      "00000000-0000-0000-0000-000000000000",
			name:        toPointer("name"),
			publicKey:   nil,
			requiredMocks: func(ctx context.Context) {
				mock.On("DeviceGetByUID", ctx, models.UID("d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:  "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
						Name: "name",
					}, nil).Once()
			},
			expected: nil,
		},
		{
			description: "fails when name does not meet the validatino requirements",
			uid:         "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
			tenant:      "00000000-0000-0000-0000-000000000000",
			name:        toPointer(""),
			publicKey:   nil,
			requiredMocks: func(ctx context.Context) {
				mock.On("DeviceGetByUID", ctx, models.UID("d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:  "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
						Name: "name",
					}, nil).Once()
			},
			expected: NewErrDeviceInvalid(map[string]interface{}{"name": ""}, nil),
		},
		{
			description: "fails when could not get the device by name",
			uid:         "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
			tenant:      "00000000-0000-0000-0000-000000000000",
			name:        toPointer("same"),
			publicKey:   nil,
			requiredMocks: func(ctx context.Context) {
				mock.On("DeviceGetByUID", ctx, models.UID("d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:  "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
						Name: "name",
					}, nil).Once()

				mock.On("DeviceGetByName", ctx, "same", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound(models.UID("same"), fmt.Errorf("failed to get device by name: %w", errors.New("error", "", 0))),
		},
		{
			description: "fails when already exists a device with same name",
			uid:         "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
			tenant:      "00000000-0000-0000-0000-000000000000",
			name:        toPointer("same"),
			publicKey:   nil,
			requiredMocks: func(ctx context.Context) {
				mock.On("DeviceGetByUID", ctx, models.UID("d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:  "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
						Name: "name",
					}, nil).Once()

				mock.On("DeviceGetByName", ctx, "same", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(&models.Device{
						UID:  "fb2de504e98d3ccab342b53d83395cd7fda297c71e8da550c31478bae0dbb8c5",
						Name: "same",
					}, nil).Once()
			},
			expected: NewErrDeviceDuplicated("same", nil),
		},
		{
			description: "success when udpate device for a different name",
			uid:         "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
			tenant:      "00000000-0000-0000-0000-000000000000",
			name:        other,
			publicKey:   new(bool),
			requiredMocks: func(ctx context.Context) {
				mock.On("DeviceGetByUID", ctx, models.UID("d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:  "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
						Name: "name",
					}, nil).Once()

				mock.On("DeviceGetByName", ctx, "other", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).Once()

				mock.On("DeviceUpdate", ctx, "00000000-0000-0000-0000-000000000000", models.UID("d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e"), other, new(bool)).
					Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			ctx := context.Background()
			test.requiredMocks(ctx)

			err := service.UpdateDevice(ctx, test.tenant, models.UID(test.uid), test.name, test.publicKey)
			assert.Equal(t, test.expected, err)
		})
	}
}

func TestUpdateDeviceStatus_other_than_accepted(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	cases := []struct {
		description   string
		uid           models.UID
		status        models.DeviceStatus
		tenant        string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when could not get the namespace",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrNamespaceNotFound("00000000-0000-0000-0000-000000000000", errors.New("error", "", 0)),
		},
		{
			description: "fails when could not get the devcie",
			uid:         models.UID("uid"),
			tenant:      "00000000-0000-0000-0000-000000000000",
			status:      "accepted",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound("uid", errors.New("error", "", 0)),
		},
		{
			description: "fails when device already accepted",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "accepted",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()
			},
			expected: NewErrDeviceStatusAccepted(nil),
		},
		{
			description: "fails when the intended status is pending, but store update fails",
			uid:         models.UID("uid"),
			status:      "pending",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("pending")).
					Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "success to update device status when the intended status is pending",
			uid:         models.UID("uid"),
			status:      "pending",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("pending")).
					Return(nil).Once()
			},
			expected: nil,
		},
		{
			description: "fails when the intended status is rejected, but store update fails",
			uid:         models.UID("uid"),
			status:      "rejected",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("rejected")).
					Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "success to update device status when the intended status is rejected",
			uid:         models.UID("uid"),
			status:      "rejected",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()

				mock.On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("rejected")).
					Return(nil).Once()
			},
			expected: nil,
		},
		{
			description: "fails when the device is already accepted",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "accepted",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()
			},
			expected: NewErrDeviceStatusAccepted(nil),
		},
		{
			description: "fails when the intended status is removed",
			uid:         models.UID("uid"),
			status:      "removed",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()
			},
			expected: NewErrDeviceStatusInvalid("removed", nil),
		},
		{
			description: "fails when the intended status is unused",
			uid:         models.UID("uid"),
			status:      "unused",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()
			},
			expected: NewErrDeviceStatusInvalid("unused", nil),
		},
		{
			description: "fails when the intended status is unknown",
			uid:         models.UID("uid"),
			status:      "unused",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", true).
					Return(&models.Namespace{
						TenantID: "00000000-0000-0000-0000-000000000000",
					}, nil).Once()

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(&models.Device{
						UID:       "uid",
						Name:      "name",
						TenantID:  "00000000-0000-0000-0000-000000000000",
						Status:    "pending",
						Identity:  &models.DeviceIdentity{MAC: "mac"},
						CreatedAt: time.Time{},
					}, nil).Once()
			},
			expected: NewErrDeviceStatusInvalid("unused", nil),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			err := service.UpdateDeviceStatus(ctx, tc.tenant, tc.uid, tc.status)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
