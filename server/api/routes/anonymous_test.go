package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/envs/envstest"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
	serviceMocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func authenticatedRouter(t *testing.T) (*echo.Echo, *routesmiddleware.Authenticator) {
	t.Helper()

	envstest.SetEdition(t, envs.Community)

	service := serviceMocks.NewMockService(t)
	// Reading the JWT signing key is incidental to what these tests assert; a nil
	// key still rejects every token, which is what an unauthenticated request
	// needs.
	service.On("PublicKey").Return(nil).Maybe()

	authn := routesmiddleware.NewAuthenticator(service)

	return NewRouter(service, WithAuthentication(authn)), authn
}

// TestAnonymousAllowlistMatchesRegisteredRoutes catches a typo or a stale entry
// in the anonymous allowlist. A key matching no route is dead: the route it meant
// to open stays authenticated, which fails safe but breaks a public endpoint in a
// way no other test would notice.
func TestAnonymousAllowlistMatchesRegisteredRoutes(t *testing.T) {
	router, authn := authenticatedRouter(t)

	registered := make(map[string]struct{})
	paths := make(map[string]struct{})

	for _, route := range router.Routes() {
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
	router, _ := authenticatedRouter(t)

	tests := []struct {
		description string
		headers     map[string]string
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
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
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
	router, authn := authenticatedRouter(t)

	const probe = "/api/anonymous-probe"

	var seen http.Header

	router.GET(probe, func(c echo.Context) error {
		seen = c.Request().Header.Clone()

		return c.NoContent(http.StatusOK)
	})
	authn.AllowAnonymous(http.MethodGet, probe)

	req := httptest.NewRequest(http.MethodGet, probe, nil)
	req.Header.Set("X-ID", "forged-user")
	req.Header.Set("X-Role", "owner")
	req.Header.Set("X-Tenant-ID", "forged-tenant")
	req.Header.Set("X-Admin", "true")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	require.NotNil(t, seen)
	assert.Empty(t, seen.Get("X-ID"))
	assert.Empty(t, seen.Get("X-Role"))
	assert.Empty(t, seen.Get("X-Tenant-ID"))
	assert.Empty(t, seen.Get("X-Admin"))
}
