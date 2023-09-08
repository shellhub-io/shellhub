package main

import (
	"errors"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/shellhub-io/shellhub/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/routes"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/api/workers"
	requests "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use: "server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		cfg, ok := ctx.Value("cfg").(*config)
		if !ok {
			log.Fatal("Failed to retrieve environment config from context")
		}

		log.Trace("Connecting to Redis")

		cache, err := storecache.NewRedisCache(cfg.RedisURI)
		if err != nil {
			log.WithError(err).Error("Failed to configure redis store cache")
		}

		log.Info("Connected to Redis")

		log.Trace("Connecting to MongoDB")

		store, err := mongo.NewStoreMongo(ctx, cache, cfg.MongoURI)
		if err != nil {
			log.WithError(err).Fatal("failed to create the store")
		}

		log.Info("Connected to MongoDB")

		go func() {
			cleaner := workers.NewCleaner(store)
			heartbeater := workers.NewHeartbeater(store)

			msgsCleaner := make(chan workers.WorkerMessage)
			go cleaner.Start(ctx, msgsCleaner)

			msgsHeartbeater := make(chan workers.WorkerMessage)
			go heartbeater.Start(ctx, msgsHeartbeater)

			for {
				select {
				case msg := <-msgsCleaner:
					log.WithError(msg.Error).WithFields(
						log.Fields{
							"worker": "cleaner",
						},
					).Info(msg.Message)
				case msg := <-msgsHeartbeater:
					log.WithError(msg.Error).WithFields(log.Fields{
						"worker": "heartbeater",
					}).Info(msg.Message)
				}
			}
		}()

		return startServer(cfg, store, cache)
	},
}

// Provides the configuration for the API service.
// The values are load from the system environment variables.
type config struct {
	// MongoDB connection string (URI format)
	MongoURI string `envconfig:"mongo_uri" default:"mongodb://mongo:27017/main"`
	// Redis connection string (URI format)
	RedisURI string `envconfig:"redis_uri" default:"redis://redis:6379"`
	// Enable GeoIP feature.
	//
	// GeoIP features enable the ability to get the logitude and latitude of the client from the IP address.
	// The feature is disabled by default. To enable it, it is required to have a `MAXMIND` database license and feed it
	// to `SHELLHUB_MAXMIND_LICENSE` with it, and `SHELLHUB_GEOIP=true`.
	GeoIP bool `envconfig:"geoip" default:"false"`
	// Session record cleanup worker schedule
	SessionRecordCleanupSchedule string `envconfig:"session_record_cleanup_schedule" default:"@daily"`
	// Sentry DSN.
	SentryDSN string `envconfig:"sentry_dsn" default:""`
}

func init() {
	if value, ok := os.LookupEnv("SHELLHUB_ENV"); ok && value == "development" {
		log.SetLevel(log.TraceLevel)
		log.Debug("Log level set to Trace")
	} else {
		log.Debug("Log level default")
	}
}

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
			return nil, err
		}

		return reporter, nil
	}

	return nil, errors.New("sentry DSN not provided")
}

func startServer(cfg *config, store store.Store, cache storecache.Cache) error {
	log.Info("Starting Sentry client")

	reporter, err := startSentry(cfg.SentryDSN)
	if err != nil {
		log.WithField("DSN", cfg.SentryDSN).WithError(err).Warn("Failed to start Sentry")
	} else {
		log.Info("Sentry client started")
	}

	log.Info("Starting API server")

	requestClient := requests.NewClient()

	var locator geoip.Locator
	if cfg.GeoIP {
		log.Info("GeoIP feature is enable")
		locator, err = geoip.NewGeoLite2()
		if err != nil {
			log.WithError(err).Fatal("Failed to init GeoIP")
		}
	} else {
		log.Info("GeoIP is disabled")
		locator = geoip.NewNullGeoLite()
	}

	service := services.NewService(store, nil, nil, cache, requestClient, locator)

	e := routes.NewRouter(service)
	e.Use(middleware.Log)
	e.Use(echoMiddleware.RequestID())
	e.HTTPErrorHandler = handlers.NewErrors(reporter)

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apicontext := gateway.NewContext(service, c)

			return next(apicontext)
		}
	})

	e.Logger.Fatal(e.Start(":8080"))

	return nil
}
