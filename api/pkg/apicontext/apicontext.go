package apicontext

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	// HeaderTenant is the header to get namespace's tenant.
	HeaderTenant = "X-Tenant-ID"
	// HeaderUserID is the header to get user's ID.
	HeaderUserID = "X-ID"
	// HeaderUserType is the header to get user's type inside a namespace.
	// User's type has association with authorizer's package
	HeaderUserType = "X-Type"
	// HeaderUserName is the header to get user's name.
	HeaderUserName = "X-Username"
)

type Context struct {
	service interface{}
	echo.Context
}

func NewContext(service interface{}, c echo.Context) *Context {
	return &Context{service: service, Context: c}
}

// Value gets a value from api's context.
// header parameter must be one of Header constants in the api's context package.
// It returns an empty string when the header does not exist.
func (c *Context) Value(header string) string {
	return c.Request().Header.Get(header)
}

func (c *Context) Service() interface{} {
	return c.service
}

func (c *Context) Tenant() *models.Tenant {
	return &models.Tenant{c.Value(HeaderTenant)}
}

func (c *Context) Username() *models.Username {
	return &models.Username{c.Value(HeaderUserName)}
}

func (c *Context) ID() *models.ID {
	return &models.ID{c.Value(HeaderUserID)}
}

func (c *Context) Ctx() context.Context {
	return c.Request().Context()
}
