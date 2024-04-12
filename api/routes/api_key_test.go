package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestCreateAPIKey(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedSession string
		expectedStatus  int
	}
	cases := []struct {
		title         string
		tenantID      string
		key           string
		id            string
		requestBody   *requests.CreateAPIKey
		requiredMocks func()
		expected      Expected
	}{
		{
			title:    "success when trying to create a valid APIKey",
			key:      "key",
			tenantID: "00000000-0000-4000-0000-000000000000",
			id:       "id",
			requestBody: &requests.CreateAPIKey{
				Name:      "nameAPIKey",
				ExpiresAt: 30,
				TenantParam: requests.TenantParam{
					Tenant: "00000000-0000-4000-0000-000000000000",
				},
			},
			requiredMocks: func() {
				req := &requests.CreateAPIKey{
					Name:      "nameAPIKey",
					ExpiresAt: 30,
					TenantParam: requests.TenantParam{
						Tenant: "00000000-0000-4000-0000-000000000000",
					},
				}
				mock.On("CreateAPIKey", gomock.Anything, "id", "", "", "owner", req).Return("APIKey", nil).Once()
			},
			expected: Expected{
				expectedSession: "APIKey",
				expectedStatus:  http.StatusOK,
			},
		},
		{
			title:         "failure when request body is nil",
			id:            "id",
			tenantID:      "00000000-0000-4000-0000-000000000000",
			requestBody:   nil,
			requiredMocks: func() {},
			expected: Expected{
				expectedStatus: http.StatusBadRequest,
			},
		},
		{
			title:    "failure when validation fails",
			id:       "id",
			tenantID: "00000000-0000-4000-0000-000000000000",
			requestBody: &requests.CreateAPIKey{
				Name:      "",
				ExpiresAt: 30,
			},
			requiredMocks: func() {},
			expected: Expected{
				expectedStatus: http.StatusBadRequest,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			jsonData, err := json.Marshal(tc.requestBody)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/namespaces/%s/api-key", tc.tenantID), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			c := gateway.NewContext(mock, e.NewContext(req, rec))
			c.Request().Header.Set("X-ID", "id")
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestListAPIKey(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedResponse []models.APIKey
		expectedStatus   int
	}
	cases := []struct {
		title         string
		requestParams *requests.APIKeyList
		requiredMocks func()
		expected      Expected
	}{
		{
			title: "success when trying to get a  APIKey",
			requestParams: &requests.APIKeyList{
				TenantParam: requests.TenantParam{Tenant: "00000000-0000-4000-0000-000000000000"},
			},
			requiredMocks: func() {
				mock.On("ListAPIKeys", gomock.Anything, gomock.Anything).Return([]models.APIKey{}, 0, nil).Once()
			},
			expected: Expected{
				expectedResponse: []models.APIKey{},
				expectedStatus:   http.StatusOK,
			},
		},
		{
			title: "failure when request body is nil",
			requestParams: &requests.APIKeyList{
				TenantParam: requests.TenantParam{Tenant: ""},
			},
			requiredMocks: func() {},
			expected: Expected{
				expectedStatus: http.StatusBadRequest,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			jsonData, err := json.Marshal(tc.requestParams)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/namespaces/%s/api-key", tc.requestParams.Tenant), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			c := gateway.NewContext(mock, e.NewContext(req, rec))
			c.Request().Header.Set("X-ID", "id")
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestDeleteAPIKey(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		requestParams  *requests.APIKeyID
		tenantID       string
		requiredMocks  func()
		expectedStatus int
	}{
		{
			title: "failure when id is nil",
			requestParams: &requests.APIKeyID{
				ID: "",
			},
			tenantID:       "00000000-0000-4000-0000-000000000000",
			requiredMocks:  func() {},
			expectedStatus: http.StatusNotFound,
		},
		{
			title:    "success",
			tenantID: "00000000-0000-4000-0000-000000000000",
			requestParams: &requests.APIKeyID{
				ID: "id",
			},
			requiredMocks: func() {
				mock.On("DeleteAPIKey", gomock.Anything, "id", "").Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			jsonData, err := json.Marshal(tc.requestParams)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/namespaces/%s/api-key/%s", tc.tenantID, tc.requestParams.ID), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			c := gateway.NewContext(mock, e.NewContext(req, rec))
			c.Request().Header.Set("X-ID", "id")
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestEditAPIKey(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedSession *models.APIKey
		expectedStatus  int
	}
	cases := []struct {
		title         string
		requestParams *requests.APIKeyChanges
		tenantID      string
		requiredMocks func()
		expected      Expected
	}{
		{
			title: "failure when request body is nil",
			requestParams: &requests.APIKeyChanges{
				ID:   "",
				Name: "",
			},
			tenantID:      "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title:    "success",
			tenantID: "00000000-0000-4000-0000-000000000000",
			requestParams: &requests.APIKeyChanges{
				ID:   "id",
				Name: "newName",
			},
			requiredMocks: func() {
				mock.On("EditAPIKey", gomock.Anything, gomock.Anything, gomock.Anything).Return(&models.APIKey{}, nil).Once()
			},
			expected: Expected{
				expectedSession: &models.APIKey{},
				expectedStatus:  http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			jsonData, err := json.Marshal(tc.requestParams)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/namespaces/%s/api-key/%s", tc.tenantID, tc.requestParams.ID), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			c := gateway.NewContext(mock, e.NewContext(req, rec))
			c.Request().Header.Set("X-ID", "id")
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)
		})
	}
}
