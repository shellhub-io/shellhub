package gateway

import (
	"context"

	"github.com/labstack/echo/v5"
)

func adapt(next func(*Context) error) echo.HandlerFunc {
	return func(c *echo.Context) error {
		gCtx, ok := From(c)
		if !ok {
			return echo.ErrInternalServerError
		}

		stash(c, gCtx)

		return next(gCtx)
	}
}

func stash(c *echo.Context, gCtx *Context) {
	c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), "ctx", gCtx)))
}

// Middleware adapts echo middleware so it runs with a gateway [Context] in place.
func Middleware(m echo.MiddlewareFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			return adapt(func(c *Context) error {
				return m(next)(c.Context)
			})(c)
		}
	}
}
