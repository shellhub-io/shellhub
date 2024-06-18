package middleware

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/auth"
)

func Authorize(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := context.WithValue(c.Request().Context(), "ctx", c.(*gateway.Context)) //nolint:revive

		id := gateway.IDFromContext(ctx)
		tenant := gateway.TenantFromContext(ctx)
		if id != nil && tenant == nil {
			return c.NoContent(http.StatusForbidden)
		}

		return next(c)
	}
}

// BlockAPIKey blocks request using API keys to continue.
func BlockAPIKey(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if key := c.Request().Header.Get("X-API-Key"); key != "" {
			return c.NoContent(http.StatusForbidden)
		}

		return next(c)
	}
}

// RequiresPermission reports whether the client has the specified permission.
// If not, it returns an [http.StatusForbidden] response. Otherwise, it executes
// the next handler.
func RequiresPermission(permission auth.Permission) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if ctx, ok := c.(*gateway.Context); !ok || !ctx.Role().HasPermission(permission) {
				return c.NoContent(http.StatusForbidden)
			}

			return next(c)
		}
	}
}
