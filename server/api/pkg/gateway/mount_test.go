package gateway_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rootOf(router *echo.Echo) *gateway.Mounter {
	return gateway.MountOn(router, router)
}

func okRoute() gateway.Route {
	return gateway.None(func(_ context.Context, _ scope.Scope, _ gateway.Actor, _ *probeRequest) error {
		return nil
	})
}

func TestMountingCompletesTheDeclarationWithItsAddress(t *testing.T) {
	e := probeRouter(t, true)

	gateway.GET(gateway.MountOn(e, e.Group("/api")), "/devices", okRoute())
	gateway.POST(rootOf(e), "/root", okRoute())

	addresses := make([]string, 0)
	for _, declaration := range gateway.Declarations(e) {
		addresses = append(addresses, declaration.Address())
	}

	assert.Equal(t, []string{"GET /api/devices", "POST /root"}, addresses)

	registered := make([]string, 0)
	for _, route := range e.Router().Routes() {
		registered = append(registered, route.Method+" "+route.Path)
	}

	assert.ElementsMatch(t, addresses, registered)
}

func TestDeclarationsBelongToTheRouterThatMountedThem(t *testing.T) {
	first, second := probeRouter(t, true), probeRouter(t, true)

	gateway.GET(rootOf(first), "/first", okRoute())
	gateway.GET(rootOf(second), "/second", okRoute())

	require.Len(t, gateway.Declarations(first), 1)
	require.Len(t, gateway.Declarations(second), 1)

	assert.Equal(t, "GET /first", gateway.Declarations(first)[0].Address())
	assert.Equal(t, "GET /second", gateway.Declarations(second)[0].Address())
}

func TestRequiresDeclaresThePermitItEnforces(t *testing.T) {
	cases := []struct {
		description    string
		role           string
		expectedStatus int
	}{
		{
			description:    "refuses a role without the permission",
			role:           "observer",
			expectedStatus: http.StatusForbidden,
		},
		{
			description:    "admits a role holding it",
			role:           "owner",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := probeRouter(t, true)
			gateway.GET(rootOf(e), "/probe", okRoute(), gateway.Requires(authorizer.DeviceRemove))

			declarations := gateway.Declarations(e)
			require.Len(t, declarations, 1)
			assert.Equal(t, gateway.AuthorityPermission, declarations[0].Authority)
			assert.Equal(t, []authorizer.Permission{authorizer.DeviceRemove}, declarations[0].Permissions)

			assert.Equal(t, tc.expectedStatus, probe(t, e, map[string]string{
				"X-Tenant-ID": probeTenant,
				"X-ID":        "user-id",
				"X-Role":      tc.role,
			}))
		})
	}
}

func TestNoAPIKeyDeclaresTheBlockItEnforces(t *testing.T) {
	e := probeRouter(t, true)
	gateway.GET(rootOf(e), "/probe", okRoute(), gateway.NoAPIKey())

	declarations := gateway.Declarations(e)
	require.Len(t, declarations, 1)
	assert.True(t, declarations[0].BlocksAPIKey)

	assert.Equal(t, http.StatusForbidden, probe(t, e, map[string]string{
		"X-Tenant-ID": probeTenant,
		"X-API-Key":   "a-key",
	}))

	assert.Equal(t, http.StatusOK, probe(t, e, map[string]string{
		"X-Tenant-ID": probeTenant,
		"X-ID":        "user-id",
	}))
}

func TestGuardsRunInTheOrderTheyAreWritten(t *testing.T) {
	order := make([]string, 0)

	mark := func(name string) echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				order = append(order, name)

				return next(c)
			}
		}
	}

	e := probeRouter(t, true)
	gateway.GET(rootOf(e), "/probe", okRoute(),
		gateway.Guard(mark("first")),
		gateway.NoAPIKey(),
		gateway.Guard(mark("second")),
		gateway.Requires(authorizer.DeviceRemove),
		gateway.Guard(mark("third")))

	require.Equal(t, http.StatusOK, probe(t, e, map[string]string{
		"X-Tenant-ID": probeTenant,
		"X-ID":        "user-id",
		"X-Role":      "owner",
	}))

	assert.Equal(t, []string{"first", "second", "third"}, order)
}

func TestGuardDeclaresNothing(t *testing.T) {
	refuse := func(_ echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			return c.NoContent(http.StatusTeapot)
		}
	}

	e := probeRouter(t, true)
	gateway.GET(rootOf(e), "/probe", okRoute(), gateway.Guard(refuse))

	declarations := gateway.Declarations(e)
	require.Len(t, declarations, 1)
	assert.Equal(t, gateway.AuthorityUnstated, declarations[0].Authority)
	assert.False(t, declarations[0].BlocksAPIKey)

	assert.Equal(t, http.StatusTeapot, probe(t, e, map[string]string{"X-Tenant-ID": probeTenant, "X-ID": "user-id"}))
}

func TestRequiresAnyDeclaresThePermitsItEnforces(t *testing.T) {
	cases := []struct {
		description    string
		permissions    []authorizer.Permission
		role           string
		expectedStatus int
	}{
		{
			description:    "refuses a role holding neither",
			permissions:    []authorizer.Permission{authorizer.NamespaceDelete, authorizer.BillingCreateCustomer},
			role:           "administrator",
			expectedStatus: http.StatusForbidden,
		},
		{
			description:    "admits a role holding only the first",
			permissions:    []authorizer.Permission{authorizer.SessionPlay, authorizer.NamespaceDelete},
			role:           "administrator",
			expectedStatus: http.StatusOK,
		},
		{
			description:    "admits a role holding only the second",
			permissions:    []authorizer.Permission{authorizer.NamespaceDelete, authorizer.SessionPlay},
			role:           "administrator",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := probeRouter(t, true)
			gateway.GET(rootOf(e), "/probe", okRoute(), gateway.RequiresAny(tc.permissions...))

			declarations := gateway.Declarations(e)
			require.Len(t, declarations, 1)
			assert.Equal(t, gateway.AuthorityPermission, declarations[0].Authority)
			assert.Equal(t, tc.permissions, declarations[0].Permissions)

			assert.Equal(t, tc.expectedStatus, probe(t, e, map[string]string{
				"X-Tenant-ID": probeTenant,
				"X-ID":        "user-id",
				"X-Role":      tc.role,
			}))
		})
	}
}

func TestRequiresAnyRefusesEveryoneWhenItNamesNothing(t *testing.T) {
	e := probeRouter(t, true)
	gateway.GET(rootOf(e), "/probe", okRoute(), gateway.RequiresAny())

	assert.Equal(t, http.StatusForbidden, probe(t, e, map[string]string{
		"X-Tenant-ID": probeTenant,
		"X-ID":        "user-id",
		"X-Role":      "owner",
	}))
}

func TestStatedAbsenceOfAuthorityInstallsNothing(t *testing.T) {
	cases := []struct {
		description       string
		option            gateway.RouteOption
		expectedAuthority gateway.Authority
	}{
		{
			description:       "the handler decides",
			option:            gateway.PermissionInHandler("the probe reads the namespace from its path"),
			expectedAuthority: gateway.AuthorityInHandler,
		},
		{
			description:       "the route demands none",
			option:            gateway.NoPermission("the probe acts on the caller's own record"),
			expectedAuthority: gateway.AuthorityNone,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := probeRouter(t, true)
			gateway.GET(rootOf(e), "/probe", okRoute(), tc.option)

			declarations := gateway.Declarations(e)
			require.Len(t, declarations, 1)
			assert.Equal(t, tc.expectedAuthority, declarations[0].Authority)
			assert.NotEmpty(t, declarations[0].AuthorityReason)
			assert.Empty(t, declarations[0].Permissions)

			assert.Equal(t, http.StatusOK, probe(t, e, map[string]string{
				"X-Tenant-ID": probeTenant,
				"X-ID":        "user-id",
				"X-Role":      "observer",
			}))
		})
	}
}

func probe(t *testing.T, e *echo.Echo, headers map[string]string) int {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/probe", nil)
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	return rec.Code
}
