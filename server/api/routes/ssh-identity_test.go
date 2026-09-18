package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	servicemock "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListSSHIdentitiesScopesByPermission(t *testing.T) {
	const (
		userID   = "11111111-1111-4111-1111-111111111111"
		tenantID = "00000000-0000-4000-0000-000000000000"
	)

	cases := []struct {
		description string
		role        string
		url         string
		expectedAll bool
	}{
		{"a member without the manage permission sees their own", "operator", "/api/ssh-identities", false},
		{"a caller with the manage permission sees every member's", "owner", "/api/ssh-identities", true},
		{"all=true does not widen a caller who may not see every member's", "operator", "/api/ssh-identities?all=true", false},
		{"all=false does not narrow a caller who may", "owner", "/api/ssh-identities?all=false", true},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := servicemock.NewMockService(t)
			svcMock.On("ListSSHIdentities", mock.Anything, mock.MatchedBy(func(req *requests.SSHIdentityList) bool {
				return req.AllPrincipals == tc.expectedAll && req.UserID == userID && req.TenantID == tenantID
			})).Return([]models.SSHIdentity{}, nil).Once()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.url, nil)
			req.Header.Set("X-ID", userID)
			req.Header.Set("X-Tenant-ID", tenantID)
			req.Header.Set("X-Role", tc.role)

			rec := httptest.NewRecorder()
			NewRouter(svcMock).ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Result().StatusCode)
		})
	}
}

// TestCreateAPIKeySSHIdentityNeedsTheManagePermission locks the gate the route was given.
// SSHIdentityAdd means "my own key" and every role holds it, so enrolling for an API key had to
// ask for something else: an operator's key must not be able to open an SSH door for itself.
func TestCreateAPIKeySSHIdentityNeedsTheManagePermission(t *testing.T) {
	const (
		userID   = "11111111-1111-4111-1111-111111111111"
		tenantID = "00000000-0000-4000-0000-000000000000"
	)

	cases := []struct {
		description string
		role        string
		expected    int
	}{
		{"an operator holds SSHIdentityAdd and is still refused", "operator", http.StatusForbidden},
		{"an observer is refused", "observer", http.StatusForbidden},
		{"an administrator holds SSHIdentityManage", "administrator", http.StatusOK},
		{"an owner holds SSHIdentityManage", "owner", http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := servicemock.NewMockService(t)
			if tc.expected == http.StatusOK {
				svcMock.On("CreateAPIKeySSHIdentity", mock.Anything, mock.MatchedBy(func(req *requests.APIKeySSHIdentityCreate) bool {
					return req.KeyName == "ci" && req.TenantID == tenantID
				})).Return(&models.SSHIdentity{ID: "id1"}, nil).Once()
			}

			body := strings.NewReader(`{"data":"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILBk key","name":"deploy"}`)
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/namespaces/api-key/ci/ssh-identities", body)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-ID", userID)
			req.Header.Set("X-Tenant-ID", tenantID)
			req.Header.Set("X-Role", tc.role)

			rec := httptest.NewRecorder()
			NewRouter(svcMock).ServeHTTP(rec, req)

			require.Equal(t, tc.expected, rec.Result().StatusCode)
		})
	}
}
