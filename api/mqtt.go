package main

import (
	"fmt"
	"net"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// GET /mqtt/auth
func AuthenticateMqttClient(c echo.Context) error {
	q := AuthQuery{}

	if err := c.Bind(&q); err != nil {
		return err
	}

	ipaddr, err := net.LookupIP("ssh-server")
	if err != nil {
		return err
	}

	// Authorize connection from internal ssh-server client
	if q.IPAddr == ipaddr[0].String() {
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
	default:
		return nil
	}

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

	return nil
}
