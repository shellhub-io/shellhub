package deviceadm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func TestListDevices(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

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
	s := NewService(store.Store(mock))

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

func TestIsNamespaceOwner(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))
	ctx := context.TODO()

	Err := errors.New("error")

	namespace := &models.Namespace{
		Name:     "namespace1",
		Owner:    "id1",
		TenantID: "tenant1",
	}

	user1 := &models.User{
		Name:     "user1",
		Username: "username1",
		ID:       "id1",
	}

	user2 := &models.User{
		Name:     "user2",
		Username: "username2",
		ID:       "id2",
	}

	cases := []struct {
		name             string
		requiredMocks    func()
		tenant, username string
		err              error
	}{
		{
			name: "isNamespaceOwner fails when store namespace get fails",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenant2").
					Return(nil, Err).Once()
			},
			tenant: "tenant2",
			err:    Err,
		},
		{
			name: "isNamespaceOwner fails when store user get fails",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByUsername", ctx, "username3").
					Return(nil, Err).Once()
			},
			tenant:   namespace.TenantID,
			username: "username3",
			err:      Err,
		},
		{
			name: "isNamespaceOwner when the user is not the owner",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByUsername", ctx, user2.Username).
					Return(user2, nil).Once()
			},
			username: user2.Username,
			tenant:   namespace.TenantID,
			err:      ErrUnauthorized,
		},
		{
			name: "isNamespaceOwner succeeds",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByUsername", ctx, user1.Username).
					Return(user1, nil).Once()
			},
			username: user1.Username,
			tenant:   namespace.TenantID,
			err:      nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			err := s.isNamespaceOwner(ctx, tc.tenant, tc.username)
			assert.Equal(t, tc.err, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestDeleteDevice(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	user := &models.User{Name: "name", Email: "", Username: "username", ID: "id"}
	user2 := &models.User{Name: "name2", Email: "", Username: "username2", ID: "id2"}
	namespace := &models.Namespace{Name: "group1", Owner: "id", TenantID: "tenant"}
	device := &models.Device{UID: "uid", TenantID: "tenant"}

	Err := errors.New("error")

	cases := []struct {
		name             string
		requiredMocks    func()
		uid              models.UID
		tenant, username string
		expected         error
	}{
		{
			name: "DeleteDevice fails when the user is not the owner",
			uid:  models.UID("uid"),
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByUsername", ctx, user2.Username).
					Return(user2, nil).Once()
			},
			username: user2.Username,
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
				mock.On("UserGetByUsername", ctx, user.Username).
					Return(user, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID("_uid"), namespace.TenantID).
					Return(nil, Err).Once()
			},
			username: user.Username,
			expected: Err,
		},
		{
			name:   "DeleteDevice fails when the store device delete fails",
			uid:    models.UID(device.UID),
			tenant: namespace.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByUsername", ctx, user.Username).
					Return(user, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), namespace.TenantID).
					Return(nil, nil).Once()
				mock.On("DeviceDelete", ctx, models.UID(device.UID)).
					Return(Err).Once()
			},
			username: user.Username,
			expected: Err,
		},
		{
			name:   "DeleteDevice succeeds",
			uid:    models.UID(device.UID),
			tenant: namespace.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByUsername", ctx, user.Username).
					Return(user, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), namespace.TenantID).
					Return(nil, nil).Once()
				mock.On("DeviceDelete", ctx, models.UID(device.UID)).
					Return(nil).Once()
			},
			username: user.Username,
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			err := s.DeleteDevice(ctx, tc.uid, tc.tenant, tc.username)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestRenameDevice(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	user := &models.User{Name: "name", Email: "email", Username: "username", ID: "id"}
	user2 := &models.User{Name: "name2", Email: "email2", Username: "username2", ID: "id2"}
	namespace := &models.Namespace{Name: "group1", Owner: "id", TenantID: "tenant"}
	device := &models.Device{UID: "uid", Name: "name", TenantID: "tenant"}
	device2 := &models.Device{UID: "uid2", Name: "newname", TenantID: "tenant2"}
	Err := errors.New("error")

	cases := []struct {
		name             string
		requiredMocks    func()
		uid              models.UID
		expected         error
		deviceNewName    string
		tenant, username string
	}{
		{
			name:     "RenameDevice fails when the user is not the owner",
			tenant:   namespace.TenantID,
			username: user2.Username,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByUsername", ctx, user2.Username).
					Return(user2, nil).Once()
			},
			expected: ErrUnauthorized,
		},
		{
			name:     "RenameDevice fails when store device get fails",
			tenant:   namespace.TenantID,
			uid:      models.UID(device.UID),
			username: user.Username,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByUsername", ctx, user.Username).
					Return(user, nil).Once()
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
			username:      user.Username,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByUsername", ctx, user.Username).
					Return(user, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), namespace.TenantID).
					Return(device, nil).Once()
			},
			expected: validator.ErrBadRequest,
		},
		{
			name:          "RenameDevice returns nil if the name is the same",
			tenant:        namespace.TenantID,
			deviceNewName: "name",
			uid:           models.UID(device.UID),
			username:      user.Username,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByUsername", ctx, user.Username).
					Return(user, nil).Once()
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
			username:      user.Username,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByUsername", ctx, user.Username).
					Return(user, nil).Once()
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
			username:      user.Username,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByUsername", ctx, user.Username).
					Return(user, nil).Once()
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
			username:      user.Username,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByUsername", ctx, user.Username).
					Return(user, nil).Once()
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
			username:      user.Username,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("UserGetByUsername", ctx, user.Username).
					Return(user, nil).Once()
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
			err := s.RenameDevice(ctx, tc.uid, tc.deviceNewName, tc.tenant, tc.username)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestLookupDevice(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

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
	s := NewService(store.Store(mock))

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
	s := NewService(store.Store(mock))

	user := &models.User{Name: "name", Username: "username", ID: "id"}
	user2 := &models.User{Name: "name2", Username: "username2", ID: "id2"}
	namespace := &models.Namespace{Name: "group1", Owner: "id", TenantID: "tenant", MaxDevices: -1}
	identity := &models.DeviceIdentity{MAC: "mac"}
	device := &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: identity}

	Err := errors.New("error")

	ctx := context.TODO()

	cases := []struct {
		name                     string
		uid                      models.UID
		status, tenant, username string
		requiredMocks            func()
		expected                 error
	}{
		{
			name:     "UpdatePendingStatus fails when the user is not the owner",
			uid:      models.UID("uid"),
			tenant:   namespace.TenantID,
			username: user2.Username,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user2.Username).
					Return(user2, nil).Once()
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
			},
			expected: ErrUnauthorized,
		},
		{
			name:     "UpdatePendingStatus fails when the store get by uid fails",
			uid:      models.UID("uid"),
			tenant:   namespace.TenantID,
			username: user.Username,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).
					Return(user, nil).Once()
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID("uid"), namespace.TenantID).
					Return(nil, Err).Once()
			},
			expected: Err,
		},
		{
			name:     "UpdatePendingStatus fails when the limit is exceeded",
			uid:      models.UID("uid_limit"),
			tenant:   "tenant_max",
			username: user.Username,
			requiredMocks: func() {
				namespaceExceedLimit := &models.Namespace{Name: "group1", Owner: "id", TenantID: "tenant_max", MaxDevices: 3, DevicesCount: 3}
				deviceExceed := &models.Device{UID: "uid_limit", Name: "name", TenantID: "tenant_max", Identity: identity, Status: "pending"}
				mock.On("UserGetByUsername", ctx, user.Username).
					Return(user, nil).Once()
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
			name:     "UpdatePendingStatus succeeds",
			uid:      models.UID("uid"),
			tenant:   namespace.TenantID,
			username: user.Username,
			requiredMocks: func() {
				oldDevice := &models.Device{UID: "old_uid", Name: "name", TenantID: "tenant", Identity: identity}
				mock.On("UserGetByUsername", ctx, user.Username).
					Return(user, nil).Once()
				mock.On("NamespaceGet", ctx, device.TenantID).
					Return(namespace, nil).Once()
				mock.On("DeviceGetByUID", ctx, models.UID(device.UID), device.TenantID).
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
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()
			err := s.UpdatePendingStatus(ctx, tc.uid, "accepted", tc.tenant, tc.username)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
