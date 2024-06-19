package main

import (
	"context"
	"errors"

	"github.com/shellhub-io/shellhub/api/routes"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/api/store/mongo/options"
	"github.com/shellhub-io/shellhub/api/workers"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	log "github.com/sirupsen/logrus"
)

var (
	ErrConnectRedis   = errors.New("failed to connect Redis store cache")
	ErrConnectMongoDB = errors.New("failed to connecto to MongoDB")
	ErrStoreCreation  = errors.New("failed to create the store drom MongoDB and Redis")
	ErrLoadKeys       = errors.New("failed to load the private and the public key")
	ErrStartWorkers   = errors.New("failed to start the workers")
)

func Server(ctx context.Context, config *Config) error {
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
		return errors.Join(ErrStoreCreation, err)
	}

	log.Info("Connected to MongoDB")

	worker, err := workers.New(store)
	if err != nil {
		return errors.Join(ErrStartWorkers, err)
	}

	worker.Start(ctx)

	keys, err := services.LoadKeys(config.PrivateKey)
	if err != nil {
		return errors.Join(ErrLoadKeys, err)
	}

	service := services.NewService(keys, store, cache)

	if config.GeoIP {
		log.Info("GeoIP feature is enable")
		locator, err := geoip.NewGeoLite2()
		if err != nil {
			log.WithError(err).Fatal("Failed to init GeoIP")
		}

		service.WithLocator(locator)
	}

	log.Info("Starting API server")

	routes := routes.NewRouter(service, &routes.Config{
		SentryDNS: config.SentryDSN,
	})

	go func() {
		<-ctx.Done()

		log.Debug("Closing HTTP server due context cancellation")

		routes.Close()
	}()

	err = routes.Start(":8080") //nolint:errcheck

	log.WithError(err).Info("HTTP server closed")

	return nil
}
