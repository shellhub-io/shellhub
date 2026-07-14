package routes

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/envs/envstest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestCreateNamespace(t *testing.T) {
	envstest.SetEdition(t, envs.Enterprise)

	mock := mocks.NewMockService(t)

	type Expected struct {
		expectedSession *models.Namespace
		expectedStatus  int
	}
	cases := []struct {
		title         string
		userID        string
		req           string
		expected      Expected
		requiredMocks func()
	}{
		{
			title:  "fails when try to creating a namespace",
			userID: "00000000-0000-4000-0000-000000000000",
			req:    `{ "name": "namespace", "tenant": "36512069-be88-497a-b0ec-03ed05b1f7e7"}`,
			requiredMocks: func() {
				mock.On("CreateNamespace", gomock.Anything, gomock.AnythingOfType("*requests.NamespaceCreate")).Return(nil, svc.ErrNotFound).Once()
			},
			expected: Expected{
				expectedStatus:  http.StatusNotFound,
				expectedSession: &models.Namespace{},
			},
		},
		{
			title:  "success when try to creating a namespace",
			userID: "123",
			req:    `{ "name": "namespace", "tenant": "36512069-be88-497a-b0ec-03ed05b1f7e7"}`,
			requiredMocks: func() {
				mock.On("CreateNamespace", gomock.Anything, gomock.AnythingOfType("*requests.NamespaceCreate")).Return(&models.Namespace{}, nil).Once()
			},
			expected: Expected{
				expectedStatus:  http.StatusOK,
				expectedSession: &models.Namespace{},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodPost, "/api/namespaces", strings.NewReader(tc.req))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			req.Header.Set("X-ID", "00000000-0000-4000-0000-000000000000")
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session models.Namespace
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}
			assert.Equal(t, tc.expected.expectedSession, &session)
		})
	}

	mock.AssertExpectations(t)
}

func TestGetNamespace(t *testing.T) {
	mock := mocks.NewMockService(t)

	type Expected struct {
		expectedSession *models.Namespace
		expectedStatus  int
	}
	cases := []struct {
		title         string
		uid           string
		req           string
		expected      Expected
		requiredMocks func()
	}{
		{
			title: "fails when validate because the tenant does not have a min of 3 characters",
			uid:   "123",
			req:   "tg",
			expected: Expected{
				expectedStatus: http.StatusBadRequest,
			}, requiredMocks: func() {},
		},

		{
			title: "fails when validate because the tenant does not have a max of 255 characters",
			uid:   "123",
			req:   "BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9",
			expected: Expected{
				expectedStatus: http.StatusBadRequest,
			}, requiredMocks: func() {},
		},
		{
			title: "fails when validate because have a '/' with in your characters",
			uid:   "123",
			req:   "tes/t",
			expected: Expected{
				expectedStatus: http.StatusNotFound,
			}, requiredMocks: func() {},
		},
		{
			title: "success when try to get a existing namespace",
			uid:   "123",
			req:   "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				mock.On("GetNamespace", gomock.Anything, "00000000-0000-4000-0000-000000000000").Return(&models.Namespace{}, nil)
			},
			expected: Expected{
				expectedStatus:  http.StatusOK,
				expectedSession: &models.Namespace{},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/namespaces/%s", tc.req), nil)

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			req.Header.Set("X-Tenant-ID", tc.req)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session *models.Namespace
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}
			assert.Equal(t, tc.expected.expectedSession, session)
		})
	}

	mock.AssertExpectations(t)
}

func TestDeleteNamespace(t *testing.T) {
	mock := mocks.NewMockService(t)

	cases := []struct {
		title          string
		uid            string
		req            string
		requiredMocks  func()
		expectedStatus int
	}{
		{
			title:          "fails when bind fails to validate uid",
			uid:            "123",
			req:            "",
			expectedStatus: http.StatusNotFound,
			requiredMocks:  func() {},
		},
		{
			title:          "fails when validate because the tenant does not have a min of 3 characters",
			uid:            "123",
			req:            "tg",
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func() {},
		},
		{
			title:          "fails when validate because the tenant does not have a max of 255 characters",
			uid:            "123",
			req:            "BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9",
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func() {},
		},
		{
			title:          "fails when validate because have a '/' with in your characters",
			uid:            "123",
			req:            "tes/t",
			expectedStatus: http.StatusNotFound,
			requiredMocks:  func() {},
		},
		{
			title: "fails when try to deleting a existing namespace",
			uid:   "123",
			req:   "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				mock.On("DeleteNamespace", gomock.Anything, "00000000-0000-4000-0000-000000000000").Return(svc.ErrNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to deleting a existing namespace",
			uid:   "123",
			req:   "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				mock.On("DeleteNamespace", gomock.Anything, "00000000-0000-4000-0000-000000000000").Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/namespaces/%s", tc.req), nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			req.Header.Set("X-ID", tc.uid)
			req.Header.Set("X-Tenant-ID", tc.req)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}

	mock.AssertExpectations(t)
}

func TestEditNamespace(t *testing.T) {
	svcMock := mocks.NewMockService(t)

	cases := []struct {
		description   string
		headers       map[string]string
		body          map[string]interface{}
		requiredMocks func()
		expected      int
	}{
		{
			description: "fails when role is observer",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"session_record": true,
			},
			requiredMocks: func() {
			},
			expected: http.StatusForbidden,
		},
		{
			description: "fails when role is operator",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"session_record": true,
			},
			requiredMocks: func() {
			},
			expected: http.StatusForbidden,
		},
		{
			description: "fails when try to editing an non-existing namespace",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"session_record": true,
			},
			requiredMocks: func() {
				svcMock.
					On("EditSessionRecordStatus", gomock.Anything, true, "00000000-0000-4000-0000-000000000000").
					Return(svc.ErrNotFound).
					Once()
			},
			expected: http.StatusNotFound,
		},
		{
			description: "success when try to editing an existing namespace",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"session_record": true,
				"tenant":         "00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				svcMock.
					On("EditSessionRecordStatus", gomock.Anything, true, "00000000-0000-4000-0000-000000000000").
					Return(nil).
					Once()
			},
			expected: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			jsonData, err := json.Marshal(tc.body)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/security/%s", tc.headers["X-Tenant-ID"]), strings.NewReader(string(jsonData)))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected, rec.Result().StatusCode)
		})
	}

	svcMock.AssertExpectations(t)
}

func TestHandler_LeaveNamespace(t *testing.T) {
	svcMock := mocks.NewMockService(t)

	cases := []struct {
		description   string
		tenantID      string
		headers       map[string]string
		requiredMocks func()
		expected      int
	}{
		{
			description: "fails with api key",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			headers: map[string]string{
				"X-API-KEY":   "b2f7cc0e-d933-4aad-9ab2-b557f2f2554f",
				"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {},
			expected:      http.StatusForbidden,
		},
		{
			description: "fails to leave the namespace",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			headers: map[string]string{
				"X-ID":        "000000000000000000000000",
				"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				svcMock.
					On("LeaveNamespace", gomock.Anything, &requests.LeaveNamespace{UserID: "000000000000000000000000", TenantID: "00000000-0000-4000-0000-000000000000", AuthenticatedTenantID: "00000000-0000-4000-0000-000000000000"}).
					Return(nil, errors.New("error")).
					Once()
			},
			expected: http.StatusInternalServerError,
		},
		{
			description: "success to leave the namespace",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			headers: map[string]string{
				"X-ID":        "000000000000000000000000",
				"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
			},
			requiredMocks: func() {
				svcMock.
					On("LeaveNamespace", gomock.Anything, &requests.LeaveNamespace{UserID: "000000000000000000000000", TenantID: "00000000-0000-4000-0000-000000000000", AuthenticatedTenantID: "00000000-0000-4000-0000-000000000000"}).
					Return(nil, nil).
					Once()
			},
			expected: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodDelete, "/api/namespaces/"+tc.tenantID+"/members", nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			assert.Equal(tt, tc.expected, rec.Result().StatusCode)
		})
	}

	svcMock.AssertExpectations(t)
}

// TestNamespaceCrossTenantAccess ensures that callers cannot read, edit, delete
// or toggle session recording of a namespace they are not scoped to. Covers the
// regression described in GHSA-vwx9-7qcf-gg7f against both auth shapes: the
// API-key caller (no X-ID) that the advisory reproduced and the JWT caller.
func TestNamespaceCrossTenantAccess(t *testing.T) {
	const (
		callerTenant = "00000000-0000-4000-0000-000000000000"
		victimTenant = "7e7389a9-55be-4e14-8c47-817a1552774f"
		victimEmail  = "victim@example.com"
		victimOwner  = "victim-owner"
	)

	routes := []struct {
		description string
		method      string
		url         string
		body        string
	}{
		{
			description: "GET /namespaces/:tenant is blocked cross-tenant",
			method:      http.MethodGet,
			url:         "/api/namespaces/" + victimTenant,
		},
		{
			description: "PUT /namespaces/:tenant is blocked cross-tenant",
			method:      http.MethodPut,
			url:         "/api/namespaces/" + victimTenant,
			body:        `{"name":"pwned"}`,
		},
		{
			description: "DELETE /namespaces/:tenant is blocked cross-tenant",
			method:      http.MethodDelete,
			url:         "/api/namespaces/" + victimTenant,
		},
		{
			description: "PUT /users/security/:tenant is blocked cross-tenant",
			method:      http.MethodPut,
			url:         "/api/users/security/" + victimTenant,
			body:        `{"session_record":false}`,
		},
	}

	shapes := []struct {
		name    string
		headers map[string]string
	}{
		{
			// The advisory PoC: X-API-Key + X-Tenant-ID, no X-ID.
			name: "api-key shape",
			headers: map[string]string{
				"X-API-Key":   "caller-api-key",
				"X-Tenant-ID": callerTenant,
			},
		},
		{
			name: "jwt shape",
			headers: map[string]string{
				"X-ID":        "caller-id",
				"X-Tenant-ID": callerTenant,
			},
		},
	}

	for _, route := range routes {
		for _, shape := range shapes {
			t.Run(route.description+" ("+shape.name+")", func(t *testing.T) {
				mock := mocks.NewMockService(t)

				// Seed the mock with a realistic victim namespace so the
				// handler would have something to serialize on vulnerable
				// code; the body-leak assertion below then catches any
				// regression that bypasses the tenant guard.
				victim := &models.Namespace{
					Name:     "victim",
					Owner:    victimOwner,
					TenantID: victimTenant,
					Members:  []models.Member{{ID: "victim-user", Email: victimEmail, Role: authorizer.RoleOwner}},
				}
				mock.On("GetNamespace", gomock.Anything, victimTenant).Return(victim, nil).Maybe()
				mock.On("EditNamespace", gomock.Anything, gomock.Anything).Return(victim, nil).Maybe()
				mock.On("DeleteNamespace", gomock.Anything, victimTenant).Return(nil).Maybe()
				mock.On("EditSessionRecordStatus", gomock.Anything, gomock.Anything, victimTenant).Return(nil).Maybe()

				var body io.Reader
				if route.body != "" {
					body = strings.NewReader(route.body)
				}
				req := httptest.NewRequest(route.method, route.url, body)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Role", authorizer.RoleOwner.String())
				for k, v := range shape.headers {
					req.Header.Set(k, v)
				}

				rec := httptest.NewRecorder()
				NewRouter(mock).ServeHTTP(rec, req)

				assert.Equal(t, http.StatusForbidden, rec.Result().StatusCode)
				// Even if the status assertion regresses, no sensitive field
				// from the victim namespace may leak into the response body.
				assert.NotContains(t, rec.Body.String(), victimEmail)
				assert.NotContains(t, rec.Body.String(), victimOwner)
			})
		}
	}
}

func TestGetNamespaceListBlocksAPIKey(t *testing.T) {
	mock := mocks.NewMockService(t)

	req := httptest.NewRequest(http.MethodGet, "/api/namespaces", nil)
	req.Header.Set("X-API-KEY", "b2f7cc0e-d933-4aad-9ab2-b557f2f2554f")
	req.Header.Set("X-Tenant-ID", "00000000-0000-4000-0000-000000000000")

	rec := httptest.NewRecorder()
	NewRouter(mock).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Result().StatusCode)
	mock.AssertExpectations(t)
}

func TestCreateNamespaceBlocksAPIKey(t *testing.T) {
	envstest.SetEdition(t, envs.Enterprise)

	mock := mocks.NewMockService(t)

	req := httptest.NewRequest(http.MethodPost, "/api/namespaces", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", "b2f7cc0e-d933-4aad-9ab2-b557f2f2554f")
	req.Header.Set("X-Tenant-ID", "00000000-0000-4000-0000-000000000000")

	rec := httptest.NewRecorder()
	NewRouter(mock).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Result().StatusCode)
	mock.AssertExpectations(t)
}

// encodeFilter serialises filters as the JSON array the API expects and
// returns it base64-encoded, ready to be used as the "filter" query param.
func encodeFilter(t *testing.T, filters []query.Filter) string {
	t.Helper()

	raw, err := json.Marshal(filters)
	if err != nil {
		t.Fatalf("encodeFilter: marshal: %v", err)
	}

	return base64.StdEncoding.EncodeToString(raw)
}

// namespaceListHasFilterName returns a testify MatchedBy matcher that
// verifies the first property filter in a *requests.NamespaceList carries the
// given field name un-rewritten, ensuring the handler passes the API-level
// field name to the service layer without translating it to a database column.
func namespaceListHasFilterName(name string) interface{} {
	return gomock.MatchedBy(func(req *requests.NamespaceList) bool {
		for _, f := range req.Filters.Data {
			if f.Type != query.FilterTypeProperty {
				continue
			}

			prop, ok := f.Params.(*query.FilterProperty)
			if !ok {
				return false
			}

			return prop.Name == name
		}

		return false
	})
}

func TestGetNamespaceList(t *testing.T) {
	svcMock := mocks.NewMockService(t)

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
			// Unknown field not present in NamespaceFilterFields must yield 400
			// before reaching the service layer.
			description: "fails with unknown filter field",
			query: "filter=" + encodeFilter(t, []query.Filter{
				{
					Type: query.FilterTypeProperty,
					Params: &query.FilterProperty{
						Name:     "unknown_field",
						Operator: "eq",
						Value:    "test",
					},
				},
			}),
			requiredMocks:  func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			// Operator "contains" is not allowed on the "type" field (enum), so
			// the handler must return 400 without calling the service.
			description: "fails with disallowed operator for type field",
			query: "filter=" + encodeFilter(t, []query.Filter{
				{
					Type: query.FilterTypeProperty,
					Params: &query.FilterProperty{
						Name:     "type",
						Operator: "contains",
						Value:    "test",
					},
				},
			}),
			requiredMocks:  func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			// A valid name+contains filter must reach the service and return 200
			// with the correct X-Total-Count header.
			description: "succeeds with valid name+contains filter",
			query: "filter=" + encodeFilter(t, []query.Filter{
				{
					Type: query.FilterTypeProperty,
					Params: &query.FilterProperty{
						Name:     "name",
						Operator: "contains",
						Value:    "test",
					},
				},
			}),
			requiredMocks: func() {
				svcMock.
					On("ListNamespaces", gomock.Anything, gomock.AnythingOfType("*requests.NamespaceList")).
					Return([]models.Namespace{}, 5, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedCount:  5,
		},
		{
			// A valid type+eq filter must reach the service with the field name
			// "type" un-rewritten (the store layer, not the handler, translates
			// "type" → "scope").
			description: "succeeds with valid type+eq filter and name reaches service un-rewritten",
			query: "filter=" + encodeFilter(t, []query.Filter{
				{
					Type: query.FilterTypeProperty,
					Params: &query.FilterProperty{
						Name:     "type",
						Operator: "eq",
						Value:    "personal",
					},
				},
			}),
			requiredMocks: func() {
				svcMock.
					On("ListNamespaces", gomock.Anything, namespaceListHasFilterName("type")).
					Return([]models.Namespace{}, 2, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			description: "succeeds and returns X-Total-Count header",
			query:       "",
			requiredMocks: func() {
				svcMock.
					On("ListNamespaces", gomock.Anything, gomock.AnythingOfType("*requests.NamespaceList")).
					Return([]models.Namespace{}, 3, nil).
					Once()
			},
			expectedStatus: http.StatusOK,
			expectedCount:  3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			url := "/api/namespaces"
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
