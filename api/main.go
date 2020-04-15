package main

import (
	"context"
	"crypto/rsa"
	"fmt"

	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/mitchellh/mapstructure"
	"github.com/shellhub-io/shellhub/api/pkg/services/authsvc"
	"github.com/shellhub-io/shellhub/api/pkg/services/deviceadm"
	"github.com/shellhub-io/shellhub/api/pkg/services/sessionmngr"
	"github.com/shellhub-io/shellhub/api/pkg/services/ssh2ws"
	"github.com/shellhub-io/shellhub/api/pkg/store/mongo"
	"github.com/shellhub-io/shellhub/pkg/models"
	mgo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/websocket"
)

type config struct {
	MongoHost string `envconfig:"mongo_host" default:"mongo"`
	MongoPort int    `envconfig:"mongo_port" default:"27017"`
}

var verifyKey *rsa.PublicKey

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

	signBytes, err := ioutil.ReadFile(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		panic(err)
	}

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		panic(err)
	}

	verifyBytes, err := ioutil.ReadFile(os.Getenv("PUBLIC_KEY"))
	if err != nil {
		panic(err)
	}

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		panic(err)
	}

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tenant := c.Request().Header.Get("X-Tenant-ID")

			ctx := context.WithValue(c.Request().Context(), "db", client.Database("main"))

			if tenant != "" {
				ctx = context.WithValue(ctx, "tenant", tenant)
			}

			c.Set("ctx", ctx)
			c.Set("db", client.Database("main"))

			return next(c)
		}
	})

	publicAPI := e.Group("/api")
	internalAPI := e.Group("/internal")
	internalAPI.POST("/devices/:uid/offline", func(c echo.Context) error {
		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := deviceadm.NewService(store)

		err := svc.UpdateDeviceStatus(ctx, models.UID(c.Param("uid")), false)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, nil)
	})

	publicAPI.POST("/devices/auth", func(c echo.Context) error {
		var req models.DeviceAuthRequest

		err := c.Bind(&req)
		if err != nil {
			return err
		}

		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := authsvc.NewService(store, signKey)

		res, err := svc.AuthDevice(ctx, &req)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, res)
	})

	publicAPI.GET("/devices", func(c echo.Context) error {
		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := deviceadm.NewService(store)

		var query struct {
			Page    int `query:"page"`
			PerPage int `query:"per_page"`
		}
		c.Bind(&query)

		page := query.Page
		perPage := query.PerPage
		if perPage < 1 || perPage > 100 {
			perPage = 10
		}
		if page < 1 {
			page = 1
		}

		devices, err := svc.ListDevices(ctx, perPage, page)
		if err != nil {
			return err
		}
		count, err := svc.CountDevices(ctx)
		if err != nil {
			return err
		}
		c.Response().Header().Set("X-Total-Count", strconv.FormatInt(count, 10))
		return c.JSON(http.StatusOK, devices)
	})

	publicAPI.GET("/devices/:uid", func(c echo.Context) error {
		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := deviceadm.NewService(store)

		device, err := svc.GetDevice(ctx, models.UID(c.Param("uid")))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, device)
	})

	publicAPI.DELETE("/devices/:uid", func(c echo.Context) error {
		tenant := c.Request().Header.Get("X-Tenant-ID")
		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := deviceadm.NewService(store)

		if err := svc.DeleteDevice(ctx, models.UID(c.Param("uid")), tenant); err != nil {
			if err == deviceadm.UnauthorizedErr {
				return c.NoContent(http.StatusForbidden)
			} else {
				return err
			}
		}
		return nil
	})

	publicAPI.PATCH("/devices/:uid", func(c echo.Context) error {
		tenant := c.Request().Header.Get("X-Tenant-ID")
		var req struct {
			Name string `json:"name"`
		}

		err := c.Bind(&req)
		if err != nil {
			return err
		}

		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := deviceadm.NewService(store)

		if err := svc.RenameDevice(ctx, models.UID(c.Param("uid")), req.Name, tenant); err != nil {
			if err == deviceadm.UnauthorizedErr {
				return c.NoContent(http.StatusForbidden)
			} else {
				return err
			}
		}
		return nil
	})

	publicAPI.POST("/login", func(c echo.Context) error {
		var req models.UserAuthRequest

		err := c.Bind(&req)
		if err != nil {
			return err
		}

		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := authsvc.NewService(store, signKey)

		res, err := svc.AuthUser(ctx, req)
		if err != nil {
			return echo.ErrUnauthorized
		}

		return c.JSON(http.StatusOK, res)
	})

	internalAPI.GET("/auth", func(c echo.Context) error {
		token := c.Get("user").(*jwt.Token)
		rawClaims := token.Claims.(*jwt.MapClaims)

		switch claims := (*rawClaims)["claims"]; claims {
		case "user":
			var claims models.UserAuthClaims

			if err := DecodeMap(rawClaims, &claims); err != nil {
				return err
			}

			// Extract tenant from JWT
			c.Response().Header().Set("X-Tenant-ID", claims.Tenant)

			return nil
		case "device":
			var claims models.DeviceAuthClaims

			if err := DecodeMap(rawClaims, &claims); err != nil {
				return err
			}

			// Extract device UID from JWT
			c.Response().Header().Set("X-Device-UID", claims.UID)

			return nil
		}

		return echo.ErrUnauthorized
	}, middleware.JWTWithConfig(middleware.JWTConfig{
		Claims:        &jwt.MapClaims{},
		SigningKey:    verifyKey,
		SigningMethod: "RS256",
	}))

	publicAPI.GET("/stats", func(c echo.Context) error {
		ctx := c.Get("ctx").(context.Context)

		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		stats, err := store.GetStats(ctx)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, stats)
	})

	publicAPI.GET("/sessions", func(c echo.Context) error {
		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := sessionmngr.NewService(store)

		var query struct {
			Page    int `query:"page"`
			PerPage int `query:"per_page"`
		}
		c.Bind(&query)

		page := query.Page
		perPage := query.PerPage
		if perPage < 1 || perPage > 100 {
			perPage = 10
		}
		if page < 1 {
			page = 1
		}

		sessions, err := svc.ListSessions(ctx, perPage, page)
		if err != nil {
			return err
		}

		count, err := svc.CountSessions(ctx)
		if err != nil {
			return err
		}

		c.Response().Header().Set("X-Total-Count", strconv.FormatInt(count, 10))

		return c.JSON(http.StatusOK, sessions)
	})

	publicAPI.GET("/sessions/:uid", func(c echo.Context) error {
		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := sessionmngr.NewService(store)

		session, err := svc.GetSession(ctx, models.UID(c.Param("uid")))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, session)
	})

	internalAPI.PATCH("/sessions/:uid", func(c echo.Context) error {
		var req struct {
			Authenticated bool `json:"authenticated"`
		}

		err := c.Bind(&req)
		if err != nil {
			return err
		}

		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := sessionmngr.NewService(store)

		return svc.SetSessionAuthenticated(ctx, models.UID(c.Param("uid")), req.Authenticated)
	})

	internalAPI.POST("/sessions", func(c echo.Context) error {
		session := new(models.Session)
		err := c.Bind(&session)
		if err != nil {
			return err
		}

		ctx := c.Get("ctx").(context.Context)

		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := sessionmngr.NewService(store)

		session, err = svc.CreateSession(ctx, *session)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, session)
	})

	publicAPI.GET("/ws/ssh", func(c echo.Context) error {
		websocket.Handler(ssh2ws.Handler).ServeHTTP(c.Response(), c.Request())
		return nil
	})

	internalAPI.GET("/lookup", func(c echo.Context) error {
		var query struct {
			Domain string `query:"domain"`
			Name   string `query:"name"`
		}

		if err := c.Bind(&query); err != nil {
			return err
		}

		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := deviceadm.NewService(store)

		device, err := svc.LookupDevice(ctx, query.Domain, query.Name)
		if err != nil {
			return nil
		}

		return c.JSON(http.StatusOK, device)
	})

	internalAPI.POST("/sessions/:uid/finish", func(c echo.Context) error {
		ctx := c.Get("ctx").(context.Context)

		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := sessionmngr.NewService(store)

		return svc.DeactivateSession(ctx, models.UID(c.Param("uid")))
	})

	e.Logger.Fatal(e.Start(":8080"))
}

func DecodeMap(input interface{}, output interface{}) error {
	config := &mapstructure.DecoderConfig{
		TagName:  "json",
		Metadata: nil,
		Result:   output,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}
