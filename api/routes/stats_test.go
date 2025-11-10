package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/responses"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestGetSystemInfo(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		request        requests.GetSystemInfo
		requiredMocks  func(updatePayloadMock requests.GetSystemInfo)
		expectedStatus int
	}{
		{
			title: "success when try to get infos of a existing system",
			request: requests.GetSystemInfo{
				Host: "example.com",
				Port: 0,
			},
			requiredMocks: func(_ requests.GetSystemInfo) {
				mock.
					On(
						"GetSystemInfo",
						gomock.Anything,
						&requests.GetSystemInfo{Host: "example.com", Port: 0},
					).
					Return(&responses.SystemInfo{}, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.request)

			jsonData, err := json.Marshal(tc.request)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/info", strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}

	mock.AssertExpectations(t)
}

func TestGetStats(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		reqStats       *models.Stats
		headers        map[string]string
		expectedStatus int
		requiredMocks  func()
	}{
		{
			title: "success when try to get stats without tenantID",
			reqStats: &models.Stats{
				RegisteredDevices: 10,
				OnlineDevices:     5,
				ActiveSessions:    20,
				PendingDevices:    3,
				RejectedDevices:   2,
			},
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       authorizer.RoleOwner.String(),
			},
			requiredMocks: func() {
				mock.On("GetStats", gomock.Anything, &requests.GetStats{TenantID: ""}).Return(&models.Stats{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			title: "success when try to get stats with tenantID",
			reqStats: &models.Stats{
				RegisteredDevices: 5,
				OnlineDevices:     2,
				ActiveSessions:    10,
				PendingDevices:    1,
				RejectedDevices:   0,
			},
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       authorizer.RoleOwner.String(),
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				mock.On("GetStats", gomock.Anything, &requests.GetStats{TenantID: "00000000-0000-4000-0000-000000000000"}).Return(&models.Stats{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)

			for key, value := range tc.headers {
				req.Header.Set(key, value)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}

	mock.AssertExpectations(t)
}
