package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetTags(t *testing.T) {
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

func TestDeleteTag(t *testing.T) {
	// Only observer lacks tag permissions; operator, administrator, and owner all hold them — do not test those as forbidden.
	cases := []struct {
		description    string
		url            string
		headers        map[string]string
		requiredMocks  func(svcMock *mocks.MockService)
		expectedStatus int
	}{
		{
			description: "fails when role is observer (new URL)",
			url:         "/api/tags/production",
			headers: map[string]string{
				"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
				"X-Role":      authorizer.RoleObserver.String(),
				"X-ID":        "000000000000000000000000",
			},
			requiredMocks:  func(_ *mocks.MockService) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			description: "fails when role is observer (legacy URL)",
			url:         "/api/namespaces/00000000-0000-4000-0000-000000000000/tags/production",
			headers: map[string]string{
				"X-Role": authorizer.RoleObserver.String(),
				"X-ID":   "000000000000000000000000",
			},
			requiredMocks:  func(_ *mocks.MockService) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			description: "fails when service returns ErrNamespaceNotFound",
			url:         "/api/tags/production",
			headers: map[string]string{
				"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
				"X-Role":      authorizer.RoleOwner.String(),
				"X-ID":        "000000000000000000000000",
			},
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
			description: "succeeds with 200 and no body (new URL)",
			url:         "/api/tags/production",
			headers: map[string]string{
				"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
				"X-Role":      authorizer.RoleOwner.String(),
				"X-ID":        "000000000000000000000000",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("DeleteTag", gomock.Anything, &requests.DeleteTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
					}).
					Return(nil).
					Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			description: "succeeds via legacy URL (tenant from path param)",
			url:         "/api/namespaces/00000000-0000-4000-0000-000000000000/tags/production",
			headers: map[string]string{
				"X-Role": authorizer.RoleOwner.String(),
				"X-ID":   "000000000000000000000000",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("DeleteTag", gomock.Anything, &requests.DeleteTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
					}).
					Return(nil).
					Once()
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := mocks.NewMockService(t)
			tc.requiredMocks(svcMock)

			req := httptest.NewRequest(http.MethodDelete, tc.url, nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(svcMock)
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
		conflicts  []string
	}

	// Only observer lacks tag permissions; operator, administrator, and owner all hold them — do not test those as forbidden.
	cases := []struct {
		description   string
		url           string
		headers       map[string]string
		body          map[string]interface{}
		requiredMocks func(svcMock *mocks.MockService)
		expected      Expected
	}{
		{
			description: "fails when role is observer (new URL)",
			url:         "/api/tags",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       authorizer.RoleObserver.String(),
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"name": "production",
			},
			requiredMocks: func(_ *mocks.MockService) {},
			expected:      Expected{status: http.StatusForbidden},
		},
		{
			description: "fails when role is observer (legacy URL)",
			url:         "/api/namespaces/00000000-0000-4000-0000-000000000000/tags",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       authorizer.RoleObserver.String(),
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"name": "production",
			},
			requiredMocks: func(_ *mocks.MockService) {},
			expected:      Expected{status: http.StatusForbidden},
		},
		{
			description: "fails when name is too short (< 3 chars)",
			url:         "/api/tags",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       authorizer.RoleOwner.String(),
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"name": "a",
			},
			requiredMocks: func(_ *mocks.MockService) {},
			expected:      Expected{status: http.StatusBadRequest},
		},
		{
			description: "fails when service returns ErrNamespaceNotFound",
			url:         "/api/tags",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       authorizer.RoleOwner.String(),
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"name": "production",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("CreateTag", gomock.Anything, &requests.CreateTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
					}).
					Return("", []string{}, svc.NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", nil)).
					Once()
			},
			expected: Expected{status: http.StatusNotFound},
		},
		{
			description: "succeeds with 200 and X-Inserted-ID header (new URL)",
			url:         "/api/tags",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       authorizer.RoleOwner.String(),
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"name": "production",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("CreateTag", gomock.Anything, &requests.CreateTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
					}).
					Return("000000000000000000000001", []string{}, nil).
					Once()
			},
			expected: Expected{
				status:     http.StatusOK,
				insertedID: "000000000000000000000001",
			},
		},
		{
			description: "returns 409 with conflicts JSON body",
			url:         "/api/tags",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       authorizer.RoleOwner.String(),
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"name": "production",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("CreateTag", gomock.Anything, &requests.CreateTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
					}).
					Return("", []string{"name"}, nil).
					Once()
			},
			expected: Expected{
				status:    http.StatusConflict,
				conflicts: []string{"name"},
			},
		},
		{
			description: "succeeds via legacy URL (tenant from path param)",
			url:         "/api/namespaces/00000000-0000-4000-0000-000000000000/tags",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       authorizer.RoleOwner.String(),
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"name": "production",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("CreateTag", gomock.Anything, &requests.CreateTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
					}).
					Return("000000000000000000000002", []string{}, nil).
					Once()
			},
			expected: Expected{
				status:     http.StatusOK,
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

			req := httptest.NewRequest(http.MethodPost, tc.url, strings.NewReader(string(data)))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			res := rec.Result()
			assert.Equal(t, tc.expected.status, res.StatusCode)

			if tc.expected.insertedID != "" {
				assert.Equal(t, tc.expected.insertedID, res.Header.Get("X-Inserted-ID"))
			}

			if tc.expected.conflicts != nil {
				var body map[string][]string
				require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
				assert.Equal(t, tc.expected.conflicts, body["conflicts"])
			}

			svcMock.AssertExpectations(t)
		})
	}
}

func TestUpdateTag(t *testing.T) {
	type Expected struct {
		status    int
		conflicts []string
	}

	// Only observer lacks tag permissions; operator, administrator, and owner all hold them — do not test those as forbidden.
	cases := []struct {
		description   string
		url           string
		headers       map[string]string
		body          map[string]interface{}
		requiredMocks func(svcMock *mocks.MockService)
		expected      Expected
	}{
		{
			description: "fails when role is observer (new URL)",
			url:         "/api/tags/production",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       authorizer.RoleObserver.String(),
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"name": "staging",
			},
			requiredMocks: func(_ *mocks.MockService) {},
			expected:      Expected{status: http.StatusForbidden},
		},
		{
			description: "fails when role is observer (legacy URL)",
			url:         "/api/namespaces/00000000-0000-4000-0000-000000000000/tags/production",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       authorizer.RoleObserver.String(),
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"name": "staging",
			},
			requiredMocks: func(_ *mocks.MockService) {},
			expected:      Expected{status: http.StatusForbidden},
		},
		{
			description: "fails when NewName is too short (< 3 chars)",
			url:         "/api/tags/production",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       authorizer.RoleOwner.String(),
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"name": "ab",
			},
			requiredMocks: func(_ *mocks.MockService) {},
			expected:      Expected{status: http.StatusBadRequest},
		},
		{
			description: "fails when service returns ErrNamespaceNotFound",
			url:         "/api/tags/production",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       authorizer.RoleOwner.String(),
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"name": "staging",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("UpdateTag", gomock.Anything, &requests.UpdateTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
						NewName:  "staging",
					}).
					Return([]string{}, svc.NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", nil)).
					Once()
			},
			expected: Expected{status: http.StatusNotFound},
		},
		{
			description: "succeeds with 200 and no body (new URL)",
			url:         "/api/tags/production",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       authorizer.RoleOwner.String(),
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"name": "staging",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("UpdateTag", gomock.Anything, &requests.UpdateTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
						NewName:  "staging",
					}).
					Return([]string{}, nil).
					Once()
			},
			expected: Expected{status: http.StatusOK},
		},
		{
			description: "succeeds via legacy URL (tenant from path param)",
			url:         "/api/namespaces/00000000-0000-4000-0000-000000000000/tags/production",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       authorizer.RoleOwner.String(),
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"name": "staging",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("UpdateTag", gomock.Anything, &requests.UpdateTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
						NewName:  "staging",
					}).
					Return([]string{}, nil).
					Once()
			},
			expected: Expected{status: http.StatusOK},
		},
		{
			description: "returns 409 with conflicts JSON body",
			url:         "/api/tags/production",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       authorizer.RoleOwner.String(),
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"name": "staging",
			},
			requiredMocks: func(svcMock *mocks.MockService) {
				svcMock.
					On("UpdateTag", gomock.Anything, &requests.UpdateTag{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "production",
						NewName:  "staging",
					}).
					Return([]string{"name"}, nil).
					Once()
			},
			expected: Expected{
				status:    http.StatusConflict,
				conflicts: []string{"name"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := mocks.NewMockService(t)
			tc.requiredMocks(svcMock)

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, tc.url, strings.NewReader(string(data)))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			res := rec.Result()
			assert.Equal(t, tc.expected.status, res.StatusCode)

			if tc.expected.conflicts != nil {
				var body map[string][]string
				require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
				assert.Equal(t, tc.expected.conflicts, body["conflicts"])
			}

			svcMock.AssertExpectations(t)
		})
	}
}

func TestPushTagToDevice(t *testing.T) {
	const deviceUID = "aabbccddee00112233445566778899aabbccddee"

	// Only observer lacks tag permissions; operator, administrator, and owner all hold them — do not test those as forbidden.
	cases := []struct {
		description    string
		url            string
		headers        map[string]string
		requiredMocks  func(svcMock *mocks.MockService)
		expectedStatus int
	}{
		{
			description: "fails when role is observer (new URL)",
			url:         "/api/devices/" + deviceUID + "/tags/production",
			headers: map[string]string{
				"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
				"X-Role":      authorizer.RoleObserver.String(),
				"X-ID":        "000000000000000000000000",
			},
			requiredMocks:  func(_ *mocks.MockService) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			description: "fails when role is observer (legacy URL)",
			url:         "/api/namespaces/00000000-0000-4000-0000-000000000000/devices/" + deviceUID + "/tags/production",
			headers: map[string]string{
				"X-Role": authorizer.RoleObserver.String(),
				"X-ID":   "000000000000000000000000",
			},
			requiredMocks:  func(_ *mocks.MockService) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			description: "fails when tag name is too short (< 3 chars)",
			url:         "/api/devices/" + deviceUID + "/tags/ab",
			headers: map[string]string{
				"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
				"X-Role":      authorizer.RoleOwner.String(),
				"X-ID":        "000000000000000000000000",
			},
			requiredMocks:  func(_ *mocks.MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "fails when service returns ErrNamespaceNotFound",
			url:         "/api/devices/" + deviceUID + "/tags/production",
			headers: map[string]string{
				"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
				"X-Role":      authorizer.RoleOwner.String(),
				"X-ID":        "000000000000000000000000",
			},
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
			description: "succeeds with 200 and no body (new URL)",
			url:         "/api/devices/" + deviceUID + "/tags/production",
			headers: map[string]string{
				"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
				"X-Role":      authorizer.RoleOwner.String(),
				"X-ID":        "000000000000000000000000",
			},
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
			expectedStatus: http.StatusOK,
		},
		{
			description: "succeeds via legacy URL (tenant from path param)",
			url:         "/api/namespaces/00000000-0000-4000-0000-000000000000/devices/" + deviceUID + "/tags/production",
			headers: map[string]string{
				"X-Role": authorizer.RoleOwner.String(),
				"X-ID":   "000000000000000000000000",
			},
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
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := mocks.NewMockService(t)
			tc.requiredMocks(svcMock)

			req := httptest.NewRequest(http.MethodPost, tc.url, nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)

			svcMock.AssertExpectations(t)
		})
	}
}

func TestPullTagFromDevice(t *testing.T) {
	const deviceUID = "aabbccddee00112233445566778899aabbccddee"

	// Only observer lacks tag permissions; operator, administrator, and owner all hold them — do not test those as forbidden.
	cases := []struct {
		description    string
		url            string
		headers        map[string]string
		requiredMocks  func(svcMock *mocks.MockService)
		expectedStatus int
	}{
		{
			description: "fails when role is observer (new URL)",
			url:         "/api/devices/" + deviceUID + "/tags/production",
			headers: map[string]string{
				"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
				"X-Role":      authorizer.RoleObserver.String(),
				"X-ID":        "000000000000000000000000",
			},
			requiredMocks:  func(_ *mocks.MockService) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			description: "fails when role is observer (legacy URL)",
			url:         "/api/namespaces/00000000-0000-4000-0000-000000000000/devices/" + deviceUID + "/tags/production",
			headers: map[string]string{
				"X-Role": authorizer.RoleObserver.String(),
				"X-ID":   "000000000000000000000000",
			},
			requiredMocks:  func(_ *mocks.MockService) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			description: "fails when tag name is too short (< 3 chars)",
			url:         "/api/devices/" + deviceUID + "/tags/ab",
			headers: map[string]string{
				"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
				"X-Role":      authorizer.RoleOwner.String(),
				"X-ID":        "000000000000000000000000",
			},
			requiredMocks:  func(_ *mocks.MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "fails when service returns ErrNamespaceNotFound",
			url:         "/api/devices/" + deviceUID + "/tags/production",
			headers: map[string]string{
				"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
				"X-Role":      authorizer.RoleOwner.String(),
				"X-ID":        "000000000000000000000000",
			},
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
			description: "succeeds with 200 and no body (new URL)",
			url:         "/api/devices/" + deviceUID + "/tags/production",
			headers: map[string]string{
				"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
				"X-Role":      authorizer.RoleOwner.String(),
				"X-ID":        "000000000000000000000000",
			},
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
			expectedStatus: http.StatusOK,
		},
		{
			description: "succeeds via legacy URL (tenant from path param)",
			url:         "/api/namespaces/00000000-0000-4000-0000-000000000000/devices/" + deviceUID + "/tags/production",
			headers: map[string]string{
				"X-Role": authorizer.RoleOwner.String(),
				"X-ID":   "000000000000000000000000",
			},
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
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			svcMock := mocks.NewMockService(t)
			tc.requiredMocks(svcMock)

			req := httptest.NewRequest(http.MethodDelete, tc.url, nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)

			svcMock.AssertExpectations(t)
		})
	}
}
