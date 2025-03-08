package main

import (
	"context"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/routes"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store/pg"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/geoip/geolite2"
	"github.com/shellhub-io/shellhub/pkg/worker"
	"github.com/shellhub-io/shellhub/pkg/worker/asynq"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	env    *env
	router *echo.Echo // TODO: evaluate if we can create a custom struct in router (e.g. router.Router)
	worker worker.Server
}

// Setup initializes all server components including database connections, cache, services, API routes, and background workers.
// It prepares the server for starting but does not actually begin serving requests.
func (s *Server) Setup(ctx context.Context) error {
	log.Info("Setting up server components")

	cache, err := cache.NewRedisCache(s.env.RedisURI, s.env.RedisCachePoolSize)
	if err != nil {
		return err
	}

	log.Debug("Redis cache initialized successfully")

	dsn := pg.DSN(s.env.PostgresHost, s.env.PostgresPort, s.env.PostgresUser, s.env.PostgresPassword, s.env.PostgresDB)
	store, err := pg.New(ctx, dsn, cache)
	if err != nil {
		return err
	}

	log.Debug("Postgres store connected successfully")

	apiClient, err := internalclient.NewClient(internalclient.WithAsynqWorker(s.env.RedisURI))
	if err != nil {
		return err
	}

	servicesOptions, err := s.serviceOptions(ctx)
	if err != nil {
		return err
	}

	routerOptions, err := s.routerOptions()
	if err != nil {
		return err
	}

	service := services.NewService(store, nil, nil, cache, apiClient, servicesOptions...)
	s.router = routes.NewRouter(service, routerOptions...)
	s.worker = asynq.NewServer(s.env.RedisURI)
	s.worker.HandleTask(services.TaskDevicesHeartbeat, service.DevicesHeartbeat(), asynq.BatchTask())

	log.Info("Server setup completed successfully")

	return nil
}

// Start begins serving API requests and processing background tasks. It blocks the current goroutine until the server stops
// or encounters an error.
func (s *Server) Start() error {
	log.Info("Starting server components")

	if err := s.worker.Start(); err != nil {
		return err
	}

	if err := s.router.Start(":8080"); err != nil {
		return err
	}

	return nil
}

// Shutdown gracefully terminates all server components.
func (s *Server) Shutdown() {
	log.Info("Gracefully shutting down server")

	s.worker.Shutdown()
	s.router.Close() // nolint: errcheck

	log.Info("Server shutdown complete")
}

// serviceOptions returns configuration options for the application services.
func (s *Server) serviceOptions(ctx context.Context) ([]services.Option, error) {
	opts := []services.Option{}

	var geoipFetcher geolite2.GeoliteFetcher
	switch {
	case s.env.GeoipMirror != "":
		log.Info("Using custom mirror for GeoIP database")
		geoipFetcher = geolite2.FetchFromMirror(s.env.GeoipMirror)
	case s.env.GeoipMaxmindLicense != "":
		log.Info("Using MaxMind license key for GeoIP database")
		geoipFetcher = geolite2.FetchFromLicenseKey(s.env.GeoipMaxmindLicense)
	}

	if geoipFetcher != nil {
		locator, err := geolite2.NewLocator(ctx, geoipFetcher)
		if err != nil {
			return nil, err
		}

		log.Info("GeoIP locator initialized successfully")

		opts = append(opts, services.WithLocator(locator))
	}

	return opts, nil
}

// routerOptions returns configuration options for the HTTP router.
func (s *Server) routerOptions() ([]routes.Option, error) {
	opts := []routes.Option{}

	if s.env.SentryDSN != "" {
		log.Info("Initializing Sentry error reporting")

		sentryOpts := sentry.ClientOptions{ //nolint:exhaustruct
			Dsn:              s.env.SentryDSN,
			Release:          os.Getenv("SHELLHUB_VERSION"),
			EnableTracing:    true,
			TracesSampleRate: 1,
		}

		reporter, err := sentry.NewClient(sentryOpts)
		if err != nil {
			return nil, err
		}

		log.Info("Sentry error reporting initialized successfully")

		opts = append(opts, routes.WithReporter(reporter))
	}

	return opts, nil
}
