package server

import (
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/echo/handlers"
)

// Echo is a wrapper around the Echo HTTP server with simplified lifecycle management.
type Echo struct {
	echo *echo.Echo
}

var _ Server[*echo.Echo] = new(Echo)

// Underlying returns the underlying HTTP server.
func (s *Echo) Underlying() *echo.Echo {
	return s.echo
}

// Close closes the server.
func (s *Echo) Close() error {
	return s.echo.Close()
}

// Start starts the server at a given address.
func (s *Echo) Start(addr string) error {
	return s.echo.Start(addr)
}

// Listen starts the HTTP server, listing for connections in [ServerListenDefaultAddress].
func (s *Echo) Listen() error {
	return s.echo.Start(ServerListenDefaultAddress)
}

// SentryOption enables, if DSN is a valid value, the error reporter for a Sentry's server.
var SentryOption = func(dsn string) func(server *echo.Echo) {
	return func(server *echo.Echo) {
		if dsn != "" {
			reporter, err := sentry.NewClient(sentry.ClientOptions{ //nolint:exhaustruct
				Dsn:              dsn,
				Release:          os.Getenv("SHELLHUB_VERSION"),
				EnableTracing:    true,
				TracesSampleRate: 1,
			})
			if err != nil {
				server.HTTPErrorHandler = handlers.NewErrors(nil)

				return
			}

			server.HTTPErrorHandler = handlers.NewErrors(reporter)
		}
	}
}
