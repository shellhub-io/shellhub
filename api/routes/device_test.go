package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
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

			e := NewRouter(mock, nil)
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
		description   string
		headers       map[string]string
		uid           string
		requiredMocks func()
		expected      int
	}{
		{
			description: "fails when role is observer",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       "observer",
			},
			uid: "1234",
			requiredMocks: func() {
			},
			expected: http.StatusForbidden,
		},
		{
			description: "fails when role is operator",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       "operator",
			},
			uid: "1234",
			requiredMocks: func() {
			},
			expected: http.StatusForbidden,
		},
		{
			description: "fails when bind fails to validate uid",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       "administrator",
			},
			uid:           "",
			requiredMocks: func() {},
			expected:      http.StatusNotFound,
		},
		{
			description: "fails when try to deleting a non-existing device",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       "administrator",
			},
			uid: "1234",
			requiredMocks: func() {
				mock.On("DeleteDevice", gomock.Anything, models.UID("1234"), "").Return(svc.ErrDeviceNotFound)
			},
			expected: http.StatusNotFound,
		},
		{
			description: "success when try to deleting an existing device",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       "administrator",
			},
			uid: "123",
			requiredMocks: func() {
				mock.On("DeleteDevice", gomock.Anything, models.UID("123"), "").Return(nil)
			},
			expected: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/devices/%s", tc.uid), nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(mock, nil)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected, rec.Result().StatusCode)
		})
	}
}

func TestRenameDevice(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		description   string
		headers       map[string]string
		payload       requests.DeviceRename
		requiredMocks func()
		expected      int
	}{
		{
			description: "fails when role is observer",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			payload: requests.DeviceRename{
				DeviceParam: requests.DeviceParam{UID: "000000000000000000000000"},
				Name:        "name",
			},
			requiredMocks: func() {
			},
			expected: http.StatusForbidden,
		},
		{
			description: "fails when bind fails to validate uid",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			payload: requests.DeviceRename{
				DeviceParam: requests.DeviceParam{UID: ""},
			},
			requiredMocks: func() {},
			expected:      http.StatusNotFound,
		},
		{
			description: "fails when try to rename a non-existing device",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			payload: requests.DeviceRename{
				DeviceParam: requests.DeviceParam{UID: "000000000000000000000000"},
				Name:        "name",
			},
			requiredMocks: func() {
				mock.
					On("RenameDevice", gomock.Anything, models.UID("000000000000000000000000"), "name", "00000000-0000-4000-0000-000000000000").
					Return(svc.ErrNotFound).
					Once()
			},
			expected: http.StatusNotFound,
		},
		{
			description: "success when try to rename an existing device",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			payload: requests.DeviceRename{
				DeviceParam: requests.DeviceParam{UID: "000000000000000000000000"},
				Name:        "name",
			},
			requiredMocks: func() {
				mock.
					On("RenameDevice", gomock.Anything, models.UID("000000000000000000000000"), "name", "00000000-0000-4000-0000-000000000000").
					Return(nil).
					Once()
			},
			expected: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			jsonData, err := json.Marshal(tc.payload)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/devices/%s", tc.payload.UID), strings.NewReader(string(jsonData)))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(mock, nil)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected, rec.Result().StatusCode)
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

			e := NewRouter(mock, nil)
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

			e := NewRouter(mock, nil)
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

func TestOfflineDevice(t *testing.T) {
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
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when try to setting a non-existing device as offline",
			uid:   "1234",
			requiredMocks: func() {
				mock.On("OfflineDevice", gomock.Anything, models.UID("1234")).Return(svc.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to setting an existing device as offline",
			uid:   "123",
			requiredMocks: func() {
				mock.On("OfflineDevice", gomock.Anything, models.UID("123")).Return(nil)
			},
			expectedStatus: http.StatusOK,
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

			e := NewRouter(mock, nil)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
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

			e := NewRouter(mock, nil)
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
		description   string
		uid           string
		headers       map[string]string
		body          map[string]interface{}
		requiredMocks func()
		expected      int
	}{
		{
			description: "fails when role is observer",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			body: map[string]interface{}{
				"tag": "tag",
			},
			requiredMocks: func() {
			},
			expected: http.StatusForbidden,
		},
		{
			description: "fails when validate because the tag does not have a min of 3 characters",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tag": "ta",
			},
			expected:      http.StatusBadRequest,
			requiredMocks: func() {},
		},
		{
			description: "fails when validate because the tag does not have a max of 255 characters",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tag": "BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9",
			},
			expected:      http.StatusBadRequest,
			requiredMocks: func() {},
		},
		{
			description: "fails when validate because have a '/' with in your characters",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tag": "tag/",
			},
			expected:      http.StatusBadRequest,
			requiredMocks: func() {},
		},
		{
			description: "fails when validate because have a '&' with in your characters",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tag": "tag&",
			},
			expected:      http.StatusBadRequest,
			requiredMocks: func() {},
		},
		{
			description: "fails when validate because have a '@' with in your characters",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tag": "tag@",
			},
			expected:      http.StatusBadRequest,
			requiredMocks: func() {},
		},
		{
			description: "fails when try to remove a non-existing device tag",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tag": "tag",
			},
			requiredMocks: func() {
				mock.
					On("RemoveDeviceTag", gomock.Anything, models.UID("000000000000000000000000"), "tag").
					Return(svc.ErrNotFound).
					Once()
			},
			expected: http.StatusNotFound,
		},
		{
			description: "success when try to remove a existing device tag",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tag": "tag",
			},
			requiredMocks: func() {
				mock.
					On("RemoveDeviceTag", gomock.Anything, models.UID("000000000000000000000000"), "tag").
					Return(nil).
					Once()
			},
			expected: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			jsonData, err := json.Marshal(tc.body)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/devices/%s/tags/%s", tc.uid, tc.body["tag"]), strings.NewReader(string(jsonData)))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(mock, nil)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected, rec.Result().StatusCode)
		})
	}
}

func TestCreateDeviceTag(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		description   string
		uid           string
		headers       map[string]string
		body          map[string]interface{}
		requiredMocks func()
		expected      int
	}{
		{
			description: "fails when role is observer",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			body: map[string]interface{}{
				"tag": "tag",
			},
			requiredMocks: func() {
			},
			expected: http.StatusForbidden,
		},
		{
			description: "fails when validate because the tag does not have a min of 3 characters",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tag": "ta",
			},
			requiredMocks: func() {
			},
			expected: http.StatusBadRequest,
		},
		{
			description: "fails when validate because the tag does not have a max of 255 characters",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tag": "BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9",
			},
			requiredMocks: func() {
			},
			expected: http.StatusBadRequest,
		},
		{
			description: "fails when validate because have a '@' with in your characters",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tag": "tag@",
			},
			requiredMocks: func() {
			},
			expected: http.StatusBadRequest,
		},
		{
			description: "fails when validate because have a '/' with in your characters",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tag": "tag/",
			},
			requiredMocks: func() {
			},
			expected: http.StatusBadRequest,
		},
		{
			description: "fails when validate because have a '&' with in your characters",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tag": "tag&",
			},
			requiredMocks: func() {
			},
			expected: http.StatusBadRequest,
		},
		{
			description: "fails when try to create a non-existing device tag",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tag": "tag",
			},
			requiredMocks: func() {
				mock.
					On("CreateDeviceTag", gomock.Anything, models.UID("000000000000000000000000"), "tag").
					Return(svc.ErrNotFound).
					Once()
			},
			expected: http.StatusNotFound,
		},
		{
			description: "succeeds",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tag": "tag",
			},

			requiredMocks: func() {
				mock.
					On("CreateDeviceTag", gomock.Anything, models.UID("000000000000000000000000"), "tag").
					Return(nil).
					Once()
			},
			expected: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			jsonData, err := json.Marshal(tc.body)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/devices/%s/tags", tc.uid), strings.NewReader(string(jsonData)))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(mock, nil)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected, rec.Result().StatusCode)
		})
	}
}

func TestUpdateDeviceTag(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		description   string
		uid           string
		headers       map[string]string
		body          map[string]interface{}
		requiredMocks func()
		expected      int
	}{
		{
			description: "fails when role is observer",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			body: map[string]interface{}{
				"tags": []string{"tag"},
			},
			requiredMocks: func() {},
			expected:      http.StatusForbidden,
		},
		{
			description: "fails when validate because have a duplicate tag",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tags": []string{"tagduplicated", "tagduplicated"},
			},
			requiredMocks: func() {},
			expected:      http.StatusBadRequest,
		},
		{
			description: "fails when validate because have a '@' with in your characters",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tags": []string{"test@"},
			},
			requiredMocks: func() {},
			expected:      http.StatusBadRequest,
		},
		{
			description: "fails when validate because have a '/' with in your characters",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tags": []string{"test/"},
			},
			requiredMocks: func() {},
			expected:      http.StatusBadRequest,
		},
		{
			description: "fails when validate because have a '&' with in your characters",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tags": []string{"test&"},
			},
			requiredMocks: func() {},
			expected:      http.StatusBadRequest,
		},
		{
			description: "fails when validate because the tag does not have a min of 3 characters",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tags": []string{"tg"},
			},
			requiredMocks: func() {},
			expected:      http.StatusBadRequest,
		},
		{
			description: "fails when validate because the tag does not have a max of 255 characters",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tags": []string{"BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9"},
			},
			requiredMocks: func() {},
			expected:      http.StatusBadRequest,
		},
		{
			description: "fails when try to update a non existent device",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tags": []string{"tag1", "tag2"},
			},
			requiredMocks: func() {
				mock.
					On("UpdateDeviceTag", gomock.Anything, models.UID("000000000000000000000000"), []string{"tag1", "tag2"}).
					Return(svc.ErrNotFound).
					Once()
			},
			expected: http.StatusNotFound,
		},
		{
			description: "success when try to update a existing device tag",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"tags": []string{"tag1", "tag2"},
			},
			requiredMocks: func() {
				mock.
					On("UpdateDeviceTag", gomock.Anything, models.UID("000000000000000000000000"), []string{"tag1", "tag2"}).
					Return(nil).
					Once()
			},
			expected: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			jsonData, err := json.Marshal(tc.body)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/devices/%s/tags", tc.uid), strings.NewReader(string(jsonData)))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(mock, nil)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected, rec.Result().StatusCode)
		})
	}
}

func TestUpdateDevice(t *testing.T) {
	mock := new(mocks.Service)

	name := "name"
	publicURL := true

	cases := []struct {
		description   string
		uid           string
		headers       map[string]string
		body          map[string]interface{}
		requiredMocks func()
		expected      int
	}{
		{
			description: "fails when role is observer",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			body: map[string]interface{}{
				"name":       "name",
				"public_url": true,
			},
			requiredMocks: func() {
			},
			expected: http.StatusForbidden,
		},
		{
			description: "fails when try to update a non existent device",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"name":       "name",
				"public_url": true,
			},
			requiredMocks: func() {
				mock.
					On("UpdateDevice", gomock.Anything, "00000000-0000-4000-0000-000000000000", models.UID("000000000000000000000000"), &name, &publicURL).
					Return(svc.ErrNotFound).
					Once()
			},
			expected: http.StatusNotFound,
		},
		{
			description: "success when try to update a existing device",
			uid:         "000000000000000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"name":       "name",
				"public_url": true,
			},
			requiredMocks: func() {
				mock.
					On("UpdateDevice", gomock.Anything, "00000000-0000-4000-0000-000000000000", models.UID("000000000000000000000000"), &name, &publicURL).
					Return(nil).
					Once()
			},
			expected: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			jsonData, err := json.Marshal(tc.body)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/devices/%s", tc.uid), strings.NewReader(string(jsonData)))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(mock, nil)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected, rec.Result().StatusCode)
		})
	}
}
