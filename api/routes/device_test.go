package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestGetDevice(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedSession *models.Device
		expectedStatus  int
	}
	cases := []struct {
		title         string
		uid           string
		requiredMocks func()
		expected      Expected
	}{
		{
			title:         "fails when bind fails to validate uid",
			uid:           "",
			requiredMocks: func() {},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title: "fails when try to get a non-existing device",
			uid:   "1234",
			requiredMocks: func() {
				mock.On("GetDevice", gomock.Anything, models.UID("1234")).Return(nil, svc.ErrDeviceNotFound)
			},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title: "success when try to get a existing device",
			uid:   "123",
			requiredMocks: func() {
				mock.On("GetDevice", gomock.Anything, models.UID("123")).Return(&models.Device{}, nil)
			},
			expected: Expected{
				expectedSession: &models.Device{},
				expectedStatus:  http.StatusOK,
			},
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

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session *models.Device
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}

			assert.Equal(t, tc.expected.expectedSession, session)
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
			title:          "fails when bind fails to validate uid",
			uid:            "",
			requiredMocks:  func() {},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "fails when try to deleting a non-existing device",
			uid:   "1234",
			requiredMocks: func() {
				mock.On("DeleteDevice", gomock.Anything, models.UID("1234"), "").Return(svc.ErrDeviceNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to deleting an existing device",
			uid:   "123",
			requiredMocks: func() {
				mock.On("DeleteDevice", gomock.Anything, models.UID("123"), "").Return(nil)
			},
			expectedStatus: http.StatusOK,
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
		renamePayload  requests.DeviceRename
		tenant         string
		requiredMocks  func(req requests.DeviceRename)
		expectedStatus int
	}{
		{
			title: "fails when bind fails to validate uid",
			renamePayload: requests.DeviceRename{
				DeviceParam: requests.DeviceParam{UID: ""},
			},
			tenant:         "tenant-id",
			requiredMocks:  func(req requests.DeviceRename) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "fails when try to rename a non-existing device",
			renamePayload: requests.DeviceRename{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				Name:        "name",
			},
			tenant: "tenant-id",
			requiredMocks: func(req requests.DeviceRename) {
				mock.On("RenameDevice", gomock.Anything, models.UID("1234"), req.Name, "tenant-id").Return(svc.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to rename an existing device",
			renamePayload: requests.DeviceRename{
				DeviceParam: requests.DeviceParam{UID: "123"},
				Name:        "name",
			},
			tenant: "tenant-id",
			requiredMocks: func(req requests.DeviceRename) {
				mock.On("RenameDevice", gomock.Anything, models.UID("123"), req.Name, "tenant-id").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.renamePayload)

			jsonData, err := json.Marshal(tc.renamePayload)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/devices/%s", tc.renamePayload.UID), strings.NewReader(string(jsonData)))
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

	type Expected struct {
		expectedSession *models.Device
		expectedStatus  int
	}
	cases := []struct {
		title         string
		address       string
		requiredMocks func()
		expected      Expected
	}{
		{
			title:         "fails when bind fails to validate uid",
			address:       "",
			requiredMocks: func() {},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title:   "fails when try to searching a device by the public URL address",
			address: "exampleaddress",
			requiredMocks: func() {
				mock.On("GetDeviceByPublicURLAddress", gomock.Anything, "exampleaddress").Return(nil, svc.ErrDeviceNotFound)
			},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title:   "success when try to searching a device by the public URL address",
			address: "example",
			requiredMocks: func() {
				mock.On("GetDeviceByPublicURLAddress", gomock.Anything, "example").Return(&models.Device{}, nil)
			},
			expected: Expected{
				expectedSession: &models.Device{},
				expectedStatus:  http.StatusOK,
			},
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

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session *models.Device
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}

			assert.Equal(t, tc.expected.expectedSession, session)
		})
	}
}

func TestGetDeviceList(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		session []models.Device
		status  int
	}

	cases := []struct {
		description   string
		paginator     query.Paginator
		sorter        query.Sorter
		filters       query.Filters
		status        models.DeviceStatus
		tenant        string
		requiredMocks func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter)
		expected      Expected
	}{
		{
			description: "fails when try to get a device list existing",
			tenant:      "tenant-id",
			status:      models.DeviceStatus("online"),
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			sorter:      query.Sorter{By: "name", Order: query.OrderAsc},
			filters: query.Filters{
				Raw: "Wwp7CiAgInR5cGUiOiAicHJvcGVydHkiLAogICJwYXJhbXMiOiB7CiAgICAibmFtZSI6ICJuYW1lIiwKICAgICJvcGVyYXRvciI6ICJjb250YWlucyIsCiAgICAidmFsdWUiOiAiZXhhbXBsZXNwYWNlIgogIH0KfQpd",
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "name",
							Operator: "contains",
							Value:    "examplespace",
						},
					},
				},
			},
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				mock.On("ListDevices",
					gomock.Anything,
					"tenant-id",
					status,
					paginator,
					filters,
					sorter,
				).Return(nil, 0, svc.ErrDeviceNotFound).Once()
			},
			expected: Expected{
				session: nil,
				status:  http.StatusNotFound,
			},
		},
		{
			description: "fails when try to get a device list existing",
			tenant:      "tenant-id",
			status:      models.DeviceStatus("online"),
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			sorter:      query.Sorter{By: "name", Order: query.OrderAsc},
			filters: query.Filters{
				Raw: "Wwp7CiAgInR5cGUiOiAicHJvcGVydHkiLAogICJwYXJhbXMiOiB7CiAgICAibmFtZSI6ICJuYW1lIiwKICAgICJvcGVyYXRvciI6ICJjb250YWlucyIsCiAgICAidmFsdWUiOiAiZXhhbXBsZXNwYWNlIgogIH0KfQpd",
				Data: []query.Filter{
					{
						Type: "property",
						Params: &query.FilterProperty{
							Name:     "name",
							Operator: "contains",
							Value:    "examplespace",
						},
					},
				},
			},
			requiredMocks: func(status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter) {
				mock.On("ListDevices",
					gomock.Anything,
					"tenant-id",
					status,
					paginator,
					filters,
					sorter,
				).Return([]models.Device{}, 1, nil).Once()
			},
			expected: Expected{
				session: []models.Device{},
				status:  http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks(tc.status, tc.paginator, tc.filters, tc.sorter)

			type Query struct {
				Status models.DeviceStatus `query:"status"`
				query.Paginator
				query.Sorter
				query.Filters
			}

			b := Query{
				Status:    tc.status,
				Paginator: tc.paginator,
				Sorter:    tc.sorter,
				Filters:   tc.filters,
			}

			jsonData, err := json.Marshal(b)
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

			assert.Equal(t, tc.expected.status, rec.Result().StatusCode)

			var session []models.Device
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}

			assert.Equal(t, tc.expected.session, session)
		})
	}
}

func TestLookupDevice(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedSession *models.Device
		expectedStatus  int
	}
	tests := []struct {
		title         string
		request       requests.DeviceLookup
		requiredMocks func(requests.DeviceLookup)
		expected      Expected
	}{
		{
			title: "fails when bind fails to validate uid",
			request: requests.DeviceLookup{
				Username:  "user1",
				IPAddress: "192.168.1.100",
			},
			requiredMocks: func(req requests.DeviceLookup) {},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusBadRequest,
			},
		},
		{
			title: "fails when try to look up of a existing device",
			request: requests.DeviceLookup{
				Domain:    "example.com",
				Name:      "device1",
				Username:  "user1",
				IPAddress: "192.168.1.100",
			},
			requiredMocks: func(req requests.DeviceLookup) {
				mock.On("LookupDevice", gomock.Anything, req.Domain, req.Name).Return(nil, svc.ErrDeviceNotFound).Once()
			},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title: "success when try to look up of a existing device",
			request: requests.DeviceLookup{
				Domain:    "example.com",
				Name:      "device1",
				Username:  "user1",
				IPAddress: "192.168.1.100",
			},
			requiredMocks: func(req requests.DeviceLookup) {
				mock.On("LookupDevice", gomock.Anything, req.Domain, req.Name).Return(&models.Device{}, nil)
			},
			expected: Expected{
				expectedSession: &models.Device{},
				expectedStatus:  http.StatusOK,
			},
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

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session *models.Device
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}

			assert.Equal(t, tc.expected.expectedSession, session)
		})
	}
}

func TestRemoveDeviceTag(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		updatePayload  requests.DeviceRemoveTag
		requiredMocks  func(req requests.DeviceRemoveTag)
		expectedStatus int
	}{
		{
			title: "fails when bind fails to validate uid",
			updatePayload: requests.DeviceRemoveTag{
				DeviceParam: requests.DeviceParam{UID: ""},
				TagBody:     requests.TagBody{Tag: "tag"},
			},
			requiredMocks:  func(req requests.DeviceRemoveTag) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because the tag does not have a min of 3 characters",
			updatePayload: requests.DeviceRemoveTag{
				TagBody: requests.TagBody{Tag: "tg"},
			},
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func(req requests.DeviceRemoveTag) {},
		},
		{
			title: "fails when validate because the tag does not have a max of 255 characters",
			updatePayload: requests.DeviceRemoveTag{
				TagBody: requests.TagBody{Tag: "BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9"},
			},
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func(req requests.DeviceRemoveTag) {},
		},
		{
			title: "fails when validate because have a '/' with in your characters",
			updatePayload: requests.DeviceRemoveTag{
				TagBody: requests.TagBody{Tag: "test/"},
			},
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func(req requests.DeviceRemoveTag) {},
		},
		{
			title: "fails when validate because have a '&' with in your characters",
			updatePayload: requests.DeviceRemoveTag{
				TagBody: requests.TagBody{Tag: "test&"},
			},
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func(req requests.DeviceRemoveTag) {},
		},
		{
			title: "fails when validate because have a '@' with in your characters",
			updatePayload: requests.DeviceRemoveTag{
				TagBody: requests.TagBody{Tag: "test@"},
			},
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func(req requests.DeviceRemoveTag) {},
		},
		{
			title: "fails when try to remove a non-existing device tag",
			updatePayload: requests.DeviceRemoveTag{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				TagBody:     requests.TagBody{Tag: "tag"},
			},
			requiredMocks: func(req requests.DeviceRemoveTag) {
				mock.On("RemoveDeviceTag", gomock.Anything, models.UID("1234"), req.Tag).Return(svc.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to remove a existing device tag",
			updatePayload: requests.DeviceRemoveTag{
				DeviceParam: requests.DeviceParam{UID: "123"},
				TagBody:     requests.TagBody{Tag: "tag"},
			},

			requiredMocks: func(req requests.DeviceRemoveTag) {
				mock.On("RemoveDeviceTag", gomock.Anything, models.UID("123"), req.Tag).Return(nil)
			},
			expectedStatus: http.StatusOK,
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
			title: "fails when bind fails to validate uid",
			updatePayload: requests.DeviceCreateTag{
				DeviceParam: requests.DeviceParam{UID: ""},
				TagBody:     requests.TagBody{Tag: "tag"},
			},
			requiredMocks: func(req requests.DeviceCreateTag) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because the tag does not have a min of 3 characters",
			updatePayload: requests.DeviceCreateTag{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				TagBody:     requests.TagBody{Tag: "tg"},
			},
			requiredMocks: func(req requests.DeviceCreateTag) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because the tag does not have a max of 255 characters",
			updatePayload: requests.DeviceCreateTag{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				TagBody:     requests.TagBody{Tag: "BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9"},
			},
			requiredMocks: func(req requests.DeviceCreateTag) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because have a '@' with in your characters",
			updatePayload: requests.DeviceCreateTag{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				TagBody:     requests.TagBody{Tag: "test@"},
			},
			requiredMocks: func(req requests.DeviceCreateTag) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because have a '/' with in your characters",
			updatePayload: requests.DeviceCreateTag{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				TagBody:     requests.TagBody{Tag: "test/"},
			},
			requiredMocks: func(req requests.DeviceCreateTag) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because have a '&' with in your characters",
			updatePayload: requests.DeviceCreateTag{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				TagBody:     requests.TagBody{Tag: "test&"},
			},
			requiredMocks: func(req requests.DeviceCreateTag) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when try to create a non-existing device tag",
			updatePayload: requests.DeviceCreateTag{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				TagBody:     requests.TagBody{Tag: "tag"},
			},
			requiredMocks: func(req requests.DeviceCreateTag) {
				mock.On("CreateDeviceTag", gomock.Anything, models.UID("1234"), req.Tag).Return(svc.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "fails when try to create a existing device tag",
			updatePayload: requests.DeviceCreateTag{
				DeviceParam: requests.DeviceParam{UID: "123"},
				TagBody:     requests.TagBody{Tag: "tag"},
			},

			requiredMocks: func(req requests.DeviceCreateTag) {
				mock.On("CreateDeviceTag", gomock.Anything, models.UID("123"), req.Tag).Return(nil)
			},
			expectedStatus: http.StatusOK,
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
			title: "fails when bind fails to validate uid",
			updatePayload: requests.DeviceUpdateTag{
				DeviceParam: requests.DeviceParam{UID: ""},
				Tags:        []string{"tag1", "tag2"},
			},
			requiredMocks:  func(req requests.DeviceUpdateTag) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because have a duplicate tag",
			updatePayload: requests.DeviceUpdateTag{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				Tags:        []string{"tagduplicated", "tagduplicated"},
			},
			requiredMocks:  func(req requests.DeviceUpdateTag) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because have a '@' with in your characters",
			updatePayload: requests.DeviceUpdateTag{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				Tags:        []string{"test@"},
			},
			requiredMocks:  func(req requests.DeviceUpdateTag) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because have a '/' with in your characters",
			updatePayload: requests.DeviceUpdateTag{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				Tags:        []string{"test/"},
			},
			requiredMocks:  func(req requests.DeviceUpdateTag) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because have a '&' with in your characters",
			updatePayload: requests.DeviceUpdateTag{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				Tags:        []string{"test&"},
			},
			requiredMocks:  func(req requests.DeviceUpdateTag) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because the tag does not have a min of 3 characters",
			updatePayload: requests.DeviceUpdateTag{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				Tags:        []string{"tg"},
			},
			requiredMocks:  func(req requests.DeviceUpdateTag) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when validate because the tag does not have a max of 255 characters",
			updatePayload: requests.DeviceUpdateTag{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				Tags:        []string{"BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9"},
			},
			requiredMocks:  func(req requests.DeviceUpdateTag) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when try to update a existing device tag",
			updatePayload: requests.DeviceUpdateTag{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				Tags:        []string{"tag1", "tag2"},
			},
			requiredMocks: func(req requests.DeviceUpdateTag) {
				mock.On("UpdateDeviceTag", gomock.Anything, models.UID("1234"), req.Tags).Return(svc.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to update a existing device tag",
			updatePayload: requests.DeviceUpdateTag{
				DeviceParam: requests.DeviceParam{UID: "123"},
				Tags:        []string{"tag1", "tag2"},
			},

			requiredMocks: func(req requests.DeviceUpdateTag) {
				mock.On("UpdateDeviceTag", gomock.Anything, models.UID("123"), req.Tags).Return(nil)
			},
			expectedStatus: http.StatusOK,
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
			title: "fails when try to uodate a existing device",
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
		{
			title: "success when try to update a existing device",
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

func TestUpdateDeviceConnectionStats(t *testing.T) {
	serviceMock := new(mocks.Service)
	clockMock := new(clockmock.Clock)

	now := time.Now()
	clock.DefaultBackend = clockMock
	clockMock.On("Now").Return(now)

	type Actual struct {
		err    error
		status int
	}

	cases := []struct {
		description string
		method      string
		request     *requests.DeviceUpdateConnectionStats
		mocks       func()
		expected    Actual
	}{
		{
			description: "fails when method is other than PATCH",
			method:      http.MethodGet,
			request:     &requests.DeviceUpdateConnectionStats{},
			mocks:       func() {},
			expected: Actual{
				status: http.StatusMethodNotAllowed,
			},
		},
		{
			description: "fails when namespace does not exists",
			method:      http.MethodPatch,
			request: &requests.DeviceUpdateConnectionStats{
				UID:            "0000000000000000000000000000000000000000000000000000000000000000",
				TenantID:       "00000000-0000-4000-0000-000000000000",
				ConnectedAt:    clock.Now(),
				DisconnectedAt: clock.Now(),
			},
			mocks: func() {
				serviceMock.
					On("UpdateDeviceConnectionStats", gomock.Anything, gomock.MatchedBy(func(req *requests.DeviceUpdateConnectionStats) bool {
						return req.TenantID == "00000000-0000-4000-0000-000000000000" &&
							req.UID == "0000000000000000000000000000000000000000000000000000000000000000" &&
							req.ConnectedAt.Sub(now) == 0 &&
							req.DisconnectedAt.Sub(now) == 0

					})).
					Return(svc.NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", nil)).
					Once()
			},
			expected: Actual{
				status: http.StatusNotFound,
			},
		},
		{
			description: "succeeds",
			method:      http.MethodPatch,
			request: &requests.DeviceUpdateConnectionStats{
				UID:            "0000000000000000000000000000000000000000000000000000000000000000",
				TenantID:       "00000000-0000-4000-0000-000000000000",
				ConnectedAt:    clock.Now(),
				DisconnectedAt: clock.Now(),
			},
			mocks: func() {
				serviceMock.
					On("UpdateDeviceConnectionStats", gomock.Anything, gomock.MatchedBy(func(req *requests.DeviceUpdateConnectionStats) bool {
						return req.TenantID == "00000000-0000-4000-0000-000000000000" &&
							req.UID == "0000000000000000000000000000000000000000000000000000000000000000" &&
							req.ConnectedAt.Sub(now) == 0 &&
							req.DisconnectedAt.Sub(now) == 0

					})).
					Return(nil).
					Once()
			},
			expected: Actual{
				status: http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.mocks()

			jsonData, err := json.Marshal(map[string]interface{}{
				"connected_at":    tc.request.ConnectedAt,
				"disconnected_at": tc.request.DisconnectedAt,
			})
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(tc.method, fmt.Sprintf("/internal/devices/%s/connection-stats", tc.request.UID), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Tenant-ID", tc.request.TenantID)
			rec := httptest.NewRecorder()

			e := NewRouter(serviceMock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.status, rec.Result().StatusCode)
		})
	}

	serviceMock.AssertExpectations(t)
}
