package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/geoip"
	mocksGeoIp "github.com/shellhub-io/shellhub/pkg/geoip/mocks"

	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

const (
	invalidUID    = "Fails to find the device invalid uid"
	invalidFormat = "Fails invalid format for name"
)

func TestListDevices(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	Err := errors.New("error")

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

	filterJSON, err := json.Marshal(filters)
	assert.NoError(t, err)

	encodedFilter := base64.StdEncoding.EncodeToString(filterJSON)

	query := paginator.Query{Page: 1, PerPage: 10}

	status := []string{"pending", "accepted", "rejected"}
	sort := "name"
	order := []string{"asc", "desc"}

	type Expected struct {
		devices []models.Device
		count   int
		err     error
	}

	cases := []struct {
		name                           string
		pagination                     paginator.Query
		requiredMocks                  func()
		expected                       Expected
		filterB64, status, sort, order string
	}{
		{
			name:       "ListDevices fails when the store device list fails",
			pagination: query,
			filterB64:  encodedFilter,
			status:     status[0],
			sort:       sort,
			order:      order[0],
			requiredMocks: func() {
				mock.On("DeviceList", ctx, query, filters, status[0], sort, order[0]).
					Return(nil, 0, Err).Once()
			},
			expected: Expected{
				nil,
				0,
				Err,
			},
		},
		{
			name:       "ListDevices succeeds",
			pagination: query,
			filterB64:  encodedFilter,
			status:     status[0],
			sort:       sort,
			order:      order[0],
			requiredMocks: func() {
				mock.On("DeviceList", ctx, query, filters, status[0], sort, order[0]).
					Return(devices, len(devices), nil).Once()
			},
			expected: Expected{
				devices,
				len(devices),
				nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(*testing.T) {
			tc.requiredMocks()
			returnedDevices, count, err := s.ListDevices(ctx, tc.pagination, tc.filterB64, tc.status, tc.sort, tc.order)
			assert.Equal(t, tc.expected, Expected{returnedDevices, count, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestGetDevice(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	Err := errors.New("error")

	ctx := context.TODO()

	device := &models.Device{UID: "uid"}

	type Expected struct {
		device *models.Device
		err    error
	}

	cases := []struct {
		name          string
		requiredMocks func()
		uid           models.UID
		expected      Expected
	}{
		{
			name: "GetDevice fails when the store get device fails",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID("_uid")).
					Return(nil, Err).Once()
			},
			uid: models.UID("_uid"),
			expected: Expected{
				nil,
				Err,
			},
		},
		{
			name: "GetDevice succeeds",
			requiredMocks: func() {
				mock.On("DeviceGet", ctx, models.UID(device.UID)).
					Return(device, nil).Once()
			},
			uid: models.UID("uid"),
			expected: Expected{
				device,
				nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			returnedDevice, err := s.GetDevice(ctx, tc.uid)
			assert.Equal(t, tc.expected, Expected{returnedDevice, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestDeleteDevice(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	user := &models.User{Name: "name", Email: "", Username: "username", ID: "id"}
	user2 := &models.User{Name: "name2", Email: "", Username: "username2", ID: "id2"}
	namespace := &models.Namespace{Name: "group1", Owner: "id", TenantID: "tenant"}
	device := &models.Device{UID: "uid", TenantID: "tenant", CreatedAt: time.Time{}}

	Err := errors.New("error")

	cases := []struct {
		name          string
		requiredMocks func()
		uid           models.UID
		tenant, id    string
		expected      error
	}{
		{
			name: "DeleteDevice fails when the user is not the owner",
			uid:  models.UID("uid"),
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user2.ID, false).
					Return(user2, 0, nil).Once()
			},
			id:       user2.ID,
			tenant:   namespace.TenantID,
			expected: ErrUnauthorized,
		},
		{
			name:   "DeleteDevice fails when the store device get by uid fails",
			uid:    models.UID("_uid"),
			tenant: namespace.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).
					Return(user, 0, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID("_uid"), namespace.TenantID).
					Return(nil, Err).Once()
			},
			id:       user.ID,
			expected: Err,
		},
		{
			name:   "DeleteDevice fails when the store device delete fails",
			uid:    models.UID(device.UID),
			tenant: namespace.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).
					Return(user, 0, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), namespace.TenantID).
					Return(nil, nil).Once()
				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(namespace, nil).Once()
				mock.On("DeviceDelete", ctx, models.UID(device.UID)).
					Return(Err).Once()
			},
			id:       user.ID,
			expected: Err,
		},
		{
			name:   "DeleteDevice succeeds",
			uid:    models.UID(device.UID),
			tenant: namespace.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).
					Return(user, 0, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), namespace.TenantID).
					Return(nil, nil).Once()
				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(&models.Namespace{
						Billing: &models.Billing{
							Active: true,
						},
					}, nil).Once()
				mock.On("DeviceDelete", ctx, models.UID(device.UID)).
					Return(nil).Once()
			},
			id:       user.ID,
			expected: nil,
		},
		{
			name:   "DeleteDevice reports usage with success",
			uid:    models.UID(device.UID),
			tenant: namespace.TenantID,
			requiredMocks: func() {
				namespaceBilling := &models.Namespace{
					Name:       "namespace1",
					MaxDevices: -1,
					Billing: &models.Billing{
						Active: true,
					},
				}
				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).
					Return(user, 0, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), namespace.TenantID).
					Return(device, nil).Once()
				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(namespaceBilling, nil).Once()
				clockMock.On("Now").Return(now).Twice()
				clientMock.On("ReportUsage", &models.UsageRecord{
					UUID:      "uid",
					Namespace: namespaceBilling,
					Created:   strconv.Itoa(int(device.CreatedAt.Unix())),
					Timestamp: now.Unix(),
				},
					"billing-api").Return(200, nil).Once()
				mock.On("DeviceDelete", ctx, models.UID(device.UID)).
					Return(nil).Once()
			},
			id:       user.ID,
			expected: nil,
		},
		{
			name:   "DeleteDevice fails to report usage",
			uid:    models.UID(device.UID),
			tenant: namespace.TenantID,
			requiredMocks: func() {
				namespaceBilling := &models.Namespace{
					Name:       "namespace1",
					MaxDevices: -1,
					Billing: &models.Billing{
						Active: true,
					},
				}
				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).
					Return(user, 0, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), namespace.TenantID).
					Return(device, nil).Once()
				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(namespaceBilling, nil).Once()
				clockMock.On("Now").Return(now).Twice()
				clientMock.On("ReportUsage", &models.UsageRecord{
					UUID:      "uid",
					Namespace: namespaceBilling,
					Created:   strconv.Itoa(int(device.CreatedAt.Unix())),
					Timestamp: now.Unix(),
				},
					"billing-api").Return(500, nil).Once()
			},
			id:       user.ID,
			expected: ErrReportUsage,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			err := s.DeleteDevice(ctx, tc.uid, tc.tenant, tc.id)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestRenameDevice(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	user := &models.User{Name: "name", Email: "email", Username: "username", ID: "id"}
	user2 := &models.User{Name: "name2", Email: "email2", Username: "username2", ID: "id2"}
	namespace := &models.Namespace{Name: "group1", Owner: "id", TenantID: "tenant"}
	device := &models.Device{UID: "uid", Name: "name", TenantID: "tenant"}
	device2 := &models.Device{UID: "uid2", Name: "newname", TenantID: "tenant2"}
	Err := errors.New("error")

	cases := []struct {
		name          string
		requiredMocks func()
		uid           models.UID
		expected      error
		deviceNewName string
		tenant, id    string
	}{
		{
			name:   "RenameDevice fails when the user is not the owner",
			tenant: namespace.TenantID,
			id:     user2.ID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user2.ID, false).
					Return(user2, 0, nil).Once()
			},
			expected: ErrUnauthorized,
		},
		{
			name:   "RenameDevice fails when store device get fails",
			tenant: namespace.TenantID,
			uid:    models.UID(device.UID),
			id:     user.ID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).
					Return(user, 0, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), namespace.TenantID).
					Return(device, Err).Once()
			},
			expected: Err,
		},
		{
			name:          "RenameDevice fails when the name is invalid",
			tenant:        namespace.TenantID,
			deviceNewName: "---invalid...",
			uid:           models.UID(device.UID),
			id:            user.ID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).
					Return(user, 0, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), namespace.TenantID).
					Return(device, nil).Once()
			},
			expected: ErrInvalidFormat,
		},
		{
			name:          "RenameDevice returns nil if the name is the same",
			tenant:        namespace.TenantID,
			deviceNewName: "name",
			uid:           models.UID(device.UID),
			id:            user.ID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).
					Return(user, 0, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), namespace.TenantID).
					Return(device, nil).Once()
			},
			expected: nil,
		},
		{
			name:          "RenameDevice fails when store get by device name fails",
			tenant:        namespace.TenantID,
			deviceNewName: "newname",
			uid:           models.UID(device.UID),
			id:            user.ID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).
					Return(user, 0, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), namespace.TenantID).Return(device, nil).Once()
				mock.On("DeviceGetByName", ctx, "newname", namespace.TenantID).
					Return(device2, Err).Once()
			},
			expected: Err,
		},
		{
			name:          "RenameDevice fails when the name already exists",
			tenant:        namespace.TenantID,
			deviceNewName: "newname",
			uid:           models.UID(device.UID),
			id:            user.ID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).
					Return(user, 0, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), namespace.TenantID).Return(device, nil).Once()
				mock.On("DeviceGetByName", ctx, "newname", namespace.TenantID).
					Return(device2, nil).Once()
			},
			expected: ErrDuplicatedDeviceName,
		},
		{
			name:          "RenameDevice fails when the store device rename fails",
			tenant:        namespace.TenantID,
			deviceNewName: "anewname",
			uid:           models.UID(device.UID),
			id:            user.ID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).
					Return(user, 0, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), namespace.TenantID).Return(device, nil).Once()
				mock.On("DeviceGetByName", ctx, "anewname", namespace.TenantID).
					Return(nil, store.ErrNoDocuments).Once()
				mock.On("DeviceRename", ctx, models.UID(device.UID), "anewname").
					Return(Err).Once()
			},
			expected: Err,
		},
		{
			name:          "RenameDevice succeeds",
			tenant:        namespace.TenantID,
			deviceNewName: "anewname",
			uid:           models.UID(device.UID),
			id:            user.ID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByID", ctx, user.ID, false).
					Return(user, 0, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), namespace.TenantID).Return(device, nil).Once()
				mock.On("DeviceGetByName", ctx, "anewname", namespace.TenantID).
					Return(nil, store.ErrNoDocuments).Once()
				mock.On("DeviceRename", ctx, models.UID(device.UID), "anewname").
					Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			err := s.RenameDevice(ctx, tc.uid, tc.deviceNewName, tc.tenant, tc.id)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestLookupDevice(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	device := &models.Device{UID: "uid", Name: "name", TenantID: "tenant"}
	namespace := &models.Namespace{Name: "namespace"}
	Err := errors.New("error")

	type Expected struct {
		device *models.Device
		err    error
	}

	cases := []struct {
		name          string
		namespace     string
		deviceName    string
		requiredMocks func()
		expected      Expected
	}{
		{
			name:       "LookupDevice fails when store device lookup fails",
			namespace:  namespace.Name,
			deviceName: device.Name,
			requiredMocks: func() {
				mock.On("DeviceLookup", ctx, namespace.Name, device.Name).
					Return(nil, Err).Once()
			},
			expected: Expected{
				nil,
				Err,
			},
		},
		{
			name:       "LookupDevice fails when the device is not found",
			namespace:  namespace.Name,
			deviceName: device.Name,
			requiredMocks: func() {
				mock.On("DeviceLookup", ctx, namespace.Name, device.Name).
					Return(nil, store.ErrNoDocuments).Once()
			},
			expected: Expected{
				nil,
				ErrNotFound,
			},
		},
		{
			name:       "LookupDevice succeeds",
			namespace:  namespace.Name,
			deviceName: device.Name,
			requiredMocks: func() {
				mock.On("DeviceLookup", ctx, namespace.Name, device.Name).
					Return(device, nil).Once()
			},
			expected: Expected{
				device,
				nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			returnedDevice, err := s.LookupDevice(ctx, tc.namespace, tc.deviceName)
			assert.Equal(t, tc.expected, Expected{returnedDevice, err})
		})
	}
	mock.AssertExpectations(t)
}

func TestUpdateDeviceStatus(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	Err := errors.New("error")

	ctx := context.TODO()

	cases := []struct {
		name          string
		uid           models.UID
		online        bool
		requiredMocks func()
		expected      error
	}{
		{
			name: "UpdateDeviceStatus fails when store device online fails",
			uid:  models.UID("uid"),
			requiredMocks: func() {
				mock.On("DeviceSetOnline", ctx, models.UID("uid"), false).
					Return(Err).Once()
			},
			expected: Err,
		},
		{
			name:   "UpdateDeviceStatus succeeds",
			uid:    models.UID("uid"),
			online: true,
			requiredMocks: func() {
				online := true
				mock.On("DeviceSetOnline", ctx, models.UID("uid"), online).
					Return(Err).Once()
			},
			expected: Err,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			err := s.UpdateDeviceStatus(ctx, tc.uid, tc.online)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestUpdatePendingStatus(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	user := &models.User{Name: "name", Username: "username", ID: "id"}
	user2 := &models.User{Name: "name2", Username: "username2", ID: "id2"}
	namespace := &models.Namespace{Name: "group1", Owner: "id", TenantID: "tenant", MaxDevices: -1}
	identity := &models.DeviceIdentity{MAC: "mac"}
	device := &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: identity, CreatedAt: time.Time{}}

	Err := errors.New("error")

	ctx := context.TODO()

	cases := []struct {
		name               string
		uid                models.UID
		status, tenant, id string
		requiredMocks      func()
		expected           error
	}{
		{
			name:   "UpdatePendingStatus fails when the user is not the owner",
			uid:    models.UID("uid"),
			tenant: namespace.TenantID,
			id:     user2.ID,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user2.ID, false).
					Return(user2, 0, nil).Once()
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
			},
			expected: ErrUnauthorized,
		},
		{
			name:   "UpdatePendingStatus fails when the store get by uid fails",
			uid:    models.UID("uid"),
			tenant: namespace.TenantID,
			id:     user.ID,
			requiredMocks: func() {
				mock.On("UserGetByID", ctx, user.ID, false).
					Return(user, 0, nil).Once()
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID("uid"), namespace.TenantID).
					Return(nil, Err).Once()
			},
			expected: Err,
		},
		{
			name:   "UpdatePendingStatus fails when the limit is exceeded",
			uid:    models.UID("uid_limit"),
			tenant: "tenant_max",
			id:     user.ID,
			requiredMocks: func() {
				namespaceExceedLimit := &models.Namespace{Name: "group1", Owner: "id", TenantID: "tenant_max", MaxDevices: 3, DevicesCount: 3}
				deviceExceed := &models.Device{UID: "uid_limit", Name: "name", TenantID: "tenant_max", Identity: identity, Status: "pending"}
				mock.On("UserGetByID", ctx, user.ID, false).
					Return(user, 0, nil).Once()
				mock.On("NamespaceGet", ctx, deviceExceed.TenantID).
					Return(namespaceExceedLimit, nil).Twice()
				mock.On("DeviceGetByUID", ctx, models.UID(deviceExceed.UID), deviceExceed.TenantID).
					Return(deviceExceed, nil).Once()
				mock.On("DeviceGetByMac", ctx, "mac", deviceExceed.TenantID, "accepted").
					Return(nil, nil).Once()
			},
			expected: ErrMaxDeviceCountReached,
		},
		{
			name:   "UpdatePendingStatus succeeds",
			uid:    models.UID("uid"),
			tenant: namespace.TenantID,
			id:     user.ID,
			requiredMocks: func() {
				oldDevice := &models.Device{UID: "uid2", Name: "name", TenantID: "tenant", Identity: identity}
				mock.On("UserGetByID", ctx, user.ID, false).
					Return(user, 0, nil).Once()
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID("uid"), namespace.TenantID).
					Return(device, nil).Once()
				mock.On("DeviceGetByMac", ctx, "mac", device.TenantID, "accepted").
					Return(oldDevice, nil).Once()
				mock.On("SessionUpdateDeviceUID", ctx, models.UID(oldDevice.UID), models.UID(device.UID)).
					Return(nil).Once()
				mock.On("DeviceDelete", ctx, models.UID(oldDevice.UID)).
					Return(nil).Once()
				mock.On("DeviceRename", ctx, models.UID(device.UID), oldDevice.Name).
					Return(nil).Once()
				mock.On("DeviceUpdateStatus", ctx, models.UID(device.UID), "accepted").
					Return(nil).Once()
			},
			expected: nil,
		},
		{
			name:   "UpdatePendingStatus reports usage",
			uid:    models.UID("uid"),
			tenant: "tenant_max",
			id:     user.ID,
			requiredMocks: func() {
				namespaceBilling := &models.Namespace{Name: "group1", Owner: "id", TenantID: "tenant_max", MaxDevices: -1, DevicesCount: 10, Billing: &models.Billing{Active: true}}
				device := &models.Device{UID: "uid", Name: "name", TenantID: "tenant_max", Identity: identity, Status: "pending"}
				mock.On("UserGetByID", ctx, user.ID, false).
					Return(user, 0, nil).Once()
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespaceBilling, nil).Twice()
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), device.TenantID).
					Return(device, nil).Once()
				mock.On("DeviceGetByMac", ctx, "mac", device.TenantID, "accepted").
					Return(nil, nil).Once()
				clockMock.On("Now").Return(now).Twice()
				clientMock.On("ReportUsage", &models.UsageRecord{
					UUID:      "uid",
					Inc:       true,
					Namespace: namespaceBilling,
					Created:   strconv.Itoa(int(device.CreatedAt.Unix())),
					Timestamp: now.Unix(),
				},
					"billing-api").Return(200, nil).Once()
				mock.On("DeviceUpdateStatus", ctx, models.UID(device.UID), "accepted").
					Return(nil).Once()
			},
			expected: nil,
		},
		{
			name:   "UpdatePendingStatus fails to reports usage",
			uid:    models.UID("uid"),
			tenant: "tenant_max",
			id:     user.ID,
			requiredMocks: func() {
				namespaceBilling := &models.Namespace{Name: "group1", Owner: "id", TenantID: "tenant_max", MaxDevices: -1, DevicesCount: 10, Billing: &models.Billing{Active: true}}
				device := &models.Device{UID: "uid", Name: "name", TenantID: "tenant_max", Identity: identity, Status: "pending"}
				mock.On("UserGetByID", ctx, user.ID, false).
					Return(user, 0, nil).Once()
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespaceBilling, nil).Twice()
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), device.TenantID).
					Return(device, nil).Once()
				mock.On("DeviceGetByMac", ctx, "mac", device.TenantID, "accepted").
					Return(nil, nil).Once()
				clockMock.On("Now").Return(now).Twice()
				clientMock.On("ReportUsage", &models.UsageRecord{
					UUID:      "uid",
					Inc:       true,
					Namespace: namespaceBilling,
					Created:   strconv.Itoa(int(device.CreatedAt.Unix())),
					Timestamp: now.Unix(),
				},
					"billing-api").Return(500, nil).Once()
			},
			expected: ErrReportUsage,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			err := s.UpdatePendingStatus(ctx, tc.uid, "accepted", tc.tenant, tc.id)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestSetDevicePosition(t *testing.T) {
	locator := &mocksGeoIp.Locator{}
	mock := &mocks.Store{}
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, locator)

	ctx := context.TODO()

	device := &models.Device{UID: "uid"}

	Err := errors.New("error")

	positionGeoIP := geoip.Position{Longitude: 0, Latitude: 0}
	positionDeviceModel := models.DevicePosition{Longitude: 0, Latitude: 0}

	cases := []struct {
		name          string
		requiredMocks func()
		uid           models.UID
		ip            string
		expected      error
	}{
		{
			name: "SetDevicePosition fails when DeviceSetPosition return error",
			requiredMocks: func() {
				locator.On("GetPosition", net.ParseIP("127.0.0.1")).
					Return(positionGeoIP, nil).Once()
				mock.On("DeviceSetPosition", ctx, models.UID(device.UID), positionDeviceModel).
					Return(Err).Once()
			},
			uid:      models.UID(device.UID),
			ip:       "127.0.0.1",
			expected: Err,
		},
		{
			name: "SetDevicePosition success",
			requiredMocks: func() {
				locator.On("GetPosition", net.ParseIP("127.0.0.1")).
					Return(positionGeoIP, nil).Once()
				mock.On("DeviceSetPosition", ctx, models.UID(device.UID), positionDeviceModel).
					Return(nil).Once()
			},
			uid:      models.UID(device.UID),
			ip:       "127.0.0.1",
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()

			err := s.SetDevicePosition(ctx, tc.uid, "127.0.0.1")
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

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
