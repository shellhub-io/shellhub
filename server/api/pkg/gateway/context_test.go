package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/shellhub-io/shellhub/pkg/api/identity/identitytest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestRole(t *testing.T) {
	cases := []struct {
		description string
		caller      identity.Identity
		expected    authorizer.Role
	}{
		{
			description: "returns the value the identity carries",
			caller:      identity.Identity{Role: authorizer.RoleOwner},
			expected:    authorizer.Role("owner"),
		}, {
			description: "returns the zero value with no identity",
			caller:      identity.Identity{},
			expected:    authorizer.RoleInvalid,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", nil)

			if tc.caller != (identity.Identity{}) {
				req = identitytest.As(req, tc.caller)
			}

			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			ctxNew := Context{
				ctx,
			}

			role := ctxNew.Role()

			require.Equal(t, tc.expected, role)
		})
	}
}

func TestTenant(t *testing.T) {
	cases := []struct {
		description string
		caller      identity.Identity
		expected    *models.Tenant
	}{
		{
			description: "returns the value the identity carries",
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000"},
			expected: &models.Tenant{
				ID: "00000000-0000-4000-0000-000000000000",
			},
		}, {
			description: "returns the zero value with no identity",
			caller:      identity.Identity{},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", nil)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			if tc.caller != (identity.Identity{}) {
				req = identitytest.As(req, tc.caller)
			}

			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			ctxNew := Context{
				ctx,
			}

			tenant := ctxNew.Tenant()

			require.Equal(t, tc.expected, tenant)
		})
	}
}

func TestID(t *testing.T) {
	cases := []struct {
		description string
		caller      identity.Identity
		expected    *models.ID
	}{
		{
			description: "returns the value the identity carries",
			caller:      identity.Identity{ID: "507f191e810c19729de860ea"},
			expected: &models.ID{
				ID: "507f191e810c19729de860ea",
			},
		}, {
			description: "returns the zero value with no identity",
			caller:      identity.Identity{},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", nil)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			if tc.caller != (identity.Identity{}) {
				req = identitytest.As(req, tc.caller)
			}

			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			ctxNew := Context{
				ctx,
			}

			id := ctxNew.ID()

			require.Equal(t, tc.expected, id)
		})
	}
}

func TestGetID(t *testing.T) {
	type Expected struct {
		id string
		ok bool
	}
	cases := []struct {
		description string
		caller      identity.Identity
		expected    Expected
	}{
		{
			description: "returns the value the identity carries",
			caller:      identity.Identity{ID: "507f191e810c19729de860ea"},
			expected: Expected{
				id: "507f191e810c19729de860ea",
				ok: true,
			},
		}, {
			description: "returns the zero value with no identity",
			caller:      identity.Identity{},
			expected: Expected{
				id: "",
				ok: false,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", nil)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			if tc.caller != (identity.Identity{}) {
				req = identitytest.As(req, tc.caller)
			}

			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			ctxNew := Context{
				ctx,
			}

			idNew, validate := ctxNew.GetID()

			require.Equal(t, tc.expected, Expected{idNew, validate})
		})
	}
}

func TestGetTennat(t *testing.T) {
	type Expected struct {
		user string
		ok   bool
	}
	cases := []struct {
		description string
		caller      identity.Identity
		expected    Expected
	}{
		{
			description: "returns the value the identity carries",
			caller:      identity.Identity{TenantID: "00000000-0000-4000-0000-000000000000"},
			expected: Expected{
				user: "00000000-0000-4000-0000-000000000000",
				ok:   true,
			},
		}, {
			description: "returns the zero value with no identity",
			caller:      identity.Identity{},
			expected: Expected{
				user: "",
				ok:   false,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", nil)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			if tc.caller != (identity.Identity{}) {
				req = identitytest.As(req, tc.caller)
			}

			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			ctxNew := Context{
				ctx,
			}

			userNew, validate := ctxNew.GetTennat()

			require.Equal(t, tc.expected, Expected{userNew, validate})
		})
	}
}
