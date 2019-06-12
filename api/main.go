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
}

type Device struct {
	ID        bson.ObjectId     `json:"-" bson:"_id,omitempty"`
	UID       string            `json:"uid"`
	Identity  map[string]string `json:"identity"`
	PublicKey string            `json:"public_key" bson:"public_key"`
	LastSeen  time.Time         `json:"last_seen"`
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

		uid := sha256.Sum256(structhash.Dump(req, 1))

		d := &Device{
			UID:       hex.EncodeToString(uid[:]),
			Identity:  req.Identity,
			PublicKey: req.PublicKey,
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
			LastSeen: time.Now(),
		}

		if err := db.C("connected_devices").Insert(&cd); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, echo.Map{
			"uid":   d.UID,
			"token": signature,
		})
	})

	e.GET("/devices", func(c echo.Context) error {
		db := c.Get("db").(*mgo.Database)

		devices := make([]Device, 0)
		if err := db.C("devices").Find(bson.M{}).All(&devices); err != nil {
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

	e.GET("/mqtt/auth", AuthenticateMqttClient)
	e.GET("/mqtt/acl", AuthorizeMqttClient)
	e.POST("/mqtt/webhook", ProcessMqttEvent)

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

		resp := []bson.M{}

		if err := db.C("connected_devices").Pipe(query).All(&resp); err != nil {
			fmt.Println(err)
			return err
		}

		connectedDevices := 0
		if len(resp) > 0 {
			connectedDevices = resp[0]["count"].(int)
		}

		return c.JSON(http.StatusOK, echo.Map{
			"connected_devices": connectedDevices,
		})
	})

	e.Logger.Fatal(e.Start(":8080"))
}
