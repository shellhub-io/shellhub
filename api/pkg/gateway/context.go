package gateway

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/api/auth"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type Context struct {
	service interface{}
	echo.Context
}

func NewContext(service interface{}, c echo.Context) *Context {
	return &Context{service: service, Context: c}
}

func (c *Context) Service() interface{} {
	return c.service
}

// Role returns the user's namespace role got from JWT through gateway. It is
// empty if the user has no namespaces.
func (c *Context) Role() auth.Role {
	return auth.RoleFromString(c.Request().Header.Get("X-Role"))
}

// Tenant returns the namespace's tenant got from JWT through gateway.
func (c *Context) Tenant() *models.Tenant {
	tenant := c.Request().Header.Get("X-Tenant-ID")
	if tenant != "" {
		return &models.Tenant{tenant}
	}

	return nil
}

// Username returns the username got from JWT through gateway.
func (c *Context) Username() *models.Username {
	username := c.Request().Header.Get("X-Username")
	if username != "" {
		return &models.Username{username}
	}

	return nil
}

// ID returns the user's ID got from JWT through gateway.
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

// GetID returns the user's ID got from JWT through gateway.
func (c *Context) GetID() (string, bool) {
	ID := c.Request().Header.Get("X-ID")
	if ID != "" {
		return ID, true
	}

	return "", false
}

// GetTennat returns the namespace's tenant got from JWT through gateway.
func (c *Context) GetTennat() (string, bool) {
	tenant := c.Request().Header.Get("X-Tenant-ID")
	if tenant != "" {
		return tenant, true
	}

	return "", false
}

// GetUsername returns the username got from JWT through gateway.
func (c *Context) GetUsername() (string, bool) {
	username := c.Request().Header.Get("X-Username")
	if username != "" {
		return username, true
	}

	return "", false
}
