package gateway

import (
	"context"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	routes "github.com/shellhub-io/shellhub/server/api/routes/errors"
)

const totalCountHeader = "X-Total-Count"

// OneHandler answers with a single value, or the error the route fails with.
type OneHandler[T, R any] func(ctx context.Context, sc scope.Scope, actor Actor, req *T) (R, error)

// ListHandler answers with a page of values and the size of the whole collection. [List] writes
// that size to the X-Total-Count response header.
type ListHandler[T, R any] func(ctx context.Context, sc scope.Scope, actor Actor, req *T) (R, int64, error)

// NoneHandler answers with success alone.
type NoneHandler[T any] func(ctx context.Context, sc scope.Scope, actor Actor, req *T) error

// Claims are the exceptions a route takes to the two rules every route follows: it is bounded to
// the caller's namespace, and it is performed by an actor. Anonymous serves a request that carries
// no actor, and Unbounded spans every namespace instead of the caller's.
type Claims struct {
	Anonymous bool
	Unbounded bool
}

// Endpoint is a handler waiting for the claims of the route that mounts it.
type Endpoint func(Claims) echo.HandlerFunc

// One answers 200 with the handler's result as JSON. It fails with 500 when no gateway [Context]
// was installed, with the bind or validation error for a malformed request, with 403 when the
// request carries no namespace and the route is not [Claims.Unbounded], with 401 when it carries
// no actor and the route is not [Claims.Anonymous], and otherwise with the handler's error.
func One[T, R any](handler OneHandler[T, R]) Endpoint {
	return func(claims Claims) echo.HandlerFunc {
		return func(c *echo.Context) error {
			in, err := prepare[T](c, claims)
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
}

// List answers 200 with the page as JSON and the collection's total in X-Total-Count, written only
// once the handler succeeded. It fails as [One] does.
func List[T, R any](handler ListHandler[T, R]) Endpoint {
	return func(claims Claims) echo.HandlerFunc {
		return func(c *echo.Context) error {
			in, err := prepare[T](c, claims)
			if err != nil {
				return err
			}

			res, count, err := handler(in.ctx, in.scope, in.actor, in.req)
			if err != nil {
				return err
			}

			c.Response().Header().Set(totalCountHeader, strconv.FormatInt(count, 10))

			return c.JSON(http.StatusOK, res)
		}
	}
}

// None answers 200 with no body once the handler succeeded. It fails as [One] does.
func None[T any](handler NoneHandler[T]) Endpoint {
	return func(claims Claims) echo.HandlerFunc {
		return func(c *echo.Context) error {
			in, err := prepare[T](c, claims)
			if err != nil {
				return err
			}

			if err := handler(in.ctx, in.scope, in.actor, in.req); err != nil {
				return err
			}

			return c.NoContent(http.StatusOK)
		}
	}
}

// Legacy mounts a handler that still writes its own response through the gateway [Context]. It
// ignores the claims, which the handler resolves itself, and fails as [Handler] does.
func Legacy(next func(*Context) error) Endpoint {
	return func(Claims) echo.HandlerFunc {
		return Handler(next)
	}
}

type inputs[T any] struct {
	ctx   context.Context
	scope scope.Scope
	actor Actor
	req   *T
}

func prepare[T any](c *echo.Context, claims Claims) (inputs[T], error) {
	gCtx, ok := From(c)
	if !ok {
		return inputs[T]{}, echo.ErrInternalServerError
	}

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

	sc, err := resolveScope(c, gCtx, claims)
	if err != nil {
		return inputs[T]{}, err
	}

	actor, err := resolveActor(gCtx, claims)
	if err != nil {
		return inputs[T]{}, err
	}

	return inputs[T]{ctx: gCtx.Ctx(), scope: sc, actor: actor, req: req}, nil
}

func resolveScope(c *echo.Context, gCtx *Context, claims Claims) (scope.Scope, error) {
	if claims.Unbounded {
		return scope.NewUnbounded(c.Request().Method + " " + c.Path()), nil
	}

	return gCtx.AdminOrScope()
}

func resolveActor(gCtx *Context, claims Claims) (Actor, error) {
	identity := IdentityFrom(gCtx.Request().Header)

	actor := identity.Actor()

	if claims.Anonymous || !actor.IsZero() {
		return actor, nil
	}

	return Actor{}, routes.NewErrUnauthorized(nil)
}
