package gateway_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/shellhub-io/shellhub/pkg/api/identity/identitytest"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	userIdentity = identity.Identity{
		ID: "user-id", Username: "username", TenantID: probeTenant, Role: authorizer.RoleAdministrator,
	}
	apiKeyIdentity = identity.Identity{
		TenantID: probeTenant, Role: authorizer.RoleOperator, APIKey: "key", APIKeyID: "key-id",
	}
	deviceIdentity  = identity.Identity{TenantID: probeTenant, DeviceUID: "device-uid"}
	consoleIdentity = identity.Identity{Admin: true, AdminID: "admin-id"}
	instanceKey     = identity.Identity{Admin: true, APIKey: models.InstanceAPIKeyPrefix + "key"}
	adminOnAPI      = identity.Identity{ID: "admin-id", TenantID: probeTenant, Role: authorizer.RoleOwner, Admin: true}
)

func serveCaller[A gateway.Caller](t *testing.T, caller *identity.Identity, claims gateway.Claims) (A, bool, int) {
	t.Helper()

	var (
		got    A
		called bool
	)

	endpoint := gateway.None(func(_ context.Context, _ scope.Scope, c A, _ *probeRequest) error {
		got = c
		called = true

		return nil
	})

	e := probeRouter(t, true)
	e.GET("/probe", endpoint.Handler(claims))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/probe", nil)
	if caller != nil {
		req = identitytest.As(req, *caller)
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	return got, called, rec.Code
}

func TestShapesHandTheHandlerTheCallerItsTypeNames(t *testing.T) {
	t.Run("principal", func(t *testing.T) {
		cases := []struct {
			description string
			caller      identity.Identity
			expected    gateway.Principal
			member      bool
		}{
			{
				description: "a user acts as a principal with a person behind it",
				caller:      userIdentity,
				expected:    gateway.Principal{Role: authorizer.RoleAdministrator, UserID: "user-id", Username: "username"},
				member:      true,
			},
			{
				description: "an API key acts as a principal with no person behind it",
				caller:      apiKeyIdentity,
				expected:    gateway.Principal{Role: authorizer.RoleOperator, APIKeyID: "key-id"},
			},
			{
				description: "an administrator on the regular surface acts as a principal",
				caller:      adminOnAPI,
				expected:    gateway.Principal{Role: authorizer.RoleOwner, UserID: "admin-id"},
				member:      true,
			},
		}

		for _, tc := range cases {
			t.Run(tc.description, func(t *testing.T) {
				got, called, status := serveCaller[gateway.Principal](t, &tc.caller, gateway.Claims{})

				require.Equal(t, http.StatusOK, status)
				require.True(t, called)
				assert.Equal(t, tc.expected, got)

				_, isMember := got.Member()
				assert.Equal(t, tc.member, isMember)
			})
		}
	})

	t.Run("a principal backed by a person names that member", func(t *testing.T) {
		got, _, _ := serveCaller[gateway.Principal](t, &userIdentity, gateway.Claims{})

		member, ok := got.Member()
		require.True(t, ok)
		assert.Equal(t, gateway.Member{UserID: "user-id", Username: "username", Role: authorizer.RoleAdministrator}, member)
	})

	t.Run("member", func(t *testing.T) {
		got, called, status := serveCaller[gateway.Member](t, &userIdentity, gateway.Claims{})

		require.Equal(t, http.StatusOK, status)
		require.True(t, called)
		assert.Equal(t, gateway.Member{UserID: "user-id", Username: "username", Role: authorizer.RoleAdministrator}, got)
	})

	t.Run("admin", func(t *testing.T) {
		cases := []struct {
			description string
			caller      identity.Identity
			expected    gateway.Admin
		}{
			{
				description: "an administrator on the console is named by the admin ID",
				caller:      consoleIdentity,
				expected:    gateway.Admin{UserID: "admin-id"},
			},
			{
				description: "the instance API key acts as an administrator with no person behind it",
				caller:      instanceKey,
				expected:    gateway.Admin{InstanceAPIKey: true},
			},
		}

		for _, tc := range cases {
			t.Run(tc.description, func(t *testing.T) {
				got, called, status := serveCaller[gateway.Admin](t, &tc.caller, gateway.Claims{Unbounded: true})

				require.Equal(t, http.StatusOK, status)
				require.True(t, called)
				assert.Equal(t, tc.expected, got)
			})
		}
	})

	t.Run("anonymous admits a request with no identity", func(t *testing.T) {
		_, called, status := serveCaller[gateway.Anonymous](t, nil, gateway.Claims{Unbounded: true})

		assert.Equal(t, http.StatusOK, status)
		assert.True(t, called)
	})

	t.Run("anonymous admits a request that carries an identity", func(t *testing.T) {
		_, called, status := serveCaller[gateway.Anonymous](t, &userIdentity, gateway.Claims{Unbounded: true})

		assert.Equal(t, http.StatusOK, status)
		assert.True(t, called)
	})
}

func TestShapesRefuseACallerOfAnotherKind(t *testing.T) {
	type refusal struct {
		description string
		caller      identity.Identity
	}

	cases := []struct {
		kind     string
		serve    func(*testing.T, *identity.Identity) (bool, int)
		refusals []refusal
	}{
		{
			kind: "principal",
			serve: func(t *testing.T, caller *identity.Identity) (bool, int) {
				t.Helper()
				_, called, status := serveCaller[gateway.Principal](t, caller, gateway.Claims{Unbounded: true})

				return called, status
			},
			refusals: []refusal{
				{description: "a device", caller: deviceIdentity},
				{description: "an administrator on the console", caller: consoleIdentity},
				{description: "the instance API key", caller: instanceKey},
				{description: "an identity naming nobody", caller: identity.Identity{TenantID: probeTenant}},
			},
		},
		{
			kind: "member",
			serve: func(t *testing.T, caller *identity.Identity) (bool, int) {
				t.Helper()
				_, called, status := serveCaller[gateway.Member](t, caller, gateway.Claims{Unbounded: true})

				return called, status
			},
			refusals: []refusal{
				{description: "an API key", caller: apiKeyIdentity},
				{description: "a user identity carrying an API key", caller: identity.Identity{ID: "user-id", TenantID: probeTenant, APIKey: "key"}},
				{description: "a device", caller: deviceIdentity},
				{description: "an administrator on the console", caller: consoleIdentity},
				{description: "the instance API key", caller: instanceKey},
			},
		},
		{
			kind: "admin",
			serve: func(t *testing.T, caller *identity.Identity) (bool, int) {
				t.Helper()
				_, called, status := serveCaller[gateway.Admin](t, caller, gateway.Claims{Unbounded: true})

				return called, status
			},
			refusals: []refusal{
				{description: "a user", caller: userIdentity},
				{description: "an administrator on the regular surface", caller: adminOnAPI},
				{description: "an API key", caller: apiKeyIdentity},
				{description: "a device", caller: deviceIdentity},
				{description: "a user stripped of scope without the admin flag", caller: identity.Identity{AdminID: "admin-id"}},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			for _, r := range tc.refusals {
				t.Run(r.description, func(t *testing.T) {
					called, status := tc.serve(t, &r.caller)

					assert.Equal(t, http.StatusForbidden, status)
					assert.False(t, called)
				})
			}

			t.Run("a request with no identity", func(t *testing.T) {
				called, status := tc.serve(t, nil)

				assert.Equal(t, http.StatusUnauthorized, status)
				assert.False(t, called)
			})
		})
	}
}

func TestEndpointsNameTheCallerTheirHandlerTakes(t *testing.T) {
	cases := []struct {
		endpoint gateway.Endpoint
		expected gateway.CallerKind
	}{
		{
			endpoint: gateway.None(func(context.Context, scope.Scope, gateway.Anonymous, *probeRequest) error { return nil }),
			expected: gateway.CallerAnonymous,
		},
		{
			endpoint: gateway.One(func(context.Context, scope.Scope, gateway.Principal, *probeRequest) (string, error) { return "", nil }),
			expected: gateway.CallerPrincipal,
		},
		{
			endpoint: gateway.List(func(context.Context, scope.Scope, gateway.Member, *probeRequest) ([]string, int64, error) {
				return nil, 0, nil
			}),
			expected: gateway.CallerMember,
		},
		{
			endpoint: gateway.None(func(context.Context, scope.Scope, gateway.Admin, *probeRequest) error { return nil }),
			expected: gateway.CallerAdmin,
		},
		{
			endpoint: gateway.Legacy(func(*gateway.Context) error { return nil }),
			expected: gateway.CallerUnstated,
		},
		{
			endpoint: gateway.Raw(func(*echo.Context) error { return nil }),
			expected: gateway.CallerUnstated,
		},
	}

	for _, tc := range cases {
		t.Run(tc.expected.String(), func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.endpoint.Caller())
		})
	}
}
