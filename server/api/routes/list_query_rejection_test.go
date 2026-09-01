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
			description:  "refuses to order install keys by the webhook signing secret",
			path:         "/api/namespaces/install-key",
			sortBy:       "webhook_secret",
			uncalledFunc: "ListInstallKeys",
		},
		{
			description:  "refuses to order install keys by the key ciphertext",
			path:         "/api/namespaces/install-key",
			sortBy:       "key_encrypted",
			uncalledFunc: "ListInstallKeys",
		},
		{
			description:  "refuses to order install key history by an unexposed column",
			path:         "/api/namespaces/install-key/abc/history",
			sortBy:       "public_key",
			uncalledFunc: "ListInstallKeyEvents",
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
