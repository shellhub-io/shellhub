package main

import (
	"context"
	"errors"
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/shellhub-io/shellhub/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/routes"
	apiMiddleware "github.com/shellhub-io/shellhub/api/routes/middleware"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/api/workers"
	requests "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

var serverCmd = &cobra.Command{
	Use: "server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		cfg, ok := ctx.Value("cfg").(*config)
		if !ok {
			log.Fatal("Failed to retrieve environment config from context")
		}

		go func() {
			log.Info("Starting workers")

			if err := workers.StartCleaner(ctx); err != nil {
				log.WithError(err).Fatal("Failed to start cleaner worker")
			}

			log.Info("Workers started")
		}()

		return startServer(cfg)
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
			Dsn:     dsn,
			Release: os.Getenv("SHELLHUB_VERSION"),
		})
		if err != nil {
			return nil, err
		}

		return reporter, nil
	}

	return nil, errors.New("sentry DSN not provided")
}

func startServer(cfg *config) error {
	ctx := context.Background()

	log.Info("Starting Sentry client")

	reporter, err := startSentry(cfg.SentryDSN)
	if err != nil {
		log.WithField("DSN", cfg.SentryDSN).WithError(err).Warn("Failed to start Sentry")
	} else {
		log.Info("Sentry client started")
	}

	log.Info("Starting API server")

	e := echo.New()
	e.Use(middleware.Log)
	e.Use(echoMiddleware.RequestID())
	e.Binder = handlers.NewBinder()
	e.Validator = handlers.NewValidator()
	e.HTTPErrorHandler = handlers.NewErrors(reporter)

	log.Trace("Connecting to Redis")

	cache, err := storecache.NewRedisCache(cfg.RedisURI)
	if err != nil {
		log.WithError(err).Error("Failed to configure redis store cache")
	}

	log.Info("Connected to Redis")

	log.Trace("Connecting to MongoDB")

	connStr, err := connstring.ParseAndValidate(cfg.MongoURI)
	if err != nil {
		log.WithError(err).Fatal("Invalid Mongo URI format")
	}

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongodriver.Connect(ctx, clientOptions)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to MongoDB")
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.WithError(err).Fatal("Failed to ping MongoDB")
	}

	log.Info("Connected to MongoDB")

	log.Info("Running database migrations")

	if err := mongo.ApplyMigrations(client.Database(connStr.Database)); err != nil {
		log.WithError(err).Fatal("Failed to apply mongo migrations")
	}

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

	store := mongo.NewStore(client.Database(connStr.Database), cache)
	service := services.NewService(store, nil, nil, cache, requestClient, locator)
	handler := routes.NewHandler(service)

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apicontext := gateway.NewContext(service, c)

			return next(apicontext)
		}
	})

	// Public routes for external access through API gateway
	publicAPI := e.Group("/api")

	// Internal routes only accessible by other services in the local container network
	internalAPI := e.Group("/internal")

	internalAPI.GET(routes.AuthRequestURL, gateway.Handler(handler.AuthRequest), gateway.Middleware(routes.AuthMiddleware))
	publicAPI.POST(routes.AuthDeviceURL, gateway.Handler(handler.AuthDevice))
	publicAPI.POST(routes.AuthDeviceURLV2, gateway.Handler(handler.AuthDevice))
	publicAPI.POST(routes.AuthUserURL, gateway.Handler(handler.AuthUser))
	publicAPI.POST(routes.AuthUserURLV2, gateway.Handler(handler.AuthUser))
	publicAPI.GET(routes.AuthUserURLV2, gateway.Handler(handler.AuthUserInfo))
	internalAPI.GET(routes.AuthUserTokenURL, gateway.Handler(handler.AuthGetToken))
	publicAPI.POST(routes.AuthPublicKeyURL, gateway.Handler(handler.AuthPublicKey))
	publicAPI.GET(routes.AuthUserTokenURL, gateway.Handler(handler.AuthSwapToken))

	publicAPI.PATCH(routes.UpdateUserDataURL, gateway.Handler(handler.UpdateUserData))
	publicAPI.PATCH(routes.UpdateUserPasswordURL, gateway.Handler(handler.UpdateUserPassword))
	publicAPI.PUT(routes.EditSessionRecordStatusURL, gateway.Handler(handler.EditSessionRecordStatus))
	publicAPI.GET(routes.GetSessionRecordURL, gateway.Handler(handler.GetSessionRecord))

	publicAPI.GET(routes.GetDeviceListURL,
		apiMiddleware.Authorize(gateway.Handler(handler.GetDeviceList)))
	publicAPI.GET(routes.GetDeviceURL,
		apiMiddleware.Authorize(gateway.Handler(handler.GetDevice)))
	publicAPI.DELETE(routes.DeleteDeviceURL, gateway.Handler(handler.DeleteDevice))
	publicAPI.PUT(routes.UpdateDevice, gateway.Handler(handler.UpdateDevice))
	publicAPI.PATCH(routes.RenameDeviceURL, gateway.Handler(handler.RenameDevice))
	internalAPI.POST(routes.OfflineDeviceURL, gateway.Handler(handler.OfflineDevice))
	internalAPI.POST(routes.HeartbeatDeviceURL, gateway.Handler(handler.HeartbeatDevice))
	internalAPI.GET(routes.LookupDeviceURL, gateway.Handler(handler.LookupDevice))
	publicAPI.PATCH(routes.UpdateStatusURL, gateway.Handler(handler.UpdatePendingStatus))

	publicAPI.POST(routes.CreateTagURL, gateway.Handler(handler.CreateDeviceTag))
	publicAPI.DELETE(routes.RemoveTagURL, gateway.Handler(handler.RemoveDeviceTag))
	publicAPI.PUT(routes.UpdateTagURL, gateway.Handler(handler.UpdateDeviceTag))

	publicAPI.GET(routes.GetTagsURL, gateway.Handler(handler.GetTags))
	publicAPI.PUT(routes.RenameTagURL, gateway.Handler(handler.RenameTag))
	publicAPI.DELETE(routes.DeleteTagsURL, gateway.Handler(handler.DeleteTag))

	publicAPI.GET(routes.GetSessionsURL,
		apiMiddleware.Authorize(gateway.Handler(handler.GetSessionList)))
	publicAPI.GET(routes.GetSessionURL,
		apiMiddleware.Authorize(gateway.Handler(handler.GetSession)))
	internalAPI.PATCH(routes.SetSessionAuthenticatedURL, gateway.Handler(handler.SetSessionAuthenticated))
	internalAPI.POST(routes.CreateSessionURL, gateway.Handler(handler.CreateSession))
	internalAPI.POST(routes.FinishSessionURL, gateway.Handler(handler.FinishSession))
	internalAPI.POST(routes.KeepAliveSessionURL, gateway.Handler(handler.KeepAliveSession))
	internalAPI.POST(routes.RecordSessionURL, gateway.Handler(handler.RecordSession))
	publicAPI.GET(routes.PlaySessionURL, gateway.Handler(handler.PlaySession))
	publicAPI.DELETE(routes.RecordSessionURL, gateway.Handler(handler.DeleteRecordedSession))

	publicAPI.GET(routes.GetStatsURL,
		apiMiddleware.Authorize(gateway.Handler(handler.GetStats)))
	publicAPI.GET(routes.GetSystemInfoURL, gateway.Handler(handler.GetSystemInfo))

	publicAPI.GET(routes.GetPublicKeysURL, gateway.Handler(handler.GetPublicKeys))
	publicAPI.POST(routes.CreatePublicKeyURL, gateway.Handler(handler.CreatePublicKey))
	publicAPI.PUT(routes.UpdatePublicKeyURL, gateway.Handler(handler.UpdatePublicKey))
	publicAPI.DELETE(routes.DeletePublicKeyURL, gateway.Handler(handler.DeletePublicKey))
	internalAPI.GET(routes.GetPublicKeyURL, gateway.Handler(handler.GetPublicKey))
	internalAPI.POST(routes.CreatePrivateKeyURL, gateway.Handler(handler.CreatePrivateKey))
	internalAPI.POST(routes.EvaluateKeyURL, gateway.Handler(handler.EvaluateKey))

	publicAPI.POST(routes.AddPublicKeyTagURL, gateway.Handler(handler.AddPublicKeyTag))
	publicAPI.DELETE(routes.RemovePublicKeyTagURL, gateway.Handler(handler.RemovePublicKeyTag))
	publicAPI.PUT(routes.UpdatePublicKeyTagsURL, gateway.Handler(handler.UpdatePublicKeyTags))

	publicAPI.GET(routes.ListNamespaceURL, gateway.Handler(handler.GetNamespaceList))
	publicAPI.GET(routes.GetNamespaceURL, gateway.Handler(handler.GetNamespace))
	publicAPI.POST(routes.CreateNamespaceURL, gateway.Handler(handler.CreateNamespace))
	publicAPI.DELETE(routes.DeleteNamespaceURL, gateway.Handler(handler.DeleteNamespace))
	publicAPI.PUT(routes.EditNamespaceURL, gateway.Handler(handler.EditNamespace))
	publicAPI.POST(routes.AddNamespaceUserURL, gateway.Handler(handler.AddNamespaceUser))
	publicAPI.DELETE(routes.RemoveNamespaceUserURL, gateway.Handler(handler.RemoveNamespaceUser))
	publicAPI.PATCH(routes.EditNamespaceUserURL, gateway.Handler(handler.EditNamespaceUser))
	publicAPI.GET(routes.HealthCheckURL, gateway.Handler(handler.EvaluateHealth))

	e.Logger.Fatal(e.Start(":8080"))

	return nil
}
