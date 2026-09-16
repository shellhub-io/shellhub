package routes

import (
	"net/http"
	"net/http/httptest"
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
