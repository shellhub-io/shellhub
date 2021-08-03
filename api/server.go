package main

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/shellhub-io/shellhub/api/apicontext"
	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/routes"
	"github.com/shellhub-io/shellhub/api/routes/middlewares"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store/mongo"
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
}

func startServer() error {
	logrus.Info("Starting API server")

	e := echo.New()
	e.Use(middleware.Logger())

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

	// apply dependency injection through project layers
	store := mongo.NewStore(client.Database("main"), cache)
	service := services.NewService(store, nil, nil, cache)
	handler := routes.NewHandler(service)

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apicontext := apicontext.NewContext(service, c)

			return next(apicontext)
		}
	})

	// Public routes for external access through API gateway
	publicAPI := e.Group("/api")

	// Internal routes only accessible by other services in the local container network
	internalAPI := e.Group("/internal")

	internalAPI.GET(routes.AuthRequestURL, apicontext.Handler(handler.AuthRequest), apicontext.Middleware(routes.AuthMiddleware))
	publicAPI.POST(routes.AuthDeviceURL, apicontext.Handler(handler.AuthDevice))
	publicAPI.POST(routes.AuthDeviceURLV2, apicontext.Handler(handler.AuthDevice))
	publicAPI.POST(routes.AuthUserURL, apicontext.Handler(handler.AuthUser))
	publicAPI.POST(routes.AuthUserURLV2, apicontext.Handler(handler.AuthUser))
	publicAPI.GET(routes.AuthUserURLV2, apicontext.Handler(handler.AuthUserInfo))
	internalAPI.GET(routes.AuthUserTokenURL, apicontext.Handler(handler.AuthGetToken))
	publicAPI.POST(routes.AuthPublicKeyURL, apicontext.Handler(handler.AuthPublicKey))
	publicAPI.GET(routes.AuthUserTokenURL, apicontext.Handler(handler.AuthSwapToken))
	publicAPI.PATCH(routes.UpdateUserDataURL, apicontext.Handler(handler.UpdateUserData))
	publicAPI.PATCH(routes.UpdateUserPasswordURL, apicontext.Handler(handler.UpdateUserPassword))
	publicAPI.PUT(routes.EditSessionRecordStatusURL, apicontext.Handler(handler.EditSessionRecordStatus), apicontext.Middleware(routes.IsSessionOwner))
	publicAPI.GET(routes.GetSessionRecordURL, apicontext.Handler(handler.GetSessionRecord), apicontext.Middleware(routes.IsSessionOwner))
	publicAPI.GET(routes.GetDeviceListURL,
		middlewares.Authorize(apicontext.Handler(handler.GetDeviceList)), apicontext.Middleware(routes.IsDeviceMember))
	publicAPI.GET(routes.GetDeviceURL,
		middlewares.Authorize(apicontext.Handler(handler.GetDevice)))
	publicAPI.DELETE(routes.DeleteDeviceURL, apicontext.Handler(handler.DeleteDevice), apicontext.Middleware(routes.IsDeviceOwner))
	publicAPI.PATCH(routes.RenameDeviceURL, apicontext.Handler(handler.RenameDevice), apicontext.Middleware(routes.IsDeviceOwner))
	internalAPI.POST(routes.OfflineDeviceURL, apicontext.Handler(handler.OfflineDevice))
	internalAPI.GET(routes.LookupDeviceURL, apicontext.Handler(handler.LookupDevice))
	publicAPI.PATCH(routes.UpdateStatusURL, apicontext.Handler(handler.UpdatePendingStatus), apicontext.Middleware(routes.IsDeviceOwner))
	publicAPI.GET(routes.GetSessionsURL,
		middlewares.Authorize(apicontext.Handler(handler.GetSessionList)))
	publicAPI.GET(routes.GetSessionURL,
		middlewares.Authorize(apicontext.Handler(handler.GetSession)), apicontext.Middleware(routes.IsSessionMember))
	internalAPI.PATCH(routes.SetSessionAuthenticatedURL, apicontext.Handler(handler.SetSessionAuthenticated))
	internalAPI.POST(routes.CreateSessionURL, apicontext.Handler(handler.CreateSession))
	internalAPI.POST(routes.FinishSessionURL, apicontext.Handler(handler.FinishSession))
	internalAPI.POST(routes.RecordSessionURL, apicontext.Handler(handler.RecordSession))
	publicAPI.GET(routes.PlaySessionURL, apicontext.Handler(handler.PlaySession))
	publicAPI.DELETE(routes.RecordSessionURL, apicontext.Handler(handler.DeleteRecordedSession))

	publicAPI.GET(routes.GetStatsURL,
		middlewares.Authorize(apicontext.Handler(handler.GetStats)))

	publicAPI.GET(routes.GetPublicKeysURL, apicontext.Handler(handler.GetPublicKeys), apicontext.Middleware(routes.IsKeysMember))
	publicAPI.POST(routes.CreatePublicKeyURL, apicontext.Handler(handler.CreatePublicKey), apicontext.Middleware(routes.IsKeysOwner))
	publicAPI.PUT(routes.UpdatePublicKeyURL, apicontext.Handler(handler.UpdatePublicKey), apicontext.Middleware(routes.IsKeysOwner))
	publicAPI.DELETE(routes.DeletePublicKeyURL, apicontext.Handler(handler.DeletePublicKey), apicontext.Middleware(routes.IsKeysOwner))
	internalAPI.GET(routes.GetPublicKeyURL, apicontext.Handler(handler.GetPublicKey))
	internalAPI.POST(routes.CreatePrivateKeyURL, apicontext.Handler(handler.CreatePrivateKey))
	internalAPI.POST(routes.EvaluateKeyURL, apicontext.Handler(handler.EvaluateKeyHostname))

	publicAPI.GET(routes.ListNamespaceURL, apicontext.Handler(handler.GetNamespaceList))
	publicAPI.GET(routes.GetNamespaceURL, apicontext.Handler(handler.GetNamespace))
	publicAPI.POST(routes.CreateNamespaceURL, apicontext.Handler(handler.CreateNamespace))
	publicAPI.DELETE(routes.DeleteNamespaceURL, apicontext.Handler(handler.DeleteNamespace), apicontext.Middleware(routes.IsNamespaceOwner))
	publicAPI.PUT(routes.EditNamespaceURL, apicontext.Handler(handler.EditNamespace), apicontext.Middleware(routes.IsNamespaceOwner))
	publicAPI.PATCH(routes.AddNamespaceUserURL, apicontext.Handler(handler.AddNamespaceUser), apicontext.Middleware(routes.IsNamespaceOwner))
	publicAPI.PATCH(routes.RemoveNamespaceUserURL, apicontext.Handler(handler.RemoveNamespaceUser), apicontext.Middleware(routes.IsNamespaceOwner))

	e.Logger.Fatal(e.Start(":8080"))

	return nil
}
