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

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/shellhub-io/shellhub/pkg/api/identity/identitytest"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
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
		expectedFields  map[string]string
	}
	cases := []struct {
		description   string
		paginator     query.Paginator
		filter        string
		caller        identity.Identity
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when try to searching a session list of a existing session",
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner},
			requiredMocks: func() {
				mock.
					On("ListSessions", gomock.Anything, gomock.Anything, &requests.ListSessions{Paginator: query.Paginator{Page: 1, PerPage: 10}, TenantID: "00000000-0000-4000-0000-000000000000"}).
					Return(nil, int64(0), svc.ErrNotFound).
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
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner},
			requiredMocks: func() {
				mock.
					On("ListSessions", gomock.Anything, gomock.Anything, &requests.ListSessions{Paginator: query.Paginator{Page: 2, PerPage: 5}, TenantID: "00000000-0000-4000-0000-000000000000"}).
					Return([]models.Session{}, int64(1), nil).
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
			caller:        identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner},
			requiredMocks: func() {},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusBadRequest,
				expectedFields:  map[string]string{"filter": query.ErrFilterPropertyInvalid.Error()},
			},
		},
		{
			description:   "returns 400 when filter is malformed non-base64",
			paginator:     query.Paginator{Page: 1, PerPage: 10},
			filter:        "!!!not-base64!!!",
			caller:        identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner},
			requiredMocks: func() {},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusBadRequest,
				expectedFields:  map[string]string{"filter": query.ErrFilterInvalid.Error()},
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
			caller: identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner},
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
					Return([]models.Session{}, int64(1), nil).
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
			urlVal.Set("page", strconv.FormatInt(tc.paginator.Page, 10))
			urlVal.Set("per_page", strconv.FormatInt(tc.paginator.PerPage, 10))
			if tc.filter != "" {
				urlVal.Set("filter", tc.filter)
			}

			rawURL += "?" + urlVal.Encode()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, rawURL, nil)
			req.Header.Set("Content-Type", "application/json")
			req = identitytest.As(req, tc.caller)

			rec := httptest.NewRecorder()

			e := newTestRouter(t, mock)
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

			if tc.expected.expectedFields != nil {
				var refusal responses.Error
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &refusal))
				assert.Equal(t, tc.expected.expectedFields, refusal.Fields)
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
				mock.On("GetSession", gomock.Anything, scope.MustBounded("00000000-0000-4000-0000-000000000000"), sessionGet("123", true)).Return(session, nil)
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
				mock.On("GetSession", gomock.Anything, scope.MustBounded("00000000-0000-4000-0000-000000000000"), sessionGet("1234", true)).Return(nil, svc.NewErrSessionNotFound(models.UID("1234"), store.ErrNoDocuments))
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
				mock.On("GetSession", gomock.Anything, scope.MustBounded("00000000-0000-4000-0000-000000000000"), sessionGet("123", true)).Return(session, nil)
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
			req = identitytest.As(req, identity.Identity{Role: authorizer.RoleOwner, TenantID: tc.tenant, Admin: tc.admin})
			rec := httptest.NewRecorder()

			e := newTestRouter(t, mock)
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

func sessionGet(uid string, timeline bool) requests.SessionGet {
	return requests.SessionGet{SessionIDParam: requests.SessionIDParam{UID: uid}, Timeline: timeline}
}

func TestGetSessionGivesTheTimelineOnlyToCallersWhoMayReplay(t *testing.T) {
	const tenant = "00000000-0000-4000-0000-000000000000"

	cases := []struct {
		caller   string
		role     authorizer.Role
		admin    bool
		timeline bool
	}{
		{caller: "observer", role: authorizer.RoleObserver, timeline: false},
		{caller: "operator", role: authorizer.RoleOperator, timeline: false},
		{caller: "administrator", role: authorizer.RoleAdministrator, timeline: true},
		{caller: "owner", role: authorizer.RoleOwner, timeline: true},
		{caller: "instance admin console", admin: true, timeline: true},
	}

	for _, tc := range cases {
		t.Run(tc.caller, func(t *testing.T) {
			service := mocks.NewMockService(t)
			router := newTestRouter(t, service)
			path := "/api/sessions/123"

			if tc.admin {
				path = "/admin/api/sessions/123"
				markAdmin := func(next echo.HandlerFunc) echo.HandlerFunc {
					return func(c *echo.Context) error {
						gateway.NewContext(c).MarkAdminRoute()

						return next(c)
					}
				}
				router.GET("/admin/api/sessions/:uid", gateway.Handler(NewHandler(service).GetSession), markAdmin)
			}

			service.On("GetSession", gomock.Anything, gomock.AnythingOfType("scope.Scope"), sessionGet("123", tc.timeline)).
				Return(&models.Session{UID: "123"}, nil).Once()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
			if !tc.admin {
				req = identitytest.As(req, identity.Identity{Role: tc.role, TenantID: tenant})
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}
