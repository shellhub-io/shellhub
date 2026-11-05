package routes

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/server/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	serviceMocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
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

func TestEveryRouteStatesItsAccess(t *testing.T) {
	for _, edition := range editions {
		for _, r := range tableFor(t, edition) {
			assert.NotEqual(t, accessUnstated, r.access.kind, "%s: %s %s states no access", edition, r.method, r.path)

			if r.access.kind == accessPermission {
				assert.NotEmpty(t, r.access.permissions, "%s: %s %s requires no permission", edition, r.method, r.path)
			}
		}
	}
}

func TestNoRouteIsMountedTwice(t *testing.T) {
	for _, edition := range editions {
		seen := make(map[string]bool)

		for _, r := range tableFor(t, edition) {
			address := r.method + " " + pathParam.ReplaceAllString(r.path, ":")
			assert.False(t, seen[address], "%s: %s %s is mounted twice", edition, r.method, r.path)
			seen[address] = true
		}
	}
}

func TestEveryRouteNewRouterMountsOnTheAPIComesFromTheTable(t *testing.T) {
	for _, edition := range editions {
		router := newTestRouter(t, serviceMocks.NewMockService(t), WithEdition(edition))

		table := make(map[string]bool)
		for _, r := range tableFor(t, edition) {
			table[r.method+" "+publicAPIPrefix+r.path] = true
		}

		for _, mounted := range router.Router().Routes() {
			if !strings.HasPrefix(mounted.Path, publicAPIPrefix+"/") {
				continue
			}

			address := mounted.Method + " " + mounted.Path
			assert.True(t, table[address], "%s: %s is mounted outside the route table", edition, address)
		}
	}
}

func TestEveryGuardedRouteRefusesBeforeItsHandler(t *testing.T) {
	const tenant = "00000000-0000-4000-0000-000000000000"

	cases := []struct {
		description string
		applies     func(route) bool
		headers     map[string]string
	}{
		{
			description: "an API key where the route blocks one",
			applies: func(r route) bool {
				return r.noAPIKey || r.access.kind == accessOwnAccount || r.access.kind == accessPathCode
			},
			headers: map[string]string{"X-ID": "user-id", "X-Tenant-ID": tenant, "X-Role": "owner", "X-API-Key": "key"},
		},
		{
			description: "a user token with no namespace on a member route",
			applies: func(r route) bool {
				return r.access.kind == accessNamespaceMember && r.tenantParam == ""
			},
			headers: map[string]string{"X-ID": "user-id"},
		},
		{
			description: "a role holding none of the route's permissions",
			applies: func(r route) bool {
				return r.access.kind == accessPermission
			},
			headers: map[string]string{"X-ID": "user-id", "X-Tenant-ID": tenant},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			var covered int

			for _, edition := range editions {
				for _, r := range tableFor(t, edition) {
					if !tc.applies(r) {
						continue
					}

					covered++

					req := httptest.NewRequestWithContext(t.Context(), r.method, "/", nil)
					for name, value := range tc.headers {
						req.Header.Set(name, value)
					}

					rec := httptest.NewRecorder()
					c := echo.New().NewContext(req, rec)
					if r.tenantParam != "" {
						c.SetPathValues(echo.PathValues{{Name: r.tenantParam, Value: tenant}})
					}

					var reached bool
					next := echo.HandlerFunc(func(*echo.Context) error {
						reached = true

						return nil
					})

					for _, guard := range slices.Backward(r.middleware()) {
						next = guard(next)
					}

					if err := gateway.WithContext()(next)(c); err != nil {
						handlers.NewErrors(nil)(c, err)
					}

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
