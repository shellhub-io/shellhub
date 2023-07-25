package services

import (
	"context"
	"net"
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
							Role: guard.RoleOwner,
						},
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

func TestOffineDevice(t *testing.T) {
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
			err := service.OffineDevice(ctx, tc.uid, tc.online)
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
			description: "fails namespace has reached the limit of devices in community instance",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
				mock.On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000").
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
