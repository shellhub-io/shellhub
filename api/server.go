package main

import (
	"context"
	"os"

	"github.com/shellhub-io/shellhub/api/routers"

	"github.com/kelseyhightower/envconfig"
	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	requests "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var serverCmd = &cobra.Command{
	Use: "server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return startServer()
	},
}

// Provides the configuration for the API service.
// The values are load from the system environment variables.
type config struct {
	// MongoDB connection string (URI format)
	MongoURI string `envconfig:"mongo_uri" default:"mongodb://mongo:27017"`
	// Redis connection stirng (URI format)
	RedisURI string `envconfig:"redis_uri" default:"redis://redis:6379"`
	// Enable store cache
	StoreCache bool `envconfig:"store_cache" default:"false"`
	// Enable geoip feature
	GeoIP bool `envconfig:"geoip" default:"false"`
}

func startServer() error {
	loadCharacteristic := func(condition bool, successReturn func() interface{}, errorReturn interface{}) interface{} {
		if condition {
			return successReturn()
		} else {
			return errorReturn
		}
	}

	if os.Getenv("SHELLHUB_ENV") == "development" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Info("Starting API server")

	// Populates configuration based on environment variables prefixed with 'API_'
	var cfg config
	if err := envconfig.Process("api", &cfg); err != nil {
		logrus.WithError(err).Fatal("Failed to load environment variables")
	}

	logrus.Info("Connecting to MongoDB")

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongodriver.Connect(context.Background(), clientOptions)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to MongoDB")
	}

	if err = client.Ping(context.TODO(), nil); err != nil {
		logrus.WithError(err).Fatal("Failed to ping MongoDB")
	}

	logrus.Info("Running database migrations")

	if err := mongo.ApplyMigrations(client.Database("main")); err != nil {
		logrus.WithError(err).Fatal("Failed to apply mongo migrations")
	}

	requestClient := requests.NewClient()

	cache := loadCharacteristic(cfg.StoreCache, func() interface{} {
		cache, err := storecache.NewRedisCache(cfg.RedisURI)
		if err != nil {
			logrus.WithError(err).Fatal("Could not init cache")
		}

		return cache
	}, storecache.NewNullCache()).(storecache.Cache)

	// apply dependency injection through project layers
	store := mongo.NewStore(client.Database("main"), cache)

	locator := loadCharacteristic(cfg.GeoIP, func() interface{} {
		locator, err := geoip.NewGeoLite2()
		if err != nil {
			logrus.WithError(err).Fatal("Could not init geoip")
		}

		return locator
	}, geoip.NewNullGeoLite()).(geoip.Locator)

	service := services.NewService(store, nil, nil, cache, requestClient, locator)

	var r routers.Router
	r = routers.NewEchoRouter()
	r.LoadMiddleware(service)
	r.LoadRoutes(service)

	r.ListenAndServe(":8080")

	return nil
}
