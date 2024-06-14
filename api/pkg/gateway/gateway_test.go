package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestTenantFromContext(t *testing.T) {
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
			description: "validate empty tenant string, for ID behaviour",
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

			ctxArg := context.WithValue(context.TODO(), "ctx", &ctxNew) // nolint:revive

			tenant := TenantFromContext(ctxArg)

			require.Equal(t, tc.expected, tenant)
		})
	}
}

func TestUsernameFromContext(t *testing.T) {
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
			description: "validate empty username string, for ID behaviour",
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

			ctxArg := context.WithValue(context.TODO(), "ctx", &ctxNew) // nolint:revive

			user := UsernameFromContext(ctxArg)

			require.Equal(t, tc.expected, user)
		})
	}
}

func TestIDFromContext(t *testing.T) {
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
			description: "validate empty ID string, for ID behaviour",
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

			ctxArg := context.WithValue(context.TODO(), "ctx", &ctxNew) // nolint:revive

			id := IDFromContext(ctxArg)

			require.Equal(t, tc.expected, id)
		})
	}
}
