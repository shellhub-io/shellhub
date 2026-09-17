package middleware

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

// Authorize refuses the request unless the identity resolved by authentication holds the
// permission declared on the route.
func Authorize(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		gCtx, ok := gateway.From(c)
		if !ok {
			return c.NoContent(http.StatusForbidden)
		}

		ctx := context.WithValue(c.Request().Context(), "ctx", gCtx)

		id := gateway.IDFromContext(ctx)
		tenant := gateway.TenantFromContext(ctx)

		if id != nil && tenant == nil && !gCtx.IsAdmin() {
			return c.NoContent(http.StatusForbidden)
		}

		return next(c)
	}
}

// BlockAPIKey blocks request using API keys to continue. It is [gateway.BlockAPIKey] under the
// name its callers outside this repository still use; a route in this one states the claim with
// [gateway.NoAPIKey] on the line that mounts it.
func BlockAPIKey(next echo.HandlerFunc) echo.HandlerFunc {
	return gateway.BlockAPIKey(next)
}

// RequiresPermission reports whether the client has the specified permission. It is
// [gateway.RequiresPermission] under the name its callers outside this repository still use; a
// route in this one states the claim with [gateway.Requires] on the line that mounts it.
func RequiresPermission(permission authorizer.Permission) echo.MiddlewareFunc {
	return gateway.RequiresPermission(permission)
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
		return func(c *echo.Context) error {
			ctx, ok := gateway.From(c)
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
