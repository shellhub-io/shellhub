package gateway

import (
	"context"

	"github.com/labstack/echo/v5"
)

func Handler(next func(*Context) error) echo.HandlerFunc {
	return func(c *echo.Context) error {
		gCtx, ok := From(c)
		if !ok {
			return echo.ErrInternalServerError
		}

		ctx := context.WithValue(c.Request().Context(), "ctx", gCtx) //nolint:revive

		c.SetRequest(c.Request().WithContext(ctx))

		return next(gCtx)
	}
}

func Middleware(m echo.MiddlewareFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			return Handler(func(c *Context) error {
				return m(next)(c.Context)
			})(c)
		}
	}
}
