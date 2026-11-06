package routetable_test

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/shellhub-io/shellhub/pkg/api/identity/identitytest"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/server/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/server/api/routes/middleware"
	"github.com/shellhub-io/shellhub/server/api/routes/routetable"
	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newRouter() *echo.Echo {
	router := echo.New()
	router.HTTPErrorHandler = handlers.NewErrors(nil)
	router.Use(gateway.WithContext())

	return router
}

func ok(gateway.Claims) echo.HandlerFunc {
	return func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	}
}

func serve(t *testing.T, router *echo.Echo, method, path string, caller *identity.Identity, header http.Header) int {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, path, nil)
	maps.Copy(req.Header, header)

	if caller != nil {
		req = identitytest.As(req, *caller)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec.Code
}

func denyAll(echo.HandlerFunc) echo.HandlerFunc {
	return func(*echo.Context) error {
		return echo.ErrForbidden
	}
}

func allowAll(next echo.HandlerFunc) echo.HandlerFunc {
	return next
}

func TestMountRefusesARouteThatStatesNoAccess(t *testing.T) {
	cases := []struct {
		name   string
		access routetable.Access
	}{
		{name: "no access stated", access: routetable.Access{}},
		{name: "no permission required", access: routetable.Requires()},
		{name: "no admin guard", access: routetable.Admin(nil)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := newRouter()

			err := routetable.Mount(router.Group("/api"), nil, []routetable.Route{
				{Method: http.MethodGet, Path: "/stated", Endpoint: ok, Access: routetable.OwnAccount},
				{Method: http.MethodGet, Path: "/unstated", Endpoint: ok, Access: tc.access},
			})

			require.ErrorIs(t, err, routetable.ErrUnstatedAccess)
			assert.Empty(t, router.Router().Routes())
		})
	}
}

func TestMountGuardsEachRouteByTheAccessItStates(t *testing.T) {
	owner := &identity.Identity{ID: "user", TenantID: "tenant", Role: authorizer.RoleOwner}
	observer := &identity.Identity{ID: "user", TenantID: "tenant", Role: authorizer.RoleObserver}
	userWithoutNamespace := &identity.Identity{ID: "user"}
	apiKey := &identity.Identity{ID: "user", TenantID: "tenant", Role: authorizer.RoleOwner, APIKey: "key"}

	router := newRouter()
	require.NoError(t, routetable.Mount(router.Group("/api"), nil, []routetable.Route{
		{Method: http.MethodGet, Path: "/permission", Endpoint: ok, Access: routetable.Requires(authorizer.BillingGetCustomer)},
		{Method: http.MethodGet, Path: "/member", Endpoint: ok, Access: routetable.NamespaceMember},
		{Method: http.MethodGet, Path: "/account", Endpoint: ok, Access: routetable.OwnAccount},
		{Method: http.MethodGet, Path: "/no-api-key", Endpoint: ok, Access: routetable.NamespaceMember, NoAPIKey: true},
		{Method: http.MethodGet, Path: "/admin", Endpoint: ok, Access: routetable.Admin(denyAll)},
	}))

	cases := []struct {
		name   string
		path   string
		caller *identity.Identity
		header http.Header
		status int
	}{
		{name: "a role holding the permission passes", path: "/api/permission", caller: owner, status: http.StatusOK},
		{name: "a role lacking the permission is refused", path: "/api/permission", caller: observer, status: http.StatusForbidden},
		{name: "a namespace member passes", path: "/api/member", caller: observer, status: http.StatusOK},
		{name: "a user outside any namespace is refused a namespace route", path: "/api/member", caller: userWithoutNamespace, status: http.StatusForbidden},
		{name: "a user passes an own account route", path: "/api/account", caller: userWithoutNamespace, status: http.StatusOK},
		{name: "an API key is refused an own account route", path: "/api/account", caller: apiKey, status: http.StatusForbidden},
		{name: "an API key is refused a route closed to API keys", path: "/api/no-api-key", caller: apiKey, status: http.StatusForbidden},
		{name: "a route closed to API keys still serves a user", path: "/api/no-api-key", caller: owner, status: http.StatusOK},
		{name: "an admin route answers through its guard", path: "/api/admin", caller: owner, status: http.StatusForbidden},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.status, serve(t, router, http.MethodGet, tc.path, tc.caller, tc.header))
		})
	}
}

func TestMountOpensOnlyAnonymousRoutesToACallerWithoutACredential(t *testing.T) {
	authn := middleware.NewAuthenticator(nil)

	router := newRouter()
	router.Use(authn.Middleware)
	require.NoError(t, routetable.Mount(router.Group("/api"), authn, []routetable.Route{
		{Method: http.MethodPost, Path: "/anonymous", Endpoint: ok, Access: routetable.Anonymous},
		{Method: echo.RouteAny, Path: "/anonymous-any", Endpoint: ok, Access: routetable.Anonymous},
		{Method: http.MethodPost, Path: "/account", Endpoint: ok, Access: routetable.OwnAccount},
	}))

	assert.Equal(t, http.StatusOK, serve(t, router, http.MethodPost, "/api/anonymous", nil, nil))
	assert.Equal(t, http.StatusOK, serve(t, router, http.MethodDelete, "/api/anonymous-any", nil, nil))
	assert.Equal(t, http.StatusUnauthorized, serve(t, router, http.MethodPost, "/api/account", nil, nil))
	assert.Empty(t, authn.UnregisteredRoutes(router.Router().Routes()))
}

func TestUndeclaredNamesEveryRouteMountedOutsideATable(t *testing.T) {
	router := newRouter()
	require.NoError(t, routetable.Mount(router.Group("/api"), nil, []routetable.Route{
		{Method: http.MethodGet, Path: "/mounted", Endpoint: ok, Access: routetable.NamespaceMember},
	}))

	group := router.Group("/grouped")
	group.Use(denyAll)
	group.GET("/bare", ok(gateway.Claims{}))
	router.GET("/declared/:id", ok(gateway.Claims{}))
	router.POST("/declared/:id", ok(gateway.Claims{}))

	declared := map[string]string{
		routetable.Address(http.MethodGet, "/declared/:id"): "mounted by a table of its own",
	}

	assert.ElementsMatch(t, []string{"POST /declared/:id", "GET /grouped/bare"}, routetable.Undeclared(router.Router().Routes(), declared))
}

func TestMountRefusesAPIKeysOnAPathCodeRoute(t *testing.T) {
	router := newRouter()
	require.NoError(t, routetable.Mount(router.Group("/api"), nil, []routetable.Route{
		{Method: http.MethodGet, Path: "/code/:code", Endpoint: ok, Access: routetable.PathCode},
	}))

	user := &identity.Identity{ID: "user"}
	apiKey := &identity.Identity{ID: "user", TenantID: "tenant", Role: authorizer.RoleOwner, APIKey: "key"}

	assert.Equal(t, http.StatusOK, serve(t, router, http.MethodGet, "/api/code/abc", user, nil))
	assert.Equal(t, http.StatusForbidden, serve(t, router, http.MethodGet, "/api/code/abc", apiKey, nil))
}

func TestMountOpensOnlyDeviceTokenRoutesToADevice(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	bearer, err := jwttoken.EncodeDeviceClaims(authorizer.DeviceClaims{UID: "device", TenantID: "tenant"}, "http://localhost", privateKey, clock.New().Now())
	require.NoError(t, err)

	service := servicemocks.NewMockService(t)
	service.On("PublicKey").Return(&privateKey.PublicKey)
	service.On("AuthDeviceToken", mock.Anything, mock.Anything).Return(nil)

	authn := middleware.NewAuthenticator(service)

	router := newRouter()
	router.Use(authn.Middleware)
	require.NoError(t, routetable.Mount(router.Group("/api"), authn, []routetable.Route{
		{Method: http.MethodPost, Path: "/device", Endpoint: ok, Access: routetable.DeviceToken},
		{Method: http.MethodPost, Path: "/account", Endpoint: ok, Access: routetable.OwnAccount},
	}))

	header := http.Header{"Authorization": []string{"Bearer " + bearer}}

	assert.Equal(t, http.StatusOK, serve(t, router, http.MethodPost, "/api/device", nil, header))
	assert.Equal(t, http.StatusForbidden, serve(t, router, http.MethodPost, "/api/account", nil, header))
	assert.Empty(t, authn.UnregisteredRoutes(router.Router().Routes()))
}

func TestMountRefusesACallerNamingAnotherNamespaceInThePath(t *testing.T) {
	router := newRouter()
	require.NoError(t, routetable.Mount(router.Group("/api"), nil, []routetable.Route{
		{Method: http.MethodGet, Path: "/namespaces/:" + routetable.TenantParam, Endpoint: ok, Access: routetable.NamespaceMember},
		{Method: http.MethodGet, Path: "/namespaces/:" + routetable.TenantParam + "/leave", Endpoint: ok, Access: routetable.OwnAccount},
		{Method: http.MethodGet, Path: "/admin/namespaces/:" + routetable.TenantParam, Endpoint: ok, Access: routetable.Admin(allowAll)},
	}))

	caller := &identity.Identity{ID: "user", TenantID: "tenant", Role: authorizer.RoleObserver}

	assert.Equal(t, http.StatusOK, serve(t, router, http.MethodGet, "/api/namespaces/tenant", caller, nil))
	assert.Equal(t, http.StatusForbidden, serve(t, router, http.MethodGet, "/api/namespaces/other", caller, nil))
	assert.Equal(t, http.StatusOK, serve(t, router, http.MethodGet, "/api/namespaces/other/leave", caller, nil))
	assert.Equal(t, http.StatusOK, serve(t, router, http.MethodGet, "/api/admin/namespaces/other", caller, nil))
}

func TestMountHandsEachEndpointTheClaimsItsRowStates(t *testing.T) {
	answeringClaims := func(claims gateway.Claims) echo.HandlerFunc {
		return func(c *echo.Context) error {
			return c.String(http.StatusOK, fmt.Sprintf("%+v", claims))
		}
	}

	router := newRouter()
	require.NoError(t, routetable.Mount(router.Group("/api"), nil, []routetable.Route{
		{Method: http.MethodGet, Path: "/anonymous", Endpoint: answeringClaims, Access: routetable.Anonymous},
		{Method: http.MethodGet, Path: "/unbounded", Endpoint: answeringClaims, Access: routetable.Anonymous, Unbounded: true},
		{Method: http.MethodGet, Path: "/member", Endpoint: answeringClaims, Access: routetable.NamespaceMember},
	}))

	owner := identity.Identity{ID: "user", TenantID: "tenant", Role: authorizer.RoleOwner}

	cases := []struct {
		path   string
		claims gateway.Claims
	}{
		{path: "/api/anonymous", claims: gateway.Claims{Anonymous: true}},
		{path: "/api/unbounded", claims: gateway.Claims{Anonymous: true, Unbounded: true}},
		{path: "/api/member", claims: gateway.Claims{}},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			req := identitytest.As(httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.path, nil), owner)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)

			assert.Equal(t, fmt.Sprintf("%+v", tc.claims), rec.Body.String())
		})
	}
}
