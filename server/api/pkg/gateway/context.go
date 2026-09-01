package gateway

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type ctxKey string

const ctxAdminRoute ctxKey = "admin-route"

// Context is echo's request context extended with the service layer and with the identity the
// authentication middleware resolved. Handlers take one of these instead of echo's own.
type Context struct {
	service any
	*echo.Context
}

// MarkAdminRoute tags the request as originating from the admin route group. Only the admin
// middleware should call this — it is what lets [AdminOrScope] distinguish an admin-console
// request (no tenant by design) from a regular request by a user who happens to be an admin.
func (c *Context) MarkAdminRoute() {
	c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), ctxAdminRoute, true)))
}

func (c *Context) isAdminRoute() bool {
	v, _ := c.Request().Context().Value(ctxAdminRoute).(bool)

	return v
}

const contextKey = "gateway-context"

// NewContext wraps c, binding service as the one its handlers will reach.
func NewContext(service any, c *echo.Context) *Context {
	return &Context{service: service, Context: c}
}

// WithContext installs a gateway [Context] bound to service on every request passing through it.
//
// Echo's Context is a concrete struct, so the gateway context cannot be handed to the next
// handler in its place — it rides along in the request store instead, and [From] takes it back
// out. Register this before anything that calls [From].
func WithContext(service any) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set(contextKey, NewContext(service, c))

			return next(c)
		}
	}
}

// From returns the [Context] [WithContext] installed for this request. It reports false when
// there is none, so callers can fail closed instead of assuming an identity is present.
//
// Middleware and handlers must reach the gateway context through this function rather than
// reading the store directly: where it lives is the gateway's business.
func From(c *echo.Context) (*Context, bool) {
	gCtx, ok := c.Get(contextKey).(*Context)

	return gCtx, ok
}

// Service returns the service layer bound to this request. Callers assert it to the concrete
// interface they need, which is what lets community and enterprise share these handlers.
func (c *Context) Service() any {
	return c.service
}

// Role returns the user's namespace role got from JWT through gateway. It is
// empty if the user has no namespaces.
func (c *Context) Role() authorizer.Role {
	return authorizer.RoleFromString(c.Request().Header.Get("X-Role"))
}

// Tenant returns the namespace's tenant got from JWT through gateway.
func (c *Context) Tenant() *models.Tenant {
	tenant := c.Request().Header.Get("X-Tenant-ID")
	if tenant != "" {
		return &models.Tenant{ID: tenant}
	}

	return nil
}

// Scope bounds the request to the namespace its tenant header carries. A route that serves
// namespace-bound data calls this and returns the error rather than widening the read: an absent
// header refuses the request, matching the tenant guard's fail-closed behaviour.
func (c *Context) Scope() (scope.Scope, error) {
	tenant := c.Tenant()
	if tenant == nil {
		return scope.Scope{}, echo.NewHTTPError(http.StatusForbidden, scope.ErrEmptyTenantID.Error())
	}

	sc, err := scope.NewBounded(tenant.ID)
	if err != nil {
		return scope.Scope{}, echo.NewHTTPError(http.StatusForbidden, err.Error())
	}

	return sc, nil
}

// AdminOrScope returns an unbounded scope when the request was routed through the admin route group
// (marked by [MarkAdminRoute]), or a namespace-bounded scope otherwise. This distinction is
// deliberate: an admin user browsing the regular UI must stay bounded to their selected namespace,
// while the same user in the admin console sees every namespace.
func (c *Context) AdminOrScope() (scope.Scope, error) {
	if c.isAdminRoute() {
		return scope.NewUnbounded("admin console: cross-namespace read for system administration"), nil
	}

	return c.Scope()
}

// DeviceUID returns the device's UID got from the device JWT through gateway.
// It is empty when the request was not authenticated with a device token.
func (c *Context) DeviceUID() string {
	return c.Request().Header.Get("X-Device-UID")
}

// Username returns the username got from JWT through gateway.
func (c *Context) Username() *models.Username {
	username := c.Request().Header.Get("X-Username")
	if username != "" {
		return &models.Username{ID: username}
	}

	return nil
}

// ID returns the user's ID got from JWT through gateway.
func (c *Context) ID() *models.ID {
	ID := c.Request().Header.Get("X-ID")
	if ID != "" {
		return &models.ID{ID: ID}
	}

	return nil
}

// Ctx returns the request's [context.Context], which is cancelled when the client goes away.
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

// IsAdmin returns whether the user is an admin or not.
func (c *Context) IsAdmin() bool {
	return c.Request().Header.Get("X-Admin") == "true"
}
