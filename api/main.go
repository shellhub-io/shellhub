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
	"github.com/shellhub-io/shellhub/api/authsvc"
	"github.com/shellhub-io/shellhub/api/deviceadm"
	"github.com/shellhub-io/shellhub/api/firewall"
	"github.com/shellhub-io/shellhub/api/sessionmngr"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	api "github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	mgo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type config struct {
	MongoHost string `envconfig:"mongo_host" default:"mongo"`
	MongoPort int    `envconfig:"mongo_port" default:"27017"`
}

var verifyKey *rsa.PublicKey

const (
	TenantIDHeader = "X-Tenant-ID"
)

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
			tenant := c.Request().Header.Get(TenantIDHeader)

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
			Filter string `query:"filter"`
			paginator.Query
		}
		c.Bind(&query)

		// TODO: normalize is not required when request is privileged
		query.Normalize()

		devices, count, err := svc.ListDevices(ctx, query.Query, query.Filter)
		if err != nil {
			return err
		}

		c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))
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
		tenant := c.Request().Header.Get(TenantIDHeader)
		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := deviceadm.NewService(store)

		if err := svc.DeleteDevice(ctx, models.UID(c.Param("uid")), tenant); err != nil {
			if err == deviceadm.ErrUnauthorized {
				return c.NoContent(http.StatusForbidden)
			}
			return err
		}
		return nil
	})

	publicAPI.PATCH("/devices/:uid", func(c echo.Context) error {
		tenant := c.Request().Header.Get(TenantIDHeader)
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
			if err == deviceadm.ErrUnauthorized {
				return c.NoContent(http.StatusForbidden)
			}
			return err
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
			c.Response().Header().Set(TenantIDHeader, claims.Tenant)

			return nil
		case "device":
			var claims models.DeviceAuthClaims

			if err := DecodeMap(rawClaims, &claims); err != nil {
				return err
			}

			// Extract device UID from JWT
			c.Response().Header().Set(api.DeviceUIDHeader, claims.UID)

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

		query := paginator.NewQuery()
		c.Bind(query)

		// TODO: normalize is not required when request is privileged
		query.Normalize()

		sessions, count, err := svc.ListSessions(ctx, *query)
		if err != nil {
			return err
		}

		c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

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

	internalAPI.GET("/lookup", func(c echo.Context) error {
		var query struct {
			Domain    string `query:"domain"`
			Name      string `query:"name"`
			Username  string `query:"username"`
			IPAddress string `query:"ip_address"`
		}

		if err := c.Bind(&query); err != nil {
			return err
		}

		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := deviceadm.NewService(store)
		fw := firewall.NewService(store)

		device, err := svc.LookupDevice(ctx, query.Domain, query.Name)
		if err != nil {
			return nil
		}

		ok, err := fw.Evaluate(ctx, firewall.Request{
			Hostname:  query.Name,
			Namespace: query.Domain,
			Username:  query.Username,
			IPAddress: query.IPAddress,
		})
		if err != nil {
			return err
		}

		if !ok {
			return c.NoContent(http.StatusForbidden)
		}

		return c.JSON(http.StatusOK, device)
	})

	internalAPI.POST("/sessions/:uid/finish", func(c echo.Context) error {
		ctx := c.Get("ctx").(context.Context)

		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := sessionmngr.NewService(store)

		return svc.DeactivateSession(ctx, models.UID(c.Param("uid")))
	})

	publicAPI.GET("/firewall/rules", func(c echo.Context) error {
		ctx := c.Get("ctx").(context.Context)

		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := firewall.NewService(store)

		query := paginator.NewQuery()
		c.Bind(query)

		// TODO: normalize is not required when request is privileged
		query.Normalize()

		rules, _, _ := svc.ListRules(ctx, *query)

		return c.JSON(http.StatusOK, rules)
	})

	publicAPI.GET("/firewall/rules/:id", func(c echo.Context) error {
		ctx := c.Get("ctx").(context.Context)

		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := firewall.NewService(store)

		rule, err := svc.GetRule(ctx, c.Param("id"))
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, rule)
	})

	publicAPI.POST("/firewall/rules", func(c echo.Context) error {
		ctx := c.Get("ctx").(context.Context)

		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := firewall.NewService(store)

		var rule models.FirewallRule
		if err := c.Bind(&rule); err != nil {
			return err
		}

		if tenant, ok := ctx.Value("tenant").(string); ok {
			rule.TenantID = tenant
		}

		if err := svc.CreateRule(ctx, &rule); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, rule)
	})

	publicAPI.PUT("/firewall/rules/:id", func(c echo.Context) error {
		ctx := c.Get("ctx").(context.Context)

		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := firewall.NewService(store)

		var rule models.FirewallRuleUpdate
		if err := c.Bind(&rule); err != nil {
			return err
		}

		if value, err := svc.UpdateRule(ctx, c.Param("id"), rule); err != nil {
			return err
		} else {
			return c.JSON(http.StatusOK, value)
		}
	})

	publicAPI.DELETE("/firewall/rules/:id", func(c echo.Context) error {
		ctx := c.Get("ctx").(context.Context)

		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := firewall.NewService(store)

		if err := svc.DeleteRule(ctx, c.Param("id")); err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	})

	e.Logger.Fatal(e.Start(":8080"))
}

func DecodeMap(input, output interface{}) error {
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
