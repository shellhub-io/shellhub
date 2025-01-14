// Package server The server package provides a flexible and configurable HTTP server framework for ShellHub services.
// It offers a generic server creation mechanism with built-in middleware, error handling, and routing capabilities.
package server

import (
	"context"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/shellhub-io/shellhub/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	pkgmiddleware "github.com/shellhub-io/shellhub/pkg/middleware"
)

// ServerListenDefaultAddress is the default address used by the HTTP server to listen for connections.
const ServerListenDefaultAddress = ":8080"

// Server is the interface that should be implemented to create a HTTP server for ShellHub services.
type Server[S any] interface {
	// Underlying returns the underlying HTTP server used.
	Underlying() S
	// Close closes the server.
	Close() error
	// Start starts the server at a given address.
	Start(addr string) error
	// Listen starts the HTTP server, listing for connections in [ServerListenDefaultAddress].
	Listen() error
}

// Route represents a loader of routes to the underlying HTTP server.
type Route[H any] func(server *echo.Echo, handler H)

// Option is used to pass custom configurations to the underlying HTTP server.
type Option func(server *echo.Echo)

type Handler[S any] interface {
	GetService() S
}

// NewDefaultServer uses [echo] to create a default HTTP server meet to be used as ShellHub services, aggregating the
// middlewares, binder, validator, environmental variables, features, and anything related to service providing.
func NewDefaultServer[H Handler[any]](
	ctx context.Context,
	handler H,
	middlewares []echo.MiddlewareFunc,
	routes []Route[H],
	options []Option,
) Server[*echo.Echo] { //nolint:whitespace
	e := echo.New()

	e.Binder = handlers.NewBinder()
	e.Validator = handlers.NewValidator()
	e.HTTPErrorHandler = handlers.NewErrors(nil)
	e.IPExtractor = echo.ExtractIPFromRealIPHeader()

	for _, option := range options {
		option(e)
	}

	// NOTE: As recommend by Echo's, this middleware should be registered before any other middleware, and custom
	// context cannot be defined in a middleware before the router ran (Pre)
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := gateway.NewContext(handler.GetService(), c)

			return next(ctx)
		}
	})
	e.Use(middleware.RequestID())
	e.Use(middleware.Secure())
	e.Use(pkgmiddleware.Log)

	if middlewares != nil {
		e.Use(middlewares...)
	}

	for _, route := range routes {
		route(e, handler)
	}

	// NOTE: When context received is done, we close the HTTP server.
	go func() {
		<-ctx.Done()

		e.Close()
	}()

	return &Echo{
		echo: e,
	}
}
