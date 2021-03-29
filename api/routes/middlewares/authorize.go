package middlewares

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/apicontext"
)

func Authorize(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := context.WithValue(c.Request().Context(), "ctx", c.(*apicontext.Context))

		id := apicontext.IDFromContext(ctx)
		tenant := apicontext.TenantFromContext(ctx)
		if id != nil && tenant == nil {
			return c.NoContent(http.StatusForbidden)
		}

		return next(c)
	}
}
