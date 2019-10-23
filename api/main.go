package main

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/cnf/structhash"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type DeviceAuthRequest struct {
	Identity  map[string]string `json:"identity"`
	PublicKey string            `json:"public_key"`
	TenantID  string            `json:"tenant_id"`
	Sessions  []string          `json:"sessions,omitempty"`
}

type Device struct {
	ID        bson.ObjectId     `json:"-" bson:"_id,omitempty"`
	UID       string            `json:"uid"`
	Identity  map[string]string `json:"identity"`
	PublicKey string            `json:"public_key" bson:"public_key"`
	TenantID  string            `json:"tenant_id" bson:"tenant_id"`
	LastSeen  time.Time         `json:"last_seen"`
	Online    bool              `json:"online"`
}

type Session struct {
	ID        bson.ObjectId `json:"-" bson:"_id,omitempty"`
	UID       string        `json:"uid"`
	Device    string        `json:"device"`
	TenantID  string        `json:"tenant_id" bson:"tenant_id"`
	Username  string        `json:"username"`
	IPAddress string        `json:"ip_address" bson:"ip_address"`
	StartedAt time.Time     `json:"started_at" bson:"started_at"`
	LastSeen  time.Time     `json:"last_seen" bson:"last_seen"`
	Active    bool          `json:"active"`
}

type ActiveSession struct {
	ID       bson.ObjectId `json:"-" bson:"_id,omitempty"`
	UID      string        `json:"uid"`
	LastSeen time.Time     `json:"last_seen" bson:"last_seen"`
}

type AuthQuery struct {
	Username string `query:"username"`
	Password string `query:"password"`
	IPAddr   string `query:"ipaddr"`
}

type ACLQuery struct {
	Access   string `query:"access"`
	Username string `query:"username"`
	Topic    string `query:"topic"`
	IPAddr   string `query:"ipaddr"`
}

type User struct {
	ID       bson.ObjectId `json:"-" bson:"_id,omitempty"`
	Username string        `json:"username"`
	Password string        `json:"password"`
	TenantID string        `json:"tenant_id" bson:"tenant_id"`
}

type AuthClaims struct {
	UID string `json:"uid"`

	jwt.StandardClaims
}

type WebHookEvent struct {
	Action string `json:"action"`

	WebHookClientEvent
}

type WebHookClientEvent struct {
	ClientID string `json:"client_id"`
	Username string `json:"username"`
}

const (
	WebHookClientConnectedEventType    = "client_connected"
	WebHookClientDisconnectedEventType = "client_disconnected"
)

var verifyKey *rsa.PublicKey

func main() {
	e := echo.New()
	e.Use(middleware.Logger())

	session, err := mgo.Dial("mongodb://mongo:27017")
	if err != nil {
		panic(err)
	}

	err = session.DB("main").C("devices").EnsureIndex(mgo.Index{
		Key:        []string{"uid"},
		Unique:     true,
		Name:       "uid",
		Background: false,
	})
	if err != nil {
		panic(err)
	}

	err = session.DB("main").C("connected_devices").EnsureIndex(mgo.Index{
		Key:         []string{"last_seen"},
		Name:        "last_seen",
		ExpireAfter: time.Duration(time.Second * 30),
	})
	if err != nil {
		panic(err)
	}

	err = session.DB("main").C("connected_devices").EnsureIndex(mgo.Index{
		Key:        []string{"uid"},
		Unique:     false,
		Name:       "uid",
		Background: false,
	})
	if err != nil {
		panic(err)
	}

	err = session.DB("main").C("sessions").EnsureIndex(mgo.Index{
		Key:        []string{"uid"},
		Unique:     true,
		Name:       "uid",
		Background: false,
	})
	if err != nil {
		panic(err)
	}

	err = session.DB("main").C("active_sessions").EnsureIndex(mgo.Index{
		Key:         []string{"last_seen"},
		Name:        "last_seen",
		ExpireAfter: time.Duration(time.Second * 30),
	})
	if err != nil {
		panic(err)
	}

	err = session.DB("main").C("active_sessions").EnsureIndex(mgo.Index{
		Key:        []string{"uid"},
		Unique:     false,
		Name:       "uid",
		Background: false,
	})
	if err != nil {
		panic(err)
	}

	err = session.DB("main").C("users").EnsureIndex(mgo.Index{
		Key:        []string{"username"},
		Unique:     true,
		Name:       "username",
		Background: false,
	})
	if err != nil {
		panic(err)
	}

	err = session.DB("main").C("users").EnsureIndex(mgo.Index{
		Key:        []string{"tenant_id"},
		Unique:     true,
		Name:       "tenant_id",
		Background: false,
	})
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
			s := session.Clone()

			defer s.Close()

			c.Set("db", s.DB("main"))

			return next(c)
		}
	})

	e.POST("/devices/auth", func(c echo.Context) error {
		db := c.Get("db").(*mgo.Database)

		var req DeviceAuthRequest

		err := c.Bind(&req)
		if err != nil {
			return err
		}

		sessions := req.Sessions

		req.Sessions = []string{}

		uid := sha256.Sum256(structhash.Dump(req, 1))

		d := &Device{
			UID:       hex.EncodeToString(uid[:]),
			Identity:  req.Identity,
			PublicKey: req.PublicKey,
			TenantID:  req.TenantID,
			LastSeen:  time.Now(),
		}

		_, err = db.C("devices").Upsert(bson.M{"uid": d.UID}, d)
		if err != nil {
			return err
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, AuthClaims{
			UID: string(uid[:]),
		})

		signature, err := token.SignedString(signKey)
		if err != nil {
			return err
		}

		cd := &ConnectedDevice{
			UID:      d.UID,
			TenantID: d.TenantID,
			LastSeen: time.Now(),
		}

		if err := db.C("connected_devices").Insert(&cd); err != nil {
			return err
		}

		for _, s := range sessions {
			sess := Session{}
			err := db.C("sessions").Find(bson.M{"uid": s}).One(&sess)
			if err != nil {
				return err
			}

			sess.LastSeen = time.Now()

			_, err = db.C("sessions").Upsert(bson.M{"uid": sess.UID}, sess)
			if err != nil {
				return err
			}

			as := &ActiveSession{
				UID:      s,
				LastSeen: time.Now(),
			}

			if err := db.C("active_sessions").Insert(&as); err != nil {
				return err
			}
		}

		return c.JSON(http.StatusOK, echo.Map{
			"uid":   d.UID,
			"token": signature,
		})
	})

	e.GET("/devices", func(c echo.Context) error {
		db := c.Get("db").(*mgo.Database)

		devices := make([]Device, 0)

		query := []bson.M{
			{
				"$lookup": bson.M{
					"from":         "connected_devices",
					"localField":   "uid",
					"foreignField": "uid",
					"as":           "online",
				},
			},
			{
				"$addFields": bson.M{
					"online": bson.M{"$anyElementTrue": []interface{}{"$online"}},
				},
			},
		}

		// Only match for the respective tenant if requested
		if len(c.Request().Header.Get("X-Tenant-ID")) > 0 {
			query = append(query, bson.M{
				"$match": bson.M{
					"tenant_id": c.Request().Header.Get("X-Tenant-ID"),
				},
			})
		}

		if err := db.C("devices").Pipe(query).All(&devices); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, devices)
	})

	e.GET("/devices/:uid", func(c echo.Context) error {
		db := c.Get("db").(*mgo.Database)

		device := new(Device)
		if err := db.C("devices").Find(bson.M{"uid": c.Param("uid")}).One(&device); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, device)
	})

	e.DELETE("/devices/:uid", func(c echo.Context) error {
		db := c.Get("db").(*mgo.Database)

		if err := db.C("devices").Remove(bson.M{"uid": c.Param("uid")}); err != nil {
			return err
		}

		if err := db.C("sessions").Remove(bson.M{"device": c.Param("uid")}); err != nil {
			return err
		}

		return nil
	})

	e.GET("/mqtt/auth", AuthenticateMqttClient)
	e.GET("/mqtt/acl", AuthorizeMqttClient)
	e.POST("/mqtt/webhook", ProcessMqttEvent)

	e.POST("/login", func(c echo.Context) error {
		var login struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		c.Bind(&login)

		db := c.Get("db").(*mgo.Database)

		user := new(User)
		if err := db.C("users").Find(bson.M{"username": login.Username}).One(&user); err != nil {
			return echo.ErrUnauthorized
		}

		password := sha256.Sum256([]byte(login.Password))
		if user.Password == hex.EncodeToString(password[:]) {
			token := jwt.New(jwt.SigningMethodHS256)

			claims := token.Claims.(jwt.MapClaims)
			claims["name"] = user.Username
			claims["admin"] = true
			claims["tenant"] = user.TenantID
			claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

			t, err := token.SignedString([]byte("secret"))
			if err != nil {
				return err
			}

			return c.JSON(http.StatusOK, map[string]string{
				"token": t,
			})
		}

		return echo.ErrUnauthorized
	})

	e.GET("/auth", func(c echo.Context) error {
		token := c.Get("user").(*jwt.Token)
		claims := token.Claims.(jwt.MapClaims)

		// Extract tenant from JWT
		c.Response().Header().Set("X-Tenant-ID", claims["tenant"].(string))

		return nil
	}, middleware.JWT([]byte("secret")))

	e.GET("/users", func(c echo.Context) error {
		db := c.Get("db").(*mgo.Database)

		users := make([]Device, 0)
		if err := db.C("users").Find(bson.M{}).All(&users); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, users)
	})

	e.GET("/stats", func(c echo.Context) error {
		db := c.Get("db").(*mgo.Database)

		query := []bson.M{
			{"$group": bson.M{"_id": bson.M{"uid": "$uid"}, "count": bson.M{"$sum": 1}}},
			{"$group": bson.M{"_id": bson.M{"uid": "$uid"}, "count": bson.M{"$sum": 1}}},
		}

		// Only match for the respective tenant if requested
		if len(c.Request().Header.Get("X-Tenant-ID")) > 0 {
			query = append([]bson.M{{
				"$match": bson.M{
					"tenant_id": c.Request().Header.Get("X-Tenant-ID"),
				},
			}}, query...)
		}

		resp := []bson.M{}

		if err := db.C("connected_devices").Pipe(query).All(&resp); err != nil {
			fmt.Println(err)
			return err
		}

		fmt.Println(resp)

		connectedDevices := 0
		if len(resp) > 0 {
			connectedDevices = resp[0]["count"].(int)
		}

		return c.JSON(http.StatusOK, echo.Map{
			"connected_devices": connectedDevices,
		})
	})

	e.GET("/sessions", func(c echo.Context) error {
		db := c.Get("db").(*mgo.Database)

		sessions := make([]Session, 0)

		query := []bson.M{
			{
				"$lookup": bson.M{
					"from":         "active_sessions",
					"localField":   "uid",
					"foreignField": "uid",
					"as":           "active",
				},
			},
			{
				"$addFields": bson.M{
					"active": bson.M{"$anyElementTrue": []interface{}{"$active"}},
				},
			},
		}

		// Only match for the respective tenant if requested
		if len(c.Request().Header.Get("X-Tenant-ID")) > 0 {
			query = append(query, bson.M{
				"$match": bson.M{
					"tenant_id": c.Request().Header.Get("X-Tenant-ID"),
				},
			})
		}

		if err := db.C("sessions").Pipe(query).All(&sessions); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, sessions)
	})

	e.POST("/sessions", func(c echo.Context) error {
		db := c.Get("db").(*mgo.Database)

		var session Session
		err := c.Bind(&session)
		if err != nil {
			return err
		}

		session.StartedAt = time.Now()
		session.LastSeen = session.StartedAt

		device := new(Device)
		if err := db.C("devices").Find(bson.M{"uid": session.Device}).One(&device); err != nil {
			return err
		}

		session.TenantID = device.TenantID

		if err := db.C("sessions").Insert(session); err != nil {
			return err
		}

		as := &ActiveSession{
			UID:      session.UID,
			LastSeen: session.StartedAt,
		}

		if err := db.C("active_sessions").Insert(&as); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, session)
	})

	e.POST("/sessions/:uid/finish", func(c echo.Context) error {
		db := c.Get("db").(*mgo.Database)

		session := new(Session)
		if err := db.C("sessions").Find(bson.M{"uid": c.Param("uid")}).One(&session); err != nil {
			return err
		}

		session.LastSeen = time.Now()

		_, err = db.C("sessions").Upsert(bson.M{"uid": session.UID}, session)
		if err != nil {
			return err
		}

		_, err := db.C("active_sessions").RemoveAll(bson.M{"uid": session.UID})
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, session)
	})

	e.Logger.Fatal(e.Start(":8080"))
}
