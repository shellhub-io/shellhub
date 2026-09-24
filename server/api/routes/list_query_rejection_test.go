package routes

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	servicemock "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
)

func TestListEndpointsRejectAnUnapprovedFilterField(t *testing.T) {
	cases := []struct {
		description  string
		path         string
		filter       string
		uncalledFunc string
	}{
		{
			description:  "refuses to filter a user's invitations by the invite signature",
			path:         "/api/users/invitations",
			filter:       "sig",
			uncalledFunc: "UserMembershipInvitationList",
		},
		{
			description:  "refuses to filter a user's invitations by an unexposed column",
			path:         "/api/users/invitations",
			filter:       "invited_by",
			uncalledFunc: "UserMembershipInvitationList",
		},
		{
			description:  "refuses to filter a namespace's invitations by the invite signature",
			path:         "/api/namespaces/00000000-0000-4000-0000-000000000000/invitations",
			filter:       "sig",
			uncalledFunc: "NamespaceMembershipInvitationList",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := servicemock.NewMockService(t)

			values := url.Values{}
			values.Set("filter", encodeFilter(t, []query.Filter{
				{
					Type:   query.FilterTypeProperty,
					Params: &query.FilterProperty{Name: tc.filter, Operator: "contains", Value: "A"},
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

func TestListEndpointsRejectAnUnapprovedSortField(t *testing.T) {
	cases := []struct {
		description  string
		path         string
		sortBy       string
		uncalledFunc string
	}{
		{
			description:  "refuses to order provisioning keys by the webhook signing secret",
			path:         "/api/namespaces/provisioning-key",
			sortBy:       "webhook_secret",
			uncalledFunc: "ListProvisioningKeys",
		},
		{
			description:  "refuses to order provisioning keys by the key ciphertext",
			path:         "/api/namespaces/provisioning-key",
			sortBy:       "key_encrypted",
			uncalledFunc: "ListProvisioningKeys",
		},
		{
			description:  "refuses to order provisioning key history by an unexposed column",
			path:         "/api/namespaces/provisioning-key/abc/history",
			sortBy:       "public_key",
			uncalledFunc: "ListProvisioningKeyEvents",
		},
		{
			description:  "refuses to order API keys by the key digest",
			path:         "/api/namespaces/api-key",
			sortBy:       "key_digest",
			uncalledFunc: "ListAPIKeys",
		},
		{
			description:  "refuses to order a user's invitations by the invite signature",
			path:         "/api/users/invitations",
			sortBy:       "sig",
			uncalledFunc: "UserMembershipInvitationList",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := servicemock.NewMockService(t)

			values := url.Values{}
			values.Set("sort_by", tc.sortBy)
			values.Set("order_by", "asc")

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
