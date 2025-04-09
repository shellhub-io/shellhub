package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/geoip/geolite2"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	"github.com/shellhub-io/shellhub/pkg/worker"
	"github.com/shellhub-io/shellhub/pkg/worker/asynq"
	"github.com/shellhub-io/shellhub/server/api/routes"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/store/mongo"
	"github.com/shellhub-io/shellhub/server/api/store/mongo/options"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/tunnel"
	"github.com/shellhub-io/shellhub/server/ssh/server"
	"github.com/shellhub-io/shellhub/server/ssh/web"
	log "github.com/sirupsen/logrus"
)

type Env struct {
	// MongoURI specifies the connection string for MongoDB.
	MongoURI string `env:"API_MONGO_URI,default=mongodb://mongo:27017/main"`

	// RedisURI specifies the connection string for Redis.
	RedisURI string `env:"API_REDIS_URI,default=redis://redis:6379"`
	// RedisCachePoolSize defines the maximum number of concurrent connections to Redis cache.
	// Set to 0 for unlimited connections.
	RedisCachePoolSize int `env:"API_REDIS_CACHE_POOL_SIZE,default=0"`

	// SentryDSN specifies the Data Source Name for Sentry error tracking.
	// Leave empty to disable Sentry integration.
	SentryDSN string `env:"API_SENTRY_DSN,default="`

	// AsynqGroupMaxDelay specifies the maximum time (in seconds) to wait before
	// processing a group of tasks, regardless of other conditions.
	AsynqGroupMaxDelay int `env:"API_ASYNQ_GROUP_MAX_DELAY,default=1"`
	// AsynqGroupGracePeriod defines the grace period (in seconds) before task aggregation.
	// Tasks arriving within this period will be aggregated with existing tasks in the group.
	AsynqGroupGracePeriod int64 `env:"API_ASYNQ_GROUP_GRACE_PERIOD,default=2"`
	// AsynqGroupMaxSize specifies the maximum number of tasks that can be aggregated in a group.
	// When this limit is reached, the group will be processed immediately.
	AsynqGroupMaxSize int `env:"API_ASYNQ_GROUP_MAX_SIZE,default=1000"`
	// AsynqUniquenessTimeout defines how long (in hours) a unique job remains locked in the queue.
	// If a job doesn't complete within this period, its lock is released, allowing a new instance
	// to be enqueued and executed.
	AsynqUniquenessTimeout int `env:"API_ASYNQ_UNIQUENESS_TIMEOUT,default=24"`

	// GeoipMirror specifies an alternative URL for downloading GeoIP databases.
	// When configured, this takes precedence over GeoipMaxmindLicense.
	GeoipMirror string `env:"API_MAXMIND_MIRROR,default="`
	// GeoipMaxmindLicense is the MaxMind license key for downloading GeoIP databases directly.
	// This is used as a fallback when GeoipMirror is not configured.
	GeoipMaxmindLicense string `env:"API_MAXMIND_LICENSE,default="`

	ConnectTimeout time.Duration `env:"SSH_CONNECT_TIMEOUT,default=30s"`
	// Allows SSH to connect with an agent via a public key when the agent version is less than 0.6.0.
	// Agents 0.5.x or earlier do not validate the public key request and may panic.
	// Please refer to: https://github.com/shellhub-io/shellhub/issues/3453
	AllowPublickeyAccessBelow060 bool `env:"SSH_ALLOW_PUBLIC_KEY_ACCESS_BELLOW_0_6_0,default=false"`
}

type Server struct {
	env    *Env
	router *echo.Echo // TODO: evaluate if we can create a custom struct in router (e.g. router.Router)
	tun    *tunnel.Tunnel
	worker worker.Server
	cache  cache.Cache
}

// Setup initializes all server components including database connections, cache, services, API routes, and background workers.
// It prepares the server for starting but does not actually begin serving requests.
func (s *Server) Setup(ctx context.Context) error {
	log.Info("Setting up server components")

	cache, err := cache.NewRedisCache(s.env.RedisURI, s.env.RedisCachePoolSize)
	if err != nil {
		return err
	}

	s.cache = cache

	log.Debug("Redis cache initialized successfully")

	store, err := mongo.NewStore(ctx, s.env.MongoURI, cache, options.RunMigatrions)
	if err != nil {
		log.
			WithError(err).
			Fatal("failed to create the store")
	}

	log.Debug("MongoDB store connected successfully")

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

	tun, err := tunnel.NewTunnel(s.router, "/ssh/connection", "/ssh/revdial", s.env.RedisURI)
	if err != nil {
		log.WithError(err).
			Fatal("failed to create the internalclient")
	}

	web.NewSSHServerBridge(s.router, cache)

	s.tun = tun

	s.worker = asynq.NewServer(
		s.env.RedisURI,
		asynq.BatchConfig(s.env.AsynqGroupMaxSize, s.env.AsynqGroupMaxDelay, int(s.env.AsynqGroupGracePeriod)),
		asynq.UniquenessTimeout(s.env.AsynqUniquenessTimeout),
	)

	s.worker.HandleTask(services.TaskDevicesHeartbeat, service.DevicesHeartbeat(), asynq.BatchTask())

	log.Info("Server setup completed successfully")

	return nil
}

const ListenAddress = ":8080"

// Start begins serving API requests and processing background tasks. It blocks the current goroutine until the server stops
// or encounters an error.
func (s *Server) Start() error {
	log.Info("Starting server components")

	if err := s.worker.Start(); err != nil {
		return err
	}

	errs := make(chan error)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Debugf("listen for HTTP server on %s paniced", ListenAddress)

				errs <- fmt.Errorf("listen for HTTP on %s paniced", ListenAddress)
			}
		}()

		errs <- s.router.Start(ListenAddress) //nolint:gosec
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Debugf("listen for SSH server paniced")

				errs <- fmt.Errorf("listen for SSH server paniced")
			}
		}()

		errs <- server.NewServer(&server.Options{
			ConnectTimeout:               s.env.ConnectTimeout,
			AllowPublickeyAccessBelow060: s.env.AllowPublickeyAccessBelow060,
		}, s.tun.Tunnel, s.cache).ListenAndServe()
	}()

	return <-errs
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

func main() {
	loglevel.UseEnvs()

	ctx := context.Background()

	env, err := envs.Parse[Env]()
	if err != nil {
		log.WithError(err).
			Fatal("Failed to load environment variables")
	}

	server := &Server{env: env}

	if err := server.Setup(ctx); err != nil {
		log.WithError(err).
			Fatal("failed to setup the server")
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.WithField("signal", sig).
			Info("shutting down the server")

		server.Shutdown()
		os.Exit(0)
	}()

	if err := server.Start(); err != nil {
		log.WithError(err).
			Fatal("failed too start the server")
	}
}
