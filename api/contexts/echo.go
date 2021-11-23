package contexts

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type EchoContext struct {
	service interface{}
	echo.Context
}

func NewEchoContext(service interface{}, c echo.Context) *EchoContext {
	return &EchoContext{
		service: service,
		Context: c,
	}
}

func (c *EchoContext) Ctx() context.Context {
	return c.Request().Context()
}

func (c *EchoContext) Service() interface{} {
	return c.service
}

func (c *EchoContext) Tenant() *models.Tenant {
	tenant := c.Request().Header.Get("X-Tenant-ID")
	if tenant == "" {
		return nil
	}

	return &models.Tenant{tenant}
}

func (c *EchoContext) Username() *models.Username {
	username := c.Request().Header.Get("X-Username")
	if username == "" {
		return nil
	}

	return &models.Username{username}
}

func (c *EchoContext) ID() *models.ID {
	ID := c.Request().Header.Get("X-ID")
	if ID == "" {
		return nil
	}

	return &models.ID{ID}
}

// TODO

func IDFromContext(ctx context.Context) *models.ID {
	if c, ok := ctx.Value("ctx").(*EchoContext); ok {
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

func TenantFromContext(ctx context.Context) *models.Tenant {
	if c, ok := ctx.Value("ctx").(*EchoContext); ok {
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
