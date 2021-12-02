package apicontext

import (
	"context"

	"github.com/labstack/echo/v4"
)

func Handler(next func(Context) error) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := context.WithValue(c.Request().Context(), "ctx", c.(*Context)) //nolint:revive

		c.SetRequest(c.Request().WithContext(ctx))
		c.Set("ctx", c.(*Context))

		return next(*c.(*Context))
	}
}

func Middleware(m echo.MiddlewareFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return Handler(func(c Context) error {
				return m(next)(&c)
			})(c)
		}
	}
}
