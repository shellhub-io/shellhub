// Package routetable mounts HTTP routes from tables in which every row states who may call it, and
// finds the routes a router serves that no table declared.
package routetable

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/server/api/routes/middleware"
)

// ErrUnstatedAccess is returned by [Mount] for a row whose [Access] admits no caller it can name.
var ErrUnstatedAccess = errors.New("the route states no access")

// TenantParam is the path parameter naming a namespace. On a row whose path names it, unless the
// row's access is [OwnAccount] or [Admin], [Mount] refuses a caller with no namespace, an anonymous
// one included, and a caller whose namespace is not the one the path names. An admin-panel
// identity, which carries the admin flag and no ID, passes.
const TenantParam = "tenant"

type kind int

const (
	unstated kind = iota
	anonymous
	deviceToken
	ownAccount
	namespaceMember
	pathCode
	permission
	admin
)

var kindNames = map[kind]string{
	anonymous:       "anonymous",
	deviceToken:     "device-token",
	ownAccount:      "own-account",
	namespaceMember: "namespace-member",
	pathCode:        "path-code",
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
	// DeviceToken opens a route to a device token, which every other access but [Anonymous]
	// refuses; [Anonymous] drops the token and serves the request as anonymous. It installs no
	// guard of its own, so a user token or an API key reaches the handler too, and the handler
	// must refuse a request that carries no device.
	DeviceToken = Access{kind: deviceToken}
	// OwnAccount admits any authenticated user acting on their own account, and refuses API keys.
	OwnAccount = Access{kind: ownAccount}
	// NamespaceMember admits any member of the caller's namespace, whatever the role.
	NamespaceMember = Access{kind: namespaceMember}
	// PathCode admits any authenticated user holding the code the path carries, which the handler
	// checks, and refuses API keys.
	PathCode = Access{kind: pathCode}
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

// String names the kind of caller a admits, the same for every [Requires] whatever its
// permissions, and returns "" for an access that states none.
func (a Access) String() string {
	if !a.stated() {
		return ""
	}

	return kindNames[a.kind]
}

func (a Access) stated() bool {
	switch a.kind {
	case anonymous, deviceToken, ownAccount, namespaceMember, pathCode:
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
// it. Unbounded hands the endpoint claims spanning every namespace instead of the caller's. NoAPIKey
// refuses API keys on a row whose access admits them. Middleware runs after the access guard, in
// order.
type Route struct {
	Method     string
	Path       string
	Endpoint   gateway.Endpoint
	Access     Access
	Unbounded  bool
	NoAPIKey   bool
	Middleware []echo.MiddlewareFunc
}

func (r Route) blocksAPIKey() bool {
	return r.NoAPIKey || r.Access.kind == ownAccount || r.Access.kind == pathCode
}

func (r Route) namesTenant() bool {
	if r.Access.kind == ownAccount || r.Access.kind == admin {
		return false
	}

	return slices.Contains(strings.Split(r.Path, "/"), ":"+TenantParam)
}

func (r Route) chain() []echo.MiddlewareFunc {
	var chain []echo.MiddlewareFunc

	if r.blocksAPIKey() {
		chain = append(chain, middleware.BlockAPIKey)
	}

	if r.namesTenant() {
		chain = append(chain, middleware.RequiresTenant(TenantParam))
	}

	switch r.Access.kind {
	case permission:
		chain = append(chain, middleware.RequiresAnyPermission(r.Access.permissions...))
	case namespaceMember:
		chain = append(chain, middleware.Authorize)
	case admin:
		chain = append(chain, r.Access.guard)
	case unstated, anonymous, deviceToken, ownAccount, pathCode:
	}

	return append(chain, r.Middleware...)
}

const namePrefix = "routetable:"

// Mount adds every row of table to group behind the guard its access states, refusing API keys on
// the rows that block them, and guarding a row whose path names [TenantParam] as that constant
// describes. It declares the anonymous and device-token rows on authn so they open without a
// credential or to a device.
// authn may be nil when the router authenticates nothing. It mounts nothing and returns an error
// wrapping [ErrUnstatedAccess] when any row states no access, and returns echo's error when echo
// refuses a row, leaving the rows before it mounted.
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
			Handler:     r.Endpoint(gateway.Claims{Anonymous: r.Access.kind == anonymous, Unbounded: r.Unbounded}),
			Middlewares: r.chain(),
		})
		if err != nil {
			return fmt.Errorf("%s %s: %w", r.Method, r.Path, err)
		}

		if authn == nil {
			continue
		}

		switch r.Access.kind {
		case anonymous:
			authn.AllowAnonymous(allowlistMethod(info.Method), info.Path)
		case deviceToken:
			authn.AllowDevice(allowlistMethod(info.Method), info.Path)
		case unstated, ownAccount, namespaceMember, pathCode, permission, admin:
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
