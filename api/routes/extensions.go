package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/worker"
	log "github.com/sirupsen/logrus"
)

// RouteExtension is a function that registers additional HTTP routes on the Echo router.
// This is the extension point for enterprise/cloud features to add their own HTTP endpoints.
//
// Extensions receive:
// - router: The Echo instance to register routes on
// - service: The core service providing access to shared infrastructure (store, cache, etc.)
//
// Extensions should return an error if route registration fails.
//
// Example usage:
//
//	routes.RegisterRouteExtension(func(router *echo.Echo, service services.Service) error {
//	    adminGroup := router.Group("/api/admin")
//	    adminGroup.POST("/users", adminHandler.CreateUser)
//	    return nil
//	})
type RouteExtension func(router *echo.Echo, service services.Service) error

// routeExtensions holds all registered route extensions.
// Extensions are typically registered by enterprise/cloud builds in init() or main().
var routeExtensions []RouteExtension

// RegisterRouteExtension registers a route extension function.
// This must be called before NewRouter() is invoked.
//
// Extensions are applied in the order they are registered.
func RegisterRouteExtension(ext RouteExtension) {
	routeExtensions = append(routeExtensions, ext)
}

// applyExtensions invokes all registered extensions on the router.
// Returns an error if any extension fails to register its routes.
func applyExtensions(router *echo.Echo, service services.Service) error {
	for _, ext := range routeExtensions {
		if err := ext(router, service); err != nil {
			log.WithError(err).Error("failed to apply route extension")

			return err
		}
	}

	return nil
}

// WorkerExtension is a function that registers additional task handlers on the
// worker server. Cloud packages register handlers for cloud-specific background
// tasks (e.g., member invitations) via this extension point.
type WorkerExtension func(server worker.Server, store store.Store, cache cache.Cache)

// workerExtensions holds all registered worker extensions.
var workerExtensions []WorkerExtension

// RegisterWorkerExtension registers a worker extension. Must be called before
// server.Setup() — typically from a cloud package's init() function.
func RegisterWorkerExtension(ext WorkerExtension) {
	workerExtensions = append(workerExtensions, ext)
}

// ApplyWorkerExtensions invokes all registered worker extensions, passing the
// worker server so each extension can register its own task handlers.
func ApplyWorkerExtensions(server worker.Server, store store.Store, cache cache.Cache) {
	for _, ext := range workerExtensions {
		ext(server, store, cache)
	}
}
