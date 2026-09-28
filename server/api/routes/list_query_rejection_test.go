package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	servicemock "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const listQueryTenant = "00000000-0000-4000-0000-000000000000"

type listRoute struct {
	description      string
	path             string
	serviceCall      string
	serviceArguments int

	refusedFilterField  string
	refusedSortField    string
	acceptedFilterField string
	acceptedFilterOp    string
	acceptedFilterValue string
	acceptedSortField   string
}

func listRoutes() []listRoute {
	return []listRoute{
		{
			description:         "devices",
			path:                "/api/devices",
			serviceCall:         "ListDevices",
			serviceArguments:    3,
			refusedFilterField:  "custom_fields.token",
			refusedSortField:    "id",
			acceptedFilterField: "name",
			acceptedFilterOp:    "contains",
			acceptedSortField:   "last_seen",
		},
		{
			description:       "api keys",
			path:              "/api/namespaces/api-key",
			serviceCall:       "ListAPIKeys",
			serviceArguments:  2,
			refusedSortField:  "key_digest",
			acceptedSortField: "expires_in",
		},
		{
			description:       "provisioning keys",
			path:              "/api/namespaces/provisioning-key",
			serviceCall:       "ListProvisioningKeys",
			serviceArguments:  2,
			refusedSortField:  "webhook_secret",
			acceptedSortField: "used_times",
		},
		{
			description:       "provisioning key history",
			path:              "/api/namespaces/provisioning-key/abc/history",
			serviceCall:       "ListProvisioningKeyEvents",
			serviceArguments:  2,
			refusedSortField:  "public_key",
			acceptedSortField: "decided_at",
		},
		{
			description:         "a user's invitations",
			path:                "/api/users/invitations",
			serviceCall:         "UserMembershipInvitationList",
			serviceArguments:    2,
			refusedFilterField:  "sig",
			refusedSortField:    "sig",
			acceptedFilterField: "status",
			acceptedFilterOp:    "eq",
			acceptedFilterValue: "pending",
			acceptedSortField:   "expires_at",
		},
		{
			description:         "a namespace's invitations",
			path:                "/api/namespaces/" + listQueryTenant + "/invitations",
			serviceCall:         "NamespaceMembershipInvitationList",
			serviceArguments:    2,
			refusedFilterField:  "sig",
			refusedSortField:    "sig",
			acceptedFilterField: "role",
			acceptedFilterOp:    "eq",
			acceptedFilterValue: "operator",
			acceptedSortField:   "expires_at",
		},
		{
			description:        "tags",
			path:               "/api/tags",
			serviceCall:        "ListTags",
			serviceArguments:   2,
			refusedFilterField: "name",
			refusedSortField:   "id",
			acceptedSortField:  "updated_at",
		},
		{
			description:        "tags under the deprecated namespace path",
			path:               "/api/namespaces/" + listQueryTenant + "/tags",
			serviceCall:        "ListTags",
			serviceArguments:   2,
			refusedFilterField: "name",
			refusedSortField:   "id",
			acceptedSortField:  "updated_at",
		},
		{
			description:         "sessions",
			path:                "/api/sessions",
			serviceCall:         "ListSessions",
			serviceArguments:    3,
			refusedFilterField:  "username",
			acceptedFilterField: "device_uid",
			acceptedFilterOp:    "eq",
		},
		{
			description:         "public keys",
			path:                "/api/sshkeys/public-keys",
			serviceCall:         "ListPublicKeys",
			serviceArguments:    2,
			refusedFilterField:  "data",
			acceptedFilterField: "fingerprint",
			acceptedFilterOp:    "contains",
		},
		{
			description:         "namespaces",
			path:                "/api/namespaces",
			serviceCall:         "ListNamespaces",
			serviceArguments:    2,
			refusedFilterField:  "tenant_id",
			acceptedFilterField: "name",
			acceptedFilterOp:    "contains",
		},
		{
			description:      "namespace members",
			path:             "/api/namespaces/" + listQueryTenant + "/members",
			serviceCall:      "ListNamespaceMembers",
			serviceArguments: 2,
		},
		{
			description:      "access policies",
			path:             "/api/access-policies",
			serviceCall:      "ListAccessPolicies",
			serviceArguments: 2,
		},
		{
			description:      "ssh identities",
			path:             "/api/ssh-identities",
			serviceCall:      "ListSSHIdentities",
			serviceArguments: 2,
		},
	}
}

func filterOn(t *testing.T, name, operator, value string) string {
	t.Helper()

	return encodeFilter(t, []query.Filter{
		{
			Type:   query.FilterTypeProperty,
			Params: &query.FilterProperty{Name: name, Operator: operator, Value: value},
		},
	})
}

func serveList(t *testing.T, svcMock *servicemock.MockService, path string, values url.Values) *httptest.ResponseRecorder {
	t.Helper()

	target := path
	if encoded := values.Encode(); encoded != "" {
		target += "?" + encoded
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	req.Header.Set("X-ID", "000000000000000000000000")
	req.Header.Set("X-Tenant-ID", listQueryTenant)
	req.Header.Set("X-Role", "owner")

	rec := httptest.NewRecorder()
	NewRouter(svcMock).ServeHTTP(rec, req)

	return rec
}

func assertRefusedField(t *testing.T, rec *httptest.ResponseRecorder, field string) {
	t.Helper()

	require.Equal(t, http.StatusBadRequest, rec.Result().StatusCode, rec.Body.String())

	var body responses.Error
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Contains(t, body.Fields, field, "the refusal must name what the client got wrong")
}

func TestListEndpointsRejectAnUnapprovedFilterField(t *testing.T) {
	for _, tc := range listRoutes() {
		if tc.refusedFilterField == "" {
			continue
		}

		t.Run(tc.description, func(t *testing.T) {
			svcMock := servicemock.NewMockService(t)

			values := url.Values{"filter": {filterOn(t, tc.refusedFilterField, "contains", "A")}}

			assertRefusedField(t, serveList(t, svcMock, tc.path, values), "filter")
			svcMock.AssertNumberOfCalls(t, tc.serviceCall, 0)
		})
	}
}

func TestListEndpointsRejectAnUnapprovedSortField(t *testing.T) {
	for _, tc := range listRoutes() {
		if tc.refusedSortField == "" {
			continue
		}

		t.Run(tc.description, func(t *testing.T) {
			svcMock := servicemock.NewMockService(t)

			values := url.Values{"sort_by": {tc.refusedSortField}, "order_by": {"asc"}}

			assertRefusedField(t, serveList(t, svcMock, tc.path, values), "sort_by")
			svcMock.AssertNumberOfCalls(t, tc.serviceCall, 0)
		})
	}
}

// TestListEndpointsAcceptTheirOwnContract is the direction the route table's invariant cannot check.
// That invariant proves a list route names a contract; only driving the route with a field its own
// resource allows proves it names the right one, and a route wired to a neighbour's contract fails
// here rather than in production.
//
// The three lists that take no query at all — members, access policies and SSH identities — have
// nothing to send, so for them this only reaches the service and the header. Their contract is
// pinned by the rejection tests above being inapplicable, not by this one.
func TestListEndpointsAcceptTheirOwnContract(t *testing.T) {
	for _, tc := range listRoutes() {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := servicemock.NewMockService(t)

			arguments := make([]any, tc.serviceArguments)
			for i := range arguments {
				arguments[i] = mock.Anything
			}

			svcMock.On(tc.serviceCall, arguments...).Return(nil, 0, nil).Once()

			values := url.Values{}
			if tc.acceptedFilterField != "" {
				value := tc.acceptedFilterValue
				if value == "" {
					value = "A"
				}

				values.Set("filter", filterOn(t, tc.acceptedFilterField, tc.acceptedFilterOp, value))
			}

			if tc.acceptedSortField != "" {
				values.Set("sort_by", tc.acceptedSortField)
			}

			rec := serveList(t, svcMock, tc.path, values)

			require.Equal(t, http.StatusOK, rec.Result().StatusCode, rec.Body.String())
			assert.Equal(t, "0", rec.Header().Get("X-Total-Count"))
			svcMock.AssertExpectations(t)
		})
	}
}

func TestListEndpointsRejectAValueTheColumnCannotHold(t *testing.T) {
	cases := []struct {
		description  string
		path         string
		field        string
		value        string
		uncalledFunc string
	}{
		{
			description:  "refuses a device status no device can hold",
			path:         "/api/devices",
			field:        "status",
			value:        "theprimeagen",
			uncalledFunc: "ListDevices",
		},
		{
			description:  "refuses an invitation status no invitation can hold",
			path:         "/api/users/invitations",
			field:        "status",
			value:        "theprimeagen",
			uncalledFunc: "UserMembershipInvitationList",
		},
		{
			description:  "refuses a membership role nobody can hold",
			path:         "/api/users/invitations",
			field:        "role",
			value:        "theprimeagen",
			uncalledFunc: "UserMembershipInvitationList",
		},
		{
			description:  "refuses an invitation status on a namespace's invitations",
			path:         "/api/namespaces/00000000-0000-4000-0000-000000000000/invitations",
			field:        "status",
			value:        "theprimeagen",
			uncalledFunc: "NamespaceMembershipInvitationList",
		},
		{
			description:  "refuses a namespace type no namespace can hold",
			path:         "/api/namespaces",
			field:        "type",
			value:        "theprimeagen",
			uncalledFunc: "ListNamespaces",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := servicemock.NewMockService(t)

			values := url.Values{}
			values.Set("filter", encodeFilter(t, []query.Filter{
				{
					Type:   query.FilterTypeProperty,
					Params: &query.FilterProperty{Name: tc.field, Operator: "eq", Value: tc.value},
				},
			}))

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.path+"?"+values.Encode(), nil)
			req.Header.Set("X-ID", "000000000000000000000000")
			req.Header.Set("X-Tenant-ID", "00000000-0000-4000-0000-000000000000")
			req.Header.Set("X-Role", "owner")

			rec := httptest.NewRecorder()
			NewRouter(svcMock).ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Result().StatusCode)
			svcMock.AssertNotCalled(t, tc.uncalledFunc)
		})
	}
}
