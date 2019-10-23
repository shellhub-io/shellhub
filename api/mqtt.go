package main

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type ConnectedDevice struct {
	ID       bson.ObjectId `json:"-" bson:"_id,omitempty"`
	UID      string        `json:"uid"`
	TenantID string        `json:"tenant_id" bson:"tenant_id"`
	LastSeen time.Time     `json:"last_seen" bson:"last_seen"`
}

// GET /mqtt/auth
func AuthenticateMqttClient(c echo.Context) error {
	q := AuthQuery{}

	if err := c.Bind(&q); err != nil {
		return err
	}

	// Authorize connection from internal ssh client
	if q.Username == "ssh" {
		return nil
	}

	token, err := jwt.ParseWithClaims(q.Password, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return verifyKey, nil
	})
	if err != nil {
		return err
	}

	if _, ok := token.Claims.(AuthClaims); ok && token.Valid {
		return nil
	}

	return nil
}

// GET /mqtt/acl
func AuthorizeMqttClient(c echo.Context) error {
	q := ACLQuery{}

	if err := c.Bind(&q); err != nil {
		return err
	}

	return nil
}

// POST /mqtt/webhook
func ProcessMqttEvent(c echo.Context) error {
	db := c.Get("db").(*mgo.Database)

	evt := WebHookEvent{}

	if err := c.Bind(&evt); err != nil {
		return err
	}

	switch evt.Action {
	case WebHookClientConnectedEventType:
		d := Device{}
		err := db.C("devices").Find(bson.M{"uid": evt.WebHookClientEvent.Username}).One(&d)
		if err != nil {
			return err
		}

		d.LastSeen = time.Now()

		_, err = db.C("devices").Upsert(bson.M{"uid": d.UID}, d)
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
	case WebHookClientDisconnectedEventType:
		_, err := db.C("connected_devices").RemoveAll(bson.M{"uid": evt.WebHookClientEvent.Username})
		return err
	default:
		return nil
	}

	return nil
}
