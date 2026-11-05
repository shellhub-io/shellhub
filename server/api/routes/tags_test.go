package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/shellhub-io/shellhub/pkg/api/identity/identitytest"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	svc "github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetLegacyTagsTenantScope(t *testing.T) {
	const (
		tenant      = "00000000-0000-4000-0000-000000000000"
		otherTenant = "7e7389a9-55be-4e14-8c47-817a1552774f"
		userID      = "000000000000000000000000"
	)

	cases := []struct {
		description string
		tokenTenant string
		path        string
		admin       bool
		anonymous   bool
		expected    int
	}{
		{
			description: "refuses the tenantless token issued after leaving",
			path:        "/api/namespaces/" + tenant + "/tags",
			expected:    http.StatusForbidden,
		},
		{
			description: "refuses a token scoped to another namespace",
			tokenTenant: otherTenant,
			path:        "/api/namespaces/" + tenant + "/tags",
			expected:    http.StatusForbidden,
		},
		{
			description: "refuses an admin on the regular API without tenant scope",
			path:        "/api/namespaces/" + tenant + "/tags",
			admin:       true,
			expected:    http.StatusForbidden,
		},
		{
			description: "allows an observer in the selected namespace",
			tokenTenant: tenant,
			path:        "/api/namespaces/" + tenant + "/tags",
			expected:    http.StatusOK,
		},
		{
			description: "modern route still refuses a tenantless token",
			path:        "/api/tags",
			expected:    http.StatusForbidden,
		},
		{
			description: "refuses an unauthenticated request",
			path:        "/api/namespaces/" + tenant + "/tags",
			anonymous:   true,
			expected:    http.StatusUnauthorized,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			router, _, service := authenticatedRouter(t)
			if !tc.anonymous {
				service.On("AuthUserClaims", gomock.Anything, gomock.MatchedBy(func(claims *authorizer.UserClaims) bool {
					return claims.ID == userID
				})).Return(&models.User{ID: userID, Admin: tc.admin}, nil).Once()
				if tc.tokenTenant != "" {
					service.On("ResolveNamespaceRole", gomock.Anything, tc.tokenTenant, userID).
						Return(&models.Namespace{TenantID: tc.tokenTenant}, "observer", nil).Once()
				}
			}

			if tc.expected == http.StatusOK {
				service.On("ListTags", gomock.Anything, gomock.MatchedBy(func(req *requests.ListTags) bool {
					return req.TenantID == tenant
				})).Return([]models.Tag{{TenantID: tenant, Name: "CvdSecretTag"}}, int64(1), nil).Once()
			}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.path, nil)
			if !tc.anonymous {
				privateKey, err := testSigningKey()
				require.NoError(t, err)
				bearer, err := jwttoken.EncodeUserClaims(authorizer.UserClaims{
					ID: userID, TenantID: tc.tokenTenant,
				}, "http://localhost", privateKey, clock.New().Now())
				require.NoError(t, err)
				req.Header.Set("Authorization", "Bearer "+bearer)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected, rec.Code)
			if tc.expected == http.StatusOK {
				assert.Contains(t, rec.Body.String(), "CvdSecretTag")
				assert.Equal(t, "1", rec.Header().Get("X-Total-Count"))
			} else {
				assert.NotContains(t, rec.Body.String(), "CvdSecretTag")
				service.AssertNotCalled(t, "ListTags", gomock.Anything, gomock.Anything)
			}
		})
	}
}

func TestGetTags(t *testing.T) {
	svcMock := mocks.NewMockService(t)

	cases := []struct {
		description    string
		query          string
		requiredMocks  func()
		expectedStatus int
		expectedCount  int
		expectedFields map[string]string
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
			expectedFields: map[string]string{"sort_by": query.ErrSorterFieldInvalid.Error()},
		},
		{
			description: "succeeds and returns X-Total-Count header",
			query:       "",
			requiredMocks: func() {
				svcMock.
					On("ListTags", gomock.Anything, gomock.AnythingOfType("*requests.ListTags")).
					Return([]models.Tag{}, int64(5), nil).
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
					Return([]models.Tag{}, int64(3), nil).
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

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
			req = identitytest.As(req, identity.Identity{Role: authorizer.RoleOwner, ID: "000000000000000000000000", TenantID: "00000000-0000-4000-0000-000000000000"})

			rec := httptest.NewRecorder()
			newTestRouter(t, svcMock).ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)

			if tc.expectedStatus == http.StatusOK {
				assert.Equal(t, strconv.Itoa(tc.expectedCount), rec.Result().Header.Get("X-Total-Count"))
			}

			if tc.expectedFields != nil {
				var refusal responses.Error
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &refusal))
				assert.Equal(t, tc.expectedFields, refusal.Fields)
			}
		})
	}

	svcMock.AssertExpectations(t)
}

func TestDeleteTag(t *testing.T) {
	cases := []struct {
		description    string
		url            string
		caller         identity.Identity
		requiredMocks  func(svcMock *mocks.MockService)
		expectedStatus int
	}{
		{
			description:    "fails when role is observer (new URL)",
			url:            "/api/tags/production",
			caller:         identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleObserver, ID: "000000000000000000000000"},
			requiredMocks:  func(_ *mocks.MockService) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			description:    "fails when role is observer (legacy URL)",
			url:            "/api/namespaces/00000000-0000-4000-0000-000000000000/tags/production",
			caller:         identity.Identity{Role: authorizer.RoleObserver, ID: "000000000000000000000000"},
			requiredMocks:  func(_ *mocks.MockService) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			description: "fails when service returns ErrNamespaceNotFound",
			url:         "/api/tags/production",
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner, ID: "000000000000000000000000"},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("DeleteTag", gomock.Anything, &requests.DeleteTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
					}).
					Return(svc.NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", nil)).
					Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			description: "fails with 409 when a device filter holds the tag",
			url:         "/api/tags/production",
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner, ID: "000000000000000000000000"},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("DeleteTag", gomock.Anything, &requests.DeleteTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
					}).
					Return(svc.NewErrTagInUse([]string{`public key "deploy"`})).
					Once()
			},
			expectedStatus: http.StatusConflict,
		},
		{
			description: "succeeds with 204 and no body (new URL)",
			url:         "/api/tags/production",
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner, ID: "000000000000000000000000"},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("DeleteTag", gomock.Anything, &requests.DeleteTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
					}).
					Return(nil).
					Once()
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			description: "succeeds via legacy URL",
			url:         "/api/namespaces/00000000-0000-4000-0000-000000000000/tags/production",
			caller:      identity.Identity{Role: authorizer.RoleOwner, ID: "000000000000000000000000", TenantID: "00000000-0000-4000-0000-000000000000"},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("DeleteTag", gomock.Anything, &requests.DeleteTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
					}).
					Return(nil).
					Once()
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := mocks.NewMockService(t)
			tc.requiredMocks(svcMock)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, tc.url, nil)
			req = identitytest.As(req, tc.caller)

			rec := httptest.NewRecorder()

			e := newTestRouter(t, svcMock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)

			svcMock.AssertExpectations(t)
		})
	}
}

func TestCreateTag(t *testing.T) {
	type Expected struct {
		status     int
		insertedID string
	}

	cases := []struct {
		description   string
		url           string
		headers       map[string]string
		caller        identity.Identity
		body          map[string]any
		requiredMocks func(svcMock *mocks.MockService)
		expected      Expected
	}{
		{
			description: "fails when role is observer (new URL)",
			url:         "/api/tags",
			headers:     map[string]string{"Content-Type": "application/json"},
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleObserver, ID: "000000000000000000000000"},
			body: map[string]any{
				"name": "production",
			},
			requiredMocks: func(_ *mocks.MockService) {},
			expected:      Expected{status: http.StatusForbidden},
		},
		{
			description: "fails when role is observer (legacy URL)",
			url:         "/api/namespaces/00000000-0000-4000-0000-000000000000/tags",
			headers:     map[string]string{"Content-Type": "application/json"},
			caller:      identity.Identity{Role: authorizer.RoleObserver, ID: "000000000000000000000000"},
			body: map[string]any{
				"name": "production",
			},
			requiredMocks: func(_ *mocks.MockService) {},
			expected:      Expected{status: http.StatusForbidden},
		},
		{
			description: "fails when name is too short (< 3 chars)",
			url:         "/api/tags",
			headers:     map[string]string{"Content-Type": "application/json"},
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner, ID: "000000000000000000000000"},
			body: map[string]any{
				"name": "a",
			},
			requiredMocks: func(_ *mocks.MockService) {},
			expected:      Expected{status: http.StatusBadRequest},
		},
		{
			description: "fails when service returns ErrNamespaceNotFound",
			url:         "/api/tags",
			headers:     map[string]string{"Content-Type": "application/json"},
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner, ID: "000000000000000000000000"},
			body: map[string]any{
				"name": "production",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("CreateTag", gomock.Anything, &requests.CreateTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
					}).
					Return("", svc.NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", nil)).
					Once()
			},
			expected: Expected{status: http.StatusNotFound},
		},
		{
			description: "succeeds with 201 and X-Inserted-ID header (new URL)",
			url:         "/api/tags",
			headers:     map[string]string{"Content-Type": "application/json"},
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner, ID: "000000000000000000000000"},
			body: map[string]any{
				"name": "production",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("CreateTag", gomock.Anything, &requests.CreateTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
					}).
					Return("000000000000000000000001", nil).
					Once()
			},
			expected: Expected{
				status:     http.StatusCreated,
				insertedID: "000000000000000000000001",
			},
		},
		{
			description: "returns 409 when the name is already taken",
			url:         "/api/tags",
			headers:     map[string]string{"Content-Type": "application/json"},
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner, ID: "000000000000000000000000"},
			body: map[string]any{
				"name": "production",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("CreateTag", gomock.Anything, &requests.CreateTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
					}).
					Return("", svc.NewErrTagDuplicated([]string{"name"}, nil)).
					Once()
			},
			expected: Expected{status: http.StatusConflict},
		},
		{
			description: "succeeds via legacy URL",
			url:         "/api/namespaces/00000000-0000-4000-0000-000000000000/tags",
			headers:     map[string]string{"Content-Type": "application/json"},
			caller:      identity.Identity{Role: authorizer.RoleOwner, ID: "000000000000000000000000", TenantID: "00000000-0000-4000-0000-000000000000"},
			body: map[string]any{
				"name": "production",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("CreateTag", gomock.Anything, &requests.CreateTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
					}).
					Return("000000000000000000000002", nil).
					Once()
			},
			expected: Expected{
				status:     http.StatusCreated,
				insertedID: "000000000000000000000002",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := mocks.NewMockService(t)
			tc.requiredMocks(svcMock)

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, tc.url, strings.NewReader(string(data)))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			req = identitytest.As(req, tc.caller)

			rec := httptest.NewRecorder()

			e := newTestRouter(t, svcMock)
			e.ServeHTTP(rec, req)

			res := rec.Result()
			assert.Equal(t, tc.expected.status, res.StatusCode)

			if tc.expected.insertedID != "" {
				assert.Equal(t, tc.expected.insertedID, res.Header.Get("X-Inserted-ID"))
			}

			svcMock.AssertExpectations(t)
		})
	}
}

func TestUpdateTag(t *testing.T) {
	type Expected struct {
		status int
		body   string
	}

	renamed := &models.Tag{
		ID:        "tag_00000000-0000-4000-0000-000000000000",
		TenantID:  "00000000-0000-4000-0000-000000000000",
		Name:      "staging",
		CreatedAt: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC),
	}

	renamedBody := `{
		"tenant_id": "00000000-0000-4000-0000-000000000000",
		"name": "staging",
		"created_at": "2026-01-01T12:00:00Z",
		"updated_at": "2026-01-02T12:00:00Z"
	}`

	cases := []struct {
		description   string
		url           string
		headers       map[string]string
		caller        identity.Identity
		body          map[string]any
		requiredMocks func(svcMock *mocks.MockService)
		expected      Expected
	}{
		{
			description: "fails when role is observer (new URL)",
			url:         "/api/tags/production",
			headers:     map[string]string{"Content-Type": "application/json"},
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleObserver, ID: "000000000000000000000000"},
			body: map[string]any{
				"name": "staging",
			},
			requiredMocks: func(_ *mocks.MockService) {},
			expected:      Expected{status: http.StatusForbidden},
		},
		{
			description: "fails when role is observer (legacy URL)",
			url:         "/api/namespaces/00000000-0000-4000-0000-000000000000/tags/production",
			headers:     map[string]string{"Content-Type": "application/json"},
			caller:      identity.Identity{Role: authorizer.RoleObserver, ID: "000000000000000000000000"},
			body: map[string]any{
				"name": "staging",
			},
			requiredMocks: func(_ *mocks.MockService) {},
			expected:      Expected{status: http.StatusForbidden},
		},
		{
			description: "fails when NewName is too short (< 3 chars)",
			url:         "/api/tags/production",
			headers:     map[string]string{"Content-Type": "application/json"},
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner, ID: "000000000000000000000000"},
			body: map[string]any{
				"name": "ab",
			},
			requiredMocks: func(_ *mocks.MockService) {},
			expected:      Expected{status: http.StatusBadRequest},
		},
		{
			description: "fails when service returns ErrNamespaceNotFound",
			url:         "/api/tags/production",
			headers:     map[string]string{"Content-Type": "application/json"},
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner, ID: "000000000000000000000000"},
			body: map[string]any{
				"name": "staging",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("UpdateTag", gomock.Anything, &requests.UpdateTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
						NewName:  "staging",
					}).
					Return(nil, svc.NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", nil)).
					Once()
			},
			expected: Expected{status: http.StatusNotFound},
		},
		{
			description: "succeeds with 200 and the renamed tag (new URL)",
			url:         "/api/tags/production",
			headers:     map[string]string{"Content-Type": "application/json"},
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner, ID: "000000000000000000000000"},
			body: map[string]any{
				"name": "staging",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("UpdateTag", gomock.Anything, &requests.UpdateTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
						NewName:  "staging",
					}).
					Return(renamed, nil).
					Once()
			},
			expected: Expected{status: http.StatusOK, body: renamedBody},
		},
		{
			description: "succeeds via legacy URL",
			url:         "/api/namespaces/00000000-0000-4000-0000-000000000000/tags/production",
			headers:     map[string]string{"Content-Type": "application/json"},
			caller:      identity.Identity{Role: authorizer.RoleOwner, ID: "000000000000000000000000", TenantID: "00000000-0000-4000-0000-000000000000"},
			body: map[string]any{
				"name": "staging",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("UpdateTag", gomock.Anything, &requests.UpdateTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
						NewName:  "staging",
					}).
					Return(renamed, nil).
					Once()
			},
			expected: Expected{status: http.StatusOK, body: renamedBody},
		},
		{
			description: "returns 409 when the name is already taken",
			url:         "/api/tags/production",
			headers:     map[string]string{"Content-Type": "application/json"},
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner, ID: "000000000000000000000000"},
			body: map[string]any{
				"name": "staging",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("UpdateTag", gomock.Anything, &requests.UpdateTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
						NewName:  "staging",
					}).
					Return(nil, svc.NewErrTagDuplicated([]string{"name"}, nil)).
					Once()
			},
			expected: Expected{status: http.StatusConflict},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := mocks.NewMockService(t)
			tc.requiredMocks(svcMock)

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPatch, tc.url, strings.NewReader(string(data)))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			req = identitytest.As(req, tc.caller)

			rec := httptest.NewRecorder()

			e := newTestRouter(t, svcMock)
			e.ServeHTTP(rec, req)

			res := rec.Result()
			assert.Equal(t, tc.expected.status, res.StatusCode)

			if tc.expected.body != "" {
				assert.Equal(t, "application/json", res.Header.Get("Content-Type"))
				assert.JSONEq(t, tc.expected.body, rec.Body.String())
			}

			svcMock.AssertExpectations(t)
		})
	}
}

func TestPushTagToDevice(t *testing.T) {
	const deviceUID = "aabbccddee00112233445566778899aabbccddee"

	cases := []struct {
		description    string
		url            string
		caller         identity.Identity
		requiredMocks  func(svcMock *mocks.MockService)
		expectedStatus int
	}{
		{
			description:    "fails when role is observer (new URL)",
			url:            "/api/devices/" + deviceUID + "/tags/production",
			caller:         identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleObserver, ID: "000000000000000000000000"},
			requiredMocks:  func(_ *mocks.MockService) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			description:    "fails when role is observer (legacy URL)",
			url:            "/api/namespaces/00000000-0000-4000-0000-000000000000/devices/" + deviceUID + "/tags/production",
			caller:         identity.Identity{Role: authorizer.RoleObserver, ID: "000000000000000000000000"},
			requiredMocks:  func(_ *mocks.MockService) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			description:    "fails when tag name is too short (< 3 chars)",
			url:            "/api/devices/" + deviceUID + "/tags/ab",
			caller:         identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner, ID: "000000000000000000000000"},
			requiredMocks:  func(_ *mocks.MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "fails when service returns ErrNamespaceNotFound",
			url:         "/api/devices/" + deviceUID + "/tags/production",
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner, ID: "000000000000000000000000"},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("PushTagTo", gomock.Anything, store.TagTargetDevice, &requests.PushTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
						TargetID: deviceUID,
					}).
					Return(svc.NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", nil)).
					Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			description: "succeeds with 204 and no body (new URL)",
			url:         "/api/devices/" + deviceUID + "/tags/production",
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner, ID: "000000000000000000000000"},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("PushTagTo", gomock.Anything, store.TagTargetDevice, &requests.PushTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
						TargetID: deviceUID,
					}).
					Return(nil).
					Once()
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			description: "succeeds via legacy URL",
			url:         "/api/namespaces/00000000-0000-4000-0000-000000000000/devices/" + deviceUID + "/tags/production",
			caller:      identity.Identity{Role: authorizer.RoleOwner, ID: "000000000000000000000000", TenantID: "00000000-0000-4000-0000-000000000000"},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("PushTagTo", gomock.Anything, store.TagTargetDevice, &requests.PushTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
						TargetID: deviceUID,
					}).
					Return(nil).
					Once()
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := mocks.NewMockService(t)
			tc.requiredMocks(svcMock)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, tc.url, nil)
			req = identitytest.As(req, tc.caller)

			rec := httptest.NewRecorder()

			e := newTestRouter(t, svcMock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)

			svcMock.AssertExpectations(t)
		})
	}
}

func TestPullTagFromDevice(t *testing.T) {
	const deviceUID = "aabbccddee00112233445566778899aabbccddee"

	cases := []struct {
		description    string
		url            string
		caller         identity.Identity
		requiredMocks  func(svcMock *mocks.MockService)
		expectedStatus int
	}{
		{
			description:    "fails when role is observer (new URL)",
			url:            "/api/devices/" + deviceUID + "/tags/production",
			caller:         identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleObserver, ID: "000000000000000000000000"},
			requiredMocks:  func(_ *mocks.MockService) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			description:    "fails when role is observer (legacy URL)",
			url:            "/api/namespaces/00000000-0000-4000-0000-000000000000/devices/" + deviceUID + "/tags/production",
			caller:         identity.Identity{Role: authorizer.RoleObserver, ID: "000000000000000000000000"},
			requiredMocks:  func(_ *mocks.MockService) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			description:    "fails when tag name is too short (< 3 chars)",
			url:            "/api/devices/" + deviceUID + "/tags/ab",
			caller:         identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner, ID: "000000000000000000000000"},
			requiredMocks:  func(_ *mocks.MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "fails when service returns ErrNamespaceNotFound",
			url:         "/api/devices/" + deviceUID + "/tags/production",
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner, ID: "000000000000000000000000"},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("PullTagFrom", gomock.Anything, store.TagTargetDevice, &requests.PullTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
						TargetID: deviceUID,
					}).
					Return(svc.NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", nil)).
					Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			description: "succeeds with 204 and no body (new URL)",
			url:         "/api/devices/" + deviceUID + "/tags/production",
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000", Role: authorizer.RoleOwner, ID: "000000000000000000000000"},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("PullTagFrom", gomock.Anything, store.TagTargetDevice, &requests.PullTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
						TargetID: deviceUID,
					}).
					Return(nil).
					Once()
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			description: "succeeds via legacy URL",
			url:         "/api/namespaces/00000000-0000-4000-0000-000000000000/devices/" + deviceUID + "/tags/production",
			caller:      identity.Identity{Role: authorizer.RoleOwner, ID: "000000000000000000000000", TenantID: "00000000-0000-4000-0000-000000000000"},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("PullTagFrom", gomock.Anything, store.TagTargetDevice, &requests.PullTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
						TargetID: deviceUID,
					}).
					Return(nil).
					Once()
			},
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := mocks.NewMockService(t)
			tc.requiredMocks(svcMock)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodDelete, tc.url, nil)
			req = identitytest.As(req, tc.caller)

			rec := httptest.NewRecorder()

			e := newTestRouter(t, svcMock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)

			svcMock.AssertExpectations(t)
		})
	}
}
