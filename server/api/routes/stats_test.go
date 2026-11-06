package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/shellhub-io/shellhub/pkg/api/identity/identitytest"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/origin"
	"github.com/shellhub-io/shellhub/server/api/pkg/responses"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestGetSystemInfo(t *testing.T) {
	cases := []struct {
		title    string
		peer     string
		expected origin.Origin
	}{
		{
			title:    "describes the address the client reached, not the one its headers name",
			peer:     "203.0.113.9:54321",
			expected: origin.Origin{IP: "203.0.113.9", Host: "shellhub.example.com", Scheme: "http"},
		},
		{
			title:    "describes the address a trusted proxy names",
			peer:     "127.0.0.1:54321",
			expected: origin.Origin{IP: "127.0.0.1", Host: "elsewhere.example.com", Scheme: "https", Port: "8443"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			mock := mocks.NewMockService(t)
			mock.On("GetSystemInfo", gomock.Anything, tc.expected).Return(&responses.SystemInfo{}, nil).Once()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/info", nil)
			req.RemoteAddr = tc.peer
			req.Host = "shellhub.example.com"
			req.Header.Set("X-Forwarded-Host", "elsewhere.example.com")
			req.Header.Set("X-Forwarded-Proto", "https")
			req.Header.Set("X-Forwarded-Port", "8443")
			rec := httptest.NewRecorder()

			newTestRouter(t, mock).ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
		})
	}
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
