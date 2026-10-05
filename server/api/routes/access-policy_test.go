package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func accessPolicyBody(filter map[string]any) map[string]any {
	body := map[string]any{
		"name":    "policy",
		"subject": map[string]any{"type": "all-members"},
		"logins":  []string{"root"},
	}

	if filter != nil {
		body["filter"] = filter
	}

	return body
}

func TestAccessPolicyFilterTagLimit(t *testing.T) {
	cases := []struct {
		description    string
		filter         map[string]any
		reachesService bool
		expectedStatus int
	}{
		{
			description:    "fails with more than 3 tags",
			filter:         map[string]any{"tags": []string{"tag1", "tag2", "tag3", "tag4"}},
			reachesService: false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			description:    "succeeds with 3 tags",
			filter:         map[string]any{"tags": []string{"tag1", "tag2", "tag3"}},
			reachesService: true,
			expectedStatus: http.StatusOK,
		},
		{
			description:    "succeeds with an empty tag list",
			filter:         map[string]any{"tags": []string{}},
			reachesService: true,
			expectedStatus: http.StatusOK,
		},
		{
			description:    "succeeds without a filter",
			filter:         nil,
			reachesService: true,
			expectedStatus: http.StatusOK,
		},
	}

	routes := []struct {
		method        string
		url           string
		serviceMethod string
		requestType   string
	}{
		{
			method:        http.MethodPost,
			url:           "/api/access-policies",
			serviceMethod: "CreateAccessPolicy",
			requestType:   "*requests.AccessPolicyCreate",
		},
		{
			method:        http.MethodPut,
			url:           "/api/access-policies/00000000-0000-4000-0000-000000000001",
			serviceMethod: "UpdateAccessPolicy",
			requestType:   "*requests.AccessPolicyUpdate",
		},
	}

	for _, route := range routes {
		for _, tc := range cases {
			t.Run(route.serviceMethod+" "+tc.description, func(t *testing.T) {
				svcMock := mocks.NewMockService(t)
				if tc.reachesService {
					svcMock.
						On(route.serviceMethod, gomock.Anything, gomock.AnythingOfType(route.requestType)).
						Return(&models.AccessPolicy{}, nil).
						Once()
				}

				data, err := json.Marshal(accessPolicyBody(tc.filter))
				require.NoError(t, err)

				req := httptest.NewRequestWithContext(t.Context(), route.method, route.url, strings.NewReader(string(data)))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Tenant-ID", "00000000-0000-4000-0000-000000000000")
				req.Header.Set("X-Role", authorizer.RoleOwner.String())
				req.Header.Set("X-ID", "000000000000000000000000")

				rec := httptest.NewRecorder()
				NewRouter(svcMock).ServeHTTP(rec, req)

				assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
				svcMock.AssertExpectations(t)
			})
		}
	}
}
