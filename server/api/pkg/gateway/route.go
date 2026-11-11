package gateway

import (
	"context"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/binding"
	"github.com/shellhub-io/shellhub/pkg/api/identity"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	routes "github.com/shellhub-io/shellhub/server/api/routes/errors"
)

const totalCountHeader = "X-Total-Count"

// OneHandler answers with a single value, or the error the route fails with.
type OneHandler[A Caller, T, R any] func(ctx context.Context, sc scope.Scope, caller A, req *T) (R, error)

// ListHandler answers with a page of values and the size of the whole collection. [List] writes
// that size to the X-Total-Count response header.
type ListHandler[A Caller, T, R any] func(ctx context.Context, sc scope.Scope, caller A, req *T) (R, int64, error)

// NoneHandler answers with success alone.
type NoneHandler[A Caller, T any] func(ctx context.Context, sc scope.Scope, caller A, req *T) error

// Claims are the exception a route takes to the rule that it is bounded to the caller's
// namespace: Unbounded spans every namespace instead.
type Claims struct {
	Unbounded bool
}

// Endpoint is a handler waiting for the claims of the route that mounts it, together with the
// kind of caller the handler takes. Build one with [One], [List], [None], [Legacy] or [Raw].
type Endpoint struct {
	caller CallerKind
	build  func(Claims) echo.HandlerFunc
}

// Caller returns the kind of caller e's handler takes, [CallerUnstated] for [Legacy] and [Raw].
func (e Endpoint) Caller() CallerKind {
	return e.caller
}

// Handler returns e's handler for a route making claims. It panics on the zero Endpoint.
func (e Endpoint) Handler(claims Claims) echo.HandlerFunc {
	return e.build(claims)
}

// One answers 200 with the handler's result as JSON. It panics when T breaks the rule of package
// binding, so a misdeclared request stops the route table from being built. It fails with 500 when
// no gateway [Context] was installed, with the bind or validation error for a malformed request,
// with 403 when the request carries no namespace and the route is not [Claims.Unbounded], with 401
// when the request carries no identity and A is not [Anonymous], with 403 when the identity is not
// of the kind A names, and otherwise with the handler's error.
func One[A Caller, T, R any](handler OneHandler[A, T, R]) Endpoint {
	binding.MustCheck[T]()

	return Endpoint{caller: callerKind[A](), build: func(claims Claims) echo.HandlerFunc {
		return func(c *echo.Context) error {
			in, err := prepare[A, T](c, claims)
			if err != nil {
				return err
			}

			res, err := handler(in.ctx, in.scope, in.caller, in.req)
			if err != nil {
				return err
			}

			return c.JSON(http.StatusOK, res)
		}
	}}
}

// List answers 200 with the page as JSON and the collection's total in X-Total-Count, written only
// once the handler succeeded. It panics and fails as [One] does.
func List[A Caller, T, R any](handler ListHandler[A, T, R]) Endpoint {
	binding.MustCheck[T]()

	return Endpoint{caller: callerKind[A](), build: func(claims Claims) echo.HandlerFunc {
		return func(c *echo.Context) error {
			in, err := prepare[A, T](c, claims)
			if err != nil {
				return err
			}

			res, count, err := handler(in.ctx, in.scope, in.caller, in.req)
			if err != nil {
				return err
			}

			c.Response().Header().Set(totalCountHeader, strconv.FormatInt(count, 10))

			return c.JSON(http.StatusOK, res)
		}
	}}
}

// None answers 200 with no body once the handler succeeded. It panics and fails as [One] does.
func None[A Caller, T any](handler NoneHandler[A, T]) Endpoint {
	binding.MustCheck[T]()

	return Endpoint{caller: callerKind[A](), build: func(claims Claims) echo.HandlerFunc {
		return func(c *echo.Context) error {
			in, err := prepare[A, T](c, claims)
			if err != nil {
				return err
			}

			if err := handler(in.ctx, in.scope, in.caller, in.req); err != nil {
				return err
			}

			return c.NoContent(http.StatusOK)
		}
	}}
}

// Legacy mounts a handler that still writes its own response through the gateway [Context]. It
// names no caller, ignores the claims, which the handler resolves itself, and fails as [Handler]
// does.
func Legacy(next func(*Context) error) Endpoint {
	return Endpoint{caller: CallerUnstated, build: func(Claims) echo.HandlerFunc {
		return Handler(next)
	}}
}

// Raw mounts a handler that writes its own response from the echo context. It names no caller and
// ignores the claims, and needs no gateway [Context].
func Raw(handler echo.HandlerFunc) Endpoint {
	return Endpoint{caller: CallerUnstated, build: func(Claims) echo.HandlerFunc {
		return handler
	}}
}

type inputs[A Caller, T any] struct {
	ctx    context.Context
	scope  scope.Scope
	caller A
	req    *T
}

func prepare[A Caller, T any](c *echo.Context, claims Claims) (inputs[A, T], error) {
	gCtx, ok := From(c)
	if !ok {
		return inputs[A, T]{}, echo.ErrInternalServerError
	}

	req := new(T)
	if err := c.Bind(req); err != nil {
		return inputs[A, T]{}, err
	}

	if paginated, ok := any(req).(query.Paginated); ok {
		paginated.GetPaginator().Normalize()
	}

	if sorted, ok := any(req).(query.Sorted); ok {
		sorted.GetSorter().Normalize()
	}

	if err := c.Validate(req); err != nil {
		return inputs[A, T]{}, err
	}

	sc, err := resolveScope(c, gCtx, claims)
	if err != nil {
		return inputs[A, T]{}, err
	}

	caller, err := resolveCaller[A](gCtx)
	if err != nil {
		return inputs[A, T]{}, err
	}

	return inputs[A, T]{ctx: gCtx.Ctx(), scope: sc, caller: caller, req: req}, nil
}

func resolveScope(c *echo.Context, gCtx *Context, claims Claims) (scope.Scope, error) {
	if claims.Unbounded {
		return scope.NewUnbounded(c.Request().Method + " " + c.Path()), nil
	}

	return gCtx.AdminOrScope()
}

func resolveCaller[A Caller](gCtx *Context) (A, error) {
	var zero A

	resolved, authenticated := identity.FromContext(gCtx.Ctx())
	if !authenticated && callerKind[A]() != CallerAnonymous {
		return zero, routes.NewErrUnauthorized(nil)
	}

	caller, ok := callerOf[A](resolved)
	if !ok {
		return zero, routes.NewErrForbidden(nil)
	}

	return caller, nil
}
