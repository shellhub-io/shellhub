package api

import (
	"context"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/geoip/geolite2"
	"github.com/shellhub-io/shellhub/pkg/worker"
	"github.com/shellhub-io/shellhub/pkg/worker/asynq"
	"github.com/shellhub-io/shellhub/server/api/routes"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/store/mongo"
	"github.com/shellhub-io/shellhub/server/api/store/mongo/options"
)

type Env struct {
	MongoURI               string `env:"MONGO_URI,default=mongodb://mongo:27017/main"`
	RedisURI               string `env:"REDIS_URI,default=redis://redis:6379"`
	RedisCachePoolSize     int    `env:"REDIS_CACHE_POOL_SIZE,default=0"`
	SentryDSN              string `env:"SENTRY_DSN,default="`
	AsynqGroupMaxDelay     int    `env:"ASYNQ_GROUP_MAX_DELAY,default=1"`
	AsynqGroupGracePeriod  int64  `env:"ASYNQ_GROUP_GRACE_PERIOD,default=2"`
	AsynqGroupMaxSize      int    `env:"ASYNQ_GROUP_MAX_SIZE,default=1000"`
	AsynqUniquenessTimeout int    `env:"ASYNQ_UNIQUENESS_TIMEOUT,default=24"`
	GeoipMirror            string `env:"MAXMIND_MIRROR,default="`
	GeoipMaxmindLicense    string `env:"MAXMIND_LICENSE,default="`
}

type Server struct {
	env    *Env
	router *echo.Echo
	worker worker.Server
}

func New() *Server {
	return &Server{}
}

func (s *Server) Setup(ctx context.Context) error {
	env, err := envs.ParseWithPrefix[Env]("API_")
	if err != nil {
		return err
	}

	s.env = env

	cache, err := cache.NewRedisCache(s.env.RedisURI, s.env.RedisCachePoolSize)
	if err != nil {
		return err
	}

	store, err := mongo.NewStore(ctx, s.env.MongoURI, cache, options.RunMigatrions)
	if err != nil {
		return err
	}

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

	s.worker = asynq.NewServer(
		s.env.RedisURI,
		asynq.BatchConfig(s.env.AsynqGroupMaxSize, s.env.AsynqGroupMaxDelay, int(s.env.AsynqGroupGracePeriod)),
		asynq.UniquenessTimeout(s.env.AsynqUniquenessTimeout),
	)

	s.worker.HandleTask(services.TaskDevicesHeartbeat, service.DevicesHeartbeat(), asynq.BatchTask())

	return nil
}

func (s *Server) Start() error {
	if err := s.worker.Start(); err != nil {
		return err
	}
	return s.router.Start(":8080")
}

func (s *Server) Shutdown() {
	s.worker.Shutdown()
	s.router.Close() // nolint: errcheck
}

func (s *Server) serviceOptions(ctx context.Context) ([]services.Option, error) {
	opts := []services.Option{}

	var geoipFetcher geolite2.GeoliteFetcher
	switch {
	case s.env.GeoipMirror != "":
		geoipFetcher = geolite2.FetchFromMirror(s.env.GeoipMirror)
	case s.env.GeoipMaxmindLicense != "":
		geoipFetcher = geolite2.FetchFromLicenseKey(s.env.GeoipMaxmindLicense)
	}

	if geoipFetcher != nil {
		locator, err := geolite2.NewLocator(ctx, geoipFetcher)
		if err != nil {
			return nil, err
		}
		opts = append(opts, services.WithLocator(locator))
	}

	return opts, nil
}

func (s *Server) routerOptions() ([]routes.Option, error) {
	opts := []routes.Option{}

	if s.env.SentryDSN != "" {
		sentryOpts := sentry.ClientOptions{
			Dsn:              s.env.SentryDSN,
			Release:          os.Getenv("SHELLHUB_VERSION"),
			EnableTracing:    true,
			TracesSampleRate: 1,
		}

		reporter, err := sentry.NewClient(sentryOpts)
		if err != nil {
			return nil, err
		}

		opts = append(opts, routes.WithReporter(reporter))
	}

	return opts, nil
}
