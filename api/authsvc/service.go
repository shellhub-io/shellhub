package authsvc

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/cnf/structhash"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type Service interface {
	AuthDevice(ctx context.Context, req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error)
	AuthUser(ctx context.Context, req models.UserAuthRequest) (*models.UserAuthResponse, error)
	PublicKey() *rsa.PublicKey
}

type service struct {
	store   store.Store
	privKey *rsa.PrivateKey
	pubKey  *rsa.PublicKey
}

func NewService(store store.Store, privKey *rsa.PrivateKey, pubKey *rsa.PublicKey) Service {
	if privKey == nil || pubKey == nil {
		var err error
		privKey, pubKey, err = loadKeys()
		if err != nil {
			panic(err)
		}
	}

	return &service{store, privKey, pubKey}
}

func (s *service) AuthDevice(ctx context.Context, req *models.DeviceAuthRequest) (*models.DeviceAuthResponse, error) {
	uid := sha256.Sum256(structhash.Dump(req.DeviceAuth, 1))

	device := models.Device{
		UID:       hex.EncodeToString(uid[:]),
		Identity:  req.Identity,
		Info:      req.Info,
		PublicKey: req.PublicKey,
		TenantID:  req.TenantID,
		LastSeen:  time.Now(),
	}
	sameMacDev, err := s.store.GetDeviceByMac(ctx, device.Identity.MAC, device.TenantID)
	if sameMacDev != nil && sameMacDev.UID != device.UID {
		return nil, errors.New("device with this mac address already authored")
	}

	if err := s.store.AddDevice(ctx, device); err != nil {
		return nil, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.DeviceAuthClaims{
		UID: hex.EncodeToString(uid[:]),
		AuthClaims: models.AuthClaims{
			Claims: "device",
		},
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

	dev, err := s.store.GetDevice(ctx, models.UID(device.UID))
	if err != nil {
		return nil, err
	}

	user, err := s.store.GetUserByTenant(ctx, device.TenantID)
	if err != nil {
		return nil, err
	}

	return &models.DeviceAuthResponse{
		UID:       hex.EncodeToString(uid[:]),
		Token:     tokenStr,
		Name:      dev.Name,
		Namespace: user.Username,
	}, nil
}

func (s *service) AuthUser(ctx context.Context, req models.UserAuthRequest) (*models.UserAuthResponse, error) {
	user, err := s.store.GetUserByUsername(ctx, strings.ToLower(req.Username))
	if err != nil {
		user, err = s.store.GetUserByEmail(ctx, strings.ToLower(req.Username))
		if err != nil {
			return nil, err
		}
	}

	password := sha256.Sum256([]byte(req.Password))
	if user.Password == hex.EncodeToString(password[:]) {
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, models.UserAuthClaims{
			Username: user.Username,
			Admin:    true,
			Tenant:   user.TenantID,
			AuthClaims: models.AuthClaims{
				Claims: "user",
			},
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
			Name:   user.Name,
			User:   user.Username,
			Tenant: user.TenantID,
			Email:  user.Email,
		}, nil
	}

	return nil, errors.New("unauthorized")
}

func (s *service) PublicKey() *rsa.PublicKey {
	return s.pubKey
}

func loadKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	signBytes, err := ioutil.ReadFile(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		return nil, nil, err
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return nil, nil, err
	}

	verifyBytes, err := ioutil.ReadFile(os.Getenv("PUBLIC_KEY"))
	if err != nil {
		return nil, nil, err
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return nil, nil, err
	}

	return privKey, pubKey, nil
}
