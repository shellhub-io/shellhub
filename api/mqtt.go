package main

import (
	"context"
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/shellhub-io/shellhub/api/pkg/models"
	"github.com/shellhub-io/shellhub/api/pkg/services/deviceadm"
	"github.com/shellhub-io/shellhub/api/pkg/store/mongo"
	mgo "gopkg.in/mgo.v2"
)

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

	token, err := jwt.ParseWithClaims(q.Password, &models.DeviceAuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return verifyKey, nil
	})
	if err != nil {
		return err
	}

	if _, ok := token.Claims.(models.DeviceAuthClaims); ok && token.Valid {
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
	evt := WebHookEvent{}

	if err := c.Bind(&evt); err != nil {
		return err
	}

	switch evt.Action {
	case WebHookClientConnectedEventType:
		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := deviceadm.NewService(store)

		if err := svc.UpdateDeviceStatus(ctx, models.UID(evt.WebHookClientEvent.Username), true); err != nil {
			return err
		}
	case WebHookClientDisconnectedEventType:
		ctx := c.Get("ctx").(context.Context)
		store := mongo.NewStore(ctx.Value("db").(*mgo.Database))
		svc := deviceadm.NewService(store)

		if err := svc.UpdateDeviceStatus(ctx, models.UID(evt.WebHookClientEvent.Username), false); err != nil {
			return err
		}
	default:
		return nil
	}

	return nil
}
