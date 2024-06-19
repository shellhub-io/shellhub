package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/getsentry/sentry-go"
	"github.com/shellhub-io/shellhub/api/routes"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/api/store/mongo/options"
	"github.com/shellhub-io/shellhub/api/workers"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	log "github.com/sirupsen/logrus"
)

// startSentry initializes the Sentry client.
//
// The Sentry client is used to report errors to the Sentry server, and is initialized only if the `SHELLHUB_SENTRY_DSN`
// environment variable is set. Else, the function returns a error with a not initialized Sentry client.
func startSentry(dsn string) (*sentry.Client, error) {
	if dsn != "" {
		var err error
		reporter, err := sentry.NewClient(sentry.ClientOptions{ //nolint:exhaustruct
			Dsn:              dsn,
			Release:          os.Getenv("SHELLHUB_VERSION"),
			EnableTracing:    true,
			TracesSampleRate: 1,
		})
		if err != nil {
			log.WithError(err).Error("Failed to create Sentry client")

			return nil, err
		}
		log.Info("Sentry client started")

		return reporter, nil
	}

	return nil, errors.New("sentry DSN not provided")
}

var (
	ErrConnectRedis        = errors.New("failed to connect Redis store cache")
	ErrConnectMongoDB      = errors.New("failed to connecto to MongoDB")
	ErrFailedStoreCreation = errors.New("failed to create the store drom MongoDB and Redis")
	ErrStartWorkers        = errors.New("failed to start the workers")
)

const (
	FeatureGeoIP uint = iota + 1
	FeatureSentry
)

func Server(ctx context.Context, config Config) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Trace("Connecting to Redis")

	cache, err := storecache.NewRedisCache(config.RedisURI, config.RedisCachePoolSize)
	if err != nil {
		return errors.Join(ErrConnectRedis, err)
	}

	log.Info("Connected to Redis")

	log.Trace("Connecting to MongoDB")

	_, db, err := mongo.Connect(ctx, config.MongoURI)
	if err != nil {
		return errors.Join(ErrConnectMongoDB, err)
	}

	store, err := mongo.NewStore(ctx, db, cache, options.RunMigatrions)
	if err != nil {
		return errors.Join(ErrFailedStoreCreation, err)
	}

	log.Info("Connected to MongoDB")

	worker, err := workers.New(store)
	if err != nil {
		return errors.Join(ErrStartWorkers, err)
	}

	worker.Start(ctx)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs

		log.WithFields(log.Fields{
			"signal": sig,
		}).Info("signal received to terminate API")

		cancel()
	}()

	service := services.NewService(services.Keys{}, store, cache)

	/*log.Info("Starting Sentry client")

	reporter, err := startSentry(cfg.SentryDSN)
	if err != nil {
		log.WithField("DSN", cfg.SentryDSN).WithError(err).Warn("Failed to start Sentry")
	} else {
		log.Info("Sentry client started")
	}*/

	if config.GeoIP {
		log.Info("GeoIP feature is enable")
		locator, err := geoip.NewGeoLite2()
		if err != nil {
			log.WithError(err).Fatal("Failed to init GeoIP")
		}

		service.WithLocator(locator)
	}

	log.Info("Starting API server")

	routes := routes.NewRouter(service)

	go func() {
		<-ctx.Done()

		log.Debug("Closing HTTP server due context cancellation")

		routes.Close()
	}()

	err = routes.Start(":8080") //nolint:errcheck

	log.WithError(err).Info("HTTP server closed")

	return nil
}
