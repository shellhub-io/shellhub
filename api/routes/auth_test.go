package routes

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthDevice(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedResponse *models.DeviceAuthResponse
		expectedStatus   int
	}
	cases := []struct {
		title         string
		requestBody   *requests.DeviceAuth
		requiredMocks func()
		expected      Expected
	}{
		{
			title: "success when device has a preferred hostname and it is uppercase",
			requestBody: &requests.DeviceAuth{
				Info: &requests.DeviceInfo{
					ID:         "device_id",
					PrettyName: "Device Name",
					Version:    "1.0",
					Arch:       "amd64",
					Platform:   "Linux",
				},
				Hostname:  "TEST",
				PublicKey: "your_public_key",
				TenantID:  "your_tenant_id",
			},
			requiredMocks: func() {
				mock.On("AuthDevice", gomock.Anything, gomock.AnythingOfType("requests.DeviceAuth")).Return(&models.DeviceAuthResponse{}, nil).Once()
			},
			expected: Expected{
				expectedResponse: &models.DeviceAuthResponse{},
				expectedStatus:   http.StatusOK,
			},
		},
		{
			title: "success when device has a preferred hostname and it is lowercase",
			requestBody: &requests.DeviceAuth{
				Info: &requests.DeviceInfo{
					ID:         "device_id",
					PrettyName: "Device Name",
					Version:    "1.0",
					Arch:       "amd64",
					Platform:   "Linux",
				},
				Hostname:  "test",
				PublicKey: "your_public_key",
				TenantID:  "your_tenant_id",
			},
			requiredMocks: func() {
				mock.On("AuthDevice", gomock.Anything, gomock.AnythingOfType("requests.DeviceAuth")).Return(&models.DeviceAuthResponse{}, nil).Once()
			},
			expected: Expected{
				expectedResponse: &models.DeviceAuthResponse{},
				expectedStatus:   http.StatusOK,
			},
		},
		{
			title: "success when try auth a device",
			requestBody: &requests.DeviceAuth{
				Info: &requests.DeviceInfo{
					ID:         "device_id",
					PrettyName: "Device Name",
					Version:    "1.0",
					Arch:       "amd64",
					Platform:   "Linux",
				},
				Identity: &requests.DeviceIdentity{
					MAC: "00:11:22:33:44:55",
				},
				PublicKey: "your_public_key",
				TenantID:  "your_tenant_id",
			},
			requiredMocks: func() {
				mock.On("AuthDevice", gomock.Anything, gomock.AnythingOfType("requests.DeviceAuth")).Return(&models.DeviceAuthResponse{}, nil).Once()
				mock.On("SetDevicePosition", gomock.Anything, models.UID(""), "").Return(nil).Once()
			},
			expected: Expected{
				expectedResponse: &models.DeviceAuthResponse{},
				expectedStatus:   http.StatusOK,
			},
		},
		{
			title:         "fails when try validate request",
			requestBody:   &requests.DeviceAuth{},
			requiredMocks: func() {},
			expected: Expected{
				expectedResponse: nil,
				expectedStatus:   http.StatusBadRequest,
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

			req := httptest.NewRequest(http.MethodPost, "/api/devices/auth", strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			if tc.expected.expectedResponse != nil {
				var response models.DeviceAuthResponse
				if err := json.NewDecoder(rec.Result().Body).Decode(&response); err != nil {
					assert.ErrorIs(t, io.EOF, err)
				}

				assert.Equal(t, tc.expected.expectedResponse, &response)
			}
		})
	}
}

func TestAuthLocalUser(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		body    *models.UserAuthResponse
		headers map[string]string
		status  int
	}

	cases := []struct {
		description string
		req         *requests.AuthLocalUser
		mocks       func()
		expected    Expected
	}{
		{
			description: "fails when the identifier is empty",
			req: &requests.AuthLocalUser{
				Identifier: "",
				Password:   "secret",
			},
			mocks: func() {},
			expected: Expected{
				body:    nil,
				headers: map[string]string{},
				status:  http.StatusBadRequest,
			},
		},
		{
			description: "fails when the password is empty",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "",
			},
			mocks: func() {},
			expected: Expected{
				body:    nil,
				headers: map[string]string{},
				status:  http.StatusBadRequest,
			},
		},
		{
			description: "fails when the user is not found",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "wrong_password",
			},
			mocks: func() {
				mock.
					On("AuthLocalUser", gomock.Anything, &requests.AuthLocalUser{
						Identifier: "john_doe",
						Password:   "wrong_password",
					}, gomock.Anything).
					Return(nil, int64(0), "", svc.ErrUserNotFound).
					Once()
			},
			expected: Expected{
				body:    nil,
				headers: map[string]string{},
				status:  http.StatusUnauthorized,
			},
		},
		{
			description: "fails when the password is wrong",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "wrong_password",
			},
			mocks: func() {
				mock.
					On("AuthLocalUser", gomock.Anything, &requests.AuthLocalUser{
						Identifier: "john_doe",
						Password:   "wrong_password",
					}, gomock.Anything).
					Return(nil, int64(0), "", svc.ErrAuthUnathorized).
					Once()
			},
			expected: Expected{
				body: nil,
				headers: map[string]string{
					"X-Account-Lockout": "0",
					"X-MFA-Token":       "",
				},
				status: http.StatusUnauthorized,
			},
		},
		{
			description: "fails when reaching the attempt limits",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "wrong_password",
			},
			mocks: func() {
				mock.
					On("AuthLocalUser", gomock.Anything, &requests.AuthLocalUser{
						Identifier: "john_doe",
						Password:   "wrong_password",
					}, gomock.Anything).
					Return(nil, int64(1711176851), "", svc.ErrAuthUnathorized).
					Once()
			},
			expected: Expected{
				body: nil,
				headers: map[string]string{
					"X-Account-Lockout": "1711176851",
					"X-MFA-Token":       "",
				},
				status: http.StatusTooManyRequests,
			},
		},
		{
			description: "fails when mfa is enable",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "wrong_password",
			},
			mocks: func() {
				mock.
					On("AuthLocalUser", gomock.Anything, &requests.AuthLocalUser{
						Identifier: "john_doe",
						Password:   "wrong_password",
					}, gomock.Anything).
					Return(nil, int64(0), "00000000-0000-4000-0000-000000000000", svc.ErrAuthUnathorized).
					Once()
			},
			expected: Expected{
				body: nil,
				headers: map[string]string{
					"X-Account-Lockout": "0",
					"X-MFA-Token":       "00000000-0000-4000-0000-000000000000",
				},
				status: http.StatusUnauthorized,
			},
		},
		{
			description: "success when try to auth a user",
			req: &requests.AuthLocalUser{
				Identifier: "john_doe",
				Password:   "secret",
			},
			mocks: func() {
				mock.
					On("AuthLocalUser", gomock.Anything, &requests.AuthLocalUser{
						Identifier: "john_doe",
						Password:   "secret",
					}, gomock.Anything).
					Return(&models.UserAuthResponse{
						ID:     "65fdd16b5f62f93184ec8a39",
						Name:   "john doe",
						User:   "john_doe",
						Email:  "john.doe@test.com",
						Tenant: "00000000-0000-4000-0000-000000000000",
						Token:  "not-empty",
					}, int64(0), "", nil).
					Once()
			},
			expected: Expected{
				body: &models.UserAuthResponse{
					ID:     "65fdd16b5f62f93184ec8a39",
					Name:   "john doe",
					User:   "john_doe",
					Email:  "john.doe@test.com",
					Tenant: "00000000-0000-4000-0000-000000000000",
					Token:  "not-empty",
				},
				headers: map[string]string{
					"X-Account-Lockout": "0",
					"X-MFA-Token":       "",
				},
				status: http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.mocks()

			jsonData, err := json.Marshal(tc.req)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/auth/user", strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			var body *models.UserAuthResponse

			if tc.expected.body != nil {
				if err := json.NewDecoder(rec.Result().Body).Decode(&body); err != nil {
					assert.ErrorIs(t, io.EOF, err)
				}
			}

			assert.Equal(t, tc.expected.body, body)
			assert.Equal(t, tc.expected.status, rec.Result().StatusCode)
			for k, v := range tc.expected.headers {
				assert.Equal(t, v, rec.Result().Header.Get(k))
			}
		})
	}
}

func TestCreateUserToken(t *testing.T) {
	svcMock := new(mocks.Service)

	type Expected struct {
		body   *models.UserAuthResponse
		status int
	}

	cases := []struct {
		description string
		tenantID    string
		headers     map[string]string
		mocks       func()
		expected    Expected
	}{
		{
			description: "success without tenant_id",
			tenantID:    "",
			headers:     map[string]string{"X-ID": "000000000000000000000000"},
			mocks: func() {
				svcMock.
					On("CreateUserToken", gomock.Anything, &requests.CreateUserToken{
						UserID:   "000000000000000000000000",
						TenantID: "",
					}).
					Return(&models.UserAuthResponse{
						ID:     "000000000000000000000000",
						Name:   "john doe",
						User:   "john_doe",
						Email:  "john.doe@test.com",
						Tenant: "00000000-0000-4000-0000-000000000000",
						Token:  "not-empty",
					}, nil).
					Once()
			},
			expected: Expected{
				body: &models.UserAuthResponse{
					ID:     "000000000000000000000000",
					Name:   "john doe",
					User:   "john_doe",
					Email:  "john.doe@test.com",
					Tenant: "00000000-0000-4000-0000-000000000000",
					Token:  "not-empty",
				},
				status: http.StatusOK,
			},
		},
		{
			description: "success with tenant_id",
			tenantID:    "00000000-0000-4000-0000-000000000001",
			headers:     map[string]string{"X-ID": "000000000000000000000000"},
			mocks: func() {
				svcMock.
					On("CreateUserToken", gomock.Anything, &requests.CreateUserToken{
						UserID:   "000000000000000000000000",
						TenantID: "00000000-0000-4000-0000-000000000001",
					}).
					Return(&models.UserAuthResponse{
						ID:     "000000000000000000000000",
						Name:   "john doe",
						User:   "john_doe",
						Email:  "john.doe@test.com",
						Tenant: "00000000-0000-4000-0000-000000000001",
						Token:  "not-empty",
					}, nil).
					Once()
			},
			expected: Expected{
				body: &models.UserAuthResponse{
					ID:     "000000000000000000000000",
					Name:   "john doe",
					User:   "john_doe",
					Email:  "john.doe@test.com",
					Tenant: "00000000-0000-4000-0000-000000000001",
					Token:  "not-empty",
				},
				status: http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.mocks()

			req := new(http.Request)
			if tc.tenantID == "" {
				req = httptest.NewRequest(http.MethodGet, "/api/auth/user", nil)
			} else {
				req = httptest.NewRequest(http.MethodGet, "/api/auth/token/"+tc.tenantID, nil)
			}

			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			body := new(models.UserAuthResponse)
			if tc.expected.body != nil {
				if err := json.NewDecoder(rec.Result().Body).Decode(&body); err != nil {
					require.ErrorIs(t, io.EOF, err)
				}
			}

			assert.Equal(t, tc.expected.body, body)
			assert.Equal(t, tc.expected.status, rec.Result().StatusCode)
		})
	}
}

func TestAuthPublicKey(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedResponse *models.PublicKeyAuthResponse
		expectedStatus   int
	}

	cases := []struct {
		title         string
		requestBody   *requests.PublicKeyAuth
		requiredMocks func()
		expected      Expected
	}{
		{
			title: "success when try to auth a public key",
			requestBody: &requests.PublicKeyAuth{
				Fingerprint: "fingerprint",
				Data:        "data",
			},
			requiredMocks: func() {
				req := requests.PublicKeyAuth{
					Fingerprint: "fingerprint",
					Data:        "data",
				}
				mock.On("AuthPublicKey", gomock.Anything, req).Return(&models.PublicKeyAuthResponse{}, nil).Once()
			},
			expected: Expected{
				expectedResponse: &models.PublicKeyAuthResponse{},
				expectedStatus:   http.StatusOK,
			},
		},
		{
			title:         "fails when try to validate a request",
			requestBody:   &requests.PublicKeyAuth{},
			requiredMocks: func() {},
			expected: Expected{
				expectedResponse: nil,
				expectedStatus:   http.StatusBadRequest,
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

			req := httptest.NewRequest(http.MethodPost, "/api/auth/ssh", strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			if tc.expected.expectedResponse != nil {
				var response models.PublicKeyAuthResponse
				if err := json.NewDecoder(rec.Result().Body).Decode(&response); err != nil {
					assert.ErrorIs(t, io.EOF, err)
				}

				assert.Equal(t, tc.expected.expectedResponse, &response)
			}
		})
	}
}

func TestHandler_AuthRequest_with_authorization_header(t *testing.T) {
	type Expected struct {
		status  int
		headers map[string]string
	}

	svcMock := new(mocks.Service)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	cases := []struct {
		description   string
		token         func() (string, error)
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "failed when token is invalid",
			token: func() (string, error) {
				return "", nil
			},
			requiredMocks: func() {
				svcMock.On("PublicKey").Return(&privateKey.PublicKey).Once()
			},
			expected: Expected{
				status:  401,
				headers: map[string]string{},
			},
		},
		{
			description: "succeeds to authenticate a user",
			token: func() (string, error) {
				claims := authorizer.UserClaims{
					ID:       "000000000000000000000000",
					TenantID: "00000000-0000-4000-0000-000000000000",
					Role:     authorizer.RoleOwner,
					Username: "john_doe",
				}

				return jwttoken.EncodeUserClaims(claims, privateKey)
			},
			requiredMocks: func() {
				svcMock.On("PublicKey").Return(&privateKey.PublicKey).Once()
				svcMock.On("GetUserRole", gomock.Anything, "00000000-0000-4000-0000-000000000000", "000000000000000000000000").Return("owner", nil).Once()
			},
			expected: Expected{
				status: 200,
				headers: map[string]string{
					"X-ID":        "000000000000000000000000",
					"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
					"X-Role":      authorizer.RoleOwner.String(),
					"X-Username":  "john_doe",
				},
			},
		},
		{
			description: "succeeds to authenticate a device",
			token: func() (string, error) {
				claims := authorizer.DeviceClaims{
					UID:      "0000000000000000000000000000000000000000000000000000000000000000",
					TenantID: "00000000-0000-4000-0000-000000000000",
				}

				return jwttoken.EncodeDeviceClaims(claims, privateKey)
			},
			requiredMocks: func() {
				svcMock.On("PublicKey").Return(&privateKey.PublicKey).Once()
			},
			expected: Expected{
				status: 200,
				headers: map[string]string{
					"X-Device-UID": "0000000000000000000000000000000000000000000000000000000000000000",
					"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodGet, "/internal/auth", nil)

			token, err := tc.token()
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", token)

			rec := httptest.NewRecorder()

			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			require.Equal(t, tc.expected.status, rec.Result().StatusCode)
			for k, v := range tc.expected.headers {
				require.Equal(t, rec.Result().Header.Get(k), v)
			}
		})
	}
}
