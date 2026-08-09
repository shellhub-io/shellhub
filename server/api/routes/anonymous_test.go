package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/envs/envstest"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
	"github.com/shellhub-io/shellhub/server/api/services"
	serviceMocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func authenticatedRouter(t *testing.T) (*echo.Echo, *routesmiddleware.Authenticator, *serviceMocks.MockService) {
	t.Helper()

	envstest.SetEdition(t, envs.Community)

	service := serviceMocks.NewMockService(t)
	// Reading the JWT signing key is incidental to what these tests assert; a nil
	// key still rejects every token, which is what an unauthenticated request
	// needs.
	service.On("PublicKey").Return(nil).Maybe()

	authn := routesmiddleware.NewAuthenticator(service)

	return NewRouter(service, WithAuthentication(authn)), authn, service
}

// TestAnonymousAllowlistMatchesRegisteredRoutes catches a typo or a stale entry
// in the anonymous allowlist. A key matching no route is dead: the route it meant
// to open stays authenticated, which fails safe but breaks a public endpoint in a
// way no other test would notice.
func TestAnonymousAllowlistMatchesRegisteredRoutes(t *testing.T) {
	router, authn, _ := authenticatedRouter(t)

	registered := make(map[string]struct{})
	paths := make(map[string]struct{})

	for _, route := range router.Router().Routes() {
		registered[route.Method+" "+route.Path] = struct{}{}
		paths[route.Path] = struct{}{}
	}

	for _, entry := range authn.AnonymousRoutes() {
		method, path, found := strings.Cut(entry, " ")
		require.True(t, found, "malformed allowlist entry %q", entry)

		if method == routesmiddleware.AnyMethod {
			assert.Contains(t, paths, path,
				"allowlist names %q but no route is registered on that path", entry)

			continue
		}

		assert.Contains(t, registered, entry,
			"allowlist names %q but that method/path pair is not registered", entry)
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
			// A key the store cannot resolve is a rejected credential, not a
			// missing resource: the edge proxy turned every non-2xx from the
			// authentication subrequest into a 401, and a client holding a
			// revoked key needs to be told to reauthenticate.
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

			req := httptest.NewRequest(http.MethodGet, "/api/devices", nil)
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

	// Every header gateway.Identity writes, so a new one cannot be added to that
	// set without being scrubbed here too.
	forged := map[string]string{
		"X-ID":         "forged-user",
		"X-Username":   "forged-username",
		"X-Tenant-ID":  "forged-tenant",
		"X-Device-UID": "forged-device",
		"X-API-Key":    "forged-key",
		"X-Role":       "owner",
		"X-Admin":      "true",
	}

	req := httptest.NewRequest(http.MethodGet, probe, nil)
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
