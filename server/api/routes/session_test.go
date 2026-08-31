package routes

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	svc "github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetSessionList(t *testing.T) {
	mock := mocks.NewMockService(t)

	type Expected struct {
		expectedSession []models.Session
		expectedStatus  int
		expectedCount   int // total count from X-Total-Count header; only checked for 200 responses
	}
	cases := []struct {
		description   string
		paginator     query.Paginator
		filter        string
		headers       map[string]string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when try to searching a session list of a existing session",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			headers:     map[string]string{"X-Tenant-ID": "00000000-0000-4000-0000-000000000000"},
			requiredMocks: func() {
				mock.
					On("ListSessions", gomock.Anything, gomock.Anything, &requests.ListSessions{Paginator: query.Paginator{Page: 1, PerPage: 10}, TenantID: "00000000-0000-4000-0000-000000000000"}).
					Return(nil, 0, svc.ErrNotFound).
					Once()
			},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			description: "success when try to searching a session list of a existing session",
			paginator:   query.Paginator{Page: 2, PerPage: 5},
			headers:     map[string]string{"X-Tenant-ID": "00000000-0000-4000-0000-000000000000"},
			requiredMocks: func() {
				mock.
					On("ListSessions", gomock.Anything, gomock.Anything, &requests.ListSessions{Paginator: query.Paginator{Page: 2, PerPage: 5}, TenantID: "00000000-0000-4000-0000-000000000000"}).
					Return([]models.Session{}, 1, nil).
					Once()
			},
			expected: Expected{
				expectedSession: []models.Session{},
				expectedStatus:  http.StatusOK,
				expectedCount:   1,
			},
		},
		{
			description: "returns 400 when filter uses a disallowed field",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			filter: func() string {
				filters := []query.Filter{
					{
						Type: query.FilterTypeProperty,
						Params: &query.FilterProperty{
							Name:     "username",
							Operator: "eq",
							Value:    "johndoe",
						},
					},
				}
				b, err := json.Marshal(filters)
				require.NoError(t, err)

				return base64.StdEncoding.EncodeToString(b)
			}(),
			headers:       map[string]string{"X-Tenant-ID": "00000000-0000-4000-0000-000000000000"},
			requiredMocks: func() {},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusBadRequest,
			},
		},
		{
			description:   "returns 400 when filter is malformed non-base64",
			paginator:     query.Paginator{Page: 1, PerPage: 10},
			filter:        "!!!not-base64!!!",
			headers:       map[string]string{"X-Tenant-ID": "00000000-0000-4000-0000-000000000000"},
			requiredMocks: func() {},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusBadRequest,
			},
		},
		{
			description: "returns 200 for a valid allowed filter",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			filter: func() string {
				filters := []query.Filter{
					{
						Type: query.FilterTypeProperty,
						Params: &query.FilterProperty{
							Name:     "device_uid",
							Operator: "eq",
							Value:    "abc123",
						},
					},
				}
				b, err := json.Marshal(filters)
				require.NoError(t, err)

				return base64.StdEncoding.EncodeToString(b)
			}(),
			headers: map[string]string{"X-Tenant-ID": "00000000-0000-4000-0000-000000000000"},
			requiredMocks: func() {
				mock.
					On("ListSessions", gomock.Anything, gomock.Anything, gomock.MatchedBy(func(req *requests.ListSessions) bool {
						if len(req.Filters.Data) != 1 {
							return false
						}

						prop, ok := req.Filters.Data[0].Params.(*query.FilterProperty)
						if !ok {
							return false
						}

						return prop.Name == "device_uid" && prop.Operator == "eq" && prop.Value == "abc123"
					})).
					Return([]models.Session{}, 1, nil).
					Once()
			},
			expected: Expected{
				expectedSession: []models.Session{},
				expectedStatus:  http.StatusOK,
				expectedCount:   1,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			rawURL := "/api/sessions"
			urlVal := url.Values{}
			urlVal.Set("page", strconv.Itoa(tc.paginator.Page))
			urlVal.Set("per_page", strconv.Itoa(tc.paginator.PerPage))
			if tc.filter != "" {
				urlVal.Set("filter", tc.filter)
			}

			rawURL += "?" + urlVal.Encode()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, rawURL, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session []models.Session
			if rec.Result().StatusCode < http.StatusBadRequest {
				if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
					assert.ErrorIs(t, io.EOF, err)
				}
			}
			assert.Equal(t, tc.expected.expectedSession, session)

			if rec.Result().StatusCode == http.StatusOK {
				assert.Equal(t, strconv.Itoa(tc.expected.expectedCount), rec.Result().Header.Get("X-Total-Count"))
			}
		})
	}

	mock.AssertExpectations(t)
}

func TestGetSession(t *testing.T) {
	mock := mocks.NewMockService(t)

	type Expected struct {
		expectedSession *models.Session
		expectedStatus  int
	}
	cases := []struct {
		title         string
		uid           string
		tenant        string
		admin         bool
		requiredMocks func(session *models.Session)
		expected      Expected
	}{
		{
			title:         "fails when try to get session don't existing",
			uid:           "",
			tenant:        "00000000-0000-4000-0000-000000000000",
			requiredMocks: func(*models.Session) {},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title:         "refuses the request when the caller carries no tenant",
			uid:           "1234",
			tenant:        "",
			requiredMocks: func(*models.Session) {},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusForbidden,
			},
		},
		{
			title:  "admin user on the regular path stays bounded to their namespace",
			uid:    "123",
			tenant: "00000000-0000-4000-0000-000000000000",
			admin:  true,
			requiredMocks: func(session *models.Session) {
				mock.On("GetSession", gomock.Anything, scope.MustBounded("00000000-0000-4000-0000-000000000000"), models.UID("123")).Return(session, nil)
			},
			expected: Expected{
				expectedSession: &models.Session{UID: "123"},
				expectedStatus:  http.StatusOK,
			},
		},
		{
			title:  "fails when try to get session don't existing",
			uid:    "1234",
			tenant: "00000000-0000-4000-0000-000000000000",
			requiredMocks: func(*models.Session) {
				mock.On("GetSession", gomock.Anything, scope.MustBounded("00000000-0000-4000-0000-000000000000"), models.UID("1234")).Return(nil, svc.NewErrSessionNotFound(models.UID("1234"), store.ErrNoDocuments))
			},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title:  "success when try to get a session exists",
			uid:    "123",
			tenant: "00000000-0000-4000-0000-000000000000",
			requiredMocks: func(session *models.Session) {
				mock.On("GetSession", gomock.Anything, scope.MustBounded("00000000-0000-4000-0000-000000000000"), models.UID("123")).Return(session, nil)
			},
			expected: Expected{
				expectedSession: &models.Session{UID: "123"},
				expectedStatus:  http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.expected.expectedSession)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/sessions/"+tc.uid, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			if tc.tenant != "" {
				req.Header.Set("X-Tenant-ID", tc.tenant)
			}
			if tc.admin {
				req.Header.Set("X-Admin", "true")
			}
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session *models.Session
			if rec.Result().StatusCode < http.StatusBadRequest {
				if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
					assert.ErrorIs(t, io.EOF, err)
				}
			}

			assert.Equal(t, tc.expected.expectedSession, session)
		})
	}

	mock.AssertExpectations(t)
}
