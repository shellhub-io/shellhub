package routes

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	svc "github.com/shellhub-io/shellhub/api/services"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestGetDevice(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title           string
		uid             string
		requiredMocks   func()
		expectedSession *models.Device
		expectedStatus  int
	}{
		{
			title: "returns Ok for a existing device",
			uid:   "123",
			requiredMocks: func() {
				mock.On("GetDevice", gomock.Anything, models.UID("123")).Return(&models.Device{}, nil)
			},
			expectedSession: &models.Device{},
			expectedStatus:  http.StatusOK,
		},
		{
			title: "returns Not Found for a non-existing device",
			uid:   "1234",
			requiredMocks: func() {
				mock.On("GetDevice", gomock.Anything, models.UID("1234")).Return(nil, svc.ErrDeviceNotFound)
			},
			expectedSession: nil,
			expectedStatus:  http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/devices/%s", tc.uid), nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)

			var session *models.Device
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}

			assert.Equal(t, tc.expectedSession, session)
		})
	}
}

func TestDeleteDevice(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		uid            string
		requiredMocks  func()
		expectedStatus int
	}{
		{
			title: "returns Ok when deleting an existing device",
			uid:   "123",
			requiredMocks: func() {
				mock.On("DeleteDevice", gomock.Anything, models.UID("123"), "").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			title: "returns Not Found when deleting a non-existing device ",
			uid:   "1234",
			requiredMocks: func() {
				mock.On("DeleteDevice", gomock.Anything, models.UID("1234"), "").Return(svc.ErrDeviceNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/devices/%s", tc.uid), nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestRenameDevice(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		updatePayload  requests.DeviceRename
		tenant         string
		requiredMocks  func(req requests.DeviceRename)
		expectedStatus int
	}{
		{
			title: "returns Ok when renaming an existing device",
			updatePayload: requests.DeviceRename{
				DeviceParam: requests.DeviceParam{UID: "123"},
				Name:        "name",
			},
			tenant: "tenant-id",
			requiredMocks: func(req requests.DeviceRename) {
				mock.On("RenameDevice", gomock.Anything, models.UID("123"), req.Name, "tenant-id").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			title: "returns Not Found when renaming a non-existing device",
			updatePayload: requests.DeviceRename{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				Name:        "name",
			},
			tenant: "tenant-id",
			requiredMocks: func(req requests.DeviceRename) {
				mock.On("RenameDevice", gomock.Anything, models.UID("1234"), req.Name, "tenant-id").Return(svc.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.updatePayload)

			jsonData, err := json.Marshal(tc.updatePayload)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/devices/%s", tc.updatePayload.UID), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-Tenant-ID", tc.tenant)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestGetDeviceByPublicURLAddress(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title           string
		address         string
		requiredMocks   func()
		expectedStatus  int
		expectedSession *models.Device
	}{
		{
			title:   "returns Ok when searching a device by the public URL address",
			address: "example",
			requiredMocks: func() {
				mock.On("GetDeviceByPublicURLAddress", gomock.Anything, "example").Return(&models.Device{}, nil)
			},
			expectedSession: &models.Device{},
			expectedStatus:  http.StatusOK,
		},
		{
			title:   "returns Not Found when searching a device by the public URL address",
			address: "exampleaddress",
			requiredMocks: func() {
				mock.On("GetDeviceByPublicURLAddress", gomock.Anything, "exampleaddress").Return(nil, svc.ErrDeviceNotFound)
			},
			expectedSession: nil,
			expectedStatus:  http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/internal/devices/public/%s", tc.address), nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)

			var session *models.Device
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}

			assert.Equal(t, tc.expectedSession, session)
		})
	}
}

func TestGetDeviceList(t *testing.T) {
	mock := new(mocks.Service)

	filter := []map[string]interface{}{
		{
			"type": "property",
			"params": map[string]interface{}{
				"name":     "name",
				"operator": "contains",
				"value":    "examplespace",
			},
		},
	}

	jsonData, err := json.Marshal(filter)
	if err != nil {
		assert.NoError(t, err)
	}

	filteb64 := base64.StdEncoding.EncodeToString(jsonData)

	cases := []struct {
		title           string
		filter          string
		queryPayload    filterQuery
		tenant          string
		requiredMocks   func(query filterQuery)
		expectedSession []models.Device
		expectedStatus  int
	}{
		{
			title: "returns Ok when a device list existing",
			queryPayload: filterQuery{
				Filter:  filteb64,
				Status:  models.DeviceStatus("online"),
				SortBy:  "name",
				OrderBy: "asc",
				Query: paginator.Query{
					Page:    1,
					PerPage: 10,
				},
			},
			tenant: "tenant-id",
			requiredMocks: func(query filterQuery) {
				query.Normalize()
				raw, err := base64.StdEncoding.DecodeString(query.Filter)
				if err != nil {
					assert.NoError(t, err)
				}

				var filters []models.Filter
				if err := json.Unmarshal(raw, &filters); len(raw) > 0 && err != nil {
					assert.NoError(t, err)
				}

				mock.On("ListDevices", gomock.Anything, "tenant-id", query.Query, filters, query.Status, query.SortBy, query.OrderBy).Return([]models.Device{}, 1, nil)
			},
			expectedSession: []models.Device{},
			expectedStatus:  http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.queryPayload)

			jsonData, err := json.Marshal(tc.queryPayload)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/devices", strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-Tenant-ID", tc.tenant)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)

			var session []models.Device
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}

			assert.Equal(t, tc.expectedSession, session)
		})
	}
}

func TestOfflineDevice(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		uid            string
		requiredMocks  func()
		expectedStatus int
	}{
		{
			title: "returns Ok for setting an existing device as offline",
			uid:   "123",
			requiredMocks: func() {
				mock.On("UpdateDeviceStatus", gomock.Anything, models.UID("123"), false).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			title: "returns Not Found for setting a non-existing device as offline",
			uid:   "1234",
			requiredMocks: func() {
				mock.On("UpdateDeviceStatus", gomock.Anything, models.UID("1234"), false).Return(svc.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/internal/devices/%s/offline", tc.uid), nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-Tenant-ID", "tenant-id")
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestLookupDevice(t *testing.T) {
	mock := new(mocks.Service)

	tests := []struct {
		title           string
		request         requests.DeviceLookup
		requiredMocks   func(requests.DeviceLookup)
		expectedSession *models.Device
		expectedStatus  int
	}{
		{
			title: "returns Ok for look up of a existing device",
			request: requests.DeviceLookup{
				Domain:    "example.com",
				Name:      "device1",
				Username:  "user1",
				IPAddress: "192.168.1.100",
			},

			requiredMocks: func(req requests.DeviceLookup) {
				mock.On("LookupDevice", gomock.Anything, req.Domain, req.Name).Return(&models.Device{}, nil)
			},
			expectedSession: &models.Device{},
			expectedStatus:  http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.request)

			jsonData, err := json.Marshal(tc.request)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodGet, "/internal/lookup", strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)

			var session *models.Device
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}

			assert.Equal(t, tc.expectedSession, session)
		})
	}
}

func TestHeartbeatDevice(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		uid            string
		requiredMocks  func()
		expectedStatus int
	}{
		{
			title: "returns Ok for heartbeat of a existing device",
			uid:   "123",
			requiredMocks: func() {
				mock.On("DeviceHeartbeat", gomock.Anything, models.UID("123")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			title: "returns Not Found for heartbeat non-existing device",
			uid:   "1234",
			requiredMocks: func() {
				mock.On("DeviceHeartbeat", gomock.Anything, models.UID("1234")).Return(svc.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/internal/devices/%s/heartbeat", tc.uid), nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-Tenant-ID", "tenant-id")
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestRemoveDeviceTag(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		updatePayload  requests.DeviceCreateTag
		requiredMocks  func(req requests.DeviceCreateTag)
		expectedStatus int
	}{
		{
			title: "returns Ok for remove device tag of a existing device",
			updatePayload: requests.DeviceCreateTag{
				DeviceParam: requests.DeviceParam{UID: "123"},
				TagBody:     requests.TagBody{Tag: "tag"},
			},

			requiredMocks: func(req requests.DeviceCreateTag) {
				mock.On("RemoveDeviceTag", gomock.Anything, models.UID("123"), req.Tag).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			title: "returns Not Found for remove device tag of a non-existing device",
			updatePayload: requests.DeviceCreateTag{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				TagBody:     requests.TagBody{Tag: "tag"},
			},
			requiredMocks: func(req requests.DeviceCreateTag) {
				mock.On("RemoveDeviceTag", gomock.Anything, models.UID("1234"), req.Tag).Return(svc.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.updatePayload)

			jsonData, err := json.Marshal(tc.updatePayload)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/devices/%s/tags/%s", tc.updatePayload.UID, tc.updatePayload.Tag), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-Tenant-ID", "tenant-id")
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestCreateDeviceTag(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		updatePayload  requests.DeviceCreateTag
		requiredMocks  func(req requests.DeviceCreateTag)
		expectedStatus int
	}{
		{
			title: "returns Ok for create device tag of a existing device",
			updatePayload: requests.DeviceCreateTag{
				DeviceParam: requests.DeviceParam{UID: "123"},
				TagBody:     requests.TagBody{Tag: "tag"},
			},

			requiredMocks: func(req requests.DeviceCreateTag) {
				mock.On("CreateDeviceTag", gomock.Anything, models.UID("123"), req.Tag).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			title: "returns Not Found for create device tag of a non-existing device",
			updatePayload: requests.DeviceCreateTag{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				TagBody:     requests.TagBody{Tag: "tag"},
			},
			requiredMocks: func(req requests.DeviceCreateTag) {
				mock.On("CreateDeviceTag", gomock.Anything, models.UID("1234"), req.Tag).Return(svc.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.updatePayload)

			jsonData, err := json.Marshal(tc.updatePayload)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/devices/%s/tags", tc.updatePayload.UID), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-Tenant-ID", "tenant-id")
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestUpdateDeviceTag(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		updatePayload  requests.DeviceUpdateTag
		requiredMocks  func(req requests.DeviceUpdateTag)
		expectedStatus int
	}{
		{
			title: "returns Ok for update device tag of a existing device",
			updatePayload: requests.DeviceUpdateTag{
				DeviceParam: requests.DeviceParam{UID: "123"},
				Tags:        []string{"tag1", "tag2"},
			},

			requiredMocks: func(req requests.DeviceUpdateTag) {
				mock.On("UpdateDeviceTag", gomock.Anything, models.UID("123"), req.Tags).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			title: "returns Not Found for update device tag of a non-existing device",
			updatePayload: requests.DeviceUpdateTag{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				Tags:        []string{"tag1", "tag2"},
			},
			requiredMocks: func(req requests.DeviceUpdateTag) {
				mock.On("UpdateDeviceTag", gomock.Anything, models.UID("1234"), req.Tags).Return(svc.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.updatePayload)

			jsonData, err := json.Marshal(tc.updatePayload)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/devices/%s/tags", tc.updatePayload.UID), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-Tenant-ID", "tenant-id")
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestUpdateDevice(t *testing.T) {
	mock := new(mocks.Service)
	name := "new device name"
	url := true

	cases := []struct {
		title          string
		updatePayload  requests.DeviceUpdate
		requiredMocks  func(req requests.DeviceUpdate)
		expectedStatus int
	}{
		{
			title: "returns Ok for updating device of a existing device",
			updatePayload: requests.DeviceUpdate{
				DeviceParam: requests.DeviceParam{UID: "123"},
				Name:        &name,
				PublicURL:   &url,
			},

			requiredMocks: func(req requests.DeviceUpdate) {
				mock.On("UpdateDevice", gomock.Anything, "tenant-id", models.UID("123"), req.Name, req.PublicURL).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			title: "returns Not Found for updating device of a non-existing device",
			updatePayload: requests.DeviceUpdate{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				Name:        &name,
				PublicURL:   &url,
			},
			requiredMocks: func(req requests.DeviceUpdate) {
				mock.On("UpdateDevice", gomock.Anything, "tenant-id", models.UID("1234"), req.Name, req.PublicURL).Return(svc.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.updatePayload)

			jsonData, err := json.Marshal(tc.updatePayload)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/devices/%s", tc.updatePayload.UID), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-Tenant-ID", "tenant-id")
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}
