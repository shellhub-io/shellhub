package gateway

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/origin"
)

type ctxKey string

const ctxAdminRoute ctxKey = "admin-route"

// Context is echo's request context extended with the identity the authentication middleware
// resolved. Handlers take one of these instead of echo's own.
type Context struct {
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

// NewContext wraps c. The identity accessors read c's request each time they are called, so a
// Context built before authentication still sees what authentication writes afterwards.
func NewContext(c *echo.Context) *Context {
	return &Context{Context: c}
}

// WithContext installs a gateway [Context] on every request passing through it.
//
// Echo's Context is a concrete struct, so the gateway context cannot be handed to the next
// handler in its place — it rides along in the request store instead, and [From] takes it back
// out. Register this before anything that calls [From].
func WithContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set(contextKey, NewContext(c))

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

func (c *Context) identity() identity.Identity {
	caller, _ := identity.FromContext(c.Request().Context())

	return caller
}

// Origin returns how the client reached the server, as the router resolved it, and the zero
// Origin when the request did not pass through the router.
func (c *Context) Origin() origin.Origin {
	reached, _ := origin.FromContext(c.Request().Context())

	return reached
}

// Role returns the user's namespace role resolved by authentication. It is
// empty if the user has no namespaces.
func (c *Context) Role() authorizer.Role {
	return c.identity().Role
}

// Tenant returns the namespace's tenant resolved by authentication.
func (c *Context) Tenant() *models.Tenant {
	tenant := c.identity().TenantID
	if tenant != "" {
		return &models.Tenant{ID: tenant}
	}

	return nil
}

// Scope bounds the request to the namespace its identity carries. A route that serves
// namespace-bound data calls this and returns the error rather than widening the read: an absent
// tenant refuses the request, matching the tenant guard's fail-closed behaviour.
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
		return scope.NewUnbounded(reasonAdminConsole), nil
	}

	return c.Scope()
}

// DeviceUID returns the device's UID the device token authenticated as.
// It is empty when the request was not authenticated with a device token.
func (c *Context) DeviceUID() string {
	return c.identity().DeviceUID
}

// ID returns the user's ID resolved by authentication.
func (c *Context) ID() *models.ID {
	ID := c.identity().ID
	if ID != "" {
		return &models.ID{ID: ID}
	}

	return nil
}

// APIKeyID returns the surrogate id of the API key the request authenticated with, and an
// empty string for every other credential. A handler reads it to tell an automation acting on
// itself from one acting on another key.
func (c *Context) APIKeyID() string {
	return c.identity().APIKeyID
}

// Ctx returns the request's [context.Context], which is cancelled when the client goes away.
func (c *Context) Ctx() context.Context {
	return c.Request().Context()
}

// GetID returns the user's ID resolved by authentication.
func (c *Context) GetID() (string, bool) {
	ID := c.identity().ID
	if ID != "" {
		return ID, true
	}

	return "", false
}

// GetTennat returns the namespace's tenant resolved by authentication.
func (c *Context) GetTennat() (string, bool) {
	tenant := c.identity().TenantID
	if tenant != "" {
		return tenant, true
	}

	return "", false
}

// IsAdmin returns whether the user is an admin or not.
func (c *Context) IsAdmin() bool {
	return c.identity().Admin
}
