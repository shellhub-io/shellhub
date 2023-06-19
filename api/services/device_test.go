package services

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	mocksGeoIp "github.com/shellhub-io/shellhub/pkg/geoip/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestListDevices(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	status := []models.DeviceStatus{models.DeviceStatusPending, models.DeviceStatusAccepted, models.DeviceStatusRejected, models.DeviceStatusRemoved}
	order := []string{"asc", "desc"}

	type Expected struct {
		devices []models.Device
		count   int
		err     error
	}

	cases := []struct {
		description   string
		tenant        string
		pagination    paginator.Query
		filter        []models.Filter
		status        models.DeviceStatus
		sort, order   string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when the store device list fails when status is pending",
			tenant:      "tenant",
			pagination:  paginator.Query{Page: 1, PerPage: 10},
			filter: []models.Filter{
				{
					Type:   "property",
					Params: &models.PropertyParams{Name: "hostname", Operator: "eq"},
				},
			},
			status: status[0],
			sort:   "name",
			order:  order[0],
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 3,
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(namespace, nil).Once()
				mock.On("DeviceRemovedCount", ctx, namespace.TenantID).
					Return(int64(1), nil).Once()
				mock.On("DeviceList", ctx, paginator.Query{Page: 1, PerPage: 10}, []models.Filter{
					{
						Type:   "property",
						Params: &models.PropertyParams{Name: "hostname", Operator: "eq"},
					},
				}, status[0], "name", order[0], store.DeviceListModeMaxDeviceReached).
					Return(nil, 0, errors.New("error", "", 0)).Once()
			},
			expected: Expected{
				nil,
				0,
				errors.New("error", "", 0),
			},
		},
		{
			description: "fails when the store device list fails when status is not pending",
			tenant:      "tenant",
			pagination:  paginator.Query{Page: 1, PerPage: 10},
			filter: []models.Filter{
				{
					Type:   "property",
					Params: &models.PropertyParams{Name: "hostname", Operator: "eq"},
				},
			},
			status: status[1],
			sort:   "name",
			order:  order[1],
			requiredMocks: func() {
				filters := []models.Filter{
					{
						Type:   "property",
						Params: &models.PropertyParams{Name: "hostname", Operator: "eq"},
					},
				}

				mock.On("DeviceList", ctx, paginator.Query{Page: 1, PerPage: 10}, filters, status[1], "name", order[1], store.DeviceListModeDefault).
					Return(nil, 0, errors.New("error", "", 0)).Once()
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
			pagination:  paginator.Query{Page: 1, PerPage: 10},
			filter: []models.Filter{
				{
					Type:   "property",
					Params: &models.PropertyParams{Name: "hostname", Operator: "eq"},
				},
			},
			status: status[0],
			sort:   "name",
			order:  order[0],
			requiredMocks: func() {
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

				filters := []models.Filter{
					{
						Type:   "property",
						Params: &models.PropertyParams{Name: "hostname", Operator: "eq"},
					},
				}

				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(namespace, nil).Once()
				mock.On("DeviceRemovedCount", ctx, namespace.TenantID).
					Return(int64(1), nil).Once()
				mock.On("DeviceList", ctx, paginator.Query{Page: 1, PerPage: 10}, filters, status[0], "name", order[0], store.DeviceListModeMaxDeviceReached).
					Return(devices, len(devices), nil).Once()
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
			description: "succeeds when status is not pending",
			tenant:      "tenant",
			pagination:  paginator.Query{Page: 1, PerPage: 10},
			filter: []models.Filter{
				{
					Type:   "property",
					Params: &models.PropertyParams{Name: "hostname", Operator: "eq"},
				},
			},
			status: status[1],
			sort:   "name",
			order:  order[1],
			requiredMocks: func() {
				filters := []models.Filter{
					{
						Type:   "property",
						Params: &models.PropertyParams{Name: "hostname", Operator: "eq"},
					},
				}

				devices := []models.Device{
					{UID: "uid"},
					{UID: "uid2"},
					{UID: "uid3"},
				}

				mock.On("DeviceList", ctx, paginator.Query{Page: 1, PerPage: 10}, filters, status[1], "name", order[1], store.DeviceListModeDefault).
					Return(devices, len(devices), nil).Once()
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
			pagination:  paginator.Query{Page: 1, PerPage: 10},
			filter: []models.Filter{
				{
					Type:   "property",
					Params: &models.PropertyParams{Name: "hostname", Operator: "eq"},
				},
			},
			status: status[3],
			sort:   "name",
			order:  order[1],
			requiredMocks: func() {
				filters := []models.Filter{
					{
						Type:   "property",
						Params: &models.PropertyParams{Name: "hostname", Operator: "eq"},
					},
				}

				mock.On("DeviceRemovedList", ctx, "tenant", paginator.Query{Page: 1, PerPage: 10}, filters, "name", order[1]).
					Return(nil, 0, errors.New("error", "", 0)).Once()
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
			pagination:  paginator.Query{Page: 1, PerPage: 10},
			filter: []models.Filter{
				{
					Type:   "property",
					Params: &models.PropertyParams{Name: "hostname", Operator: "eq"},
				},
			},
			status: status[3],
			sort:   "name",
			order:  order[1],
			requiredMocks: func() {
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

				filters := []models.Filter{
					{
						Type:   "property",
						Params: &models.PropertyParams{Name: "hostname", Operator: "eq"},
					},
				}
				mock.On("DeviceRemovedList", ctx, "tenant", paginator.Query{Page: 1, PerPage: 10}, filters, "name", order[1]).
					Return(removedDevices, len(removedDevices), nil).Once()
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
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			returnedDevices, count, err := service.ListDevices(ctx, tc.tenant, tc.pagination, tc.filter, tc.status, tc.sort, tc.order)
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
				mock.On("NamespaceGet", ctx, "tenant").
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
					Members: []models.Member{{ID: "id",
						Role: guard.RoleOwner},
						{
							ID:   "id2",
							Role: guard.RoleObserver,
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
				mock.On("NamespaceGet", ctx, "tenant").
					Return(namespace, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Once()
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
					Members: []models.Member{{ID: "id",
						Role: guard.RoleOwner},
						{
							ID:   "id2",
							Role: guard.RoleObserver,
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
				mock.On("NamespaceGet", ctx, "tenant").
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
				mock.On("NamespaceGet", ctx, "tenant").
					Return(&models.Namespace{TenantID: "tenant"}, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				mock.On("DeviceDelete", ctx, models.UID(device.UID)).
					Return(nil).Once()
			},
			expected: nil,
		},
		{
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
					Members: []models.Member{{ID: "id", Role: guard.RoleOwner}, {ID: "id2", Role: guard.RoleObserver}},
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
		},
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
				mock.On("DeviceGetByName", ctx, "newname", "tenant").Return(device2, errors.New("error", "", 0)).Once()
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
				mock.On("DeviceGetByName", ctx, "newname", "tenant").Return(device2, nil).Once()
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
				mock.On("DeviceGetByName", ctx, "anewname", "tenant").Return(nil, store.ErrNoDocuments).Once()
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
				mock.On("DeviceGetByName", ctx, "anewname", "tenant").Return(nil, store.ErrNoDocuments).Once()
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

func TestUpdateDeviceStatus(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	cases := []struct {
		name          string
		uid           models.UID
		online        bool
		requiredMocks func()
		expected      error
	}{
		{
			name: "fails when store device online fails",
			uid:  models.UID("uid"),
			requiredMocks: func() {
				mock.On("DeviceSetOnline", ctx, models.UID("uid"), false).
					Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			name:   "succeeds",
			uid:    models.UID("uid"),
			online: true,
			requiredMocks: func() {
				online := true
				mock.On("DeviceSetOnline", ctx, models.UID("uid"), online).
					Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			err := service.UpdateDeviceStatus(ctx, tc.uid, tc.online)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestUpdatePendingStatus(t *testing.T) {
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
			description: "fails when the status is invalid",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			status:      models.DeviceStatus("invalid"),
			requiredMocks: func() {
			},
			expected: NewErrDeviceStatusInvalid("invalid", nil),
		},
		{
			description: "fails when the store get by uid fails",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			status:      models.DeviceStatusAccepted,
			requiredMocks: func() {
				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound("uid", errors.New("error", "", 0)),
		},
		{
			description: "fails when device already accepted",
			uid:         models.UID("uid"),
			status:      models.DeviceStatusAccepted,
			tenant:      "tenant",
			requiredMocks: func() {
				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").
					Return(&models.Device{Status: "accepted"}, nil).Once()
			},
			expected: NewErrDeviceStatusAccepted(nil),
		},
		{
			description: "fails to update accept",
			uid:         models.UID("uid"),
			status:      models.DeviceStatusPending,
			tenant:      "tenant",
			requiredMocks: func() {
				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").Return(&models.Device{Status: "accepted"}, nil).Once()
			},
			expected: NewErrDeviceStatusAccepted(nil),
		},
		{
			description: "fail when could not get namespace",
			uid:         models.UID("uid"),
			status:      models.DeviceStatusAccepted,
			tenant:      "tenant",
			requiredMocks: func() {
				device := &models.Device{
					UID:       "uid",
					Name:      "name",
					TenantID:  "tenant",
					Identity:  &models.DeviceIdentity{MAC: "mac"},
					CreatedAt: time.Time{},
				}

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").
					Return(device, nil).Once()
				mock.On("NamespaceGet", ctx, "tenant").
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrNamespaceNotFound("tenant", errors.New("error", "", 0)),
		},
		{
			description: "Test should fail when device removed get return a error that is not store.ErrNoDocuments",
			uid:         models.UID("uid"),
			status:      models.DeviceStatusAccepted,
			tenant:      "tenant",
			requiredMocks: func() {
				device := &models.Device{
					UID:       "uid",
					Name:      "name",
					TenantID:  "tenant",
					Identity:  &models.DeviceIdentity{MAC: "mac"},
					CreatedAt: time.Time{},
				}

				namespaceWithLimit := &models.Namespace{Name: "group1",
					Owner:        "id",
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 1,
					Members: []models.Member{
						{
							ID: "id", Role: guard.RoleOwner,
						},
						{
							ID: "id2", Role: guard.RoleObserver,
						},
					},
				}

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").
					Return(device, nil).Once()
				mock.On("NamespaceGet", ctx, "tenant").
					Return(namespaceWithLimit, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Once()
				mock.On("DeviceRemovedGet", ctx, "tenant", models.UID("uid")).
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceRemovedGet(errors.New("error", "", 0)),
		},
		{
			description: "Test should fail when device is not removed, but device removed list return a error",
			uid:         models.UID("uid"),
			status:      models.DeviceStatusAccepted,
			tenant:      "tenant",
			requiredMocks: func() {
				device := &models.Device{
					UID:       "uid",
					Name:      "name",
					TenantID:  "tenant",
					Identity:  &models.DeviceIdentity{MAC: "mac"},
					CreatedAt: time.Time{},
				}

				namespaceWithLimit := &models.Namespace{Name: "group1",
					Owner:        "id",
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 1,
					Members: []models.Member{
						{
							ID: "id", Role: guard.RoleOwner,
						},
						{
							ID: "id2", Role: guard.RoleObserver,
						},
					},
				}

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").
					Return(device, nil).Once()
				mock.On("NamespaceGet", ctx, "tenant").
					Return(namespaceWithLimit, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Once()
				mock.On("DeviceRemovedGet", ctx, "tenant", models.UID("uid")).
					Return(nil, store.ErrNoDocuments).Once()
				mock.On("DeviceRemovedCount", ctx, "tenant").
					Return(int64(0), errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceRemovedCount(errors.New("error", "", 0)),
		},
		{
			description: "Test should fail when device is not removed, but the device limit has been reached",
			uid:         models.UID("uid"),
			status:      models.DeviceStatusAccepted,
			tenant:      "tenant",
			requiredMocks: func() {
				device := &models.Device{
					UID:       "uid",
					Name:      "name",
					TenantID:  "tenant",
					Identity:  &models.DeviceIdentity{MAC: "mac"},
					CreatedAt: time.Time{},
				}

				namespaceWithLimit := &models.Namespace{Name: "group1",
					Owner:        "id",
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 1,
					Members: []models.Member{
						{
							ID: "id", Role: guard.RoleOwner,
						},
						{
							ID: "id2", Role: guard.RoleObserver,
						},
					},
				}

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").
					Return(device, nil).Once()
				mock.On("NamespaceGet", ctx, "tenant").
					Return(namespaceWithLimit, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Once()
				mock.On("DeviceRemovedGet", ctx, "tenant", models.UID("uid")).
					Return(nil, store.ErrNoDocuments).Once()
				mock.On("DeviceRemovedCount", ctx, "tenant").
					Return(int64(2), nil).Once()
			},
			expected: NewErrDeviceRemovedFull(3, nil),
		},
		{
			description: "Test should fail when device was removed, but device removed delete return a error",
			uid:         models.UID("uid"),
			status:      models.DeviceStatusAccepted,
			tenant:      "tenant",
			requiredMocks: func() {
				device := &models.Device{
					UID:       "uid",
					Name:      "name",
					TenantID:  "tenant",
					Identity:  &models.DeviceIdentity{MAC: "mac"},
					CreatedAt: time.Time{},
				}

				namespaceWithLimit := &models.Namespace{Name: "group1",
					Owner:        "id",
					TenantID:     "tenant",
					MaxDevices:   3,
					DevicesCount: 1,
					Members: []models.Member{
						{
							ID: "id", Role: guard.RoleOwner,
						},
						{
							ID: "id2", Role: guard.RoleObserver,
						},
					},
				}

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").
					Return(device, nil).Once()
				mock.On("NamespaceGet", ctx, "tenant").
					Return(namespaceWithLimit, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Once()
				mock.On("DeviceRemovedGet", ctx, "tenant", models.UID("uid")).
					Return(&models.DeviceRemoved{
						Device: &models.Device{
							UID:      device.UID,
							TenantID: "tenant",
						},
					}, nil).Once()
				mock.On("DeviceRemovedDelete", ctx, "tenant", models.UID("uid")).
					Return(errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceRemovedDelete(errors.New("error", "", 0)),
		},
		{
			description: "fails when the limit is exceeded",
			uid:         models.UID("uid_limit"),
			status:      models.DeviceStatusAccepted,
			tenant:      "tenant_max",
			requiredMocks: func() {
				namespaceExceedLimit := &models.Namespace{
					Name:         "group1",
					Owner:        "id",
					TenantID:     "tenant_max",
					MaxDevices:   3,
					DevicesCount: 3,
					Members: []models.Member{
						{
							ID: "id", Role: guard.RoleOwner,
						},
						{
							ID: "id2", Role: guard.RoleObserver,
						},
					},
				}

				deviceExceed := &models.Device{UID: "uid_limit",
					Name:     "name",
					TenantID: "tenant_max",
					Identity: &models.DeviceIdentity{MAC: "mac"},
					Status:   "pending",
				}

				mock.On("DeviceGetByUID", ctx, models.UID(deviceExceed.UID), deviceExceed.TenantID).
					Return(deviceExceed, nil).Once()
				mock.On("NamespaceGet", ctx, deviceExceed.TenantID).
					Return(namespaceExceedLimit, nil).Twice()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				mock.On("DeviceGetByMac", ctx, "mac", deviceExceed.TenantID, models.DeviceStatusAccepted).
					Return(nil, nil).Once()
			},
			expected: NewErrDeviceLimit(3, nil),
		},
		{
			description: "succeeds",
			uid:         models.UID("uid"),
			status:      models.DeviceStatusAccepted,
			tenant:      "tenant",
			requiredMocks: func() {
				device := &models.Device{
					UID:       "uid",
					Name:      "name",
					TenantID:  "tenant",
					Identity:  &models.DeviceIdentity{MAC: "mac"},
					CreatedAt: time.Time{},
				}

				oldDevice := &models.Device{UID: "uid2",
					Name:     "name",
					TenantID: "tenant",
					Identity: &models.DeviceIdentity{MAC: "mac"},
				}

				namespace := &models.Namespace{Name: "group1",
					Owner:      "id",
					TenantID:   "tenant",
					MaxDevices: -1,
					Members: []models.Member{
						{
							ID: "id", Role: guard.RoleOwner,
						},
						{
							ID: "id2", Role: guard.RoleObserver,
						},
					},
				}

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").
					Return(device, nil).Once()
				mock.On("NamespaceGet", ctx, "tenant").
					Return(namespace, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				mock.On("DeviceGetByMac", ctx, "mac", device.TenantID, models.DeviceStatusAccepted).
					Return(oldDevice, nil).Once()
				mock.On("SessionUpdateDeviceUID", ctx, models.UID(oldDevice.UID), models.UID("uid")).
					Return(nil).Once()
				mock.On("DeviceDelete", ctx, models.UID(oldDevice.UID)).
					Return(nil).Once()
				mock.On("DeviceRename", ctx, models.UID("uid"), oldDevice.Name).
					Return(nil).Once()
				mock.On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatusAccepted).
					Return(nil).Once()
			},
			expected: nil,
		},
		{
			description: "reports usage",
			uid:         models.UID("uid"),
			status:      models.DeviceStatusAccepted,
			tenant:      "tenant_max",
			requiredMocks: func() {
				device := &models.Device{
					UID:       "uid",
					Name:      "name",
					TenantID:  "tenant_max",
					Identity:  &models.DeviceIdentity{MAC: "mac"},
					CreatedAt: time.Time{},
				}

				namespaceBilling := &models.Namespace{
					Name:         "group1",
					Owner:        "id",
					TenantID:     "tenant_max",
					MaxDevices:   -1,
					DevicesCount: 10,
					Billing:      &models.Billing{Active: true},
					Members: []models.Member{
						{
							ID:   "id",
							Role: guard.RoleOwner,
						},
						{
							ID:   "id2",
							Role: guard.RoleObserver,
						},
					},
				}

				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespaceBilling, nil).Once()
				mock.On("NamespaceGet", ctx, namespaceBilling.TenantID).
					Return(namespaceBilling, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				mock.On("DeviceGetByUID", ctx, models.UID("uid"), device.TenantID).
					Return(device, nil).Once()
				mock.On("DeviceGetByMac", ctx, "mac", device.TenantID, models.DeviceStatusAccepted).
					Return(nil, nil).Once()
				clockMock.On("Now").Return(now).Twice()
				envMock.On("Get", "SHELLHUB_BILLING").Return(strconv.FormatBool(true)).Once()
				clientMock.On("ReportUsage", &models.UsageRecord{
					Device:    device,
					Inc:       true,
					Namespace: namespaceBilling,
					Timestamp: now.Unix(),
				}).Return(200, nil).Once()
				mock.On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatusAccepted).
					Return(nil).Once()
			},
			expected: nil,
		},
		{
			description: "fails to reports usage",
			uid:         models.UID("uid"),
			status:      models.DeviceStatusAccepted,
			tenant:      "tenant_max",
			requiredMocks: func() {
				device := &models.Device{
					UID:       "uid",
					Name:      "name",
					TenantID:  "tenant_max",
					Identity:  &models.DeviceIdentity{MAC: "mac"},
					CreatedAt: time.Time{},
				}

				namespaceBilling := &models.Namespace{
					Name:         "group1",
					Owner:        "id",
					TenantID:     "tenant_max",
					MaxDevices:   -1,
					DevicesCount: 10,
					Billing:      &models.Billing{Active: true},
					Members: []models.Member{
						{
							ID:   "id",
							Role: guard.RoleOwner,
						},
						{
							ID:   "id2",
							Role: guard.RoleObserver,
						},
					},
				}

				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespaceBilling, nil).Once()
				mock.On("NamespaceGet", ctx, namespaceBilling.TenantID).
					Return(namespaceBilling, nil).Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				mock.On("DeviceGetByUID", ctx, models.UID("uid"), device.TenantID).
					Return(device, nil).Once()
				mock.On("DeviceGetByMac", ctx, "mac", device.TenantID, models.DeviceStatusAccepted).
					Return(nil, nil).Once()
				clockMock.On("Now").Return(now).Twice()
				envMock.On("Get", "SHELLHUB_BILLING").Return(strconv.FormatBool(true)).Once()
				clientMock.On("ReportUsage", &models.UsageRecord{
					Namespace: namespaceBilling,
					Inc:       true,
					Device:    device,
					Timestamp: now.Unix(),
				}).Return(500, nil).Once()
			},
			expected: ErrReport,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
			err := service.UpdatePendingStatus(ctx, tc.uid, tc.status, tc.tenant)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestSetDevicePosition(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	locator := &mocksGeoIp.Locator{}

	cases := []struct {
		description   string
		requiredMocks func()
		uid           models.UID
		ip            string
		expected      error
	}{
		{
			description: "fails when DeviceSetPosition return error",
			requiredMocks: func() {
				positionGeoIP := geoip.Position{
					Longitude: 0,
					Latitude:  0,
				}
				positionDeviceModel := models.DevicePosition{
					Longitude: 0,
					Latitude:  0,
				}

				locator.On("GetPosition", net.ParseIP("127.0.0.1")).
					Return(positionGeoIP, nil).Once()
				mock.On("DeviceSetPosition", ctx, models.UID("uid"), positionDeviceModel).
					Return(errors.New("error", "", 0)).Once()
			},
			uid:      models.UID("uid"),
			ip:       "127.0.0.1",
			expected: errors.New("error", "", 0),
		},
		{
			description: "success",
			requiredMocks: func() {
				positionGeoIP := geoip.Position{
					Longitude: 0,
					Latitude:  0,
				}
				positionDeviceModel := models.DevicePosition{
					Longitude: 0,
					Latitude:  0,
				}

				locator.On("GetPosition", net.ParseIP("127.0.0.1")).
					Return(positionGeoIP, nil).Once()
				mock.On("DeviceSetPosition", ctx, models.UID("uid"), positionDeviceModel).
					Return(nil).Once()
			},
			uid:      models.UID("uid"),
			ip:       "127.0.0.1",
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, locator)
			err := service.SetDevicePosition(ctx, tc.uid, tc.ip)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestDeviceHeartbeat(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	uid := models.UID("uid")

	clockMock.On("Now").Return(now).Once()

	mock.On("DeviceSetOnline", ctx, uid, true).Return(nil).Once()

	service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)
	err := service.DeviceHeartbeat(ctx, uid)
	assert.NoError(t, err)

	mock.AssertExpectations(t)
}
