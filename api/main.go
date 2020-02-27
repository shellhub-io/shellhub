package main

import (
	"context"
	"crypto/rsa"
	"io/ioutil"
	"net/http"
	"os"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/shellhub-io/shellhub/api/pkg/services/authsvc"
	"github.com/shellhub-io/shellhub/api/pkg/services/deviceadm"
	"github.com/shellhub-io/shellhub/api/pkg/services/mqtthooks"
	"github.com/shellhub-io/shellhub/api/pkg/services/sessionmngr"
	"github.com/shellhub-io/shellhub/api/pkg/services/ssh2ws"
	"github.com/shellhub-io/shellhub/api/pkg/store/mongo"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	mgo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/websocket"
)

var verifyKey *rsa.PublicKey

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://mongo:27017")
	// Connect to MongoDB
	client, err := mgo.Connect(context.TODO(), clientOptions)

	if err != nil {
		panic(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}

	mod := mgo.IndexModel{
		Keys:    bson.D{{"uid", 1}},
		Options: options.Index().SetName("uid").SetUnique(true),
	}
	_, err = client.Database("main").Collection("devices").Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		panic(err)
	}

	mod = mgo.IndexModel{
		Keys:    bson.D{{"last_seen", 1}},
		Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(30),
	}
	_, err = client.Database("main").Collection("connected_devices").Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		panic(err)
	}

	mod = mgo.IndexModel{
		Keys:    bson.D{{"uid", 1}},
		Options: options.Index().SetName("uid").SetUnique(false),
	}
	_, err = client.Database("main").Collection("connected_devices").Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		panic(err)
	}

	mod = mgo.IndexModel{
		Keys:    bson.D{{"uid", 1}},
		Options: options.Index().SetName("uid").SetUnique(true),
	}
	_, err = client.Database("main").Collection("sessions").Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		panic(err)
	}

	mod = mgo.IndexModel{
		Keys:    bson.D{{"last_seen", 1}},
		Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(30),
	}
	_, err = client.Database("main").Collection("active_sessions").Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		panic(err)
	}

	mod = mgo.IndexModel{
		Keys:    bson.D{{"uid", 1}},
		Options: options.Index().SetName("uid").SetUnique(false),
	}
	_, err = client.Database("main").Collection("active_sessions").Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		panic(err)
	}

	mod = mgo.IndexModel{
		Keys:    bson.D{{"username", 1}},
		Options: options.Index().SetName("username").SetUnique(true),
	}
	_, err = client.Database("main").Collection("users").Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
		panic(err)
	}

	mod = mgo.IndexModel{
		Keys:    bson.D{{"tenant_id", 1}},
		Options: options.Index().SetName("tenant_id").SetUnique(true),
	}
	_, err = client.Database("main").Collection("users").Indexes().CreateOne(context.TODO(), mod)
	if err != nil {
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
			ctx := context.WithValue(c.Request().Context(), "tenant", tenant)
			ctx = context.WithValue(ctx, "db", client.Database("main"))

			c.Set("ctx", ctx)
			c.Set("db", client.Database("main"))

			return next(c)
		}
	})

	publicAPI := e.Group("/api")
	internalAPI := e.Group("/internal")
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

		devices, err := svc.ListDevices(ctx)
		if err != nil {
			return err
		}

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
		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := deviceadm.NewService(store)

		return svc.DeleteDevice(ctx, models.UID(c.Param("uid")))
	})

	publicAPI.PATCH("/devices/:uid", func(c echo.Context) error {
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

		return svc.RenameDevice(ctx, models.UID(c.Param("uid")), req.Name)
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
		claims := token.Claims.(*models.UserAuthClaims)

		// Extract tenant from JWT
		c.Response().Header().Set("X-Tenant-ID", claims.Tenant)

		return nil
	}, middleware.JWTWithConfig(middleware.JWTConfig{
		Claims:        &models.UserAuthClaims{},
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
		sessions, err := svc.ListSessions(ctx)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, sessions)
	})

	publicAPI.POST("/sessions", func(c echo.Context) error {
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

	internalAPI.GET("/mqtt/auth", func(c echo.Context) error {
		q := models.MqttAuthQuery{}

		if err := c.Bind(&q); err != nil {
			return err
		}

		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := mqtthooks.NewService(store, verifyKey)

		return svc.AuthenticateClient(ctx, q)
	})

	internalAPI.GET("/mqtt/superuser", func(c echo.Context) error {
		q := models.MqttAuthQuery{}

		if err := c.Bind(&q); err != nil {
			return err
		}

		return echo.NewHTTPError(http.StatusUnauthorized)
	})

	internalAPI.GET("/mqtt/acl", func(c echo.Context) error {
		q := models.MqttACLQuery{}

		if err := c.Bind(&q); err != nil {
			return err
		}

		return nil
	})

	internalAPI.POST("/mqtt/webhook", func(c echo.Context) error {
		evt := models.MqttEvent{}

		if err := c.Bind(&evt); err != nil {
			return err
		}

		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := mqtthooks.NewService(store, verifyKey)

		return svc.ProcessEvent(ctx, evt)
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
