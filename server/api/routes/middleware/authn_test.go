package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/authctx"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testTenant = "00000000-0000-4000-0000-000000000000"
	testUserID = "6f4c1b2a1e2f3a4b5c6d7e8f"
)

var testSigningKey = sync.OnceValues(func() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
})

func userBearer(t *testing.T) (string, *rsa.PrivateKey) {
	t.Helper()

	privateKey, err := testSigningKey()
	require.NoError(t, err)

	bearer, err := jwttoken.EncodeUserClaims(
		authorizer.UserClaims{ID: testUserID, TenantID: testTenant, Username: "john"},
		"http://localhost",
		privateKey,
	)
	require.NoError(t, err)

	return bearer, privateKey
}

func authenticatedRequest(e *echo.Echo, bearer string) (*echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/namespaces", nil)
	req.Header.Set("Authorization", "Bearer "+bearer)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/namespaces")

	return c, rec
}

func TestAuthenticatorResolveUserClaims(t *testing.T) {
	bearer, privateKey := userBearer(t)

	cases := []struct {
		description   string
		requiredMocks func(service *mocks.MockService)
		expected      *gateway.Identity
	}{
		{
			description: "succeeds when both role and admin resolve",
			requiredMocks: func(service *mocks.MockService) {
				service.On("ResolveNamespaceRole", mock.Anything, testTenant, testUserID).
					Return(&models.Namespace{TenantID: testTenant}, "owner", nil).Once()
				service.On("GetUserAdmin", mock.Anything, testUserID).Return(true, nil).Once()
			},
			expected: &gateway.Identity{
				ID:       testUserID,
				Username: "john",
				TenantID: testTenant,
				Role:     authorizer.RoleOwner,
				Admin:    true,
			},
		},
		{
			description: "yields no identity when the namespace no longer exists",
			requiredMocks: func(service *mocks.MockService) {
				service.On("ResolveNamespaceRole", mock.Anything, testTenant, testUserID).Return(nil, "", store.ErrNoDocuments).Once()
			},
			expected: nil,
		},
		{
			description: "yields no identity when the user no longer exists",
			requiredMocks: func(service *mocks.MockService) {
				service.On("ResolveNamespaceRole", mock.Anything, testTenant, testUserID).
					Return(&models.Namespace{TenantID: testTenant}, "owner", nil).Once()
				service.On("GetUserAdmin", mock.Anything, testUserID).Return(false, store.ErrNoDocuments).Once()
			},
			expected: nil,
		},
		{
			description: "yields no identity when the store is unreachable",
			requiredMocks: func(service *mocks.MockService) {
				service.On("ResolveNamespaceRole", mock.Anything, testTenant, testUserID).Return(nil, "", errors.New("connection refused")).Once()
			},
			expected: nil,
		},
		{
			description: "yields no identity when the request is cancelled",
			requiredMocks: func(service *mocks.MockService) {
				service.On("ResolveNamespaceRole", mock.Anything, testTenant, testUserID).
					Return(&models.Namespace{TenantID: testTenant}, "owner", nil).Once()
				service.On("GetUserAdmin", mock.Anything, testUserID).Return(false, context.DeadlineExceeded).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			service := new(mocks.MockService)
			service.On("PublicKey").Return(&privateKey.PublicKey).Once()
			tc.requiredMocks(service)

			c, _ := authenticatedRequest(echo.New(), bearer)

			identity, err := NewAuthenticator(service).Resolve(c)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, identity)

			service.AssertExpectations(t)
		})
	}
}

// The namespace resolved to establish the caller's role is handed to the rest of the request,
// so handlers can use it instead of resolving it a second time.
func TestAuthenticatorResolveForwardsNamespace(t *testing.T) {
	bearer, privateKey := userBearer(t)

	ns := &models.Namespace{TenantID: testTenant, MaxDevices: 3, DevicesAcceptedCount: 1}

	service := new(mocks.MockService)
	service.On("PublicKey").Return(&privateKey.PublicKey).Once()
	service.On("ResolveNamespaceRole", mock.Anything, testTenant, testUserID).Return(ns, "owner", nil).Once()
	service.On("GetUserAdmin", mock.Anything, testUserID).Return(false, nil).Once()

	c, _ := authenticatedRequest(echo.New(), bearer)

	_, err := NewAuthenticator(service).Resolve(c)
	require.NoError(t, err)

	forwarded, ok := authctx.NamespaceDeviceLimit(c.Request().Context(), testTenant)
	require.True(t, ok)
	assert.Equal(t, ns.DeviceLimit(), forwarded)

	_, ok = authctx.NamespaceDeviceLimit(c.Request().Context(), "00000000-0000-4000-0000-000000000009")
	assert.False(t, ok)

	service.AssertExpectations(t)
}

func TestAuthenticatorMiddlewareStaleToken(t *testing.T) {
	bearer, privateKey := userBearer(t)

	service := new(mocks.MockService)
	service.On("PublicKey").Return(&privateKey.PublicKey).Once()
	service.On("ResolveNamespaceRole", mock.Anything, testTenant, testUserID).Return(nil, "", store.ErrNoDocuments).Once()

	c, rec := authenticatedRequest(echo.New(), bearer)

	next := func(*echo.Context) error { return c.NoContent(http.StatusOK) }
	require.NoError(t, NewAuthenticator(service).Middleware(next)(c))

	assert.Equal(t, http.StatusUnauthorized, rec.Result().StatusCode)
	service.AssertExpectations(t)
}

func TestAuthenticatorMiddlewareStaleTokenOnAnonymousRoute(t *testing.T) {
	bearer, privateKey := userBearer(t)

	service := new(mocks.MockService)
	service.On("PublicKey").Return(&privateKey.PublicKey).Once()
	service.On("ResolveNamespaceRole", mock.Anything, testTenant, testUserID).Return(nil, "", store.ErrNoDocuments).Once()

	c, rec := authenticatedRequest(echo.New(), bearer)
	c.Request().Header.Set("X-ID", "forged")

	authenticator := NewAuthenticator(service)
	authenticator.AllowAnonymous(http.MethodGet, "/api/namespaces")

	next := func(*echo.Context) error { return c.NoContent(http.StatusOK) }
	require.NoError(t, authenticator.Middleware(next)(c))

	assert.Equal(t, http.StatusOK, rec.Result().StatusCode)
	assert.Empty(t, c.Request().Header.Get("X-ID"))
	service.AssertExpectations(t)
}

func TestAuthenticatorMiddlewareDeviceToken(t *testing.T) {
	const deviceUID = "a3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	privateKey, err := testSigningKey()
	require.NoError(t, err)

	bearer, err := jwttoken.EncodeDeviceClaims(authorizer.DeviceClaims{UID: deviceUID, TenantID: testTenant}, "http://localhost", privateKey)
	require.NoError(t, err)

	cases := []struct {
		description    string
		register       func(*Authenticator)
		expectedStatus int
		expectedDevice string
	}{
		{
			description:    "refuses a route not declared for devices",
			register:       func(*Authenticator) {},
			expectedStatus: http.StatusForbidden,
		},
		{
			description: "stamps the device identity on a declared device route",
			register: func(a *Authenticator) {
				a.AllowDevice(http.MethodGet, "/api/namespaces")
			},
			expectedStatus: http.StatusOK,
			expectedDevice: deviceUID,
		},
		{
			description: "stamps the device identity on a route declared for any method",
			register: func(a *Authenticator) {
				a.AllowDevice(AnyMethod, "/api/namespaces")
			},
			expectedStatus: http.StatusOK,
			expectedDevice: deviceUID,
		},
		{
			description: "drops the device identity on an anonymous route",
			register: func(a *Authenticator) {
				a.AllowAnonymous(http.MethodGet, "/api/namespaces")
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			service := new(mocks.MockService)
			service.On("PublicKey").Return(&privateKey.PublicKey).Once()

			c, rec := authenticatedRequest(echo.New(), bearer)
			c.Request().Header.Set("X-Device-UID", "forged")

			authenticator := NewAuthenticator(service)
			tc.register(authenticator)

			var deviceSeen string

			next := func(c *echo.Context) error {
				deviceSeen = c.Request().Header.Get("X-Device-UID")

				return c.NoContent(http.StatusOK)
			}
			require.NoError(t, authenticator.Middleware(next)(c))

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
			assert.Equal(t, tc.expectedDevice, deviceSeen)

			service.AssertExpectations(t)
		})
	}
}

func TestAuthenticatorUnregisteredRoutes(t *testing.T) {
	routes := echo.Routes{
		{Method: http.MethodGet, Path: "/api/devices"},
		{Method: http.MethodPost, Path: "/api/devices/auth"},
	}

	authenticator := NewAuthenticator(nil)
	authenticator.AllowAnonymous(http.MethodPost, "/api/devices/auth")
	authenticator.AllowAnonymous(http.MethodGet, "/api/devices/auth")
	authenticator.AllowAnonymous(AnyMethod, "/api/devices")
	authenticator.AllowDevice(http.MethodGet, "/api/devices")
	authenticator.AllowDevice(AnyMethod, "/api/sessions")

	assert.ElementsMatch(t,
		[]string{"GET /api/devices/auth", "* /api/sessions"},
		authenticator.UnregisteredRoutes(routes),
	)
}
