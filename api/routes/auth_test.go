package routes

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestAuthGetToken(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedSession *models.UserAuthResponse
		expectedStatus  int
	}
	cases := []struct {
		title         string
		id            requests.AuthTokenGet
		requiredMocks func()
		expected      Expected
	}{
		{
			title:         "fails when validate fails",
			id:            requests.AuthTokenGet{UserParam: requests.UserParam{ID: ""}},
			requiredMocks: func() {},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusBadRequest,
			},
		},
		{
			title: "success when trying to get a token",
			id:    requests.AuthTokenGet{UserParam: requests.UserParam{ID: "id"}, MFA: false},
			requiredMocks: func() {
				mock.On("AuthGetToken", gomock.Anything, "id", false).Return(&models.UserAuthResponse{}, nil).Once()
			},
			expected: Expected{
				expectedSession: &models.UserAuthResponse{},
				expectedStatus:  http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			jsonData, err := json.Marshal(tc.id)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/internal/auth/token/%s", jsonData), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", guard.RoleOwner)
			req.Header.Set("X-ID", string(jsonData))
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session *models.UserAuthResponse
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}

			assert.Equal(t, tc.expected.expectedSession, session)

			mock.AssertExpectations(t)
		})
	}
}

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
				mock.On("AuthDevice", gomock.Anything, gomock.AnythingOfType("requests.DeviceAuth"), "").Return(&models.DeviceAuthResponse{}, nil).Once()
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

func TestAuthUser(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		body   *models.UserAuthResponse
		status int
	}

	cases := []struct {
		description string
		req         *requests.UserAuth
		mocks       func()
		expected    Expected
	}{
		{
			description: "fails when the identifier is empty",
			req: &requests.UserAuth{
				Identifier: "",
				Password:   "secret",
			},
			mocks: func() {},
			expected: Expected{
				body:   nil,
				status: http.StatusBadRequest,
			},
		},
		{
			description: "fails when the password is empty",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "",
			},
			mocks: func() {},
			expected: Expected{
				body:   nil,
				status: http.StatusBadRequest,
			},
		},
		{
			description: "fails when the user is not found",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "wrong_password",
			},
			mocks: func() {
				mock.
					On("AuthUser", gomock.Anything, &requests.UserAuth{
						Identifier: "john_doe",
						Password:   "wrong_password",
					}, gomock.Anything).
					Return(nil, int64(0), svc.ErrUserNotFound).
					Once()
			},
			expected: Expected{
				body:   nil,
				status: http.StatusUnauthorized,
			},
		},
		{
			description: "fails when the password is wrong",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "wrong_password",
			},
			mocks: func() {
				mock.
					On("AuthUser", gomock.Anything, &requests.UserAuth{
						Identifier: "john_doe",
						Password:   "wrong_password",
					}, gomock.Anything).
					Return(nil, int64(0), svc.ErrAuthUnathorized).
					Once()
			},
			expected: Expected{
				body:   nil,
				status: http.StatusUnauthorized,
			},
		},
		{
			description: "fails when reaching the attempt limits",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "wrong_password",
			},
			mocks: func() {
				mock.
					On("AuthUser", gomock.Anything, &requests.UserAuth{
						Identifier: "john_doe",
						Password:   "wrong_password",
					}, gomock.Anything).
					Return(nil, int64(1711176851), svc.ErrAuthUnathorized).
					Once()
			},
			expected: Expected{
				body:   nil,
				status: http.StatusTooManyRequests,
			},
		},
		{
			description: "success when try to auth a user",
			req: &requests.UserAuth{
				Identifier: "john_doe",
				Password:   "secret",
			},
			mocks: func() {
				mock.
					On("AuthUser", gomock.Anything, &requests.UserAuth{
						Identifier: "john_doe",
						Password:   "secret",
					}, gomock.Anything).
					Return(
						&models.UserAuthResponse{
							ID:     "65fdd16b5f62f93184ec8a39",
							Name:   "john doe",
							User:   "john_doe",
							Email:  "john.doe@test.com",
							Tenant: "00000000-0000-4000-0000-000000000000",
							Role:   "owner",
							Token:  "not-empty",
							MFA: models.MFA{
								Enable:   false,
								Validate: false,
							},
						},
						int64(0),
						nil,
					).
					Once()
			},
			expected: Expected{
				body: &models.UserAuthResponse{
					ID:     "65fdd16b5f62f93184ec8a39",
					Name:   "john doe",
					User:   "john_doe",
					Email:  "john.doe@test.com",
					Tenant: "00000000-0000-4000-0000-000000000000",
					Role:   "owner",
					Token:  "not-empty",
					MFA: models.MFA{
						Enable:   false,
						Validate: false,
					},
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

			status := rec.Result().StatusCode
			var body *models.UserAuthResponse

			if tc.expected.body != nil {
				if err := json.NewDecoder(rec.Result().Body).Decode(&body); err != nil {
					assert.ErrorIs(t, io.EOF, err)
				}
			}

			assert.Equal(t, tc.expected, Expected{body, status})
		})
	}
}

func TestAuthUserInfo(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedResponse *models.UserAuthResponse
		expectedStatus   int
	}

	cases := []struct {
		title          string
		requestHeaders map[string]string
		requiredMocks  func()
		expected       Expected
	}{
		{
			title: "success when try to auth a user info",
			requestHeaders: map[string]string{
				"X-Username":  "user",
				"X-Tenant-ID": "tenant",
			},
			requiredMocks: func() {
				mock.On("AuthUserInfo", gomock.Anything, "user", "tenant", gomock.Anything).Return(&models.UserAuthResponse{}, nil).Once()
			},
			expected: Expected{
				expectedResponse: &models.UserAuthResponse{},
				expectedStatus:   http.StatusOK,
			},
		},
		{
			title: "fails when try to auth a user info",
			requestHeaders: map[string]string{
				"X-Username":  "user",
				"X-Tenant-ID": "tenant",
			},
			requiredMocks: func() {
				mock.On("AuthUserInfo", gomock.Anything, "user", "tenant", gomock.Anything).Return(nil, svc.ErrAuthUnathorized).Once()
			},
			expected: Expected{
				expectedResponse: nil,
				expectedStatus:   http.StatusUnauthorized,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodGet, "/api/auth/user", nil)
			req.Header.Set("Content-Type", "application/json")

			for key, value := range tc.requestHeaders {
				req.Header.Set(key, value)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			if tc.expected.expectedResponse != nil {
				var response models.UserAuthResponse
				if err := json.NewDecoder(rec.Result().Body).Decode(&response); err != nil {
					assert.ErrorIs(t, io.EOF, err)
				}

				assert.Equal(t, tc.expected.expectedResponse, &response)
			}
		})
	}
}

func TestAuthSwapToken(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedResponse *models.UserAuthResponse
		expectedStatus   int
	}

	cases := []struct {
		title         string
		requestBody   string
		requiredMocks func()
		expected      Expected
	}{
		{
			title:       "success when try to swap token",
			requestBody: "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				mock.On("AuthSwapToken", gomock.Anything, "id", "00000000-0000-4000-0000-000000000000").Return(&models.UserAuthResponse{}, nil).Once()
			},
			expected: Expected{
				expectedResponse: &models.UserAuthResponse{},
				expectedStatus:   http.StatusOK,
			},
		},
		{
			title:         "fails when try to swap a token",
			requestBody:   "",
			requiredMocks: func() {},
			expected: Expected{
				expectedResponse: nil,
				expectedStatus:   http.StatusNotFound,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/auth/token/%s", tc.requestBody), nil)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			c := gateway.NewContext(mock, e.NewContext(req, rec))
			c.Request().Header.Set("X-ID", "id")

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			if tc.expected.expectedResponse != nil {
				var response models.UserAuthResponse
				if err := json.NewDecoder(rec.Result().Body).Decode(&response); err != nil {
					assert.ErrorIs(t, io.EOF, err)
				}

				assert.Equal(t, tc.expected.expectedResponse, &response)
			}
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

func TestAuthRequest(t *testing.T) {
	mock := new(mocks.Service)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.UserAuthClaims{
		Username: "username",
		Admin:    true,
		Tenant:   "tenant",
		Role:     "role",
		ID:       "id",
		AuthClaims: models.AuthClaims{
			Claims: "user",
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(clock.Now().Add(time.Hour * 72)),
		},
	})

	type Expected struct {
		expectedStatus int
	}
	cases := []struct {
		title         string
		requiredMocks func()
		expected      Expected
	}{
		{
			title: "success when trying to verify token authorization",
			requiredMocks: func() {
				mock.On("PublicKey").Return(&privateKey.PublicKey).Once()
				mock.On("AuthIsCacheToken", gomock.Anything, "tenant", "id").Return(true, nil).Once()
				mock.On("GetAPIKeyByUID", gomock.Anything, "").Return(&models.APIKey{
					TenantID: "tenant",
				}, nil).Once()
				mock.On("AuthMFA", gomock.Anything, "id").Return(false, nil).Once()
			},
			expected: Expected{
				expectedStatus: http.StatusOK,
			},
		},
		{
			title: "fails when token dont have cache",
			requiredMocks: func() {
				mock.On("PublicKey").Return(&privateKey.PublicKey).Once()
				mock.On("AuthIsCacheToken", gomock.Anything, "tenant", "id").Return(false, nil).Once()
			},
			expected: Expected{
				expectedStatus: http.StatusUnauthorized,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodGet, "/internal/auth", nil)
			req.Header.Set("Content-Type", "application/json")

			tokenStr, err := token.SignedString(privateKey)
			assert.NoError(t, err)

			req.Header.Add("Authorization", "Bearer "+tokenStr)

			req.Header.Set("X-Role", guard.RoleOwner)

			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)
		})
	}
}
