// Package routetable mounts HTTP routes from tables in which every row states who may call it, and
// finds the routes a router serves that no table declared.
package routetable

import (
	"errors"
	"fmt"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/server/api/routes/middleware"
)

// ErrUnstatedAccess is returned by [Mount] for a row whose [Access] admits no caller it can name.
var ErrUnstatedAccess = errors.New("the route states no access")

type kind int

const (
	unstated kind = iota
	anonymous
	ownAccount
	namespaceMember
	permission
	admin
)

var kindNames = map[kind]string{
	anonymous:       "anonymous",
	ownAccount:      "own-account",
	namespaceMember: "namespace-member",
	permission:      "permission",
	admin:           "admin",
}

// Access is who may call a route. Its zero value states nobody, and [Mount] refuses it, so a row
// cannot leave its access out and still be served.
type Access struct {
	kind        kind
	permissions []authorizer.Permission
	guard       echo.MiddlewareFunc
}

var (
	// Anonymous admits a caller without a credential. A caller presenting one still has it resolved.
	Anonymous = Access{kind: anonymous}
	// OwnAccount admits any authenticated user acting on their own account, and refuses API keys.
	OwnAccount = Access{kind: ownAccount}
	// NamespaceMember admits any member of the caller's namespace, whatever the role.
	NamespaceMember = Access{kind: namespaceMember}
)

// Requires admits a caller whose role holds at least one of permissions. Given none, it states no
// access.
func Requires(permissions ...authorizer.Permission) Access {
	return Access{kind: permission, permissions: permissions}
}

// Admin admits whoever guard lets through, for the instance administration surface. Given a nil
// guard, it states no access.
func Admin(guard echo.MiddlewareFunc) Access {
	return Access{kind: admin, guard: guard}
}

func (a Access) stated() bool {
	switch a.kind {
	case anonymous, ownAccount, namespaceMember:
		return true
	case permission:
		return len(a.permissions) > 0
	case admin:
		return a.guard != nil
	case unstated:
	}

	return false
}

// Route is one row of a table: the address it answers, the handler serving it, and who may call
// it. Middleware runs after the access guard, in order.
type Route struct {
	Method     string
	Path       string
	Endpoint   gateway.Endpoint
	Access     Access
	NoAPIKey   bool
	Middleware []echo.MiddlewareFunc
}

func (r Route) blocksAPIKey() bool {
	return r.NoAPIKey || r.Access.kind == ownAccount
}

func (r Route) chain() []echo.MiddlewareFunc {
	var chain []echo.MiddlewareFunc

	if r.blocksAPIKey() {
		chain = append(chain, middleware.BlockAPIKey)
	}

	switch r.Access.kind {
	case permission:
		chain = append(chain, middleware.RequiresAnyPermission(r.Access.permissions...))
	case namespaceMember:
		chain = append(chain, middleware.Authorize)
	case admin:
		chain = append(chain, r.Access.guard)
	case unstated, anonymous, ownAccount:
	}

	return append(chain, r.Middleware...)
}

const namePrefix = "routetable:"

// Mount adds every row of table to group behind the guard its access states, and declares the
// anonymous rows on authn so they open without a credential. authn may be nil when the router
// authenticates nothing. It mounts nothing and returns an error wrapping [ErrUnstatedAccess] when
// any row states no access, and returns echo's error when echo refuses a row, leaving the rows
// before it mounted.
func Mount(group *echo.Group, authn *middleware.Authenticator, table []Route) error {
	for _, r := range table {
		if !r.Access.stated() {
			return fmt.Errorf("%s %s: %w", r.Method, r.Path, ErrUnstatedAccess)
		}
	}

	for _, r := range table {
		info, err := group.AddRoute(echo.Route{
			Method:      r.Method,
			Path:        r.Path,
			Name:        namePrefix + kindNames[r.Access.kind],
			Handler:     r.Endpoint(gateway.Claims{Anonymous: r.Access.kind == anonymous}),
			Middlewares: r.chain(),
		})
		if err != nil {
			return fmt.Errorf("%s %s: %w", r.Method, r.Path, err)
		}

		if authn != nil && r.Access.kind == anonymous {
			authn.AllowAnonymous(allowlistMethod(info.Method), info.Path)
		}
	}

	return nil
}

func allowlistMethod(method string) string {
	if method == echo.RouteAny {
		return middleware.AnyMethod
	}

	return method
}

// Address is how [Undeclared] names a route: its method and its full path pattern.
func Address(method, path string) string {
	return method + " " + path
}

// Undeclared returns the address of every route in routes that no table mounted through [Mount]
// and that declared does not map to a non-empty reason. declared is keyed by [Address], and each
// value says why that route is served outside a table. It skips the not-found routes echo adds for
// group middleware, which serve nothing.
func Undeclared(routes echo.Routes, declared map[string]string) []string {
	var found []string

	for _, r := range routes {
		address := Address(r.Method, r.Path)

		switch {
		case r.Method == echo.RouteNotFound:
		case strings.HasPrefix(r.Name, namePrefix):
		case declared[address] != "":
		default:
			found = append(found, address)
		}
	}

	return found
}

// Endpoint adapts a handler that writes its own response from the echo context, ignoring the
// route's claims.
func Endpoint(handler echo.HandlerFunc) gateway.Endpoint {
	return func(gateway.Claims) echo.HandlerFunc {
		return handler
	}
}
