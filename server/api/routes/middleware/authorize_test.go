package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/shellhub-io/shellhub/pkg/api/identity/identitytest"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/stretchr/testify/assert"
)

func TestRequiresTenant(t *testing.T) {
	const tenant = "00000000-0000-4000-0000-000000000000"

	cases := []struct {
		description  string
		paramValue   string
		callerTenant string
		useGateway   bool
		expected     int
	}{
		{
			description:  "allows when caller tenant matches path param",
			paramValue:   tenant,
			callerTenant: tenant,
			useGateway:   true,
			expected:     http.StatusOK,
		},
		{
			description:  "blocks when caller tenant does not match path param",
			paramValue:   tenant,
			callerTenant: "7e7389a9-55be-4e14-8c47-817a1552774f",
			useGateway:   true,
			expected:     http.StatusForbidden,
		},
		{
			description: "blocks when caller tenant is missing",
			paramValue:  tenant,
			useGateway:  true,
			expected:    http.StatusForbidden,
		},
		{
			description:  "blocks when path param is empty",
			callerTenant: tenant,
			useGateway:   true,
			expected:     http.StatusForbidden,
		},
		{
			description:  "blocks when the gateway middleware is not installed",
			paramValue:   tenant,
			callerTenant: tenant,
			useGateway:   false,
			expected:     http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
			req = identitytest.As(req, identity.Identity{TenantID: tc.callerTenant})
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPathValues(echo.PathValues{{Name: "tenant", Value: tc.paramValue}})

			next := func(*echo.Context) error { return c.NoContent(http.StatusOK) }

			handler := RequiresTenant("tenant")(next)
			if tc.useGateway {
				handler = gateway.WithContext()(handler)
			}

			renderRefusal(t, c, handler(c))

			assert.Equal(t, tc.expected, rec.Result().StatusCode)

			if tc.expected != http.StatusOK {
				assertRefusalCarriesABody(t, rec)
			}
		})
	}
}

func TestRequiresAnyPermission(t *testing.T) {
	cases := []struct {
		description string
		role        authorizer.Role
		permissions []authorizer.Permission
		noGateway   bool
		expected    int
	}{
		{
			description: "admits a role holding the first permission",
			role:        authorizer.RoleOperator,
			permissions: []authorizer.Permission{authorizer.SSHIdentityAdd, authorizer.SSHIdentityManage},
			expected:    http.StatusOK,
		},
		{
			description: "refuses when the gateway middleware is not installed",
			role:        authorizer.RoleOwner,
			permissions: []authorizer.Permission{authorizer.SSHIdentityAdd},
			noGateway:   true,
			expected:    http.StatusForbidden,
		},
		{
			description: "admits a role holding only a later permission",
			role:        authorizer.RoleOperator,
			permissions: []authorizer.Permission{authorizer.SSHIdentityManage, authorizer.SSHIdentityAdd},
			expected:    http.StatusOK,
		},
		{
			description: "refuses a role holding none of them",
			role:        authorizer.RoleObserver,
			permissions: []authorizer.Permission{authorizer.SSHIdentityAdd, authorizer.SSHIdentityManage},
			expected:    http.StatusForbidden,
		},
		{
			description: "refuses every role when given no permissions",
			role:        authorizer.RoleOwner,
			expected:    http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
			req = identitytest.As(req, identity.Identity{Role: tc.role})
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			next := func(*echo.Context) error { return c.NoContent(http.StatusOK) }

			handler := RequiresAnyPermission(tc.permissions...)(next)
			if !tc.noGateway {
				handler = gateway.WithContext()(handler)
			}

			renderRefusal(t, c, handler(c))

			assert.Equal(t, tc.expected, rec.Result().StatusCode)
		})
	}
}
