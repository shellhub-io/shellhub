package gateway

import (
	"net/http"
	"slices"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
)

// RequiresPermission refuses the request with 403 unless the role the request authenticated with
// holds permission. It answers 403 rather than 401 because the caller is known and simply not
// allowed; a request carrying no credential at all never reaches it.
//
// It lives here rather than beside the other route middleware because [Requires] installs it, and
// the route middleware package imports this one.
func RequiresPermission(permission authorizer.Permission) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if ctx, ok := From(c); !ok || !ctx.Role().HasPermission(permission) {
				return c.NoContent(http.StatusForbidden)
			}

			return next(c)
		}
	}
}

// BlockAPIKey refuses with 403 a request that authenticated with an API key. It reads the header
// rather than the resolved identity because a key is refused whether or not it was honoured.
func BlockAPIKey(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		if key := c.Request().Header.Get("X-API-Key"); key != "" {
			return c.NoContent(http.StatusForbidden)
		}

		return next(c)
	}
}

// RequiresAnyPermission refuses the request with 403 unless the role the request authenticated
// with holds at least one of permissions. Naming none refuses every request, so the programming
// error fails closed rather than admitting every caller.
//
// It answers 403 for the same reason [RequiresPermission] does: the caller is known and simply not
// allowed.
func RequiresAnyPermission(permissions ...authorizer.Permission) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx, ok := From(c)
			if !ok {
				return c.NoContent(http.StatusForbidden)
			}

			if !slices.ContainsFunc(permissions, ctx.Role().HasPermission) {
				return c.NoContent(http.StatusForbidden)
			}

			return next(c)
		}
	}
}
