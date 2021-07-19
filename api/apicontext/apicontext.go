package apicontext

import (
	"context"

	"github.com/labstack/echo/v4"
	itf "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type Context struct {
	service itf.Service
	echo.Context
}

func NewContext(service itf.Service, c echo.Context) *Context {
	return &Context{service: service, Context: c}
}

func (c *Context) Service() itf.Service {
	return c.service
}

func (c *Context) Tenant() *models.Tenant {
	tenant := c.Request().Header.Get("X-Tenant-ID")
	if tenant != "" {
		return &models.Tenant{tenant}
	}

	return nil
}

func (c *Context) Username() *models.Username {
	username := c.Request().Header.Get("X-Username")
	if username != "" {
		return &models.Username{username}
	}

	return nil
}

func (c *Context) ID() *models.ID {
	ID := c.Request().Header.Get("X-ID")
	if ID != "" {
		return &models.ID{ID}
	}

	return nil
}

func (c *Context) Ctx() context.Context {
	return c.Request().Context()
}

func TenantFromContext(ctx context.Context) *models.Tenant {
	if c, ok := ctx.Value("ctx").(*Context); ok {
		tenant := c.Tenant()
		if tenant == nil {
			if value, ok := ctx.Value("tenant").(string); ok {
				tenant = &models.Tenant{value}
			}
		}

		return tenant
	}

	return nil
}

func UsernameFromContext(ctx context.Context) *models.Username {
	if c, ok := ctx.Value("ctx").(*Context); ok {
		username := c.Username()
		if username == nil {
			if value, ok := ctx.Value("username").(string); ok {
				username = &models.Username{value}
			}
		}

		return username
	}

	return nil
}

func IDFromContext(ctx context.Context) *models.ID {
	if c, ok := ctx.Value("ctx").(*Context); ok {
		ID := c.ID()
		if ID == nil {
			if value, ok := ctx.Value("ID").(string); ok {
				ID = &models.ID{value}
			}
		}

		return ID
	}

	return nil
}

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
