package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestRole(t *testing.T) {
	cases := []struct {
		description string
		headers     map[string]string
		expected    authorizer.Role
	}{
		{
			description: "verify if given value returns from header",
			headers: map[string]string{
				"X-Role": authorizer.RoleOwner.String(),
			},
			expected: authorizer.Role("owner"),
		}, {
			description: "validate empty headers, for fail function",
			headers:     map[string]string{},
			expected:    authorizer.RoleInvalid,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", nil)

			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			ctxNew := Context{
				nil,
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
		headers     map[string]string
		expected    *models.Tenant
	}{
		{
			description: "verify if given value returns from header",
			headers: map[string]string{
				"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
			},
			expected: &models.Tenant{
				ID: "00000000-0000-4000-0000-000000000000",
			},
		}, {
			description: "validate empty headers, for fail function",
			headers:     map[string]string{},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", nil)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			ctxNew := Context{
				nil,
				ctx,
			}

			tenant := ctxNew.Tenant()

			require.Equal(t, tc.expected, tenant)
		})
	}
}

func TestUsername(t *testing.T) {
	cases := []struct {
		description string
		headers     map[string]string
		expected    *models.Username
	}{
		{
			description: "verify if given value returns from header",
			headers: map[string]string{
				"X-Username": "someone",
			},
			expected: &models.Username{
				ID: "someone",
			},
		}, {
			description: "validate empty headers, for fail function",
			headers:     map[string]string{},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", nil)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			ctxNew := Context{
				nil,
				ctx,
			}

			user := ctxNew.Username()

			require.Equal(t, tc.expected, user)
		})
	}
}

func TestID(t *testing.T) {
	cases := []struct {
		description string
		headers     map[string]string
		expected    *models.ID
	}{
		{
			description: "verify if given value returns from header",
			headers: map[string]string{
				"X-ID": "507f191e810c19729de860ea",
			},
			expected: &models.ID{
				ID: "507f191e810c19729de860ea",
			},
		}, {
			description: "validate empty headers, for fail function",
			headers:     map[string]string{},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", nil)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			ctxNew := Context{
				nil,
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
		headers     map[string]string
		expected    Expected
	}{
		{
			description: "verify if given value returns from header",
			headers: map[string]string{
				"X-ID": "507f191e810c19729de860ea",
			},
			expected: Expected{
				id: "507f191e810c19729de860ea",
				ok: true,
			},
		}, {
			description: "validate empty headers, for fail function",
			headers:     map[string]string{},
			expected: Expected{
				id: "",
				ok: false,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", nil)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			ctxNew := Context{
				nil,
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
		headers     map[string]string
		expected    Expected
	}{
		{
			description: "verify if given value returns from header",
			headers: map[string]string{
				"X-Tenant-ID": "00000000-0000-4000-0000-000000000000",
			},
			expected: Expected{
				user: "00000000-0000-4000-0000-000000000000",
				ok:   true,
			},
		}, {
			description: "validate empty headers, for fail function",
			headers:     map[string]string{},
			expected: Expected{
				user: "",
				ok:   false,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", nil)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			ctxNew := Context{
				nil,
				ctx,
			}

			userNew, validate := ctxNew.GetTennat()

			require.Equal(t, tc.expected, Expected{userNew, validate})
		})
	}
}

func TestGetUsername(t *testing.T) {
	type Expected struct {
		user string
		ok   bool
	}
	cases := []struct {
		description string
		headers     map[string]string
		expected    Expected
	}{
		{
			description: "verify if given value returns from header",
			headers: map[string]string{
				"X-Username": "someone",
			},
			expected: Expected{
				user: "someone",
				ok:   true,
			},
		}, {
			description: "validate empty headers, for fail function",
			headers:     map[string]string{},
			expected: Expected{
				user: "",
				ok:   false,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/", nil)

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			ctxNew := Context{
				nil,
				ctx,
			}

			userNew, validate := ctxNew.GetUsername()

			require.Equal(t, tc.expected, Expected{userNew, validate})
		})
	}
}
