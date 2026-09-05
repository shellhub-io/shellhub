package gateway

import (
	"context"
	"net/http"
	"reflect"
	"runtime"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	routes "github.com/shellhub-io/shellhub/server/api/routes/errors"
)

const totalCountHeader = "X-Total-Count"

// Shape names the response a wrapped handler produces. Every API resource operation answers with
// one of the three; a route that fits none of them is registered directly and named in the route
// table's exempt set.
type Shape string

const (
	// ShapeOne answers with a JSON body.
	ShapeOne Shape = "one"
	// ShapeList answers with a JSON body and the total count of the collection.
	ShapeList Shape = "list"
	// ShapeNone answers with 200 and no body.
	ShapeNone Shape = "none"
	// ShapeLegacy is a handler that still writes its own response through the gateway [Context].
	// It declares its address and its guards like any other route; only its body has yet to move.
	ShapeLegacy Shape = "legacy"
)

// Authority names how a route decides what it demands of the caller's role. A route states one,
// or claims [Anonymous] — which discharges the question, because a request carrying no actor has
// no role for a permission to be checked against.
type Authority string

const (
	// AuthorityUnstated is a route that claimed nothing. It is the zero value because forgetting
	// is what the route table has to be able to see.
	AuthorityUnstated Authority = ""
	// AuthorityPermission is a route admitting a caller whose role holds one of the permissions it
	// named, refused by the guard the claim installs.
	AuthorityPermission Authority = "permission"
	// AuthorityInHandler is a route whose admission decision needs request data a middleware
	// cannot see, so the check stays in the handler and the reason says why.
	AuthorityInHandler Authority = "in-handler"
	// AuthorityNone is a route that demands no authority at all, and the reason says why that is
	// safe.
	AuthorityNone Authority = "none"
)

// OneHandler answers with a single value. It is a function of its inputs: it does not know that
// HTTP exists, and cannot be called without the namespace it is bounded to and the actor
// performing it.
type OneHandler[T, R any] func(ctx context.Context, sc scope.Scope, actor Actor, req *T) (R, error)

// ListHandler answers with a page of values and the size of the whole collection. The wrapper
// writes that count to the response, so no handler decides where the header goes.
type ListHandler[T, R any] func(ctx context.Context, sc scope.Scope, actor Actor, req *T) (R, int, error)

// NoneHandler answers with success alone.
type NoneHandler[T any] func(ctx context.Context, sc scope.Scope, actor Actor, req *T) error

// Route is a handler and the claim its registration is about to make. It is not yet a route the
// router serves: it has no address until [GET] and its siblings mount it, which is what completes
// the declaration.
type Route struct {
	declaration Declaration
	build       func(Declaration) echo.HandlerFunc
}

// One answers with a JSON body.
func One[T, R any](handler OneHandler[T, R]) Route {
	return Route{
		declaration: Declaration{Handler: handlerName(handler), Shape: ShapeOne},
		build: func(declaration Declaration) echo.HandlerFunc {
			return func(c *echo.Context) error {
				in, err := prepare[T](c, declaration)
				if err != nil {
					return err
				}

				res, err := handler(in.ctx, in.scope, in.actor, in.req)
				if err != nil {
					return err
				}

				return c.JSON(http.StatusOK, res)
			}
		},
	}
}

// List answers with a JSON body and the total-count header.
func List[T, R any](handler ListHandler[T, R]) Route {
	return Route{
		declaration: Declaration{Handler: handlerName(handler), Shape: ShapeList},
		build: func(declaration Declaration) echo.HandlerFunc {
			return func(c *echo.Context) error {
				in, err := prepare[T](c, declaration)
				if err != nil {
					return err
				}

				res, count, err := handler(in.ctx, in.scope, in.actor, in.req)
				if err != nil {
					return err
				}

				c.Response().Header().Set(totalCountHeader, strconv.Itoa(count))

				return c.JSON(http.StatusOK, res)
			}
		},
	}
}

// None answers with 200 and no body.
func None[T any](handler NoneHandler[T]) Route {
	return Route{
		declaration: Declaration{Handler: handlerName(handler), Shape: ShapeNone},
		build: func(declaration Declaration) echo.HandlerFunc {
			return func(c *echo.Context) error {
				in, err := prepare[T](c, declaration)
				if err != nil {
					return err
				}

				if err := handler(in.ctx, in.scope, in.actor, in.req); err != nil {
					return err
				}

				return c.NoContent(http.StatusOK)
			}
		},
	}
}

// Handler adapts a handler that still writes its own response, so that it can be mounted and
// declared like any other route. It fails the request when no gateway [Context] was installed,
// which means the route was registered outside the gateway's group.
func Handler(next func(*Context) error) Route {
	return Route{
		declaration: Declaration{Handler: handlerName(next), Shape: ShapeLegacy},
		build: func(_ Declaration) echo.HandlerFunc {
			return adapt(next)
		},
	}
}

type inputs[T any] struct {
	ctx   context.Context
	scope scope.Scope
	actor Actor
	req   *T
}

func prepare[T any](c *echo.Context, declaration Declaration) (inputs[T], error) {
	gCtx, ok := From(c)
	if !ok {
		return inputs[T]{}, echo.ErrInternalServerError
	}

	stash(c, gCtx)

	req := new(T)
	if err := c.Bind(req); err != nil {
		return inputs[T]{}, err
	}

	if err := applyQuery(req, declaration); err != nil {
		return inputs[T]{}, err
	}

	if err := c.Validate(req); err != nil {
		return inputs[T]{}, err
	}

	sc, err := declaration.resolveScope(gCtx)
	if err != nil {
		return inputs[T]{}, err
	}

	actor, err := declaration.resolveActor(gCtx)
	if err != nil {
		return inputs[T]{}, err
	}

	return inputs[T]{ctx: gCtx.Ctx(), scope: sc, actor: actor, req: req}, nil
}

func applyQuery[T any](req *T, declaration Declaration) error {
	if paginated, ok := any(req).(query.Paginated); ok {
		paginated.GetPaginator().Normalize()
	}

	sorted, sorts := any(req).(query.Sorted)
	if sorts {
		declaration.Query.NormalizeSorter(sorted.GetSorter())
	}

	if !declaration.AcceptsQuery {
		return nil
	}

	if filtered, ok := any(req).(query.Filtered); ok {
		filters := filtered.GetFilters()

		if err := filters.Unmarshal(); err != nil {
			return routes.NewErrInvalidEntity(map[string]string{"filter": "cannot be decoded"})
		}

		if err := query.ValidateFilters(filters, declaration.Query.Filter); err != nil {
			return routes.NewErrInvalidEntity(map[string]string{"filter": "is not valid"})
		}
	}

	if sorts {
		sorter := sorted.GetSorter()
		if err := query.ValidateSorter(sorter, declaration.Query.Sort); err != nil {
			return routes.NewErrInvalidEntity(map[string]string{"sort_by": sorter.By})
		}
	}

	return nil
}

// RouteOption states one claim a route's registration makes, and returns the guard that enforces
// it — or nil when the claim is enforced by the wrapper rather than by a middleware. Options are
// applied in the order they are written, and their guards run in that same order.
type RouteOption func(*Declaration) echo.MiddlewareFunc

// Unbounded declares that the route deliberately reads across namespaces, and records why that is
// safe. The reason is a required argument, so breadth cannot arrive by omission — only by someone
// typing why.
func Unbounded(reason string) RouteOption {
	return func(d *Declaration) echo.MiddlewareFunc {
		d.Unbounded, d.UnboundedReason = true, reason

		return nil
	}
}

// Anonymous declares that the route deliberately carries no actor, and records why that is safe.
// It is independent of [Unbounded]: a device authenticating with its own token is bounded to a
// namespace and still carries no actor.
//
// The claim frees the handler from needing an actor; it does not open the route. What lets the
// request past the credential check is the authenticator's allowlist, and the route table's tests
// are what hold the two to the same answer.
func Anonymous(reason string) RouteOption {
	return func(d *Declaration) echo.MiddlewareFunc {
		d.Anonymous, d.AnonymousReason = true, reason

		return nil
	}
}

// Requires declares the permission the route demands of the caller's role, and installs the guard
// that enforces it. Declaring and enforcing are the same act here, so the declaration is evidence
// of what the route does rather than a description of it.
func Requires(permission authorizer.Permission) RouteOption {
	return func(d *Declaration) echo.MiddlewareFunc {
		d.Authority, d.Permissions = AuthorityPermission, []authorizer.Permission{permission}

		return RequiresPermission(permission)
	}
}

// RequiresAny declares the permissions the route admits a caller for holding any one of, and
// installs the guard that enforces the set. It is [Requires] for a route whose rule is more than
// one permission, which would otherwise have to choose between an inexpressible claim and none.
//
// Naming nothing refuses every caller, because a route admitting a role that holds nothing is what
// [NoPermission] says.
func RequiresAny(permissions ...authorizer.Permission) RouteOption {
	return func(d *Declaration) echo.MiddlewareFunc {
		d.Authority, d.Permissions = AuthorityPermission, permissions

		return RequiresAnyPermission(permissions...)
	}
}

// PermissionInHandler declares that the route's permission is checked past the middleware chain,
// because the check needs request data no middleware can see: a namespace named by a code in the
// path rather than by the session, or a query parameter the permission widens the answer to rather
// than admits. The check stays where it is; the route stops reading as though it had none.
func PermissionInHandler(reason string) RouteOption {
	return func(d *Declaration) echo.MiddlewareFunc {
		d.Authority, d.AuthorityReason = AuthorityInHandler, reason

		return nil
	}
}

// NoPermission declares that the route demands no authority of its caller, and records why that is
// safe: it acts on the caller's own record, it is prior to holding any role, or the credential it
// answers to carries no role at all. The reason is a required argument, so a route cannot become
// permissionless by copy-paste.
func NoPermission(reason string) RouteOption {
	return func(d *Declaration) echo.MiddlewareFunc {
		d.Authority, d.AuthorityReason = AuthorityNone, reason

		return nil
	}
}

// NoAPIKey declares that the route is closed to API keys, and installs the guard that refuses
// one. It is for the routes that must be performed by a person: an API key authenticates a
// namespace, and names nobody to hold responsible for the act.
func NoAPIKey() RouteOption {
	return func(d *Declaration) echo.MiddlewareFunc {
		d.BlocksAPIKey = true

		return BlockAPIKey
	}
}

// Accepts declares the query contract a list route holds a client's filter and sort to, and
// installs no guard: the wrapper enforces it, before the handler is called and in one fixed order.
//
// A resource that accepts neither a filter nor a sort names a contract allowing nothing rather than
// leaving the option off, because an omission cannot be told from a forgotten one — which is the
// failure the route table's invariant exists to catch.
func Accepts(contract query.Contract) RouteOption {
	return func(d *Declaration) echo.MiddlewareFunc {
		d.Query, d.AcceptsQuery = contract, true

		return nil
	}
}

// Guard installs a middleware the declaration says nothing about. It is what the guards that are
// not claims — the tenant check, the legacy authorize middleware — are written with, and it runs
// in the position it is written in, among the guards the other options install.
func Guard(middleware echo.MiddlewareFunc) RouteOption {
	return func(_ *Declaration) echo.MiddlewareFunc {
		return middleware
	}
}

// Declaration is what one route registration claims about itself: where it is mounted, the shape
// it answers with, the authority it demands, and any exception it takes to the default rules.
//
// It is complete only once the route is mounted, because the address is the mounting's answer and
// not the registration's.
type Declaration struct {
	// Handler is the fully qualified name of the wrapped function, which is what ties a claim back
	// to the code it is about.
	Handler string
	Shape   Shape

	// Method and Path are the address the router mounted the route at, with any group prefix
	// applied. It is the same string the router reports and the authenticator matches on.
	Method string
	Path   string

	// Authority is what the route demands of the caller's role, as one of a closed set. The zero
	// value is a route that claimed nothing, which is what tells a forgotten claim from a route
	// deliberately demanding none.
	Authority Authority

	// AuthorityReason is why the route installs no permission guard — because the check is past
	// the middleware chain, or because there is nothing to check. It is set by the two claims that
	// install no guard, and empty for the one that does.
	AuthorityReason string

	// Permissions is the set the caller's role must intersect for the route to admit the call. One
	// element is [Requires]; several is [RequiresAny]. It is empty unless Authority is
	// [AuthorityPermission].
	Permissions []authorizer.Permission

	// BlocksAPIKey reports whether the route refuses a request authenticated by an API key. The
	// refusal is about the credential and not the authority: a key carrying a role that holds the
	// route's permission is refused all the same.
	BlocksAPIKey bool

	// Query is the filter and sort fields the route accepts from a client, and the order it serves
	// when the client names none. The zero value accepts neither, so AcceptsQuery is what tells a
	// route deliberately taking no query from one that named no contract at all.
	Query        query.Contract
	AcceptsQuery bool

	Unbounded       bool
	UnboundedReason string

	Anonymous       bool
	AnonymousReason string
}

// Address returns the route's method and path as the router and the authenticator both spell it,
// which is the key a claim is joined to a route by.
func (d Declaration) Address() string {
	return d.Method + " " + d.Path
}

func (d Declaration) resolveScope(c *Context) (scope.Scope, error) {
	if d.Unbounded {
		return scope.NewUnbounded(d.UnboundedReason), nil
	}

	return c.AdminOrScope()
}

func (d Declaration) resolveActor(c *Context) (Actor, error) {
	identity := IdentityFrom(c.Request().Header)

	actor := identity.Actor()

	if d.Anonymous || !actor.IsZero() {
		return actor, nil
	}

	return Actor{}, routes.NewErrUnauthorized(nil)
}

func handlerName(handler any) string {
	value := reflect.ValueOf(handler)
	if value.Kind() != reflect.Func {
		return ""
	}

	fn := runtime.FuncForPC(value.Pointer())
	if fn == nil {
		return ""
	}

	return fn.Name()
}
