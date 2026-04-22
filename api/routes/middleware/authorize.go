package middleware

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
)

func Authorize(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := context.WithValue(c.Request().Context(), "ctx", c.(*gateway.Context)) //nolint:revive

		id := gateway.IDFromContext(ctx)
		tenant := gateway.TenantFromContext(ctx)
		gCtx := c.(*gateway.Context)

		// Allow admins to access resources without tenant scope (e.g., from /admin/api endpoints)
		if id != nil && tenant == nil && !gCtx.IsAdmin() {
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
func RequiresPermission(permission authorizer.Permission) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if ctx, ok := c.(*gateway.Context); !ok || !ctx.Role().HasPermission(permission) {
				return c.NoContent(http.StatusForbidden)
			}

			return next(c)
		}
	}
}

// RequiresTenant enforces that the caller's tenant scope matches the tenant
// provided in the given URL path parameter. It fails closed: if either the
// caller's tenant or the path parameter is missing or they don't match, it
// returns [http.StatusForbidden]. Callers coming through the admin panel
// bypass this check; they are identified by the /admin/api gateway, which
// strips X-ID and keeps X-Admin: true. An admin user who hits the regular
// /api/* surface still carries X-ID and is subject to the tenant guard.
func RequiresTenant(param string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, ok := c.(*gateway.Context)
			if !ok {
				return c.NoContent(http.StatusForbidden)
			}

			if ctx.ID() == nil && ctx.IsAdmin() {
				return next(c)
			}

			path := c.Param(param)
			tenant := ctx.Tenant()
			if path == "" || tenant == nil || tenant.ID != path {
				return c.NoContent(http.StatusForbidden)
			}

			return next(c)
		}
	}
}
