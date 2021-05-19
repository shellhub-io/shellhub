package main

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/routes"
	"github.com/shellhub-io/shellhub/api/routes/middlewares"
	storecache "github.com/shellhub-io/shellhub/api/store/cache"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	mgo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type config struct {
	MongoUri   string `envconfig:"mongo_uri" default:"mongodb://mongo:27017"`
	RedisUri   string `envconfig:"redis_uri" default:"redis://redis:6379"`
	StoreCache bool   `envconfig:"store_cache" default:"false"`
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	var cfg config
	if err := envconfig.Process("api", &cfg); err != nil {
		panic(err.Error())
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(cfg.MongoUri)
	// Connect to MongoDB
	client, err := mgo.Connect(context.TODO(), clientOptions)
	if err != nil {
		panic(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}

	if err := mongo.ApplyMigrations(client.Database("main")); err != nil {
		panic(err)
	}

	var cache storecache.Cache

	if cfg.StoreCache {
		cache, err = storecache.NewRedisCache(cfg.RedisUri)
		if err != nil {
			panic(err)
		}
	} else {
		cache = storecache.NewNullCache()
	}

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			store := mongo.NewStore(client.Database("main"), cache)
			ctx := apicontext.NewContext(store, c)

			return next(ctx)
		}
	})

	publicAPI := e.Group("/api")
	internalAPI := e.Group("/internal")

	internalAPI.GET(routes.AuthRequestURL, apicontext.Handler(routes.AuthRequest), apicontext.Middleware(routes.AuthMiddleware))
	publicAPI.POST(routes.AuthDeviceURL, apicontext.Handler(routes.AuthDevice))
	publicAPI.POST(routes.AuthDeviceURLV2, apicontext.Handler(routes.AuthDevice))
	publicAPI.POST(routes.AuthUserURL, apicontext.Handler(routes.AuthUser))
	publicAPI.POST(routes.AuthUserURLV2, apicontext.Handler(routes.AuthUser))
	publicAPI.GET(routes.AuthUserURLV2, apicontext.Handler(routes.AuthUserInfo))
	internalAPI.GET(routes.AuthUserTokenURL, apicontext.Handler(routes.AuthGetToken))
	publicAPI.POST(routes.AuthPublicKeyURL, apicontext.Handler(routes.AuthPublicKey))
	publicAPI.GET(routes.AuthUserTokenURL, apicontext.Handler(routes.AuthSwapToken))

	publicAPI.PATCH(routes.UpdateUserDataURL, apicontext.Handler(routes.UpdateUserData))
	publicAPI.PATCH(routes.UpdateUserPasswordURL, apicontext.Handler(routes.UpdateUserPassword))
	publicAPI.PUT(routes.UpdateUserSecurityURL, apicontext.Handler(routes.UpdateUserSecurity))
	publicAPI.GET(routes.UserSecurityURL, apicontext.Handler(routes.GetUserSecurity))

	publicAPI.GET(routes.GetDeviceListURL,
		middlewares.Authorize(apicontext.Handler(routes.GetDeviceList)))
	publicAPI.GET(routes.GetDeviceURL,
		middlewares.Authorize(apicontext.Handler(routes.GetDevice)))
	publicAPI.DELETE(routes.DeleteDeviceURL, apicontext.Handler(routes.DeleteDevice))
	publicAPI.PATCH(routes.RenameDeviceURL, apicontext.Handler(routes.RenameDevice))
	internalAPI.POST(routes.OfflineDeviceURL, apicontext.Handler(routes.OfflineDevice))
	internalAPI.GET(routes.LookupDeviceURL, apicontext.Handler(routes.LookupDevice))
	publicAPI.PATCH(routes.UpdateStatusURL, apicontext.Handler(routes.UpdatePendingStatus))
	publicAPI.GET(routes.GetSessionsURL,
		middlewares.Authorize(apicontext.Handler(routes.GetSessionList)))
	publicAPI.GET(routes.GetSessionURL,
		middlewares.Authorize(apicontext.Handler(routes.GetSession)))
	internalAPI.PATCH(routes.SetSessionAuthenticatedURL, apicontext.Handler(routes.SetSessionAuthenticated))
	internalAPI.POST(routes.CreateSessionURL, apicontext.Handler(routes.CreateSession))
	internalAPI.POST(routes.FinishSessionURL, apicontext.Handler(routes.FinishSession))
	internalAPI.POST(routes.RecordSessionURL, apicontext.Handler(routes.RecordSession))
	publicAPI.GET(routes.PlaySessionURL, apicontext.Handler(routes.PlaySession))

	publicAPI.GET(routes.GetStatsURL,
		middlewares.Authorize(apicontext.Handler(routes.GetStats)))

	publicAPI.GET(routes.GetPublicKeysURL, apicontext.Handler(routes.GetPublicKeys))
	publicAPI.POST(routes.CreatePublicKeyURL, apicontext.Handler(routes.CreatePublicKey))
	publicAPI.PUT(routes.UpdatePublicKeyURL, apicontext.Handler(routes.UpdatePublicKey))
	publicAPI.DELETE(routes.DeletePublicKeyURL, apicontext.Handler(routes.DeletePublicKey))
	internalAPI.GET(routes.GetPublicKeyURL, apicontext.Handler(routes.GetPublicKey))
	internalAPI.POST(routes.CreatePrivateKeyURL, apicontext.Handler(routes.CreatePrivateKey))
	internalAPI.POST(routes.EvaluateKeyURL, apicontext.Handler(routes.EvaluateKeyHostname))

	publicAPI.GET(routes.ListNamespaceURL, apicontext.Handler(routes.GetNamespaceList))
	publicAPI.GET(routes.GetNamespaceURL, apicontext.Handler(routes.GetNamespace))
	publicAPI.POST(routes.CreateNamespaceURL, apicontext.Handler(routes.CreateNamespace))
	publicAPI.DELETE(routes.DeleteNamespaceURL, apicontext.Handler(routes.DeleteNamespace))
	publicAPI.PUT(routes.EditNamespaceURL, apicontext.Handler(routes.EditNamespace))
	publicAPI.PATCH(routes.AddNamespaceUserURL, apicontext.Handler(routes.AddNamespaceUser))
	publicAPI.PATCH(routes.RemoveNamespaceUserURL, apicontext.Handler(routes.RemoveNamespaceUser))

	e.Logger.Fatal(e.Start(":8080"))
}
