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
	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestGetPublicKeys(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedSession []models.PublicKey
		expectedStatus  int
	}
	cases := []struct {
		title         string
		query         paginator.Query
		requiredMocks func(query paginator.Query)
		expected      Expected
	}{
		{
			title: "success when try to list a publics keys exists",
			query: paginator.Query{
				Page:    1,
				PerPage: 10,
			},
			requiredMocks: func(query paginator.Query) {
				mock.On("ListPublicKeys", gomock.Anything, query).Return([]models.PublicKey{}, 1, nil)
			},
			expected: Expected{
				expectedSession: []models.PublicKey{},
				expectedStatus:  http.StatusOK,
			},
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

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session []models.PublicKey
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}
			assert.Equal(t, tc.expected.expectedSession, session)
		})
	}
}

func TestGetPublicKey(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedSession *models.PublicKey
		expectedStatus  int
	}
	cases := []struct {
		title         string
		query         requests.PublicKeyGet
		requiredMocks func(query requests.PublicKeyGet)
		expected      Expected
	}{
		{
			title: "fails when validate because the tag does not have a min of 3 characters",
			query: requests.PublicKeyGet{
				TenantParam: requests.TenantParam{Tenant: "tg"},
			},
			expected:      Expected{expectedStatus: http.StatusBadRequest},
			requiredMocks: func(req requests.PublicKeyGet) {},
		},
		{
			title: "fails when validate because the tag does not have a max of 255 characters",
			query: requests.PublicKeyGet{
				TenantParam: requests.TenantParam{Tenant: "BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9"},
			},
			expected:      Expected{expectedStatus: http.StatusBadRequest},
			requiredMocks: func(req requests.PublicKeyGet) {},
		},
		{
			title: "fails when validate because have a '/' with in your characters",
			query: requests.PublicKeyGet{
				TenantParam: requests.TenantParam{Tenant: "test/"},
			},
			expected:      Expected{expectedStatus: http.StatusBadRequest},
			requiredMocks: func(req requests.PublicKeyGet) {},
		},
		{
			title: "fails when validate because have a '&' with in your characters",
			query: requests.PublicKeyGet{
				TenantParam: requests.TenantParam{Tenant: "test&"},
			},
			expected:      Expected{expectedStatus: http.StatusBadRequest},
			requiredMocks: func(req requests.PublicKeyGet) {},
		},
		{
			title: "fails when validate because have a '@' with in your characters",
			query: requests.PublicKeyGet{
				TenantParam: requests.TenantParam{Tenant: "test@"},
			},
			expected:      Expected{expectedStatus: http.StatusBadRequest},
			requiredMocks: func(req requests.PublicKeyGet) {},
		},
		{
			title: "success when try to get a public key exists",
			query: requests.PublicKeyGet{
				FingerprintParam: requests.FingerprintParam{Fingerprint: "figertest"},
				TenantParam:      requests.TenantParam{Tenant: "tenant"},
			},
			requiredMocks: func(query requests.PublicKeyGet) {
				mock.On("GetPublicKey", gomock.Anything, query.Fingerprint, query.Tenant).Return(&models.PublicKey{}, nil)
			},
			expected: Expected{
				expectedSession: &models.PublicKey{},
				expectedStatus:  http.StatusOK,
			},
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

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session *models.PublicKey
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}
			assert.Equal(t, tc.expected.expectedSession, session)
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
			title: "fails when bind fails to validate uid",
			query: requests.PublicKeyDelete{
				FingerprintParam: requests.FingerprintParam{Fingerprint: ""},
			},
			requiredMocks:  func(query requests.PublicKeyDelete) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "fails when try to deleting an existing public key",
			query: requests.PublicKeyDelete{
				FingerprintParam: requests.FingerprintParam{Fingerprint: "figertest"},
			},
			requiredMocks: func(query requests.PublicKeyDelete) {
				mock.On("DeletePublicKey", gomock.Anything, query.Fingerprint, "tenant").Return(svc.ErrNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to deleting an existing public key",
			query: requests.PublicKeyDelete{
				FingerprintParam: requests.FingerprintParam{Fingerprint: "figertest"},
			},
			requiredMocks: func(query requests.PublicKeyDelete) {
				mock.On("DeletePublicKey", gomock.Anything, query.Fingerprint, "tenant").Return(nil).Once()
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
			title: "fails when validate because the tag does not have a min of 3 characters",
			query: requests.PublicKeyTagRemove{
				TagParam: requests.TagParam{Tag: "tg"},
			},
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func(req requests.PublicKeyTagRemove) {},
		},
		{
			title: "fails when validate because the tag does not have a max of 255 characters",
			query: requests.PublicKeyTagRemove{
				TagParam: requests.TagParam{Tag: "BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9"},
			},
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func(req requests.PublicKeyTagRemove) {},
		},
		{
			title: "fails when validate because have a '/' with in your characters",
			query: requests.PublicKeyTagRemove{
				TagParam: requests.TagParam{Tag: "test/"},
			},
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func(req requests.PublicKeyTagRemove) {},
		},
		{
			title: "fails when validate because have a '&' with in your characters",
			query: requests.PublicKeyTagRemove{
				TagParam: requests.TagParam{Tag: "test&"},
			},
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func(req requests.PublicKeyTagRemove) {},
		},
		{
			title: "fails when validate because have a '@' with in your characters",
			query: requests.PublicKeyTagRemove{
				TagParam: requests.TagParam{Tag: "test@"},
			},
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func(req requests.PublicKeyTagRemove) {},
		},
		{
			title: "success when try to removing an existing public key",
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
			title: "fails when validate because the tag does not have a min of 3 characters",
			query: requests.PublicKeyTagAdd{
				TagParam: requests.TagParam{Tag: "tg"},
			},
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func(req requests.PublicKeyTagAdd) {},
		},
		{
			title: "fails when validate because the tag does not have a max of 255 characters",
			query: requests.PublicKeyTagAdd{
				TagParam: requests.TagParam{Tag: "BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9"},
			},
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func(req requests.PublicKeyTagAdd) {},
		},
		{
			title: "fails when validate because have a '/' with in your characters",
			query: requests.PublicKeyTagAdd{
				TagParam: requests.TagParam{Tag: "test/"},
			},
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func(req requests.PublicKeyTagAdd) {},
		},
		{
			title: "fails when validate because have a '&' with in your characters",
			query: requests.PublicKeyTagAdd{
				TagParam: requests.TagParam{Tag: "test&"},
			},
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func(req requests.PublicKeyTagAdd) {},
		},
		{
			title: "fails when validate because have a '@' with in your characters",
			query: requests.PublicKeyTagAdd{
				TagParam: requests.TagParam{Tag: "test@"},
			},
			expectedStatus: http.StatusBadRequest,
			requiredMocks:  func(req requests.PublicKeyTagAdd) {},
		},
		{
			title: "success when try to add an existing public tag key",
			query: requests.PublicKeyTagAdd{
				FingerprintParam: requests.FingerprintParam{Fingerprint: "figertest"},
				TagParam:         requests.TagParam{Tag: "tag"},
			},
			tenant: "tenant-id",
			requiredMocks: func(query requests.PublicKeyTagAdd) {
				mock.On("AddPublicKeyTag", gomock.Anything, "tenant-id", query.Fingerprint, query.Tag).Return(nil).Once()
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
		title          string
		requiredMocks  func()
		expectedStatus int
	}{
		{
			title: "fails when try to deleting an existing public key",
			requiredMocks: func() {
				mock.On("CreatePrivateKey", gomock.Anything).Return(nil, svc.ErrNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to creating an existing private key",
			requiredMocks: func() {
				mock.On("CreatePrivateKey", gomock.Anything).Return(&models.PrivateKey{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
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
