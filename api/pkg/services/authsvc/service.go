package authsvc

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/cnf/structhash"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/api/pkg/store"
)

type Service interface {
	AuthDevice(ctx context.Context, req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error)
	AuthUser(ctx context.Context, req models.UserAuthRequest) (*models.UserAuthResponse, error)
}

type service struct {
	store   store.Store
	privKey *rsa.PrivateKey
}

func NewService(store store.Store, privKey *rsa.PrivateKey) Service {
	return &service{store, privKey}
}

func (s *service) AuthDevice(ctx context.Context, req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error) {
	uid := sha256.Sum256(structhash.Dump(req.DeviceAuth, 1))

	device := models.Device{
		UID:        hex.EncodeToString(uid[:]),
		Identity:   req.Identity,
		Attributes: req.Attributes,
		PublicKey:  req.PublicKey,
		TenantID:   req.TenantID,
		LastSeen:   time.Now(),
	}

	if err := s.store.AddDevice(ctx, device); err != nil {
		return nil, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.DeviceAuthClaims{
		UID: string(uid[:]),
	})

	tokenStr, err := token.SignedString(s.privKey)
	if err != nil {
		return nil, err
	}

	if err := s.store.UpdateDeviceStatus(ctx, models.UID(device.UID), true); err != nil {
		return nil, err
	}

	for _, uid := range req.Sessions {
		if err := s.store.KeepAliveSession(ctx, models.UID(uid)); err != nil {
			continue
		}
	}

	dev, _ := s.store.GetDevice(ctx, models.UID(device.UID))
	user, err := s.store.GetUserByTenant(ctx, device.TenantID)

	return &models.DeviceAuthResponse{
		UID:       hex.EncodeToString(uid[:]),
		Token:     tokenStr,
		Name:      dev.Name,
		Namespace: user.Username,
	}, nil
}

func (s *service) AuthUser(ctx context.Context, req models.UserAuthRequest) (*models.UserAuthResponse, error) {
	user, err := s.store.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	password := sha256.Sum256([]byte(req.Password))
	if user.Password == hex.EncodeToString(password[:]) {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.UserAuthClaims{
			Name:   user.Username,
			Admin:  true,
			Tenant: user.TenantID,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Hour * 72).Unix(),
			},
		})

		tokenStr, err := token.SignedString(s.privKey)
		if err != nil {
			return nil, err
		}

		return &models.UserAuthResponse{
			Token:  tokenStr,
			User:   user.Username,
			Tenant: user.TenantID,
		}, nil
	}

	return nil, errors.New("unauthorized")
}
