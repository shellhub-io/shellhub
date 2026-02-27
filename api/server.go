package main

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/routes"
	"github.com/shellhub-io/shellhub/api/routes/middleware"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/migrate"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	mongooptions "github.com/shellhub-io/shellhub/api/store/mongo/options"
	"github.com/shellhub-io/shellhub/api/store/pg"
	pgoptions "github.com/shellhub-io/shellhub/api/store/pg/options"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/geoip/geolite2"
	"github.com/shellhub-io/shellhub/pkg/worker"
	"github.com/shellhub-io/shellhub/pkg/worker/asynq"
	log "github.com/sirupsen/logrus"

	// Blank import triggers init() functions in cloud packages when built with -tags enterprise.
	// In CE builds this package is empty and compiles to nothing.
	_ "github.com/shellhub-io/shellhub/api/enterprise"
)

type env struct {
	Database string `env:"DATABASE,default=mongo"`

	// PostgresHost specifies the host for PostgreSQL.
	PostgresHost string `env:"POSTGRES_HOST,default=postgres"`
	// PostgresPort specifies the port for PostgreSQL.
	PostgresPort string `env:"POSTGRES_PORT,default=5432"`
	// PostgresUsername specifies the username for authenticate PostgreSQL.
	PostgresUsername string `env:"POSTGRES_USERNAME,default=admin"`
	// PostgresUser specifies the password for authenticate PostgreSQL.
	PostgresPassword string `env:"POSTGRES_PASSWORD,default=admin"`
	// PostgresDatabase especifica o nome do banco de dados PostgreSQL a ser utilizado.
	PostgresDatabase string `env:"POSTGRES_DATABASE,default=main"`

	// MongoURI specifies the connection string for MongoDB.
	MongoURI string `env:"MONGO_URI,default=mongodb://mongo:27017/main"`

	// RedisURI specifies the connection string for Redis.
	RedisURI string `env:"REDIS_URI,default=redis://redis:6379"`
	// RedisCachePoolSize defines the maximum number of concurrent connections to Redis cache.
	// Set to 0 for unlimited connections.
	RedisCachePoolSize int `env:"REDIS_CACHE_POOL_SIZE,default=0"`

	// SentryDSN specifies the Data Source Name for Sentry error tracking.
	// Leave empty to disable Sentry integration.
	SentryDSN string `env:"SENTRY_DSN,default="`

	// AsynqGroupMaxDelay specifies the maximum time (in seconds) to wait before
	// processing a group of tasks, regardless of other conditions.
	AsynqGroupMaxDelay int `env:"ASYNQ_GROUP_MAX_DELAY,default=1"`
	// AsynqGroupGracePeriod defines the grace period (in seconds) before task aggregation.
	// Tasks arriving within this period will be aggregated with existing tasks in the group.
	AsynqGroupGracePeriod int64 `env:"ASYNQ_GROUP_GRACE_PERIOD,default=2"`
	// AsynqGroupMaxSize specifies the maximum number of tasks that can be aggregated in a group.
	// When this limit is reached, the group will be processed immediately.
	AsynqGroupMaxSize int `env:"ASYNQ_GROUP_MAX_SIZE,default=1000"`
	// AsynqUniquenessTimeout defines how long (in hours) a unique job remains locked in the queue.
	// If a job doesn't complete within this period, its lock is released, allowing a new instance
	// to be enqueued and executed.
	AsynqUniquenessTimeout int `env:"ASYNQ_UNIQUENESS_TIMEOUT,default=24"`

	// GeoipMirror specifies an alternative URL for downloading GeoIP databases.
	// When configured, this takes precedence over GeoipMaxmindLicense.
	GeoipMirror string `env:"MAXMIND_MIRROR,default="`
	// GeoipMaxmindLicense is the MaxMind license key for downloading GeoIP databases directly.
	// This is used as a fallback when GeoipMirror is not configured.
	GeoipMaxmindLicense string `env:"MAXMIND_LICENSE,default="`

	// Metrics enables the /metrics endpoint.
	Metrics bool `env:"METRICS,default=false"`
}

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

	// Capture the wrapper factory before the local "store" variable shadows the
	// package name. In CE builds this returns nil.
	wrapperFactory := store.StoreWrapper()

	var store store.Store
	switch s.env.Database {
	case "mongo":
		store, err = mongo.NewStore(ctx, s.env.MongoURI, cache, mongooptions.RunMigatrions)
	case "postgres":
		uri := pg.URI(s.env.PostgresHost, s.env.PostgresPort, s.env.PostgresUsername, s.env.PostgresPassword, s.env.PostgresDatabase)
		store, err = pg.New(ctx, uri, pgoptions.Log("INFO", true), pgoptions.Migrate()) // TODO: Log envs
	case "migrate":
		mongoStore, mongoErr := mongo.NewStore(ctx, s.env.MongoURI, cache)
		if mongoErr != nil {
			log.WithError(mongoErr).Fatal("failed to connect to MongoDB for migration")
		}

		uri := pg.URI(s.env.PostgresHost, s.env.PostgresPort, s.env.PostgresUsername, s.env.PostgresPassword, s.env.PostgresDatabase)
		pgStore, pgErr := pg.New(ctx, uri, pgoptions.Log("INFO", true), pgoptions.Migrate())
		if pgErr != nil {
			log.WithError(pgErr).Fatal("failed to connect to PostgreSQL for migration")
		}

		migrator := migrate.New(mongoStore.(*mongo.Store).GetDB(), pgStore.(*pg.Pg).Driver())
		if err := migrator.Run(ctx); err != nil {
			log.WithError(err).Fatal("migration failed")
		}

		log.Info("Migration completed successfully")

		os.Exit(0)
	default:
		log.WithField("database", s.env.Database).Error("invalid database")

		return errors.New("invalid database")
	}
	if err != nil {
		log.
			WithError(err).
			Fatal("failed to create the store")
	}

	log.WithField("database", s.env.Database).Info("store connected successfully")

	// If a store wrapper factory was registered (EE/cloud build), wrap the
	// store so cloud-specific entity overrides are used by the core service.
	if wrapperFactory != nil {
		store, err = wrapperFactory(store, cache)
		if err != nil {
			return errors.Join(errors.New("failed to wrap store"), err)
		}

		log.Info("Store wrapper applied")
	}

	apiClient, err := internalclient.NewClient(nil, internalclient.WithAsynqWorker(s.env.RedisURI))
	if err != nil {
		return err
	}

	servicesOptions, err := s.serviceOptions(ctx)
	if err != nil {
		return err
	}

	// If a billing provider factory was registered (EE/cloud build), create and
	// inject the billing provider before the service is constructed.
	if factory := services.BillingFactory(); factory != nil {
		log.Info("Billing provider factory registered; initializing billing provider")

		billing, err := factory(ctx, store, cache)
		if err != nil {
			return errors.Join(errors.New("failed to initialize billing provider"), err)
		}

		if billing != nil {
			servicesOptions = append(servicesOptions, services.WithBilling(billing))
		}

		log.Info("Billing provider initialized and injected into service")
	}

	routerOptions, err := s.routerOptions()
	if err != nil {
		return err
	}

	service := services.NewService(store, nil, nil, cache, apiClient, servicesOptions...)
	s.router = routes.NewRouter(service, routerOptions...)

	s.worker = asynq.NewServer(
		s.env.RedisURI,
		asynq.BatchConfig(s.env.AsynqGroupMaxSize, s.env.AsynqGroupMaxDelay, int(s.env.AsynqGroupGracePeriod)),
		asynq.UniquenessTimeout(s.env.AsynqUniquenessTimeout),
	)

	s.worker.HandleTask(services.TaskDevicesHeartbeat, service.DevicesHeartbeat(), asynq.BatchTask())
	s.worker.HandleCron(services.CronDeviceCleanup, service.DeviceCleanup(), asynq.Unique())
	s.worker.HandleCron(services.CronNamespaceDeviceCountSync, service.NamespaceDeviceCountSync(), asynq.Unique())

	// Apply any worker extensions registered by cloud/enterprise packages.
	routes.ApplyWorkerExtensions(s.worker, store, cache)

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

	if s.env.Metrics {
		log.Info("Enabling metrics endpoint")

		opts = append(opts, routes.WithMetrics())
	}

	if envs.IsDevelopment() {
		log.Info("Enabling OpenAPI validation in development mode")

		opts = append(opts, routes.WithOpenAPIValidator(&middleware.OpenAPIValidatorConfig{
			// NOTE: By default, metrics and internal endpoints are skipped from validation for now.
			Skipper: func(ctx echo.Context) bool {
				routes := []string{"/metrics", "/internal"}

				for _, path := range routes {
					if strings.HasPrefix(ctx.Request().URL.Path, path) {
						return true
					}
				}

				return false
			},
		}))
	}

	return opts, nil
}
