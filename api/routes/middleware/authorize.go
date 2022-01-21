package middleware

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
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
