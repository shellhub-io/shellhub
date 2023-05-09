package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestGetPublicKeys(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title           string
		query           paginator.Query
		requiredMocks   func(query paginator.Query)
		expectedSession []models.PublicKey
		expectedStatus  int
	}{
		{
			title: "returns Ok if a publics keys exists",
			query: paginator.Query{
				Page:    1,
				PerPage: 10,
			},
			requiredMocks: func(query paginator.Query) {
				mock.On("ListPublicKeys", gomock.Anything, query).Return([]models.PublicKey{}, 1, nil)
			},
			expectedSession: []models.PublicKey{},
			expectedStatus:  http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.query)

			jsonData, err := json.Marshal(tc.query)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/sshkeys/public-keys", strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)

			var session []models.PublicKey
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}
			assert.Equal(t, tc.expectedSession, session)
		})
	}
}

func TestGetPublicKey(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title           string
		query           requests.PublicKeyGet
		requiredMocks   func(query requests.PublicKeyGet)
		expectedSession *models.PublicKey
		expectedStatus  int
	}{
		{
			title: "returns Ok if a public key exists",
			query: requests.PublicKeyGet{
				FingerprintParam: requests.FingerprintParam{Fingerprint: "figertest"},
				TenantParam:      requests.TenantParam{Tenant: "tenant"},
			},
			requiredMocks: func(query requests.PublicKeyGet) {
				mock.On("GetPublicKey", gomock.Anything, query.Fingerprint, query.Tenant).Return(&models.PublicKey{}, nil)
			},
			expectedSession: &models.PublicKey{},
			expectedStatus:  http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.query)

			jsonData, err := json.Marshal(tc.query)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/internal/sshkeys/public-keys/%s/%s", tc.query.Fingerprint, tc.query.Tenant), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)

			var session *models.PublicKey
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}
			assert.Equal(t, tc.expectedSession, session)
		})
	}
}

func TestDeletePublicKey(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		query          requests.PublicKeyDelete
		requiredMocks  func(query requests.PublicKeyDelete)
		expectedStatus int
	}{
		{
			title: "returns Ok when deleting an existing public key",
			query: requests.PublicKeyDelete{
				FingerprintParam: requests.FingerprintParam{Fingerprint: "figertest"},
			},
			requiredMocks: func(query requests.PublicKeyDelete) {
				mock.On("DeletePublicKey", gomock.Anything, query.Fingerprint, "tenant").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.query)

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/sshkeys/public-keys/%s", tc.query.Fingerprint), nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-Tenant-ID", "tenant")
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestRemovePublicKeyTag(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		query          requests.PublicKeyTagRemove
		tenant         string
		requiredMocks  func(query requests.PublicKeyTagRemove)
		expectedStatus int
	}{
		{
			title: "returns Ok when removing an existing public key",
			query: requests.PublicKeyTagRemove{
				FingerprintParam: requests.FingerprintParam{Fingerprint: "figertest"},
				TagParam:         requests.TagParam{Tag: "tag"},
			},
			tenant: "tenant-id",
			requiredMocks: func(query requests.PublicKeyTagRemove) {
				mock.On("RemovePublicKeyTag", gomock.Anything, "tenant-id", query.Fingerprint, query.Tag).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.query)

			jsonData, err := json.Marshal(tc.query)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/sshkeys/public-keys/%s/tags/%s", tc.query.Fingerprint, tc.query.Tag), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-Tenant-ID", tc.tenant)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestAddPublicKeyTagURL(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		query          requests.PublicKeyTagAdd
		tenant         string
		requiredMocks  func(query requests.PublicKeyTagAdd)
		expectedStatus int
	}{
		{
			title: "returns Ok when add an existing public tag key",
			query: requests.PublicKeyTagAdd{
				FingerprintParam: requests.FingerprintParam{Fingerprint: "figertest"},
				TagParam:         requests.TagParam{Tag: "tag"},
			},
			tenant: "tenant-id",
			requiredMocks: func(query requests.PublicKeyTagAdd) {
				mock.On("AddPublicKeyTag", gomock.Anything, "tenant-id", query.Fingerprint, query.Tag).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.query)

			jsonData, err := json.Marshal(tc.query)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/sshkeys/public-keys/%s/tags", tc.query.Fingerprint), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-Tenant-ID", tc.tenant)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestCreatePrivateKey(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		name           string
		requiredMocks  func()
		expectedStatus int
	}{
		{
			name: "returns Ok when creating an existing private key",
			requiredMocks: func() {
				mock.On("CreatePrivateKey", gomock.Anything).Return(&models.PrivateKey{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodPost, "/internal/sshkeys/private-keys", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}
