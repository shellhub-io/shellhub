package main

import (
	"context"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/routes"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	mongooptions "github.com/shellhub-io/shellhub/api/store/mongo/options"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/worker"
	"github.com/shellhub-io/shellhub/pkg/worker/asynq"
	log "github.com/sirupsen/logrus"
)

type env struct {
	// MongoDB connection string.
	MongoURI string `env:"MONGO_URI,default=mongodb://mongo:27017/main"`

	// Redis connection string.
	RedisURI string `env:"REDIS_URI,default=redis://redis:6379"`

	// RedisCachePoolSize is the pool size of connections available for Redis cache.
	RedisCachePoolSize int `env:"REDIS_CACHE_POOL_SIZE,default=0"`

	// Enable GeoIP feature.
	//
	// GeoIP features enable the ability to get the logitude and latitude of the client from the IP address.
	// The feature is disabled by default. To enable it, it is required to have a `MAXMIND` database license and feed it
	// to `SHELLHUB_MAXMIND_LICENSE` with it, and `SHELLHUB_GEOIP=true`.
	GeoIP               bool   `env:"GEOIP,default=false"`
	GeoIPMaxMindLicense string `env:"MAXMIND_LICENSE,default="`

	// Session record cleanup worker schedule
	// Sentry DSN.
	SentryDSN string `env:"SENTRY_DSN,default="`
}

type Server struct {
	env    *env
	router *echo.Echo // TODO: evaluate if we can create a custom struct in router (e.g. router.Router)
	worker worker.Server
}

func (s *Server) Setup(ctx context.Context, env *env) error {
	s.env = env

	cache, err := cache.NewRedisCache(s.env.RedisURI, s.env.RedisCachePoolSize)
	if err != nil {
		log.WithError(err).
			Error("Failed to configure redis store cache")

		return err
	}

	_, db, err := mongo.Connect(ctx, s.env.MongoURI)
	if err != nil {
		log.WithError(err).
			Error("unable to connect to MongoDB")

		return err
	}

	store, err := mongo.NewStore(ctx, db, cache, mongooptions.RunMigatrions)
	if err != nil {
		log.WithError(err).
			Error("failed to create the store")

		return err
	}

	apiClient, err := internalclient.NewClient(internalclient.WithAsynqWorker(env.RedisURI))
	if err != nil {
		log.WithError(err).
			Error("failed to create the api client")

		return err
	}

	servicesOptions, err := s.serviceOptions()
	if err != nil {
		log.WithError(err).
			Error("failed to create the service's options")

		return err
	}

	service := services.NewService(store, nil, nil, cache, apiClient, servicesOptions...)

	s.worker = asynq.NewServer(s.env.RedisURI)
	s.worker.HandleTask(services.TaskDevicesHeartbeat, service.DevicesHeartbeat(), asynq.BatchTask())

	routerOptions, err := s.routerOptions()
	if err != nil {
		log.WithError(err).
			Error("failed to create the router's options")

		return err
	}

	s.router = routes.NewRouter(service, routerOptions...)

	return nil
}

// Start starts the server. It will block the routine.
func (s *Server) Start() error {
	if err := s.worker.Start(); err != nil {
		log.WithError(err).
			Error("Failed to start workers.")

		return err
	}

	if err := s.router.Start(":8080"); err != nil {
		log.WithError(err).
			Error("Failed to start router.")

		return err
	}

	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() {
	s.worker.Shutdown()
	s.router.Close() // nolint: errcheck
}

// serviceOptions returns a list of options used to start the services.
func (s *Server) serviceOptions() ([]services.Option, error) {
	opts := []services.Option{}

	if s.env.GeoIP {
		locator, err := geoip.NewGeoLite2(s.env.GeoIPMaxMindLicense)
		if err != nil {
			return []services.Option{}, err
		}

		opts = append(opts, services.WithLocator(locator))
	}

	return opts, nil
}

// routerOptions returns a list of options used to start the services.
func (s *Server) routerOptions() ([]routes.Option, error) {
	opts := []routes.Option{}

	if s.env.SentryDSN != "" {
		sentryOpts := sentry.ClientOptions{ //nolint:exhaustruct
			Dsn:              s.env.SentryDSN,
			Release:          os.Getenv("SHELLHUB_VERSION"),
			EnableTracing:    true,
			TracesSampleRate: 1,
		}

		reporter, err := sentry.NewClient(sentryOpts)
		if err != nil {
			log.WithError(err).Error("Failed to create Sentry client")

			return nil, err
		}

		opts = append(opts, routes.WithReporter(reporter))
	}

	return opts, nil
}
