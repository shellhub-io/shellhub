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
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/server/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway/gatewaytest"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
	"github.com/shellhub-io/shellhub/server/api/routes/routetable"
	serviceMocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var editions = []envs.Edition{envs.Community, envs.Enterprise, envs.Cloud}

var pathParam = regexp.MustCompile(`:[^/]+`)

var (
	ownAccount      = routetable.OwnAccount.String()
	pathCode        = routetable.PathCode.String()
	namespaceMember = routetable.NamespaceMember.String()
	permission      = routetable.Requires(authorizer.DeviceRename).String()
	deviceToken     = routetable.DeviceToken.String()
	anonymous       = routetable.Anonymous.String()
)

func tableFor(t *testing.T, edition envs.Edition) []routetable.Route {
	t.Helper()

	handler := NewHandler(serviceMocks.NewMockService(t))
	handler.edition = edition

	return handler.publicRoutes()
}

func address(method, path string) string {
	return method + " " + pathParam.ReplaceAllString(path, ":")
}

func duplicateAddresses(table []routetable.Route) []string {
	var found []string

	seen := make(map[string]bool)
	for _, r := range table {
		key := address(r.Method, r.Path)
		if seen[key] {
			found = append(found, r.Method+" "+r.Path+" is mounted twice")
		}

		seen[key] = true
	}

	return found
}

func mountedWithRecordingEndpoints(t *testing.T, table []routetable.Route, authn *routesmiddleware.Authenticator, reached *bool) *echo.Echo {
	t.Helper()

	for i := range table {
		table[i].Endpoint = gatewaytest.Recording(table[i].Endpoint.Caller(), reached)
	}

	router := echo.New()
	router.Binder = handlers.NewBinder()
	router.Validator = handlers.NewValidator()
	router.HTTPErrorHandler = handlers.NewErrors(nil)
	router.Use(gateway.WithContext())

	if authn != nil {
		router.Use(authn.Middleware)
	}

	require.NoError(t, routetable.Mount(router.Group(publicAPIPrefix), authn, table))

	return router
}

func targetOf(r routetable.Route, tenant string) string {
	return publicAPIPrefix + pathParam.ReplaceAllStringFunc(r.Path, func(param string) string {
		if param[1:] == ParamNamespaceTenant {
			return tenant
		}

		return "placeholder"
	})
}

func TestEveryRouteStatesItsAccess(t *testing.T) {
	for _, edition := range editions {
		assert.NoError(t, routetable.Mount(echo.New().Group(publicAPIPrefix), nil, tableFor(t, edition)), edition)
	}
}

func TestNoRouteIsMountedTwice(t *testing.T) {
	for _, edition := range editions {
		assert.Empty(t, duplicateAddresses(tableFor(t, edition)), edition)
	}
}

func TestDuplicateAddressesFlagsTheSameAddressUnderAnotherParameterName(t *testing.T) {
	table := []routetable.Route{
		{Method: http.MethodGet, Path: "/devices/:uid"},
		{Method: http.MethodGet, Path: "/devices/:id"},
		{Method: http.MethodPost, Path: "/devices/:id"},
	}

	assert.Equal(t, []string{"GET /devices/:id is mounted twice"}, duplicateAddresses(table))
}

func TestEveryRouteTheRouterMountsIsInATableOrDeclared(t *testing.T) {
	for _, edition := range editions {
		service := serviceMocks.NewMockService(t)
		router := newTestRouter(t, service, WithAuthentication(routesmiddleware.NewAuthenticator(service)), WithEdition(edition))

		mounted := router.Router().Routes()
		assert.Empty(t, routetable.Undeclared(mounted, Declared()), edition)

		for declared := range Declared() {
			assert.True(t, slices.ContainsFunc(mounted, func(m echo.RouteInfo) bool {
				return routetable.Address(m.Method, m.Path) == declared
			}), "%s: %s is declared but not mounted", edition, declared)
		}
	}
}

func TestEveryGuardedRouteRefusesBeforeItsHandler(t *testing.T) {
	const tenant = "00000000-0000-4000-0000-000000000000"

	const otherTenant = "11111111-1111-4111-8111-111111111111"

	cases := []struct {
		description string
		applies     func(routetable.Route) bool
		headers     map[string]string
		caller      identity.Identity
		pathTenant  string
	}{
		{
			description: "an owner naming another namespace in the path",
			applies: func(r routetable.Route) bool {
				return namesTheTenant(r) && r.Access.String() != ownAccount
			},
			caller:     identity.Identity{ID: "user-id", TenantID: tenant, Role: authorizer.RoleOwner},
			pathTenant: otherTenant,
		},
		{
			description: "an API key where the route blocks one",
			applies: func(r routetable.Route) bool {
				return r.NoAPIKey || r.Access.String() == ownAccount || r.Access.String() == pathCode
			},
			headers: map[string]string{"X-API-Key": "key"},
			caller:  identity.Identity{APIKey: "key", TenantID: tenant, Role: authorizer.RoleOwner},
		},
		{
			description: "a user token with no namespace on a member route",
			applies: func(r routetable.Route) bool {
				return r.Access.String() == namespaceMember && !namesTheTenant(r)
			},
			caller: identity.Identity{ID: "user-id"},
		},
		{
			description: "a namespace member whose role holds no permission",
			applies: func(r routetable.Route) bool {
				return r.Access.String() == permission
			},
			caller: identity.Identity{ID: "user-id", TenantID: tenant},
		},
		{
			description: "a user token with neither a namespace nor a permission",
			applies: func(r routetable.Route) bool {
				return r.Access.String() == permission
			},
			caller: identity.Identity{ID: "user-id"},
		},
		{
			description: "a user token on a device-token route",
			applies: func(r routetable.Route) bool {
				return r.Access.String() == deviceToken
			},
			caller: identity.Identity{ID: "user-id", TenantID: tenant, Role: authorizer.RoleOwner},
		},
		{
			description: "an API key on a route whose handler takes a member",
			applies: func(r routetable.Route) bool {
				return r.Endpoint.Caller() == gateway.CallerMember
			},
			headers: map[string]string{"X-API-Key": "key"},
			caller:  identity.Identity{APIKey: "key", TenantID: tenant, Role: authorizer.RoleOwner},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var covered int

			for _, edition := range editions {
				var reached bool

				table := tableFor(t, edition)
				router := mountedWithRecordingEndpoints(t, table, nil, &reached)

				for _, r := range table {
					if !tc.applies(r) {
						continue
					}

					covered++
					reached = false

					pathTenant := tenant
					if tc.pathTenant != "" {
						pathTenant = tc.pathTenant
					}

					req := httptest.NewRequestWithContext(t.Context(), r.Method, targetOf(r, pathTenant), nil)
					for name, value := range tc.headers {
						req.Header.Set(name, value)
					}
					req = identitytest.As(req, tc.caller)

					rec := httptest.NewRecorder()
					router.ServeHTTP(rec, req)

					assert.False(t, reached, "%s: %s %s reached its handler", edition, r.Method, r.Path)
					assert.Equal(t, http.StatusForbidden, rec.Code, "%s: %s %s", edition, r.Method, r.Path)
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

func namesTheTenant(r routetable.Route) bool {
	return slices.Contains(strings.Split(r.Path, "/"), ":"+ParamNamespaceTenant)
}

func TestARouteActingOnTheCallersAccountReachesItsHandlerForAnyNamespace(t *testing.T) {
	const (
		tenant      = "00000000-0000-4000-0000-000000000000"
		otherTenant = "11111111-1111-4111-8111-111111111111"
	)

	var covered int

	for _, edition := range editions {
		var reached bool

		table := tableFor(t, edition)
		router := mountedWithRecordingEndpoints(t, table, nil, &reached)

		for _, r := range table {
			if r.Access.String() != ownAccount || !namesTheTenant(r) {
				continue
			}

			covered++
			reached = false

			req := httptest.NewRequestWithContext(t.Context(), r.Method, targetOf(r, otherTenant), nil)
			req = identitytest.As(req, identity.Identity{ID: "user-id", TenantID: tenant, Role: authorizer.RoleOwner})

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.True(t, reached, "%s: %s %s answered %d", edition, r.Method, r.Path, rec.Code)
		}
	}

	assert.NotZero(t, covered)
}

func TestEveryRouteButADeviceRouteRefusesADevice(t *testing.T) {
	const tenant = "00000000-0000-4000-0000-000000000000"

	privateKey, err := testSigningKey()
	require.NoError(t, err)

	bearer, err := jwttoken.EncodeDeviceClaims(authorizer.DeviceClaims{UID: "device", TenantID: tenant}, "http://localhost", privateKey, clock.New().Now())
	require.NoError(t, err)

	var covered int

	for _, edition := range editions {
		service := serviceMocks.NewMockService(t)
		service.On("PublicKey").Return(&privateKey.PublicKey).Maybe()
		service.On("AuthDeviceToken", mock.Anything, mock.Anything).Return(nil).Maybe()

		var reached bool

		table := tableFor(t, edition)
		router := mountedWithRecordingEndpoints(t, table, routesmiddleware.NewAuthenticator(service), &reached)

		for _, r := range table {
			if access := r.Access.String(); access == deviceToken || access == anonymous {
				continue
			}

			covered++
			reached = false

			req := httptest.NewRequestWithContext(t.Context(), r.Method, targetOf(r, tenant), nil)
			req.Header.Set("Authorization", "Bearer "+bearer)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.False(t, reached, "%s: %s %s reached its handler", edition, r.Method, r.Path)
			assert.Equal(t, http.StatusForbidden, rec.Code, "%s: %s %s", edition, r.Method, r.Path)
		}
	}

	assert.NotZero(t, covered)
}
