package routes

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo-contrib/v5/echoprometheus"
	"github.com/labstack/echo/v5"
	echoMiddleware "github.com/labstack/echo/v5/middleware"
	"github.com/shellhub-io/shellhub/pkg/envs"
	pkgmiddleware "github.com/shellhub-io/shellhub/pkg/middleware"
	"github.com/shellhub-io/shellhub/server/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/sirupsen/logrus"
)

const defaultBodyLimit = 1 << 20

// DefaultHTTPHandler creates an HTTP handler, using [github.com/labstack/echo/v5] package, with the default
// configuration required by ShellHub's services and the [github.com/shellhub-io/shellhub/server/api/pkg/gateway]
// loaded into the context. Its error handler reports nowhere until [WithReporter] replaces it. Request bodies are
// capped at 1 MiB: a read past the cap returns [echo.ErrStatusRequestEntityTooLarge], which the error handler
// answers with 413, and a route replaces the cap with [routesmiddleware.BodyLimit].
func DefaultHTTPHandler() http.Handler {
	server := echo.New()

	server.Binder = handlers.NewBinder()

	server.Validator = handlers.NewValidator()

	server.HTTPErrorHandler = handlers.NewErrors(nil)

	server.IPExtractor = handlers.RealIPExtractor()

	server.Logger = pkgmiddleware.NewSlogLogger(logrus.NewEntry(logrus.StandardLogger()))

	server.Use(echoMiddleware.RequestIDWithConfig(echoMiddleware.RequestIDConfig{
		RequestIDHandler: func(c *echo.Context, id string) {
			c.Request().Header.Set(echo.HeaderXRequestID, id)
		},
	}))
	server.Use(echoMiddleware.Secure())
	server.Use(gateway.WithContext())
	server.Use(pkgmiddleware.Log)
	server.Use(routesmiddleware.DefaultBodyLimit(defaultBodyLimit))

	return server
}

// Option installs a cross-cutting concern on the router. Returning an error aborts startup,
// so a misconfigured option fails the process rather than serving without it.
type Option func(e *echo.Echo, handler *Handler) error

// WithReporter sends unhandled errors to Sentry through reporter.
func WithReporter(reporter *sentry.Client) Option {
	return func(e *echo.Echo, _ *Handler) error {
		e.HTTPErrorHandler = handlers.NewErrors(reporter)

		return nil
	}
}

// WithMetrics records request metrics and serves them at /metrics.
func WithMetrics() Option {
	return func(e *echo.Echo, _ *Handler) error {
		e.Use(echoprometheus.NewMiddleware("api"))
		e.GET("/metrics", echoprometheus.NewHandler())

		return nil
	}
}

// WithOpenAPIValidator validates every response against the published spec. In strict mode the
// schema loads now, and a schema that does not load is the error the router refuses to build on.
func WithOpenAPIValidator(ctx context.Context, cfg *routesmiddleware.OpenAPIValidatorConfig) Option {
	return func(e *echo.Echo, _ *Handler) error {
		if !cfg.Strict {
			e.Use(routesmiddleware.OpenAPIValidator(cfg))

			return nil
		}

		validate, err := routesmiddleware.LoadOpenAPIValidator(ctx, cfg)
		if err != nil {
			return err
		}

		e.Use(validate)

		return nil
	}
}

// WithAuthentication authenticates every request in-process, replacing the
// per-route authentication subrequest the edge proxy used to issue.
//
// It is an option rather than a default so handler tests can keep exercising a
// route's authorization with an identity put straight on the request's context
// through identitytest.As. Every production entrypoint must pass it;
// [TestRouterRejectsUncredentialedRequests] guards that the wiring stays in place.
func WithAuthentication(authn *routesmiddleware.Authenticator) Option {
	return func(e *echo.Echo, handler *Handler) error {
		handler.WithAuthenticator(authn)
		e.Use(authn.Middleware)

		return nil
	}
}

// WithEdition sets the edition that decides which routes the router registers. Without it the
// router registers the Community routes.
func WithEdition(edition envs.Edition) Option {
	return func(_ *echo.Echo, handler *Handler) error {
		handler.edition = edition

		return nil
	}
}

// WithDeviceListMaxPageSize sets the largest page a cursor read of the device list returns. A
// larger page[size] is lowered to it, and a size below one means query.DefaultMaxPageSize.
func WithDeviceListMaxPageSize(size int64) Option {
	return func(_ *echo.Echo, handler *Handler) error {
		handler.deviceListMaxPageSize = size

		return nil
	}
}

// WithGeoIPSource names the GeoIP database the server locates devices with, dbip or maxmind, which
// the device map reads from the X-GeoIP-Source header to credit it. Empty, the header is left out.
func WithGeoIPSource(source string) Option {
	return func(_ *echo.Echo, handler *Handler) error {
		handler.geoipSource = source

		return nil
	}
}

// NewRouter builds the API router over service, applying each option in turn.
// It returns a nil router and an error wrapping the option's own error when an
// option fails, or an error when [DefaultHTTPHandler] does not build an
// [echo.Echo].
func NewRouter(service services.Service, opts ...Option) (*echo.Echo, error) {
	router, ok := DefaultHTTPHandler().(*echo.Echo)
	if !ok {
		return nil, errors.New("the default HTTP handler is not an echo router")
	}

	handler := NewHandler(service)
	for _, opt := range opts {
		if err := opt(router, handler); err != nil {
			return nil, fmt.Errorf("failed to apply a router option: %w", err)
		}
	}

	table := handler.publicRoutes()
	mount(router.Group(publicAPIPrefix), table)

	SetupMCPRoutes(router)

	registerInternalMetrics(router, handler.authn)

	if handler.authn != nil {
		allow(handler.authn, table)
	}

	if err := applyExtensions(router, handler.authn, service); err != nil {
		logrus.WithError(err).Error("failed to apply route extensions")
	}

	router.Pre(echoMiddleware.Rewrite(map[string]string{
		"/api/containers":   "/api/devices?connector=true",
		"/api/containers?*": "/api/devices?connector=true&$1",
		"/api/containers/*": "/api/devices/$1",
	}))

	router.Pre(echoMiddleware.Rewrite(rootAliases))

	return router, nil
}
