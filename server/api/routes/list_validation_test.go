package routes

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/shellhub-io/shellhub/pkg/api/identity/identitytest"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	servicemock "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
)

var listQueryCases = []struct {
	name   string
	handle func(*Handler, *gateway.Context) error
	filter bool
	sort   bool
}{
	{"GetDeviceList", (*Handler).GetDeviceList, true, true},
	{"GetDeviceLocations", (*Handler).GetDeviceLocations, true, false},
	{"GetSessionList", (*Handler).GetSessionList, true, false},
	{"GetNamespaceList", (*Handler).GetNamespaceList, true, false},
	{"GetPublicKeys", (*Handler).GetPublicKeys, true, false},
	{"GetTags", (*Handler).GetTags, true, true},
	{"ListAPIKeys", (*Handler).ListAPIKeys, false, true},
	{"ListInstanceAPIKeys", (*Handler).ListInstanceAPIKeys, false, true},
	{"ListProvisioningKeys", (*Handler).ListProvisioningKeys, false, true},
	{"HistoryProvisioningKey", (*Handler).HistoryProvisioningKey, false, true},
	{"GetUserMembershipInvitationList", (*Handler).GetUserMembershipInvitationList, true, true},
	{"GetNamespaceMembershipInvitationList", (*Handler).GetNamespaceMembershipInvitationList, true, true},
	{"TenantRegistrations", (*Handler).TenantRegistrations, false, true},
}

func TestListHandlersRejectUnexposedQueryFields(t *testing.T) {
	for _, tc := range listQueryCases {
		for _, field := range []string{"filter", "sort_by"} {
			if field == "filter" && !tc.filter || field == "sort_by" && !tc.sort {
				continue
			}
			t.Run(tc.name+"/"+field, func(t *testing.T) {
				svc := servicemock.NewMockService(t)
				handler := NewHandler(svc)
				values := url.Values{"sort_by": {"unexposed_column"}, "order_by": {"asc"}}
				if field == "filter" {
					values = url.Values{"filter": {encodeFilter(t, []query.Filter{{Type: query.FilterTypeProperty, Params: &query.FilterProperty{Name: "unexposed_column", Operator: "eq", Value: "secret"}}})}}
				}
				req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/audit/00000000-0000-4000-0000-000000000000/key?"+values.Encode(), nil)
				req = identitytest.As(req, identity.Identity{ID: "00000000-0000-4000-0000-000000000001", TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner})
				rec := httptest.NewRecorder()
				router := newTestRouter(t, svc)
				router.GET("/audit/:tenant/:id", gateway.Handler(func(c *gateway.Context) error {
					return tc.handle(handler, c)
				}))
				router.ServeHTTP(rec, req)
				assert.Equal(t, http.StatusBadRequest, rec.Code)
				assert.Contains(t, rec.Body.String(), field)
				assert.Empty(t, svc.Calls, "an invalid query must not reach the service")
			})
		}
	}
}
