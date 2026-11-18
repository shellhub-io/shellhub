package app

import (
	"context"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/worker"
	"github.com/shellhub-io/shellhub/server/api/routes/middleware"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
)

// Extension adds an edition's features to the server. [Run] opens each one once, during setup,
// after the store, the cache, the clock and the task queue client exist and before anything that
// serves requests is built. An error stops the server from starting.
type Extension func(ctx context.Context, infra Infra) (Features, error)

// Infra is what the server has built before it opens its extensions. An extension uses these
// rather than opening its own connections, so the process holds one of each.
type Infra struct {
	Env   Env
	Store store.Store
	Cache cache.Cache
	Clock clock.Clock
	// Tasks submits tasks to the worker queue the server runs.
	Tasks worker.Client
}

// Features is what an extension adds. A zero field adds nothing.
type Features struct {
	// Store replaces the store every later part of the server reads, typically a wrapper over
	// [Infra.Store].
	Store store.Store

	// MaxMind opens the GeoLite2 locator, which the server asks for only when
	// SHELLHUB_GEOIP_DATABASE is maxmind. It may return (nil, nil) when it has no licence or
	// mirror to load from; that, or an error, makes the server fall back to DB-IP.
	MaxMind func(ctx context.Context) (geoip.Locator, error)

	// Services returns the options the extension adds to the service layer, given the locator
	// the server chose.
	Services func(locator geoip.Locator) []services.Option

	// Routes mounts the extension's routes, after the core routes and the SSH routes, so a path
	// both mount answers with the extension's handler. Every route must go through
	// routetable.Mount on Mount.Router; one added any other way, or an error Routes returns,
	// stops the server from starting.
	Routes func(Mount) error

	// Workers registers the extension's background tasks.
	Workers func(worker.Server)
}

// Mount is what an extension's routes are mounted with: Router is the router its route tables are
// mounted on through routetable.Mount, Authn the authenticator that guards them, Service the core
// service its handlers call, and Tunnels the dialer that reaches a device's tunnel.
type Mount struct {
	Router  *echo.Echo
	Authn   *middleware.Authenticator
	Service services.Service
	Tunnels dialer.TunnelDialer
}

func openExtensions(ctx context.Context, extensions []Extension, infra Infra) (Features, error) {
	opened := make([]Features, 0, len(extensions))

	for _, extension := range extensions {
		features, err := extension(ctx, infra)
		if err != nil {
			return Features{}, err
		}

		if features.Store != nil {
			infra.Store = features.Store
		}

		opened = append(opened, features)
	}

	merged := Merge(opened...)
	merged.Store = infra.Store

	return merged, nil
}

// Merge joins features into one that adds what each adds, in order. The last part with a Store or
// a MaxMind wins.
func Merge(parts ...Features) Features {
	var merged Features

	for _, part := range parts {
		if part.Store != nil {
			merged.Store = part.Store
		}

		if part.MaxMind != nil {
			merged.MaxMind = part.MaxMind
		}
	}

	merged.Services = func(locator geoip.Locator) []services.Option {
		var options []services.Option

		for _, part := range parts {
			if part.Services != nil {
				options = append(options, part.Services(locator)...)
			}
		}

		return options
	}

	merged.Routes = func(m Mount) error {
		for _, part := range parts {
			if part.Routes == nil {
				continue
			}

			if err := part.Routes(m); err != nil {
				return err
			}
		}

		return nil
	}

	merged.Workers = func(server worker.Server) {
		for _, part := range parts {
			if part.Workers != nil {
				part.Workers(server)
			}
		}
	}

	return merged
}
