package main

import (
	"context"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/routes"
	apiMiddleware "github.com/shellhub-io/shellhub/api/routes/middleware"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	requests "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/middleware"
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
	// Defines environment mongo
	MongoDBName string `envconfig:"mongo_db_name" default:"main"`
}

func startServer() error {
	if os.Getenv("SHELLHUB_ENV") == "development" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Info("Starting API server")

	e := echo.New()
	e.Use(middleware.Log)
	e.Use(echoMiddleware.RequestID())

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

	if err := mongo.ApplyMigrations(client.Database(cfg.MongoDBName)); err != nil {
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
	store := mongo.NewStore(client.Database(cfg.MongoDBName), cache)

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
	publicAPI.PATCH(routes.RenameDeviceURL, gateway.Handler(handler.RenameDevice))
	internalAPI.POST(routes.OfflineDeviceURL, gateway.Handler(handler.OfflineDevice))
	internalAPI.POST(routes.HeartbeatDeviceURL, gateway.Handler(handler.HeartbeatDevice))
	internalAPI.GET(routes.LookupDeviceURL, gateway.Handler(handler.LookupDevice))
	publicAPI.PATCH(routes.UpdateStatusURL, gateway.Handler(handler.UpdatePendingStatus))

	publicAPI.POST(routes.CreateTagURL, gateway.Handler(handler.CreateTag))
	publicAPI.DELETE(routes.RemoveTagURL, gateway.Handler(handler.RemoveTag))
	publicAPI.PUT(routes.RenameTagURL, gateway.Handler(handler.RenameTag))
	publicAPI.PUT(routes.UpdateTagURL, gateway.Handler(handler.UpdateTag))
	publicAPI.GET(routes.GetTagsURL, gateway.Handler(handler.GetTags))
	publicAPI.DELETE(routes.DeleteTagsURL, gateway.Handler(handler.DeleteTags))

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

	publicAPI.GET(routes.GetPublicKeysURL, gateway.Handler(handler.GetPublicKeys))
	publicAPI.POST(routes.CreatePublicKeyURL, gateway.Handler(handler.CreatePublicKey))
	publicAPI.PUT(routes.UpdatePublicKeyURL, gateway.Handler(handler.UpdatePublicKey))
	publicAPI.DELETE(routes.DeletePublicKeyURL, gateway.Handler(handler.DeletePublicKey))
	internalAPI.GET(routes.GetPublicKeyURL, gateway.Handler(handler.GetPublicKey))
	internalAPI.POST(routes.CreatePrivateKeyURL, gateway.Handler(handler.CreatePrivateKey))
	internalAPI.POST(routes.EvaluateKeyURL, gateway.Handler(handler.EvaluateKey))

	publicAPI.GET(routes.ListNamespaceURL, gateway.Handler(handler.GetNamespaceList))
	publicAPI.GET(routes.GetNamespaceURL, gateway.Handler(handler.GetNamespace))
	publicAPI.POST(routes.CreateNamespaceURL, gateway.Handler(handler.CreateNamespace))
	publicAPI.DELETE(routes.DeleteNamespaceURL, gateway.Handler(handler.DeleteNamespace))
	publicAPI.PUT(routes.EditNamespaceURL, gateway.Handler(handler.EditNamespace))
	publicAPI.POST(routes.AddNamespaceUserURL, gateway.Handler(handler.AddNamespaceUser))
	publicAPI.DELETE(routes.RemoveNamespaceUserURL, gateway.Handler(handler.RemoveNamespaceUser))
	publicAPI.PATCH(routes.EditNamespaceUserURL, gateway.Handler(handler.EditNamespaceUser))

	e.Logger.Fatal(e.Start(":8080"))

	return nil
}
