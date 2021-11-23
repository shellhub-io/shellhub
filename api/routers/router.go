package routers

import (
	"github.com/shellhub-io/shellhub/api/services"
)

type Router interface {
	// LoadMiddleware loads all required middleware.
	LoadMiddleware(service services.Service)

	// LoadRoutes loads all required routes.
	LoadRoutes(service services.Service)

	// ListenAndServe inits the HTTP server in a port.
	ListenAndServe(port string)
}
