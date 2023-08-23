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
	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestCreateNamespace(t *testing.T) {
	mock := new(mocks.Service)

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
			title: "fails when try to creating a namespace",
			uid:   "123",
			req:   `{ "name": "example", "tenant": "tenant"}`,
			requiredMocks: func() {
				mock.On("CreateNamespace", gomock.Anything, gomock.AnythingOfType("requests.NamespaceCreate"), "123").Return(nil, svc.ErrNotFound).Once()
			},
			expected: Expected{
				expectedStatus:  http.StatusNotFound,
				expectedSession: &models.Namespace{},
			},
		},
		{
			title: "success when try to creating a namespace",
			uid:   "123",
			req:   `{ "name": "example", "tenant": "tenant"}`,
			requiredMocks: func() {
				mock.On("CreateNamespace", gomock.Anything, gomock.AnythingOfType("requests.NamespaceCreate"), "123").Return(&models.Namespace{}, nil).Once()
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
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-ID", "123")
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
	mock := new(mocks.Service)

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
			req:   "tenant",
			requiredMocks: func() {
				mock.On("GetNamespace", gomock.Anything, "tenant").Return(&models.Namespace{}, nil)
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
			req.Header.Set("X-Role", guard.RoleOwner)
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
	mock := new(mocks.Service)

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
			req:   "tenant-id",
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
				}, nil).Once()

				mock.On("DeleteNamespace", gomock.Anything, "tenant-id").Return(svc.ErrNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to deleting a existing namespace",
			uid:   "123",
			req:   "tenant-id",
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
				}, nil).Once()

				mock.On("DeleteNamespace", gomock.Anything, "tenant-id").Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/namespaces/%s", tc.req), nil)
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
			name:   "fails when try to get a session record of a non-existing session",
			tenant: "tenant",
			requiredMocks: func() {
				mock.On("GetSessionRecord", gomock.Anything, "tenant").Return(false, svc.ErrNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "success when try to get a  session record of a existing session",
			tenant: "tenant",
			requiredMocks: func() {
				mock.On("GetSessionRecord", gomock.Anything, "tenant").Return(true, nil).Once()
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
		title          string
		uid            string
		req            string
		requiredMocks  func()
		expectedStatus int
	}{
		{
			title:          "fails when bind fails to validate uid",
			uid:            "123",
			req:            `{"session_record": true, "tenant": ""}`,
			expectedStatus: http.StatusNotFound,
			requiredMocks:  func() {},
		},
		{
			title:          "fails when validate because the tenant does not have a min of 3 characters",
			uid:            "123",
			req:            `{"session_record": true, "tenant": "id"}`,
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func() {},
		},
		{
			title:          "fails when validate because the tenant does not have a max of 255 characters",
			uid:            "123",
			req:            `{"session_record": true, "tenant": "BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9"}`,
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func() {},
		},
		{
			title: "fails when try to editing an non-existing namespace",
			uid:   "123",
			req:   `{"session_record": true, "tenant": "tenant-id"}`,
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
				}, nil).Once()

				mock.On("EditSessionRecordStatus", gomock.Anything, true, "tenant-id").Return(svc.ErrNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to editing an existing namespace",
			uid:   "123",
			req:   `{"session_record": true, "tenant": "tenant-id"}`,
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
				}, nil).Once()

				mock.On("EditSessionRecordStatus", gomock.Anything, true, "tenant-id").Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			var data requests.SessionEditRecordStatus
			err := json.Unmarshal([]byte(tc.req), &data)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/security/%s", data.Tenant), strings.NewReader(tc.req))
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
