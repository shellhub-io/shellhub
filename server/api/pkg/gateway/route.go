package gateway

import (
	"context"
	"net/http"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"sync"

	"github.com/labstack/echo/v5"
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

// One registers handler as a route answering with a JSON body.
func One[T, R any](handler OneHandler[T, R], options ...RouteOption) echo.HandlerFunc {
	declaration := declare(handler, ShapeOne, options)

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
}

// List registers handler as a route answering with a JSON body and the total-count header.
func List[T, R any](handler ListHandler[T, R], options ...RouteOption) echo.HandlerFunc {
	declaration := declare(handler, ShapeList, options)

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
}

// None registers handler as a route answering with 200 and no body.
func None[T any](handler NoneHandler[T], options ...RouteOption) echo.HandlerFunc {
	declaration := declare(handler, ShapeNone, options)

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

	if paginated, ok := any(req).(query.Paginated); ok {
		paginated.GetPaginator().Normalize()
	}

	if sorted, ok := any(req).(query.Sorted); ok {
		sorted.GetSorter().Normalize()
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

// RouteOption declares an exception to the two rules every route follows: it is bounded to a
// namespace, and it is performed by an actor.
type RouteOption func(*Declaration)

// Unbounded declares that the route deliberately reads across namespaces, and records why that is
// safe. The reason is a required argument, so breadth cannot arrive by omission — only by someone
// typing why.
func Unbounded(reason string) RouteOption {
	return func(d *Declaration) {
		d.Unbounded, d.UnboundedReason = true, reason
	}
}

// Anonymous declares that the route deliberately carries no actor, and records why that is safe.
// It is independent of [Unbounded]: a device authenticating with its own token is bounded to a
// namespace and still carries no actor.
func Anonymous(reason string) RouteOption {
	return func(d *Declaration) {
		d.Anonymous, d.AnonymousReason = true, reason
	}
}

// Declaration is what one route registration claims about itself: the shape it answers with, and
// any exception it takes to the default rules.
type Declaration struct {
	// Handler is the fully qualified name of the wrapped function, which is what ties a claim back
	// to the code it is about.
	Handler string
	Shape   Shape

	Unbounded       bool
	UnboundedReason string

	Anonymous       bool
	AnonymousReason string
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

var declarations = struct {
	sync.Mutex
	set map[Declaration]struct{}
}{set: make(map[Declaration]struct{})}

func declare(handler any, shape Shape, options []RouteOption) Declaration {
	d := Declaration{Handler: handlerName(handler), Shape: shape}
	for _, option := range options {
		option(&d)
	}

	declarations.Lock()
	defer declarations.Unlock()

	declarations.set[d] = struct{}{}

	return d
}

// Declarations returns every claim the route tables built in this process have made, ordered by
// handler name.
func Declarations() []Declaration {
	declarations.Lock()
	defer declarations.Unlock()

	all := make([]Declaration, 0, len(declarations.set))
	for d := range declarations.set {
		all = append(all, d)
	}

	sort.Slice(all, func(i, j int) bool { return all[i].Handler < all[j].Handler })

	return all
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
