package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const cursorTestTenant = "00000000-0000-4000-0000-000000000000"

func newDeviceListRequest(t *testing.T, target string) *http.Request {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	req.Header.Set("X-Role", authorizer.RoleOwner.String())
	req.Header.Set("X-Tenant-ID", cursorTestTenant)

	return req
}

func cursorRequestWithSize(size int64) any {
	return gomock.MatchedBy(func(req *requests.DeviceList) bool { return req.Size == size })
}

func TestGetDeviceListCursorModeReachesTheService(t *testing.T) {
	after := query.Cursor{CreatedAt: time.Date(2026, 1, 10, 12, 0, 0, 123456000, time.UTC), ID: "abc"}

	cases := []struct {
		description string
		target      string
		after       *query.Cursor
	}{
		{
			description: "page[size] alone starts a read",
			target:      "/api/devices?page[size]=50",
		},
		{
			description: "page[after] continues a read",
			target:      "/api/devices?page[size]=50&page[after]=" + after.String(),
			after:       &after,
		},
		{
			description: "percent-encoded brackets select cursor mode too",
			target:      "/api/devices?page%5Bsize%5D=50&page%5Bafter%5D=" + after.String(),
			after:       &after,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			service := mocks.NewMockService(t)
			service.
				On("ListDevicesAfter", gomock.Anything, gomock.Anything, cursorRequestWithSize(50), gomock.MatchedBy(func(got *query.Cursor) bool {
					if tc.after == nil {
						return got == nil
					}

					return got != nil && got.ID == tc.after.ID && got.CreatedAt.Equal(tc.after.CreatedAt)
				})).
				Return([]models.Device{{UID: "d1"}}, nil, nil).
				Once()

			rec := httptest.NewRecorder()
			newTestRouter(t, service).ServeHTTP(rec, newDeviceListRequest(t, tc.target))

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

			var devices []models.Device
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &devices))
			require.Len(t, devices, 1)
			assert.Equal(t, "d1", devices[0].UID)
		})
	}
}

func TestGetDeviceListCursorModeRejectsConflicts(t *testing.T) {
	cursor := query.Cursor{CreatedAt: time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC), ID: "abc"}.String()

	cases := []struct {
		description string
		target      string
		field       string
	}{
		{description: "page with page[size]", target: "/api/devices?page[size]=10&page=2", field: "page"},
		{description: "per_page with page[size]", target: "/api/devices?page[size]=10&per_page=10", field: "per_page"},
		{description: "sort_by with page[size]", target: "/api/devices?page[size]=10&sort_by=name", field: "sort_by"},
		{description: "order_by with page[after]", target: "/api/devices?page[after]=" + cursor + "&order_by=asc", field: "order_by"},
		{description: "a page[after] that is not a cursor", target: "/api/devices?page[size]=10&page[after]=not-a-cursor", field: "page[after]"},
		{description: "an empty page[after]", target: "/api/devices?page[after]=", field: "page[after]"},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			service := mocks.NewMockService(t)

			rec := httptest.NewRecorder()
			newTestRouter(t, service).ServeHTTP(rec, newDeviceListRequest(t, tc.target))

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var body struct {
				Fields map[string]string `json:"fields"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Contains(t, body.Fields, tc.field)
			service.AssertNotCalled(t, "ListDevicesAfter")
			service.AssertNotCalled(t, "ListDevices")
		})
	}
}

func TestGetDeviceListCursorModeClampsThePageSize(t *testing.T) {
	cursor := query.Cursor{CreatedAt: time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC), ID: "abc"}.String()

	cases := []struct {
		description string
		target      string
		opts        []Option
		expected    int64
	}{
		{description: "a missing size takes the default", target: "/api/devices?page[after]=" + cursor, expected: query.DefaultPageSize},
		{description: "a zero size takes the default", target: "/api/devices?page[size]=0", expected: query.DefaultPageSize},
		{description: "a negative size becomes one", target: "/api/devices?page[size]=-3", expected: 1},
		{description: "a size above the maximum becomes the maximum", target: "/api/devices?page[size]=5000", expected: query.DefaultMaxPageSize},
		{
			description: "the operator's maximum replaces the default one",
			target:      "/api/devices?page[size]=5000",
			opts:        []Option{WithDeviceListMaxPageSize(50)},
			expected:    50,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			service := mocks.NewMockService(t)
			service.
				On("ListDevicesAfter", gomock.Anything, gomock.Anything, cursorRequestWithSize(tc.expected), gomock.Anything).
				Return([]models.Device{}, nil, nil).
				Once()

			rec := httptest.NewRecorder()
			newTestRouter(t, service, tc.opts...).ServeHTTP(rec, newDeviceListRequest(t, tc.target))

			assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		})
	}
}

func TestGetDeviceListCursorModeHeaders(t *testing.T) {
	filter := encodeFilter(t, []query.Filter{{
		Type:   query.FilterTypeProperty,
		Params: &query.FilterProperty{Name: "name", Operator: "contains", Value: "web"},
	}})

	t.Run("hands back the next page in Link and X-Next-Cursor", func(t *testing.T) {
		next := query.Cursor{CreatedAt: time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC), ID: "d2"}

		service := mocks.NewMockService(t)
		service.
			On("ListDevicesAfter", gomock.Anything, gomock.Anything, gomock.Anything, gomock.Anything).
			Return([]models.Device{{UID: "d1"}, {UID: "d2"}}, &next, nil).
			Once()

		values := url.Values{}
		values.Set("page[size]", "5000")
		values.Set("status", "accepted")
		values.Set("connector", "true")
		values.Set("filter", filter)

		rec := httptest.NewRecorder()
		newTestRouter(t, service).ServeHTTP(rec, newDeviceListRequest(t, "/api/devices?"+values.Encode()))

		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, next.String(), rec.Header().Get("X-Next-Cursor"))
		assert.Empty(t, rec.Header().Get("X-Total-Count"))

		link := rec.Header().Get("Link")
		require.True(t, strings.HasPrefix(link, "<") && strings.HasSuffix(link, `>; rel="next"`), link)

		target := strings.TrimSuffix(strings.TrimPrefix(link, "<"), `>; rel="next"`)
		assert.Contains(t, target, "page%5Bsize%5D=1000")
		assert.Contains(t, target, "page%5Bafter%5D="+next.String())

		linked, err := url.Parse(target)
		require.NoError(t, err)
		assert.Equal(t, "http", linked.Scheme)
		assert.Equal(t, "example.com", linked.Host)
		assert.Equal(t, "/api/devices", linked.Path)
		assert.Equal(t, url.Values{
			"page[size]":  {"1000"},
			"page[after]": {next.String()},
			"status":      {"accepted"},
			"connector":   {"true"},
			"filter":      {filter},
		}, linked.Query())
	})

	t.Run("links to the address the client reached the gateway on", func(t *testing.T) {
		next := query.Cursor{CreatedAt: time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC), ID: "d2"}

		cases := []struct {
			description string
			host        string
			proto       string
			port        string
			expected    string
		}{
			{description: "a non-default port", host: "shellhub.example.com", proto: "http", port: "8080", expected: "http://shellhub.example.com:8080"},
			{description: "the default https port", host: "shellhub.example.com", proto: "https", port: "443", expected: "https://shellhub.example.com"},
			{description: "the default http port", host: "shellhub.example.com", proto: "http", port: "80", expected: "http://shellhub.example.com"},
			{description: "a non-default https port", host: "shellhub.example.com", proto: "https", port: "8443", expected: "https://shellhub.example.com:8443"},
			{description: "an IPv6 host", host: "[::1]", proto: "http", port: "8080", expected: "http://[::1]:8080"},
			{description: "no forwarded port", host: "shellhub.example.com:9000", proto: "http", port: "", expected: "http://shellhub.example.com:9000"},
			{description: "a default port on the host and no forwarded port", host: "shellhub.example.com:443", proto: "https", port: "", expected: "https://shellhub.example.com"},
			{description: "an IPv6 host on a default port", host: "[::1]:80", proto: "http", port: "", expected: "http://[::1]"},
		}

		for _, tc := range cases {
			t.Run(tc.description, func(t *testing.T) {
				service := mocks.NewMockService(t)
				service.
					On("ListDevicesAfter", gomock.Anything, gomock.Anything, gomock.Anything, gomock.Anything).
					Return([]models.Device{{UID: "d1"}}, &next, nil).
					Once()

				req := newDeviceListRequest(t, "/api/devices?page[size]=1")
				req.RemoteAddr = "172.18.0.5:40000"
				req.Host = tc.host
				req.Header.Set("X-Forwarded-Proto", tc.proto)
				if tc.port != "" {
					req.Header.Set("X-Forwarded-Port", tc.port)
				}

				rec := httptest.NewRecorder()
				newTestRouter(t, service).ServeHTTP(rec, req)

				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
				assert.True(t, strings.HasPrefix(rec.Header().Get("Link"), "<"+tc.expected+"/api/devices?"), rec.Header().Get("Link"))
			})
		}
	})

	t.Run("carries neither header on the last page", func(t *testing.T) {
		service := mocks.NewMockService(t)
		service.
			On("ListDevicesAfter", gomock.Anything, gomock.Anything, gomock.Anything, gomock.Anything).
			Return([]models.Device{{UID: "d1"}}, nil, nil).
			Once()

		rec := httptest.NewRecorder()
		newTestRouter(t, service).ServeHTTP(rec, newDeviceListRequest(t, "/api/devices?page[size]=10"))

		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Empty(t, rec.Header().Values("Link"))
		assert.Empty(t, rec.Header().Values("X-Next-Cursor"))
		assert.Empty(t, rec.Header().Values("X-Total-Count"))
	})

	t.Run("page mode still sets X-Total-Count", func(t *testing.T) {
		service := mocks.NewMockService(t)
		service.
			On("ListDevices", gomock.Anything, gomock.Anything, gomock.Anything).
			Return([]models.Device{{UID: "d1"}}, int64(42), nil).
			Once()

		rec := httptest.NewRecorder()
		newTestRouter(t, service).ServeHTTP(rec, newDeviceListRequest(t, "/api/devices?page=1&per_page=10"))

		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, "42", rec.Header().Get("X-Total-Count"))
		assert.Empty(t, rec.Header().Values("Link"))
		assert.Empty(t, rec.Header().Values("X-Next-Cursor"))
	})
}

func TestGetDeviceListCursorModeOnTheAdminList(t *testing.T) {
	cursor := query.Cursor{CreatedAt: time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC), ID: "abc"}.String()

	cases := []struct {
		description string
		query       string
		field       string
	}{
		{description: "names page[size] when it was sent", query: "page[size]=10", field: "page[size]"},
		{description: "names page[after] when only it was sent", query: "page[after]=" + cursor, field: "page[after]"},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			service := mocks.NewMockService(t)

			router := newTestRouter(t, service)
			markAdmin := func(next echo.HandlerFunc) echo.HandlerFunc {
				return func(c *echo.Context) error {
					gateway.NewContext(c).MarkAdminRoute()

					return next(c)
				}
			}
			router.GET("/admin/api/devices", gateway.Handler(NewHandler(service).GetDeviceList), markAdmin)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/admin/api/devices?"+tc.query, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var body struct {
				Fields map[string]string `json:"fields"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Contains(t, body.Fields, tc.field)
			service.AssertNotCalled(t, "ListDevicesAfter")
			service.AssertNotCalled(t, "ListDevices")
		})
	}
}
