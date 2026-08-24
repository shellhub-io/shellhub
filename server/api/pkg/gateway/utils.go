package gateway

import (
	"context"

	"github.com/labstack/echo/v5"
)

// Handler adapts a gateway handler to echo's, failing the request when no gateway [Context]
// was installed — which means the route was registered outside the gateway's group.
func Handler(next func(*Context) error) echo.HandlerFunc {
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
