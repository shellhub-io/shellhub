package mqtthooks

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/api/pkg/store"
)

type Service interface {
	ProcessEvent(ctx context.Context, event models.MqttEvent) error
	AuthenticateClient(ctx context.Context, query models.MqttAuthQuery) error
}

type service struct {
	store  store.Store
	pubKey *rsa.PublicKey
}

func NewService(store store.Store, pubKey *rsa.PublicKey) Service {
	return &service{store, pubKey}
}

func (s *service) ProcessEvent(ctx context.Context, event models.MqttEvent) error {
	switch event.Action {
	case models.MqttClientConnectedEventType:
		return s.store.UpdateDeviceStatus(ctx, models.UID(event.MqttClientEvent.Username), true)
	case models.MqttClientDisconnectedEventType:
		return s.store.UpdateDeviceStatus(ctx, models.UID(event.MqttClientEvent.Username), false)
	}

	return nil
}

func (s *service) AuthenticateClient(ctx context.Context, query models.MqttAuthQuery) error {
	// Authorize connection from internal ssh service
	if query.Username == "ssh" {
		addrs, err := net.LookupHost("ssh")
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			if addr == query.IPAddr {
				return nil
			}
		}

		return errors.New("unauthorized")
	}

	token, err := jwt.ParseWithClaims(query.Password, &models.DeviceAuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return s.pubKey, nil
	})
	if err != nil {
		return err
	}

	if _, ok := token.Claims.(models.DeviceAuthClaims); ok && token.Valid {
		return nil
	}

	return nil
}
