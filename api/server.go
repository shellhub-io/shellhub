package main

import (
	"context"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	storecache "github.com/shellhub-io/shellhub/api/cache"
	apps "github.com/shellhub-io/shellhub/api/routes"
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
	client, err := mongodriver.Connect(context.TODO(), clientOptions)
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

	var cache storecache.Cache
	if cfg.StoreCache {
		logrus.Info("Using redis as store cache backend")

		cache, err = storecache.NewRedisCache(cfg.RedisURI)
		if err != nil {
			logrus.WithError(err).Error("Failed to configure redis store cache")
		}
	} else {
		logrus.Info("Store cache disabled")
		cache = storecache.NewNullCache()
	}

	requestClient := requests.NewClient()

	// apply dependency injection through project layers
	store := mongo.NewStore(client.Database("main"), cache)

	var locator geoip.Locator
	if cfg.GeoIP {
		logrus.Info("Using GeoIp for geolocation")
		locator, err = geoip.NewGeoLite2()
		if err != nil {
			logrus.WithError(err).Fatalln("Failed to init GeoIp")
		}
	} else {
		logrus.Info("GeoIp is disabled")
		locator = geoip.NewNullGeoLite()
	}

	service := services.NewService(store, nil, nil, cache, requestClient, locator)

	// instantiate echo app
	a := apps.NewEchoApp(echo.New(), service)

	a.InitRoutes()
	a.ListenAndServe(8080)

	return nil
}
