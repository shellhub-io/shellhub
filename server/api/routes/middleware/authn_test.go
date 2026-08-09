package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
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

// userBearer signs a user JWT and returns it alongside the key the
// authenticator must verify it with.
func userBearer(t *testing.T) (string, *rsa.PrivateKey) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	bearer, err := jwttoken.EncodeUserClaims(
		authorizer.UserClaims{ID: testUserID, TenantID: testTenant, Username: "john"},
		privateKey,
	)
	require.NoError(t, err)

	return bearer, privateKey
}

func authenticatedRequest(e *echo.Echo, bearer string) (*echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodGet, "/api/namespaces", nil)
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
				service.On("GetUserRole", mock.Anything, testTenant, testUserID).Return("owner", nil).Once()
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
				service.On("GetUserRole", mock.Anything, testTenant, testUserID).Return("", store.ErrNoDocuments).Once()
			},
			expected: nil,
		},
		{
			description: "yields no identity when the user no longer exists",
			requiredMocks: func(service *mocks.MockService) {
				service.On("GetUserRole", mock.Anything, testTenant, testUserID).Return("owner", nil).Once()
				service.On("GetUserAdmin", mock.Anything, testUserID).Return(false, store.ErrNoDocuments).Once()
			},
			expected: nil,
		},
		{
			description: "yields no identity when the store is unreachable",
			requiredMocks: func(service *mocks.MockService) {
				service.On("GetUserRole", mock.Anything, testTenant, testUserID).Return("", errors.New("connection refused")).Once()
			},
			expected: nil,
		},
		{
			description: "yields no identity when the request is cancelled",
			requiredMocks: func(service *mocks.MockService) {
				service.On("GetUserRole", mock.Anything, testTenant, testUserID).Return("owner", nil).Once()
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
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, identity)

			service.AssertExpectations(t)
		})
	}
}

func TestAuthenticatorMiddlewareStaleToken(t *testing.T) {
	bearer, privateKey := userBearer(t)

	service := new(mocks.MockService)
	service.On("PublicKey").Return(&privateKey.PublicKey).Once()
	service.On("GetUserRole", mock.Anything, testTenant, testUserID).Return("", store.ErrNoDocuments).Once()

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
	service.On("GetUserRole", mock.Anything, testTenant, testUserID).Return("", store.ErrNoDocuments).Once()

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
