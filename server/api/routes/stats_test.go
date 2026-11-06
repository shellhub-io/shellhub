package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/shellhub-io/shellhub/pkg/api/identity/identitytest"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/responses"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetSystemInfo(t *testing.T) {
	mock := mocks.NewMockService(t)

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
				require.NoError(t, err)
			}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/info", strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req = identitytest.As(req, identity.Identity{Role: authorizer.RoleOwner})
			rec := httptest.NewRecorder()

			e := newTestRouter(t, mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}

	mock.AssertExpectations(t)
}

func TestGetStats(t *testing.T) {
	const tenant = "00000000-0000-4000-0000-000000000000"

	cases := []struct {
		title          string
		caller         identity.Identity
		requiredMocks  func(mock *mocks.MockService)
		expectedStatus int
	}{
		{
			title:  "serves the stats of the caller's namespace",
			caller: identity.Identity{ID: "user", Role: authorizer.RoleOwner, TenantID: tenant},
			requiredMocks: func(mock *mocks.MockService) {
				mock.On("GetStats", gomock.Anything, scope.MustBounded(tenant)).Return(&models.Stats{RegisteredDevices: 3}, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			title:          "refuses an instance administrator with no namespace",
			caller:         identity.Identity{ID: "admin", Admin: true},
			requiredMocks:  func(*mocks.MockService) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			title:          "refuses an identity with no namespace",
			caller:         identity.Identity{Role: authorizer.RoleOwner},
			requiredMocks:  func(*mocks.MockService) {},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			mock := mocks.NewMockService(t)
			tc.requiredMocks(mock)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/stats", nil)
			req = identitytest.As(req, tc.caller)

			rec := httptest.NewRecorder()
			newTestRouter(t, mock).ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}
