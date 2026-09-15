package routes

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/envs/envstest"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
	"github.com/shellhub-io/shellhub/server/api/services"
	serviceMocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	sshhttp "github.com/shellhub-io/shellhub/server/ssh/http"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testSigningKey = sync.OnceValues(func() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
})

func authenticatedRouter(t *testing.T) (*echo.Echo, *routesmiddleware.Authenticator, *serviceMocks.MockService) {
	t.Helper()

	envstest.SetEdition(t, envs.Community)

	privateKey, err := testSigningKey()
	require.NoError(t, err)

	service := serviceMocks.NewMockService(t)
	service.On("PublicKey").Return(&privateKey.PublicKey).Maybe()

	authn := routesmiddleware.NewAuthenticator(service)

	return NewRouter(service, WithAuthentication(authn)), authn, service
}

// TestAllowlistsMatchRegisteredRoutes catches a typo or a stale entry in the
// anonymous and device allowlists, across the API routes and the SSH routes
// mounted on the same router. A key matching no route is dead: the route it
// meant to open stays closed, which fails safe but breaks an endpoint in a way no
// other test would notice.
func TestAllowlistsMatchRegisteredRoutes(t *testing.T) {
	router, authn, service := authenticatedRouter(t)

	sshhttp.Register(router, authn, dialer.NewDialer(nil, nil), service, &sshhttp.Config{})

	assert.Empty(t, authn.UnregisteredRoutes(router.Router().Routes()))
}

func TestRouterRefusesDeviceTokenOnManagementRoutes(t *testing.T) {
	const tenant = "00000000-0000-4000-0000-000000000000"

	privateKey, err := testSigningKey()
	require.NoError(t, err)

	bearer, err := jwttoken.EncodeDeviceClaims(authorizer.DeviceClaims{UID: "device", TenantID: tenant}, "http://localhost", privateKey)
	require.NoError(t, err)

	paths := []string{
		"/api/devices",
		"/api/devices/device",
		"/api/devices/resolve?uid=device",
		"/api/sessions",
		"/api/sessions/session",
		"/api/stats",
		"/api/tags",
		"/api/sshkeys/public-keys",
		"/api/namespaces/" + tenant,
		"/api/namespaces/" + tenant + "/members",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			router, _, _ := authenticatedRouter(t)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
			req.Header.Set("Authorization", "Bearer "+bearer)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusForbidden, rec.Code)
		})
	}
}

// TestRouterRejectsUncredentialedRequests guards the wiring: authentication is an
// option on NewRouter, so a production entrypoint that forgot to pass it would
// leave every route open.
func TestRouterRejectsUncredentialedRequests(t *testing.T) {
	tests := []struct {
		description string
		headers     map[string]string
		mock        func(*serviceMocks.MockService)
	}{
		{
			description: "no credential at all",
		},
		{
			description: "a malformed bearer token",
			headers:     map[string]string{"Authorization": "Bearer not-a-token"},
		},
		{
			description: "identity headers the client made up",
			headers: map[string]string{
				"X-ID":        "forged-user",
				"X-Role":      "owner",
				"X-Tenant-ID": "forged-tenant",
			},
		},
		{
			description: "an api key the store does not know",
			headers:     map[string]string{"X-API-Key": "not-a-key"},
			mock: func(service *serviceMocks.MockService) {
				service.
					On("AuthAPIKey", mock.Anything, "not-a-key").
					Return(nil, services.NewErrAPIKeyNotFound("", nil)).
					Once()
			},
		},
		{
			description: "an api key that is no longer valid",
			headers:     map[string]string{"X-API-Key": "expired"},
			mock: func(service *serviceMocks.MockService) {
				service.
					On("AuthAPIKey", mock.Anything, "expired").
					Return(nil, services.NewErrAPIKeyInvalid("expired")).
					Once()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			router, _, service := authenticatedRouter(t)
			if tc.mock != nil {
				tc.mock(service)
			}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/devices", nil)
			for key, value := range tc.headers {
				req.Header.Set(key, value)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	}
}

// TestAnonymousRouteReachableWithoutCredential covers the other direction, and
// pins the header scrubbing the edge proxy's configuration used to do by hand
// per route: a client cannot reach even an anonymous handler carrying an identity
// it never authenticated as.
func TestAnonymousRouteReachableWithoutCredential(t *testing.T) {
	router, authn, service := authenticatedRouter(t)

	const probe = "/api/anonymous-probe"

	var seen http.Header

	router.GET(probe, func(c *echo.Context) error {
		seen = c.Request().Header.Clone()

		return c.NoContent(http.StatusOK)
	})
	authn.AllowAnonymous(http.MethodGet, probe)

	service.
		On("AuthAPIKey", mock.Anything, "forged-key").
		Return(nil, services.NewErrAPIKeyNotFound("", nil)).
		Once()

	forged := map[string]string{
		"X-ID":         "forged-user",
		"X-Username":   "forged-username",
		"X-Tenant-ID":  "forged-tenant",
		"X-Device-UID": "forged-device",
		"X-API-Key":    "forged-key",
		"X-Role":       "owner",
		"X-Admin":      "true",
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, probe, nil)
	for key, value := range forged {
		req.Header.Set(key, value)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	require.NotNil(t, seen)
	for key := range forged {
		assert.Empty(t, seen.Get(key), "%s reached the handler", key)
	}
}
