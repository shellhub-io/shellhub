package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

const guardTenant = "00000000-0000-4000-0000-000000000000"

// TestGuardsRefuseWhatTheyRefusedBefore drives one route per guard through the built router. The
// route table now states two of these guards on the line that mounts the route instead of trailing
// them as middleware, and the order they run in moved with them — so what a caller sees is the
// thing worth pinning, not where the guard is written.
func TestGuardsRefuseWhatTheyRefusedBefore(t *testing.T) {
	cases := []struct {
		description    string
		method         string
		target         string
		headers        map[string]string
		mocks          func(*mocks.MockService)
		expectedStatus int
	}{
		{
			description: "the permission guard refuses a role that does not hold it",
			method:      http.MethodDelete,
			target:      "/api/devices/1234",
			headers: map[string]string{
				"X-ID":        "000000000000000000000000",
				"X-Tenant-ID": guardTenant,
				"X-Role":      authorizer.RoleObserver.String(),
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			description: "the permission guard admits a role that holds it",
			method:      http.MethodDelete,
			target:      "/api/devices/1234",
			headers: map[string]string{
				"X-ID":        "000000000000000000000000",
				"X-Tenant-ID": guardTenant,
				"X-Role":      authorizer.RoleOwner.String(),
			},
			mocks: func(service *mocks.MockService) {
				service.On("DeleteDevice", gomock.Anything, models.UID("1234"), guardTenant).Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			description: "the API-key block refuses a request that authenticated with a key",
			method:      http.MethodGet,
			target:      "/api/namespaces/api-key",
			headers: map[string]string{
				"X-Tenant-ID": guardTenant,
				"X-Role":      authorizer.RoleOwner.String(),
				"X-API-Key":   "a-key",
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			description: "the API-key block admits a request carrying no key",
			method:      http.MethodGet,
			target:      "/api/namespaces/api-key",
			headers: map[string]string{
				"X-ID":        "000000000000000000000000",
				"X-Tenant-ID": guardTenant,
				"X-Role":      authorizer.RoleOwner.String(),
			},
			mocks: func(service *mocks.MockService) {
				service.
					On("ListAPIKeys", gomock.Anything, gomock.AnythingOfType("*requests.ListAPIKey")).
					Return([]models.APIKey{}, 0, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			description: "the any-permission guard refuses a role holding neither of the two",
			method:      http.MethodDelete,
			target:      "/api/ssh-identities/1234",
			headers: map[string]string{
				"X-ID":        "000000000000000000000000",
				"X-Tenant-ID": guardTenant,
				"X-Role":      authorizer.RoleObserver.String(),
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			description: "the any-permission guard admits a role holding only the enrol one, and tells the service so",
			method:      http.MethodDelete,
			target:      "/api/ssh-identities/1234",
			headers: map[string]string{
				"X-ID":        "000000000000000000000000",
				"X-Tenant-ID": guardTenant,
				"X-Role":      authorizer.RoleOperator.String(),
			},
			mocks: func(service *mocks.MockService) {
				service.
					On("DeleteSSHIdentity", gomock.Anything, &requests.SSHIdentityDelete{
						SSHIdentityIDParam: requests.SSHIdentityIDParam{ID: "1234"},
						UserID:             "000000000000000000000000",
						TenantID:           guardTenant,
					}).
					Return(nil).
					Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			description: "the any-permission guard admits a role holding the manage one, and tells the service so",
			method:      http.MethodDelete,
			target:      "/api/ssh-identities/1234",
			headers: map[string]string{
				"X-ID":        "000000000000000000000000",
				"X-Tenant-ID": guardTenant,
				"X-Role":      authorizer.RoleOwner.String(),
			},
			mocks: func(service *mocks.MockService) {
				service.
					On("DeleteSSHIdentity", gomock.Anything, &requests.SSHIdentityDelete{
						SSHIdentityIDParam: requests.SSHIdentityIDParam{ID: "1234"},
						UserID:             "000000000000000000000000",
						TenantID:           guardTenant,
						Manage:             true,
					}).
					Return(nil).
					Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			description: "the authorize guard refuses an identity carrying no namespace",
			method:      http.MethodGet,
			target:      "/api/devices",
			headers: map[string]string{
				"X-ID":   "000000000000000000000000",
				"X-Role": authorizer.RoleOwner.String(),
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			description: "the tenant guard refuses a namespace the caller is not scoped to",
			method:      http.MethodGet,
			target:      "/api/namespaces/" + guardTenant,
			headers: map[string]string{
				"X-ID":        "000000000000000000000000",
				"X-Tenant-ID": "11111111-1111-4111-1111-111111111111",
				"X-Role":      authorizer.RoleOwner.String(),
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			service := mocks.NewMockService(t)
			if tc.mocks != nil {
				tc.mocks(service)
			}

			req := httptest.NewRequestWithContext(t.Context(), tc.method, tc.target, nil)
			req.Header.Set("Content-Type", "application/json")

			for name, value := range tc.headers {
				req.Header.Set(name, value)
			}

			rec := httptest.NewRecorder()
			NewRouter(service).ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code, rec.Body.String())
		})
	}
}
