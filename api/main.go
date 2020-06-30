package main

import (
	"context"
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/routes"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	mgo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type config struct {
	MongoHost string `envconfig:"mongo_host" default:"mongo"`
	MongoPort int    `envconfig:"mongo_port" default:"27017"`
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	var cfg config
	if err := envconfig.Process("api", &cfg); err != nil {
		panic(err.Error())
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%d", cfg.MongoHost, cfg.MongoPort))
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

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			store := mongo.NewStore(client.Database("main"))
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

	publicAPI.GET(routes.GetDeviceListURL, apicontext.Handler(routes.GetDeviceList))
	publicAPI.GET(routes.GetDeviceURL, apicontext.Handler(routes.GetDevice))
	publicAPI.DELETE(routes.DeleteDeviceURL, apicontext.Handler(routes.DeleteDevice))
	publicAPI.PATCH(routes.RenameDeviceURL, apicontext.Handler(routes.RenameDevice))
	internalAPI.POST(routes.OfflineDeviceURL, apicontext.Handler(routes.OfflineDevice))
	internalAPI.GET(routes.LookupDeviceURL, apicontext.Handler(routes.LookupDevice))
	publicAPI.PATCH(routes.UpdateStatusURL, apicontext.Handler(routes.UpdatePendingStatus))

	publicAPI.GET(routes.GetSessionsURL, apicontext.Handler(routes.GetSessionList))
	publicAPI.GET(routes.GetSessionURL, apicontext.Handler(routes.GetSession))
	internalAPI.PATCH(routes.SetSessionAuthenticatedURL, apicontext.Handler(routes.SetSessionAuthenticated))
	internalAPI.POST(routes.CreateSessionURL, apicontext.Handler(routes.CreateSession))
	internalAPI.POST(routes.FinishSessionURL, apicontext.Handler(routes.FinishSession))

	publicAPI.GET(routes.GetFirewallRuleListURL, apicontext.Handler(routes.GetFirewallRuleList))
	publicAPI.GET(routes.GetFirewallRuleURL, apicontext.Handler(routes.GetFirewallRule))
	publicAPI.POST(routes.CreateFirewallRuleURL, apicontext.Handler(routes.CreateFirewallRule))
	publicAPI.PUT(routes.UpdateFirewallRuleURL, apicontext.Handler(routes.UpdateFirewallRule))
	publicAPI.DELETE(routes.DeleteFirewallRuleURL, apicontext.Handler(routes.DeleteFirewallRule))

	publicAPI.GET(routes.GetStatsURL, apicontext.Handler(routes.GetStats))

	e.Logger.Fatal(e.Start(":8080"))
}
