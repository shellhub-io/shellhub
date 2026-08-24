package routes

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	svc "github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetDevice(t *testing.T) {
	mock := mocks.NewMockService(t)

	type Expected struct {
		expectedSession *models.Device
		expectedStatus  int
	}
	cases := []struct {
		title         string
		uid           string
		tenant        string
		admin         bool
		requiredMocks func()
		expected      Expected
	}{
		{
			title:         "fails when bind fails to validate uid",
			uid:           "",
			tenant:        "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title:         "refuses the request when the caller carries no tenant",
			uid:           "1234",
			tenant:        "",
			requiredMocks: func() {},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusForbidden,
			},
		},
		{
			title:  "admin user on the regular path stays bounded to their namespace",
			uid:    "123",
			tenant: "00000000-0000-4000-0000-000000000000",
			admin:  true,
			requiredMocks: func() {
				mock.On("GetDevice", gomock.Anything, scope.MustBounded("00000000-0000-4000-0000-000000000000"), models.UID("123")).Return(&models.Device{}, nil)
			},
			expected: Expected{
				expectedSession: &models.Device{},
				expectedStatus:  http.StatusOK,
			},
		},
		{
			title:  "fails when try to get a non-existing device",
			uid:    "1234",
			tenant: "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				mock.On("GetDevice", gomock.Anything, scope.MustBounded("00000000-0000-4000-0000-000000000000"), models.UID("1234")).Return(nil, svc.ErrDeviceNotFound)
			},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title:  "success when try to get a existing device",
			uid:    "123",
			tenant: "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				mock.On("GetDevice", gomock.Anything, scope.MustBounded("00000000-0000-4000-0000-000000000000"), models.UID("123")).Return(&models.Device{}, nil)
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

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/devices/"+tc.uid, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			if tc.tenant != "" {
				req.Header.Set("X-Tenant-ID", tc.tenant)
			}
			if tc.admin {
				req.Header.Set("X-Admin", "true")
			}
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session *models.Device
			if rec.Result().StatusCode < http.StatusBadRequest {
				if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
					assert.ErrorIs(t, io.EOF, err)
				}
			}

			assert.Equal(t, tc.expected.expectedSession, session)
		})
	}
}

func TestResolveDevice(t *testing.T) {
	mock := mocks.NewMockService(t)

	type Expected struct {
		device *models.Device
		status int
	}

	cases := []struct {
		description   string
		hostname      string
		uid           string
		headers       map[string]string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "succeeds when resolver is uid",
			hostname:    "",
			uid:         "uid",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       authorizer.RoleOwner.String(),
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				mock.
					On("ResolveDevice", gomock.Anything, &requests.ResolveDevice{TenantID: "00000000-0000-4000-0000-000000000000", UID: "uid"}).
					Return(&models.Device{}, nil).
					Once()
			},
			expected: Expected{
				device: &models.Device{},
				status: http.StatusOK,
			},
		},
		{
			description: "succeeds when resolver is hostname",
			hostname:    "hostname",
			uid:         "",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       authorizer.RoleOwner.String(),
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				mock.
					On("ResolveDevice", gomock.Anything, &requests.ResolveDevice{TenantID: "00000000-0000-4000-0000-000000000000", Hostname: "hostname"}).
					Return(&models.Device{}, nil).
					Once()
			},
			expected: Expected{
				device: &models.Device{},
				status: http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, fmt.Sprintf("/api/devices/resolve?hostname=%s&uid=%s", tc.hostname, tc.uid), nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.status, rec.Result().StatusCode)

			var session *models.Device
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}

			assert.Equal(t, tc.expected.device, session)
		})
	}
}

func TestDeleteDevice(t *testing.T) {
	mock := mocks.NewMockService(t)

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

			req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/api/devices/"+tc.uid, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestRenameDevice(t *testing.T) {
	mock := mocks.NewMockService(t)

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
			requiredMocks:  func(_ requests.DeviceRename) {},
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
				require.NoError(t, err)
			}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPatch, "/api/devices/"+tc.renamePayload.UID, strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			req.Header.Set("X-Tenant-ID", tc.tenant)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestGetDeviceList(t *testing.T) {
	mock := mocks.NewMockService(t)

	type Expected struct {
		devices []models.Device
		status  int
	}

	cases := []struct {
		description   string
		req           *requests.DeviceList
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when try to get a device list existing",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatus("online"),
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "name", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func() {
				mock.
					On("ListDevices", gomock.Anything, gomock.Anything, gomock.AnythingOfType("*requests.DeviceList")).
					Return(nil, 0, svc.ErrDeviceNotFound).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				status:  http.StatusNotFound,
			},
		},
		{
			description: "fails when try to get a device list existing",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatus("online"),
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "name", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func() {
				mock.
					On("ListDevices", gomock.Anything, gomock.Anything, gomock.AnythingOfType("*requests.DeviceList")).
					Return([]models.Device{}, 0, nil).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				status:  http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			urlVal := &url.Values{}
			urlVal.Set("page", strconv.Itoa(tc.req.Page))
			urlVal.Set("per_page", strconv.Itoa(tc.req.PerPage))
			urlVal.Set("sort_by", tc.req.By)
			urlVal.Set("order_by", tc.req.Order)
			urlVal.Set("status", string(tc.req.DeviceStatus))

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/devices?"+urlVal.Encode(), nil)
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			req.Header.Set("X-Tenant-ID", tc.req.TenantID)

			rec := httptest.NewRecorder()
			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			devices := make([]models.Device, 0)
			if len(tc.expected.devices) != 0 {
				if err := json.NewDecoder(rec.Result().Body).Decode(&devices); err != nil {
					require.ErrorIs(t, io.EOF, err)
				}
			}

			require.Equal(t, tc.expected.status, rec.Result().StatusCode)
			require.Equal(t, tc.expected.devices, devices)
		})
	}
}

func TestGetDeviceListBadFilter(t *testing.T) {
	cases := []struct {
		description string
		filter      string
	}{
		{
			description: "returns 400 when filter is not valid base64",
			filter:      "!!!not-base64!!!",
		},
		{
			description: "returns 400 when filter decodes to invalid JSON",
			filter:      "bm90LWpzb24=", // base64("not-json")
		},
		{
			description: "returns 400 when filter contains an unknown field",
			filter: func() string {
				filters := []query.Filter{
					{
						Type: query.FilterTypeProperty,
						Params: &query.FilterProperty{
							Name:     "nonexistent_field",
							Operator: "eq",
							Value:    "foo",
						},
					},
				}
				b, err := json.Marshal(filters)
				require.NoError(t, err)

				return base64.StdEncoding.EncodeToString(b)
			}(),
		},
		{
			description: "returns 400 when filter uses a disallowed operator for the field",
			filter: func() string {
				filters := []query.Filter{
					{
						Type: query.FilterTypeProperty,
						Params: &query.FilterProperty{
							Name:     "status",
							Operator: "contains", // disallowed: status only allows eq/ne
							Value:    "accepted",
						},
					},
				}
				b, err := json.Marshal(filters)
				require.NoError(t, err)

				return base64.StdEncoding.EncodeToString(b)
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			mock := mocks.NewMockService(t)

			urlVal := url.Values{}
			urlVal.Set("page", "1")
			urlVal.Set("per_page", "10")
			urlVal.Set("sort_by", "name")
			urlVal.Set("order_by", "asc")
			urlVal.Set("filter", tc.filter)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/devices?"+urlVal.Encode(), nil)
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			req.Header.Set("X-Tenant-ID", "00000000-0000-4000-0000-000000000000")

			rec := httptest.NewRecorder()
			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
			mock.AssertNotCalled(t, "ListDevices")
		})
	}
}

// TestContainerAliasCarriesTheConnectorIntent pins the two behaviours the /api/containers rewrite
// depends on: the container list is the device list carrying the connector intent, and a single
// container is the plain device route carrying none.
//
// Which comparison that intent becomes is the service's decision, and is asserted there.
func TestContainerAliasCarriesTheConnectorIntent(t *testing.T) {
	const tenantID = "00000000-0000-4000-0000-000000000000"

	get := func(t *testing.T, mock *mocks.MockService, target string) int {
		t.Helper()

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
		req.Header.Set("X-Role", authorizer.RoleOwner.String())
		req.Header.Set("X-ID", "000000000000000000000000")
		req.Header.Set("X-Tenant-ID", tenantID)

		rec := httptest.NewRecorder()
		NewRouter(mock).ServeHTTP(rec, req)

		return rec.Result().StatusCode
	}

	listExpecting := func(t *testing.T, connector bool) *mocks.MockService {
		t.Helper()

		mock := mocks.NewMockService(t)
		mock.
			On("ListDevices", gomock.Anything, scope.MustBounded(tenantID), gomock.MatchedBy(func(req *requests.DeviceList) bool {
				return req.Connector == connector
			})).
			Return([]models.Device{}, 0, nil).
			Once()

		return mock
	}

	t.Run("the container list asks for connector devices", func(t *testing.T) {
		mock := listExpecting(t, true)
		require.Equal(t, http.StatusOK, get(t, mock, "/api/containers"))
	})

	t.Run("the container list keeps the intent alongside a query string", func(t *testing.T) {
		mock := listExpecting(t, true)
		require.Equal(t, http.StatusOK, get(t, mock, "/api/containers?status=accepted"))
	})

	t.Run("the container list keeps its intent against a connector the caller sent", func(t *testing.T) {
		mock := listExpecting(t, true)
		require.Equal(t, http.StatusOK, get(t, mock, "/api/containers?connector=false"))
	})

	t.Run("the device list asks for none", func(t *testing.T) {
		mock := listExpecting(t, false)
		require.Equal(t, http.StatusOK, get(t, mock, "/api/devices"))
	})

	t.Run("a single container resolves to the plain device route", func(t *testing.T) {
		mock := mocks.NewMockService(t)
		mock.
			On("GetDevice", gomock.Anything, scope.MustBounded(tenantID), models.UID("uid1")).
			Return(&models.Device{UID: "uid1"}, nil).
			Once()

		require.Equal(t, http.StatusOK, get(t, mock, "/api/containers/uid1"))
	})
}

func TestUpdateDevice(t *testing.T) {
	mock := mocks.NewMockService(t)

	cases := []struct {
		description    string
		req            requests.DeviceUpdate
		requiredMocks  func()
		expectedStatus int
	}{
		{
			description: "fails when try to update a existing device",
			req: requests.DeviceUpdate{
				TenantID: "00000000-0000-4000-0000-000000000000",
				UID:      "1234",
				Name:     "name",
			},
			requiredMocks: func() {
				mock.On("UpdateDevice", gomock.Anything, &requests.DeviceUpdate{TenantID: "00000000-0000-4000-0000-000000000000", UID: "1234", Name: "name"}).
					Return(svc.ErrNotFound).
					Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			description: "success when try to update a existing device",
			req: requests.DeviceUpdate{
				TenantID: "00000000-0000-4000-0000-000000000000",
				UID:      "1234",
				Name:     "name",
			},
			requiredMocks: func() {
				mock.On("UpdateDevice", gomock.Anything, &requests.DeviceUpdate{TenantID: "00000000-0000-4000-0000-000000000000", UID: "1234", Name: "name"}).
					Return(nil).
					Once()
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			jsonData, err := json.Marshal(tc.req)
			if err != nil {
				require.NoError(t, err)
			}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/devices/"+tc.req.UID, strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			req.Header.Set("X-Tenant-ID", "00000000-0000-4000-0000-000000000000")
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}
