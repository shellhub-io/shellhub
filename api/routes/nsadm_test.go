package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestCreateNamespace(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		name            string
		uid             string
		req             requests.NamespaceCreate
		requiredMocks   func(req requests.NamespaceCreate)
		expectedStatus  int
		expectedSession *models.Namespace
	}{
		{
			name: "returns Ok when creating a namespace",
			uid:  "123",
			req: requests.NamespaceCreate{
				Name:     "example",
				TenantID: "tenant-id",
			},
			requiredMocks: func(req requests.NamespaceCreate) {
				mock.On("CreateNamespace", gomock.Anything, req, "123").Return(&models.Namespace{}, nil)
			},
			expectedStatus:  http.StatusOK,
			expectedSession: &models.Namespace{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks(tc.req)

			jsonData, err := json.Marshal(tc.req)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/namespaces", strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-ID", "123")
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)

			var session models.Namespace
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}
			assert.Equal(t, tc.expectedSession, &session)
		})
	}

	mock.AssertExpectations(t)
}

func TestGetNamespace(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		name            string
		uid             string
		req             string
		requiredMocks   func()
		expectedStatus  int
		expectedSession *models.Namespace
	}{
		{
			name: "returns Ok for a existing namespace",
			uid:  "123",
			req:  "tenant",
			requiredMocks: func() {
				mock.On("GetNamespace", gomock.Anything, "tenant").Return(&models.Namespace{}, nil)
			},
			expectedStatus:  http.StatusOK,
			expectedSession: &models.Namespace{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/namespaces/%s", tc.req), nil)

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)

			var session *models.Namespace
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}
			assert.Equal(t, tc.expectedSession, session)
		})
	}

	mock.AssertExpectations(t)
}

func TestDeleteNamespace(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		name           string
		uid            string
		req            requests.NamespaceDelete
		requiredMocks  func()
		expectedStatus int
	}{
		{
			name: "returns Ok when deleting a existing namespace",
			uid:  "123",
			req: requests.NamespaceDelete{
				TenantParam: requests.TenantParam{Tenant: "tenant-id"},
			},
			requiredMocks: func() {
				mock.On("GetNamespace", gomock.Anything, "tenant-id").Return(&models.Namespace{
					Name:     "namespace-name",
					Owner:    "owner-name",
					TenantID: "tenant-id",
					Members: []models.Member{
						{ID: "123", Username: "userexemple", Role: "owner"},
					},
					Settings:     &models.NamespaceSettings{},
					Devices:      10,
					Sessions:     5,
					MaxDevices:   100,
					DevicesCount: 50,
					CreatedAt:    time.Now(),
					Billing:      &models.Billing{},
				}, nil)

				mock.On("DeleteNamespace", gomock.Anything, "tenant-id").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()

			jsonData, err := json.Marshal(tc.req)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/namespaces/%s", tc.req.Tenant), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-ID", tc.uid)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}

	mock.AssertExpectations(t)
}

func TestGetSessionRecord(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		name           string
		tenant         string
		requiredMocks  func()
		expectedStatus int
	}{
		{
			name:   "returns Ok for session record of a existing session",
			tenant: "tenant",
			requiredMocks: func() {
				mock.On("GetSessionRecord", gomock.Anything, "tenant").Return(true, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodGet, "/api/users/security", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-Tenant-ID", tc.tenant)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}

	mock.AssertExpectations(t)
}

func TestEditNamespace(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		name           string
		uid            string
		req            requests.SessionEditRecordStatus
		requiredMocks  func(req requests.SessionEditRecordStatus)
		expectedStatus int
	}{
		{
			name: "returns OK for editing an existing namespace",
			uid:  "123",
			req: requests.SessionEditRecordStatus{
				SessionRecord: true,
				TenantParam:   requests.TenantParam{Tenant: "tenant-id"},
			},
			requiredMocks: func(req requests.SessionEditRecordStatus) {
				mock.On("GetNamespace", gomock.Anything, req.Tenant).Return(&models.Namespace{
					Name:     "namespace-name",
					Owner:    "owner-name",
					TenantID: req.Tenant,
					Members: []models.Member{
						{ID: "123", Username: "userexemple", Role: "owner"},
					},
					Settings:     &models.NamespaceSettings{},
					Devices:      10,
					Sessions:     5,
					MaxDevices:   100,
					DevicesCount: 50,
					CreatedAt:    time.Now(),
					Billing:      &models.Billing{},
				}, nil)

				mock.On("EditSessionRecordStatus", gomock.Anything, req.SessionRecord, req.Tenant).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks(tc.req)

			jsonData, err := json.Marshal(tc.req)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/security/%s", tc.req.Tenant), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-ID", tc.uid)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}

	mock.AssertExpectations(t)
}
