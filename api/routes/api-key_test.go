package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	servicemock "github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateAPIKey(t *testing.T) {
	type Expected struct {
		body   *responses.CreateAPIKey
		status int
	}

	svcMock := new(servicemock.Service)

	cases := []struct {
		description   string
		headers       map[string]string
		body          map[string]interface{}
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails with api key",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-API-KEY":    "b2f7cc0e-d933-4aad-9ab2-b557f2f2554f",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
			},
			body: map[string]interface{}{
				"name":       "dev",
				"expires_at": 30,
			},
			requiredMocks: func() {
			},
			expected: Expected{body: nil, status: http.StatusForbidden},
		},
		{
			description: "fails when role is observer",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-ID":         "000000000000000000000000",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			body: map[string]interface{}{
				"name":       "dev",
				"expires_at": 30,
			},
			requiredMocks: func() {
			},
			expected: Expected{body: nil, status: http.StatusForbidden},
		},
		{
			description: "fails when role is operator",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-ID":         "000000000000000000000000",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]interface{}{
				"name":       "dev",
				"expires_at": 30,
			},
			requiredMocks: func() {
			},
			expected: Expected{body: nil, status: http.StatusForbidden},
		},
		{
			description: "fails when name is invalid due to length < 3",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-ID":         "000000000000000000000000",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			body: map[string]interface{}{
				"name":       "de",
				"expires_at": 30,
			},
			requiredMocks: func() {
			},
			expected: Expected{body: nil, status: http.StatusBadRequest},
		},
		{
			description: "fails when name is invalid due to length > 20",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-ID":         "000000000000000000000000",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			body: map[string]interface{}{
				"name":       "devdevdevdevdevdevdev",
				"expires_at": 30,
			},
			requiredMocks: func() {
			},
			expected: Expected{body: nil, status: http.StatusBadRequest},
		},
		{
			description: "fails when name is invalid due to whitespaces",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-ID":         "000000000000000000000000",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			body: map[string]interface{}{
				"name":       "dev dev",
				"expires_at": 30,
			},
			requiredMocks: func() {
			},
			expected: Expected{body: nil, status: http.StatusBadRequest},
		},
		{
			description: "fails when expires_at is invalid",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-ID":         "000000000000000000000000",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			body: map[string]interface{}{
				"name":       "dev",
				"expires_at": 0,
			},
			requiredMocks: func() {
			},
			expected: Expected{body: nil, status: http.StatusBadRequest},
		},
		{
			description: "fails when key is provided but invalid",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-ID":         "000000000000000000000000",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			body: map[string]interface{}{
				"name":       "dev",
				"expires_at": 0,
				"key":        "invalid",
			},
			requiredMocks: func() {
			},
			expected: Expected{body: nil, status: http.StatusBadRequest},
		},
		{
			description: "fails when optional role is provided but invalid",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-ID":         "000000000000000000000000",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			body: map[string]interface{}{
				"name":       "dev",
				"expires_at": 0,
				"role":       "invalid",
			},
			requiredMocks: func() {
			},
			expected: Expected{body: nil, status: http.StatusBadRequest},
		},
		{
			description: "succeeds",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-ID":         "000000000000000000000000",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
			},
			body: map[string]interface{}{
				"name":       "dev",
				"expires_at": 30,
			},
			requiredMocks: func() {
				svcMock.On(
					"CreateAPIKey",
					mock.Anything,
					&requests.CreateAPIKey{
						UserID:    "000000000000000000000000",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Name:      "dev",
						Role:      "owner",
						ExpiresAt: 30,
					}).
					Return(&responses.CreateAPIKey{}, nil).
					Once()
			},
			expected: Expected{
				body:   &responses.CreateAPIKey{},
				status: http.StatusOK,
			},
		},
		{
			description: "succeeds with optional body",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-ID":         "000000000000000000000000",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
			},
			body: map[string]interface{}{
				"name":       "dev",
				"expires_at": 30,
				"key":        "3d7a3ea1-d1eb-4ffc-8c14-f7bfd1b7c550",
				"role":       "administrator",
			},
			requiredMocks: func() {
				svcMock.On(
					"CreateAPIKey",
					mock.Anything,
					&requests.CreateAPIKey{
						UserID:    "000000000000000000000000",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Name:      "dev",
						Role:      "owner",
						ExpiresAt: 30,
						Key:       "3d7a3ea1-d1eb-4ffc-8c14-f7bfd1b7c550",
						OptRole:   "administrator",
					}).
					Return(&responses.CreateAPIKey{}, nil).
					Once()
			},
			expected: Expected{
				body:   &responses.CreateAPIKey{},
				status: http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/namespaces/api-key", strings.NewReader(string(data)))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			e := NewRouter(svcMock, nil)
			e.ServeHTTP(rec, req)

			require.Equal(t, tc.expected.status, rec.Result().StatusCode)
			if tc.expected.body != nil {
				responseBody := new(responses.CreateAPIKey)
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&responseBody))
				require.Equal(t, tc.expected.body, responseBody)
			}
		})
	}
}

func TestListAPIKey(t *testing.T) {
	type Expected struct {
		body   []models.APIKey
		status int
	}

	svcMock := new(servicemock.Service)

	cases := []struct {
		description   string
		headers       map[string]string
		query         func() string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "success",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
			},
			query: func() string {
				url := &url.Values{}
				url.Add("page", "1")
				url.Add("per_page", "10")
				url.Add("sort_by", "created_at")
				url.Add("order_by", "asc")

				return url.Encode()
			},
			requiredMocks: func() {
				svcMock.On(
					"ListAPIKeys",
					mock.Anything,
					&requests.ListAPIKey{
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Paginator: query.Paginator{Page: 1, PerPage: 10},
						Sorter:    query.Sorter{By: "created_at", Order: "asc"},
					}).
					Return(
						[]models.APIKey{
							{
								ID:        "f23a2e56cd3fcfba002c72675c870e1e7813292adc40bbf14cea479a2e07976a",
								Name:      "dev",
								CreatedBy: "507f1f77bcf86cd799439011",
								TenantID:  "00000000-0000-4000-0000-000000000000",
								Role:      "admin",
								CreatedAt: time.Date(2023, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC),
								UpdatedAt: time.Date(2023, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC),
								ExpiresIn: 0,
							},
						},
						0,
						nil,
					).
					Once()
			},
			expected: Expected{
				body: []models.APIKey{
					{
						Name:      "dev",
						CreatedBy: "507f1f77bcf86cd799439011",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Role:      "admin",
						CreatedAt: time.Date(2023, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC),
						UpdatedAt: time.Date(2023, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC),
						ExpiresIn: 0,
					},
				},
				status: http.StatusOK,
			},
		},
		{
			description: "success when page and per_page are invalid",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
			},
			query: func() string {
				url := &url.Values{}
				url.Add("page", "-1")
				url.Add("per_page", "1000")
				url.Add("sort_by", "created_at")
				url.Add("order_by", "asc")

				return url.Encode()
			},
			requiredMocks: func() {
				svcMock.On(
					"ListAPIKeys",
					mock.Anything,
					&requests.ListAPIKey{
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Paginator: query.Paginator{Page: 1, PerPage: 100},
						Sorter:    query.Sorter{By: "created_at", Order: "asc"},
					}).
					Return(
						[]models.APIKey{
							{
								ID:        "f23a2e56cd3fcfba002c72675c870e1e7813292adc40bbf14cea479a2e07976a",
								Name:      "dev",
								CreatedBy: "507f1f77bcf86cd799439011",
								TenantID:  "00000000-0000-4000-0000-000000000000",
								Role:      "admin",
								CreatedAt: time.Date(2023, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC),
								UpdatedAt: time.Date(2023, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC),
								ExpiresIn: 0,
							},
						},
						0,
						nil,
					).
					Once()
			},
			expected: Expected{
				body: []models.APIKey{
					{
						Name:      "dev",
						CreatedBy: "507f1f77bcf86cd799439011",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Role:      "admin",
						CreatedAt: time.Date(2023, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC),
						UpdatedAt: time.Date(2023, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC),
						ExpiresIn: 0,
					},
				},
				status: http.StatusOK,
			},
		},
		{
			description: "success when order_by is an empty string",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
			},
			query: func() string {
				url := &url.Values{}
				url.Add("page", "1")
				url.Add("per_page", "10")
				url.Add("sort_by", "created_at")

				return url.Encode()
			},
			requiredMocks: func() {
				svcMock.On(
					"ListAPIKeys",
					mock.Anything,
					&requests.ListAPIKey{
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Paginator: query.Paginator{Page: 1, PerPage: 10},
						Sorter:    query.Sorter{By: "created_at", Order: "desc"},
					}).
					Return(
						[]models.APIKey{
							{
								Name:      "dev",
								CreatedBy: "507f1f77bcf86cd799439011",
								TenantID:  "00000000-0000-4000-0000-000000000000",
								Role:      "admin",
								CreatedAt: time.Date(2023, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC),
								UpdatedAt: time.Date(2023, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC),
								ExpiresIn: 0,
							},
						},
						0,
						nil,
					).
					Once()
			},
			expected: Expected{
				body: []models.APIKey{
					{
						Name:      "dev",
						CreatedBy: "507f1f77bcf86cd799439011",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Role:      "admin",
						CreatedAt: time.Date(2023, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC),
						UpdatedAt: time.Date(2023, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC),
						ExpiresIn: 0,
					},
				},
				status: http.StatusOK,
			},
		},
		{
			description: "success when sort_by is an empty string",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
			},
			query: func() string {
				url := &url.Values{}
				url.Add("page", "1")
				url.Add("per_page", "10")
				url.Add("order_by", "asc")

				return url.Encode()
			},
			requiredMocks: func() {
				svcMock.On(
					"ListAPIKeys",
					mock.Anything,
					&requests.ListAPIKey{
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Paginator: query.Paginator{Page: 1, PerPage: 10},
						Sorter:    query.Sorter{By: "expires_in", Order: "asc"},
					}).
					Return(
						[]models.APIKey{
							{
								ID:        "f23a2e56cd3fcfba002c72675c870e1e7813292adc40bbf14cea479a2e07976a",
								Name:      "dev",
								CreatedBy: "507f1f77bcf86cd799439011",
								TenantID:  "00000000-0000-4000-0000-000000000000",
								Role:      "admin",
								CreatedAt: time.Date(2023, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC),
								UpdatedAt: time.Date(2023, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC),
								ExpiresIn: 0,
							},
						},
						0,
						nil,
					).
					Once()
			},
			expected: Expected{
				body: []models.APIKey{
					{
						Name:      "dev",
						CreatedBy: "507f1f77bcf86cd799439011",
						TenantID:  "00000000-0000-4000-0000-000000000000",
						Role:      "admin",
						CreatedAt: time.Date(2023, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC),
						UpdatedAt: time.Date(2023, 0o1, 0o1, 12, 0o0, 0o0, 0o0, time.UTC),
						ExpiresIn: 0,
					},
				},
				status: http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodGet, "/api/namespaces/api-key?"+tc.query(), nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			e := NewRouter(svcMock, nil)
			e.ServeHTTP(rec, req)

			require.Equal(t, tc.expected.status, rec.Result().StatusCode)
			if tc.expected.body != nil {
				responseBody := make([]models.APIKey, 0)
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&responseBody))
				require.Equal(t, tc.expected.body, responseBody)
			}
		})
	}
}

func TestUpdateAPIKey(t *testing.T) {
	type Expected struct {
		status int
	}

	svcMock := new(servicemock.Service)

	cases := []struct {
		description   string
		name          string
		headers       map[string]string
		body          map[string]string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails with api key",
			name:        "dev",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-API-KEY":    "b2f7cc0e-d933-4aad-9ab2-b557f2f2554f",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
			},
			body: map[string]string{
				"name": "prod",
			},
			requiredMocks: func() {
			},
			expected: Expected{status: http.StatusForbidden},
		},
		{
			description: "fails when role is observer",
			name:        "dev",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			body: map[string]string{
				"name": "prod",
			},
			requiredMocks: func() {
			},
			expected: Expected{status: http.StatusForbidden},
		},
		{
			description: "fails when role is operator",
			name:        "dev",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			body: map[string]string{
				"name": "prod",
			},
			requiredMocks: func() {
			},
			expected: Expected{status: http.StatusForbidden},
		},
		{
			description: "fails when name is invalid due to length < 3",
			name:        "dev",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			body: map[string]string{
				"name": "pr",
			},
			requiredMocks: func() {
			},
			expected: Expected{status: http.StatusBadRequest},
		},
		{
			description: "fails when name is invalid due to length > 20",
			name:        "dev",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			body: map[string]string{
				"name": "prodprodprodprodprodprod",
			},
			requiredMocks: func() {
			},
			expected: Expected{status: http.StatusBadRequest},
		},
		{
			description: "fails when name is invalid due to whitespaces",
			name:        "dev",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			body: map[string]string{
				"name": "prod prod",
			},
			requiredMocks: func() {
			},
			expected: Expected{status: http.StatusBadRequest},
		},
		{
			description: "fails when role is invalid",
			name:        "dev",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			body: map[string]string{
				"role": "invalid",
			},
			requiredMocks: func() {
			},
			expected: Expected{status: http.StatusBadRequest},
		},
		{
			description: "succeeds",
			name:        "dev",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
			},
			body: map[string]string{
				"name": "prod",
				"role": "administrator",
			},
			requiredMocks: func() {
				svcMock.On(
					"UpdateAPIKey",
					mock.Anything,
					&requests.UpdateAPIKey{
						TenantID:    "00000000-0000-4000-0000-000000000000",
						CurrentName: "dev",
						Name:        "prod",
						Role:        "administrator",
					}).
					Return(nil).
					Once()
			},
			expected: Expected{status: http.StatusOK},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/api/namespaces/api-key/"+tc.name, strings.NewReader(string(data)))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			e := NewRouter(svcMock, nil)
			e.ServeHTTP(rec, req)

			require.Equal(t, tc.expected.status, rec.Result().StatusCode)
		})
	}
}

func TestDeleteAPIKey(t *testing.T) {
	type Expected struct {
		status int
	}

	svcMock := new(servicemock.Service)

	cases := []struct {
		description   string
		name          string
		headers       map[string]string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails with api key",
			name:        "dev",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-API-KEY":    "b2f7cc0e-d933-4aad-9ab2-b557f2f2554f",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
			},
			requiredMocks: func() {
			},
			expected: Expected{status: http.StatusForbidden},
		},
		{
			description: "fails when role is observer",
			name:        "dev",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
			},
			requiredMocks: func() {
			},
			expected: Expected{status: http.StatusForbidden},
		},
		{
			description: "fails when role is operator",
			name:        "dev",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
			},
			requiredMocks: func() {
			},
			expected: Expected{status: http.StatusForbidden},
		},
		{
			description: "succeeds",
			name:        "dev",
			headers: map[string]string{
				"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
				"X-Role":      "owner",
			},
			requiredMocks: func() {
				svcMock.On(
					"DeleteAPIKey",
					mock.Anything,
					&requests.DeleteAPIKey{
						TenantID: "00000000-0000-4000-0000-000000000000",
						Name:     "dev",
					}).
					Return(nil).
					Once()
			},
			expected: Expected{status: http.StatusOK},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodDelete, "/api/namespaces/api-key/"+tc.name, nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			e := NewRouter(svcMock, nil)
			e.ServeHTTP(rec, req)

			require.Equal(t, tc.expected.status, rec.Result().StatusCode)
		})
	}
}
