package routes

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/shellhub-io/shellhub/pkg/api/identity/identitytest"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/server/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
	serviceMocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	sshhttp "github.com/shellhub-io/shellhub/server/ssh/http"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	"github.com/stretchr/testify/assert"
)

var editions = []envs.Edition{envs.Community, envs.Enterprise, envs.Cloud}

var pathParam = regexp.MustCompile(`:[^/]+`)

func tableFor(t *testing.T, edition envs.Edition) []route {
	t.Helper()

	handler := NewHandler(serviceMocks.NewMockService(t))
	handler.edition = edition

	return handler.publicRoutes()
}

var mountedOutsideTheTable = map[string]string{
	echo.RouteAny + " /mcp":         "the MCP server replays each tool call against the table's own routes",
	echo.RouteAny + " /mcp/*":       "the MCP server replays each tool call against the table's own routes",
	"GET " + InternalMetricsURL:     "Prometheus scrapes it from inside the cluster",
	"GET /ssh/connection":           "the SSH server's web terminal, mounted by sshhttp.Register",
	"GET /ssh/revdial":              "the agent's reverse dial, mounted by sshhttp.Register",
	"GET /agent/connection":         "the agent's connection, mounted by sshhttp.Register",
	"POST /api/sessions/:uid/close": "the SSH server closes its own sessions, mounted by sshhttp.Register",
}

func address(method, path string) string {
	return method + " " + pathParam.ReplaceAllString(path, ":")
}

func unstatedAccess(table []route) []string {
	var found []string

	for _, r := range table {
		switch {
		case r.access.kind == accessUnstated:
			found = append(found, r.method+" "+r.path+" states no access")
		case r.access.kind == accessPermission && len(r.access.permissions) == 0:
			found = append(found, r.method+" "+r.path+" requires no permission")
		}
	}

	return found
}

func duplicateAddresses(table []route) []string {
	var found []string

	seen := make(map[string]bool)
	for _, r := range table {
		key := address(r.method, r.path)
		if seen[key] {
			found = append(found, r.method+" "+r.path+" is mounted twice")
		}

		seen[key] = true
	}

	return found
}

func mountedOutside(mounted echo.Routes, table []route, exempt map[string]string) []string {
	listed := make(map[string]bool)
	for _, r := range table {
		listed[r.method+" "+publicAPIPrefix+r.path] = true
	}

	var found []string

	for _, m := range mounted {
		key := m.Method + " " + m.Path
		if !listed[key] && exempt[key] == "" {
			found = append(found, key+" is mounted outside the route table")
		}
	}

	return found
}

func TestEveryRouteStatesItsAccess(t *testing.T) {
	for _, edition := range editions {
		assert.Empty(t, unstatedAccess(tableFor(t, edition)), edition)
	}
}

func TestUnstatedAccessFlagsARowThatStatesNone(t *testing.T) {
	table := []route{
		{method: http.MethodGet, path: "/unstated"},
		{method: http.MethodGet, path: "/no-permission", access: requires()},
		{method: http.MethodGet, path: "/anonymous", access: anonymous},
		{method: http.MethodGet, path: "/permission", access: requires(authorizer.DeviceRename)},
	}

	assert.Equal(t, []string{
		"GET /unstated states no access",
		"GET /no-permission requires no permission",
	}, unstatedAccess(table))
}

func TestNoRouteIsMountedTwice(t *testing.T) {
	for _, edition := range editions {
		assert.Empty(t, duplicateAddresses(tableFor(t, edition)), edition)
	}
}

func TestDuplicateAddressesFlagsTheSameAddressUnderAnotherParameterName(t *testing.T) {
	table := []route{
		{method: http.MethodGet, path: "/devices/:uid"},
		{method: http.MethodGet, path: "/devices/:id"},
		{method: http.MethodPost, path: "/devices/:id"},
	}

	assert.Equal(t, []string{"GET /devices/:id is mounted twice"}, duplicateAddresses(table))
}

func TestEveryRouteTheRouterMountsIsInTheTableOrExemptByName(t *testing.T) {
	for _, edition := range editions {
		service := serviceMocks.NewMockService(t)
		authn := routesmiddleware.NewAuthenticator(service)
		router := newTestRouter(t, service, WithAuthentication(authn), WithEdition(edition))
		sshhttp.Register(router, authn, dialer.NewDialer(nil, nil, clock.New()), service, session.NewRegistry(), &sshhttp.Config{})

		mounted := router.Router().Routes()
		assert.Empty(t, mountedOutside(mounted, tableFor(t, edition), mountedOutsideTheTable), edition)

		for exempt := range mountedOutsideTheTable {
			assert.True(t, slices.ContainsFunc(mounted, func(m echo.RouteInfo) bool {
				return m.Method+" "+m.Path == exempt
			}), "%s: %s is exempt but not mounted", edition, exempt)
		}
	}
}

func TestMountedOutsideFlagsARouteNeitherListedNorExempt(t *testing.T) {
	router := echo.New()
	for _, path := range []string{"/api/listed", "/api/unlisted", "/exempt", "/unexempt"} {
		router.GET(path, func(*echo.Context) error { return nil })
	}

	table := []route{{method: http.MethodGet, path: "/listed"}}
	exempt := map[string]string{"GET /exempt": "a reason"}

	assert.ElementsMatch(t, []string{
		"GET /api/unlisted is mounted outside the route table",
		"GET /unexempt is mounted outside the route table",
	}, mountedOutside(router.Router().Routes(), table, exempt))
}

func TestEveryGuardedRouteRefusesBeforeItsHandler(t *testing.T) {
	const tenant = "00000000-0000-4000-0000-000000000000"

	const otherTenant = "11111111-1111-4111-8111-111111111111"

	cases := []struct {
		description string
		applies     func(route) bool
		headers     map[string]string
		caller      identity.Identity
		pathTenant  string
	}{
		{
			description: "an owner naming another namespace in the path",
			applies: func(r route) bool {
				return namesTheTenant(r) && r.access.kind != accessOwnAccount
			},
			caller:     identity.Identity{ID: "user-id", TenantID: tenant, Role: authorizer.RoleOwner},
			pathTenant: otherTenant,
		},
		{
			description: "an API key where the route blocks one",
			applies: func(r route) bool {
				return r.noAPIKey || r.access.kind == accessOwnAccount || r.access.kind == accessPathCode
			},
			headers: map[string]string{"X-API-Key": "key"},
			caller:  identity.Identity{APIKey: "key", TenantID: tenant, Role: authorizer.RoleOwner},
		},
		{
			description: "a user token with no namespace on a member route",
			applies: func(r route) bool {
				return r.access.kind == accessNamespaceMember && !namesTheTenant(r)
			},
			caller: identity.Identity{ID: "user-id"},
		},
		{
			description: "a namespace member whose role holds no permission",
			applies: func(r route) bool {
				return r.access.kind == accessPermission
			},
			caller: identity.Identity{ID: "user-id", TenantID: tenant},
		},
		{
			description: "a user token with neither a namespace nor a permission",
			applies: func(r route) bool {
				return r.access.kind == accessPermission
			},
			caller: identity.Identity{ID: "user-id"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var covered int

			for _, edition := range editions {
				var reached bool
				recording := func(gateway.Claims) echo.HandlerFunc {
					return func(*echo.Context) error {
						reached = true

						return nil
					}
				}

				table := tableFor(t, edition)
				for i := range table {
					table[i].endpoint = recording
				}

				router := echo.New()
				router.HTTPErrorHandler = handlers.NewErrors(nil)
				router.Use(gateway.WithContext())
				mount(router.Group(publicAPIPrefix), table)

				for _, r := range table {
					if !tc.applies(r) {
						continue
					}

					covered++
					reached = false

					target := publicAPIPrefix + pathParam.ReplaceAllStringFunc(r.path, func(param string) string {
						if param[1:] == ParamNamespaceTenant {
							if tc.pathTenant != "" {
								return tc.pathTenant
							}

							return tenant
						}

						return "placeholder"
					})

					req := httptest.NewRequestWithContext(t.Context(), r.method, target, nil)
					for name, value := range tc.headers {
						req.Header.Set(name, value)
					}
					req = identitytest.As(req, tc.caller)

					rec := httptest.NewRecorder()
					router.ServeHTTP(rec, req)

					assert.False(t, reached, "%s: %s %s reached its handler", edition, r.method, r.path)
					assert.Equal(t, http.StatusForbidden, rec.Code, "%s: %s %s", edition, r.method, r.path)
				}
			}

			assert.NotZero(t, covered)
		})
	}
}

func TestAuthenticationRoutesServeAnonymousCallers(t *testing.T) {
	router, _, _ := authenticatedRouter(t)

	for _, path := range []string{AuthDeviceURL, AuthDeviceURLV2, AuthLocalUserURL, AuthLocalUserURLV2} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, publicAPIPrefix+path, strings.NewReader("{}"))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

func namesTheTenant(r route) bool {
	return slices.Contains(strings.Split(r.path, "/"), ":"+ParamNamespaceTenant)
}

func TestARouteActingOnTheCallersAccountReachesItsHandlerForAnyNamespace(t *testing.T) {
	const (
		tenant      = "00000000-0000-4000-0000-000000000000"
		otherTenant = "11111111-1111-4111-8111-111111111111"
	)

	var covered int

	for _, edition := range editions {
		var reached bool
		recording := func(gateway.Claims) echo.HandlerFunc {
			return func(c *echo.Context) error {
				reached = true

				return c.NoContent(http.StatusOK)
			}
		}

		table := tableFor(t, edition)
		for i := range table {
			table[i].endpoint = recording
		}

		router := echo.New()
		router.HTTPErrorHandler = handlers.NewErrors(nil)
		router.Use(gateway.WithContext())
		mount(router.Group(publicAPIPrefix), table)

		for _, r := range table {
			if r.access.kind != accessOwnAccount || !namesTheTenant(r) {
				continue
			}

			covered++
			reached = false

			target := publicAPIPrefix + pathParam.ReplaceAllStringFunc(r.path, func(param string) string {
				if param[1:] == ParamNamespaceTenant {
					return otherTenant
				}

				return "placeholder"
			})

			req := httptest.NewRequestWithContext(t.Context(), r.method, target, nil)
			req = identitytest.As(req, identity.Identity{ID: "user-id", TenantID: tenant, Role: authorizer.RoleOwner})

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.True(t, reached, "%s: %s %s answered %d", edition, r.method, r.path, rec.Code)
		}
	}

	assert.NotZero(t, covered)
}
