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

var (
	// ErrUnstatedAccess is returned by [Mount] for a row whose [Access] admits no caller it can name.
	ErrUnstatedAccess = errors.New("the route states no access")
	// ErrCallerMismatch is returned by [Mount] for a row whose [Access] admits a kind of caller its
	// endpoint's handler cannot take.
	ErrCallerMismatch = errors.New("the route's access admits a caller its endpoint cannot take")
	// ErrAPIKeyColumnOnTypedRow is returned by [Mount] for a row setting [Route.NoAPIKey] on an
	// endpoint that names its caller, whose type already decides whether API keys pass.
	ErrAPIKeyColumnOnTypedRow = errors.New("the route refuses API keys by column while its endpoint names its caller")
)

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
	// refuses; [Anonymous] drops the token and serves the request as anonymous. It refuses with
	// 403 every request that did not authenticate as a device.
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
// refuses API keys on a row whose access admits them, and only on a row whose endpoint names no
// caller: a typed endpoint taking a [gateway.Member] refuses them by its type. Middleware runs after
// the access guard, in order.
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
	return r.NoAPIKey || r.Access.kind == ownAccount || r.Access.kind == pathCode || r.Endpoint.Caller() == gateway.CallerMember
}

var admittedCallers = map[kind][]gateway.CallerKind{
	anonymous:       {gateway.CallerAnonymous},
	namespaceMember: {gateway.CallerPrincipal, gateway.CallerMember},
	permission:      {gateway.CallerPrincipal, gateway.CallerMember},
	admin:           {gateway.CallerAdmin},
}

func (r Route) check() error {
	if !r.Access.stated() {
		return fmt.Errorf("%s %s: %w", r.Method, r.Path, ErrUnstatedAccess)
	}

	caller := r.Endpoint.Caller()
	if caller == gateway.CallerUnstated {
		return nil
	}

	if r.NoAPIKey {
		return fmt.Errorf("%s %s: %w", r.Method, r.Path, ErrAPIKeyColumnOnTypedRow)
	}

	if !slices.Contains(admittedCallers[r.Access.kind], caller) {
		return fmt.Errorf("%s %s: %s access, %s endpoint: %w", r.Method, r.Path, r.Access, caller, ErrCallerMismatch)
	}

	return nil
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
	case deviceToken:
		chain = append(chain, middleware.RequiresDevice)
	case unstated, anonymous, ownAccount, pathCode:
	}

	return append(chain, r.Middleware...)
}

const namePrefix = "routetable:"

// Mount adds every row of table to group behind the guard its access states, refusing API keys on
// the rows that block them and on every row whose endpoint takes a [gateway.Member], and guarding a
// row whose path names [TenantParam] as that constant describes. It declares the anonymous and
// device-token rows on authn so they open without a credential or to a device.
//
// An endpoint naming its caller must take one the row's access admits: [gateway.Anonymous] on
// [Anonymous], [gateway.Principal] or [gateway.Member] on [NamespaceMember] and [Requires], and
// [gateway.Admin] on [Admin]. An endpoint naming none, [gateway.Legacy] or [gateway.Raw], mounts on
// any access.
//
// authn may be nil when the router authenticates nothing. It mounts nothing and returns an error
// naming the row and wrapping [ErrUnstatedAccess] when any row states no access,
// [ErrCallerMismatch] when its access admits a caller its endpoint cannot take, or
// [ErrAPIKeyColumnOnTypedRow] when it sets [Route.NoAPIKey] on an endpoint naming its caller. It
// returns echo's error when echo refuses a row, leaving the rows before it mounted.
func Mount(group *echo.Group, authn *middleware.Authenticator, table []Route) error {
	for _, r := range table {
		if err := r.check(); err != nil {
			return err
		}
	}

	for _, r := range table {
		info, err := group.AddRoute(echo.Route{
			Method:      r.Method,
			Path:        r.Path,
			Name:        namePrefix + kindNames[r.Access.kind],
			Handler:     r.Endpoint.Handler(gateway.Claims{Unbounded: r.Unbounded}),
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
