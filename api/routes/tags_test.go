package routes

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestGetTags(t *testing.T) {
	svcMock := new(mocks.Service)

	cases := []struct {
		description    string
		query          string
		requiredMocks  func()
		expectedStatus int
		expectedCount  int
	}{
		{
			description:    "fails with bad filter query param",
			query:          "filter=!!!notbase64!!!",
			requiredMocks:  func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "fails when filter contains a name+contains property filter",
			query: "filter=" + encodeFilter(t, []query.Filter{
				{
					Type: query.FilterTypeProperty,
					Params: &query.FilterProperty{
						Name:     "name",
						Operator: "contains",
						Value:    "foo",
					},
				},
			}),
			requiredMocks:  func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "fails when filter contains a foobar+eq property filter",
			query: "filter=" + encodeFilter(t, []query.Filter{
				{
					Type: query.FilterTypeProperty,
					Params: &query.FilterProperty{
						Name:     "foobar",
						Operator: "eq",
						Value:    "baz",
					},
				},
			}),
			requiredMocks:  func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description:    "fails when sort_by is an unknown field (badcolumn)",
			query:          "sort_by=badcolumn",
			requiredMocks:  func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "succeeds and returns X-Total-Count header",
			query:       "",
			requiredMocks: func() {
				svcMock.
					On("ListTags", gomock.Anything, gomock.AnythingOfType("*requests.ListTags")).
					Return([]models.Tag{}, 5, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedCount:  5,
		},
		{
			description: "succeeds with sort_by=name and returns X-Total-Count header",
			query:       "sort_by=name",
			requiredMocks: func() {
				svcMock.
					On("ListTags", gomock.Anything, gomock.AnythingOfType("*requests.ListTags")).
					Return([]models.Tag{}, 3, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			url := "/api/tags"
			if tc.query != "" {
				url += "?" + tc.query
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			req.Header.Set("X-ID", "000000000000000000000000")
			req.Header.Set("X-Tenant-ID", "00000000-0000-4000-0000-000000000000")

			rec := httptest.NewRecorder()
			NewRouter(svcMock).ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)

			if tc.expectedStatus == http.StatusOK {
				assert.Equal(t, strconv.Itoa(tc.expectedCount), rec.Result().Header.Get("X-Total-Count"))
			}
		})
	}

	svcMock.AssertExpectations(t)
}
