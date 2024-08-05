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
	// MongoURI is the mongodb connection uri.
	MongoURI string `env:"MONGO_URI,default=mongodb://mongo:27017/main"`
	// RedisURI is the redis connection uri.
	RedisURI string `env:"REDIS_URI,default=redis://redis:6379"`
	// RedisCachePoolSize is the pool size of connections available for Redis cache.
	RedisCachePoolSize int `env:"REDIS_CACHE_POOL_SIZE,default=0"`
	// GeoIP defines whether the GeoIP feature is enabled or not.
	//
	// GeoIP features enable the ability to get the logitude and latitude of the client from the IP address.
	// The feature is disabled by default. To enable it, it is required to have a `MAXMIND` database license and feed it
	// to `SHELLHUB_MAXMIND_LICENSE` with it, and `SHELLHUB_GEOIP=true`.
	GeoIP               bool   `env:"GEOIP,default=false"`
	GeoIPMaxMindLicense string `env:"MAXMIND_LICENSE,default="`
	SentryDSN           string `env:"SENTRY_DSN,default="`
	// AsynqGroupMaxDelay is the maximum duration to wait before processing a group of tasks.
	//
	// Its time unit is second.
	//
	// Check [https://github.com/hibiken/asynq/wiki/Task-aggregation] for more information.
	AsynqGroupMaxDelay int `env:"ASYNQ_GROUP_MAX_DELAY,default=1"`
	// AsynqGroupGracePeriod is the grace period has configurable upper bound: you can set a maximum aggregation delay, after which Asynq server
	// will aggregate the tasks regardless of the remaining grace period.
	///
	// Its time unit is second.
	//
	// Check [https://github.com/hibiken/asynq/wiki/Task-aggregation] for more information.
	AsynqGroupGracePeriod int64 `env:"ASYNQ_GROUP_GRACE_PERIOD,default=1"`
	// AsynqGroupMaxSize is the maximum number of tasks that can be aggregated together. If that number is reached, Asynq
	// server will aggregate the tasks immediately.
	//
	// Check [https://github.com/hibiken/asynq/wiki/Task-aggregation] for more information.
	AsynqGroupMaxSize             int    `env:"ASYNQ_GROUP_MAX_SIZE,default=500"`
	SessionRecordCleanupSchedule  string `env:"SESSION_RECORD_CLEANUP_SCHEDULE,default=@daily"`
	SessionRecordCleanupRetention int    `env:"RECORD_RETENTION,default=0"`
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

	apiClient, err := internalclient.NewClient()
	if err != nil {
		log.WithError(err).
			Error("failed to create the api client")

		return err
	}

	servicesOptions := []services.Option{}
	if s.env.GeoIP {
		locator, err := geoip.NewGeoLite2(s.env.GeoIPMaxMindLicense)
		if err != nil {
			log.WithError(err).
				Error("failed to init GeoIP")

			return err
		}

		servicesOptions = append(servicesOptions, services.WithLocator(locator))
	}

	service := services.NewService(store, nil, nil, cache, apiClient, servicesOptions...)

	s.worker = asynq.NewServer(s.env.RedisURI)
	s.worker.HandleTask(services.TaskDevicesHeartbeat, service.DevicesHeartbeat(), asynq.BatchTask())
	if s.env.SessionRecordCleanupRetention > 0 {
		s.worker.HandleCron(worker.CronSpec(s.env.SessionRecordCleanupSchedule), service.CleanupSessions(s.env.SessionRecordCleanupRetention))
	}

	routerOptions := []routes.Option{}
	if s.env.SentryDSN != "" {
		opts := sentry.ClientOptions{ //nolint:exhaustruct
			Dsn:              s.env.SentryDSN,
			Release:          os.Getenv("SHELLHUB_VERSION"),
			EnableTracing:    true,
			TracesSampleRate: 1,
		}

		reporter, err := sentry.NewClient(opts)
		if err != nil {
			log.WithError(err).
				Error("Failed to create Sentry client")

			return err
		}

		routerOptions = append(routerOptions, routes.WithReporter(reporter))
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
