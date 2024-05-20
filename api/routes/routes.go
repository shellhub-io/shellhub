package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/routes/middleware"
	"github.com/shellhub-io/shellhub/api/services"
	log "github.com/sirupsen/logrus"
)

// Group specifies the visibility level of a given route.
type Group int

func (g Group) String() string {
	switch g {
	case GroupPublic:
		return "public"
	case GroupInternal:
		return "internal"
	default:
		return "disabled"
	}
}

const (
	GroupPublic   Group = iota // GroupPublic indicates that a route is accessible to everyone.
	GroupInternal              // GroupPrivate indicates that a route is restricted and can only be accessed by other services within the local container network.
	GroupDisable               // GroupDisable indicates that the route is disabled.
)

// HTTPMethod is a type that represents an HTTP method, similar to http.Method[method].
type HTTPMethod string

const (
	MethodGet    HTTPMethod = http.MethodGet    // MethodGet is the equivalent of http.MethodGet
	MethodPost   HTTPMethod = http.MethodPost   // MethodPost is the equivalent of http.MethodPost
	MethodPatch  HTTPMethod = http.MethodPatch  // MethodPatch is the equivalent of http.MethodPatch
	MethodPut    HTTPMethod = http.MethodPut    // MethodPut is the equivalent of http.MethodPut
	MethodDelete HTTPMethod = http.MethodDelete // MethodDelete is the equivalent of http.MethodDelete
)

type Route struct {
	endpoint              string                // endpoint specifies the URL path where the handler is attached.
	deprecatedEndpoints   []string              // deprecatedEndpoints is a list of deprecated endpoints that use the same method and handler as [Route.endpoint].
	method                HTTPMethod            // method defines the HTTP method for which the handler is associated.
	group                 Group                 // group specifies the visibility level of the route for external requests.
	blockAPIKey           bool                  // blockAPIKey specifies whether the route allows or not authentication via api key.
	requiresAuthorization bool                  // requiresAuthorization specifies whether the handler should be called within middleware.Authorize
	middlewares           []echo.MiddlewareFunc // middlewares is a list of middleware functions to be applied to the route.
	handler               gateway.Handler       // handler is the callback that is invoked when the route is accessed.
}

func NewRouter(service services.Service) *echo.Echo {
	e := echo.New()
	e.Binder = handlers.NewBinder()
	e.Validator = handlers.NewValidator()
	e.HTTPErrorHandler = handlers.NewErrors(nil)
	e.IPExtractor = echo.ExtractIPFromRealIPHeader()

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apicontext := gateway.NewContext(service, c)

			return next(apicontext)
		}
	})

	handler := NewHandler(service)

	internalGroup := e.Group("/internal") // Internal routes are restricted and can only be accessed by other services within the local container network.
	publicGroup := e.Group("/api")        // Public routes are accessible externally through API gateway
	publicGroup.GET("/healthcheck", gateway.Handle(func(c gateway.Context) error {
		return c.NoContent(200)
	}))

	for _, h := range handler.all() {
		middlewares := make([]echo.MiddlewareFunc, 0)
		if h.blockAPIKey {
			middlewares = append(middlewares, middleware.BlockAPIKey)
		}
		for _, m := range h.middlewares {
			middlewares = append(middlewares, gateway.Middleware(m))
		}

		handler := gateway.Handle(h.handler)
		if h.requiresAuthorization {
			handler = middleware.Authorize(gateway.Handle(h.handler))
		}

		log.WithField("endpoint", h.endpoint).
			WithField("method", h.method).
			WithField("group", h.group.String()).
			Trace("registering route.")

		switch h.group {
		case GroupPublic:
			for _, e := range append(h.deprecatedEndpoints, h.endpoint) {
				publicGroup.Add(string(h.method), e, handler, middlewares...)
			}
		case GroupInternal:
			for _, e := range append(h.deprecatedEndpoints, h.endpoint) {
				internalGroup.Add(string(h.method), e, handler, middlewares...)
			}
		default:
			log.WithField("endpoint", h.endpoint).Trace("route is disabled.")

			continue
		}
	}

	return e
}
