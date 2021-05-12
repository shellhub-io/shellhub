package main

import (
	"context"

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

func InitializeMongoDB(mongoURI string) (*mgo.Client, error) {
	// Connect to MongoDB

	clientOptions := options.Client().ApplyURI(mongoURI)

	client, err := mgo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	if err := mongo.ApplyMigrations(client.Database("main")); err != nil {
		return nil, err
	}

	return client, nil
}

func InitializeRedis(hasStoreCache bool) storecache.Cache {
	var cache storecache.Cache

	if hasStoreCache {
		cache = storecache.NewRedisCache()
	} else {
		cache = storecache.NewNullCache()
	}

	return cache
}

func InitializeRoutes(publicRoute, internalRoute *echo.Group) {
	internalRoute.GET(routes.AuthRequestURL, apicontext.Handler(routes.AuthRequest), apicontext.Middleware(routes.AuthMiddleware))
	publicRoute.POST(routes.AuthDeviceURL, apicontext.Handler(routes.AuthDevice))
	publicRoute.POST(routes.AuthDeviceURLV2, apicontext.Handler(routes.AuthDevice))
	publicRoute.POST(routes.AuthUserURL, apicontext.Handler(routes.AuthUser))
	publicRoute.POST(routes.AuthUserURLV2, apicontext.Handler(routes.AuthUser))
	publicRoute.GET(routes.AuthUserURLV2, apicontext.Handler(routes.AuthUserInfo))
	internalRoute.GET(routes.AuthUserTokenURL, apicontext.Handler(routes.AuthGetToken))
	publicRoute.POST(routes.AuthPublicKeyURL, apicontext.Handler(routes.AuthPublicKey))
	publicRoute.GET(routes.AuthUserTokenURL, apicontext.Handler(routes.AuthSwapToken))

	publicRoute.PUT(routes.UpdateUserURL, apicontext.Handler(routes.UpdateUser))
	publicRoute.PUT(routes.UpdateUserSecurityURL, apicontext.Handler(routes.UpdateUserSecurity))
	publicRoute.GET(routes.UserSecurityURL, apicontext.Handler(routes.GetUserSecurity))

	publicRoute.GET(routes.GetDeviceListURL,
		middlewares.Authorize(apicontext.Handler(routes.GetDeviceList)))
	publicRoute.GET(routes.GetDeviceURL,
		middlewares.Authorize(apicontext.Handler(routes.GetDevice)))
	publicRoute.DELETE(routes.DeleteDeviceURL, apicontext.Handler(routes.DeleteDevice))
	publicRoute.PATCH(routes.RenameDeviceURL, apicontext.Handler(routes.RenameDevice))
	internalRoute.POST(routes.OfflineDeviceURL, apicontext.Handler(routes.OfflineDevice))
	internalRoute.GET(routes.LookupDeviceURL, apicontext.Handler(routes.LookupDevice))
	publicRoute.PATCH(routes.UpdateStatusURL, apicontext.Handler(routes.UpdatePendingStatus))
	publicRoute.GET(routes.GetSessionsURL,
		middlewares.Authorize(apicontext.Handler(routes.GetSessionList)))
	publicRoute.GET(routes.GetSessionURL,
		middlewares.Authorize(apicontext.Handler(routes.GetSession)))
	internalRoute.PATCH(routes.SetSessionAuthenticatedURL, apicontext.Handler(routes.SetSessionAuthenticated))
	internalRoute.POST(routes.CreateSessionURL, apicontext.Handler(routes.CreateSession))
	internalRoute.POST(routes.FinishSessionURL, apicontext.Handler(routes.FinishSession))
	internalRoute.POST(routes.RecordSessionURL, apicontext.Handler(routes.RecordSession))
	publicRoute.GET(routes.PlaySessionURL, apicontext.Handler(routes.PlaySession))

	publicRoute.GET(routes.GetStatsURL,
		middlewares.Authorize(apicontext.Handler(routes.GetStats)))

	publicRoute.GET(routes.GetPublicKeysURL, apicontext.Handler(routes.GetPublicKeys))
	publicRoute.POST(routes.CreatePublicKeyURL, apicontext.Handler(routes.CreatePublicKey))
	publicRoute.PUT(routes.UpdatePublicKeyURL, apicontext.Handler(routes.UpdatePublicKey))
	publicRoute.DELETE(routes.DeletePublicKeyURL, apicontext.Handler(routes.DeletePublicKey))
	internalRoute.GET(routes.GetPublicKeyURL, apicontext.Handler(routes.GetPublicKey))
	internalRoute.POST(routes.CreatePrivateKeyURL, apicontext.Handler(routes.CreatePrivateKey))
	internalRoute.POST(routes.EvaluateKeyURL, apicontext.Handler(routes.EvaluateKeyHostname))

	publicRoute.GET(routes.ListNamespaceURL, apicontext.Handler(routes.GetNamespaceList))
	publicRoute.GET(routes.GetNamespaceURL, apicontext.Handler(routes.GetNamespace))
	publicRoute.POST(routes.CreateNamespaceURL, apicontext.Handler(routes.CreateNamespace))
	publicRoute.DELETE(routes.DeleteNamespaceURL, apicontext.Handler(routes.DeleteNamespace))
	publicRoute.PUT(routes.EditNamespaceURL, apicontext.Handler(routes.EditNamespace))
	publicRoute.PATCH(routes.AddNamespaceUserURL, apicontext.Handler(routes.AddNamespaceUser))
	publicRoute.PATCH(routes.RemoveNamespaceUserURL, apicontext.Handler(routes.RemoveNamespaceUser))
}

func InitializeAPI(c config) {
	e := echo.New()
	e.Use(middleware.Logger())

	mongoClient, err := InitializeMongoDB(c.MongoUri)

	if err != nil {
		panic(err.Error())
	}

	cache := InitializeRedis(c.StoreCache)

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			store := mongo.NewStore(mongoClient.Database("main"), cache)
			ctx := apicontext.NewContext(store, c)

			return next(ctx)
		}
	})

	publicRoute := e.Group("/api")
	internalRoute := e.Group("/internal")

	InitializeRoutes(publicRoute, internalRoute)

	e.Logger.Fatal(e.Start(":8080"))
}
