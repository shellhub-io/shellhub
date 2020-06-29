package apicontext

import (
	"context"

	"github.com/labstack/echo"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type Context struct {
	store store.Store
	echo.Context
}

func NewContext(store store.Store, c echo.Context) *Context {
	return &Context{store: store, Context: c}
}

func (c *Context) Store() store.Store {
	return c.store
}

func (c *Context) Tenant() *models.Tenant {
	tenant := c.Request().Header.Get("X-Tenant-ID")
	if tenant != "" {
		return &models.Tenant{tenant}
	}

	return nil
}

func (c *Context) Ctx() context.Context {
	return c.Request().Context()
}

func TenantFromContext(ctx context.Context) *models.Tenant {
	c := ctx.Value("ctx").(*Context)
	return c.Tenant()
}

func Handler(next func(Context) error) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := context.WithValue(c.Request().Context(), "ctx", c.(*Context))

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
