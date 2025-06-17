package services

import (
	"context"
	"crypto/rsa"
	"os"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func LoadKeys() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	signBytes, err := os.ReadFile(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		return nil, nil, err
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		return nil, nil, err
	}

	verifyBytes, err := os.ReadFile(os.Getenv("PUBLIC_KEY"))
	if err != nil {
		return nil, nil, err
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		return nil, nil, err
	}

	return privKey, pubKey, nil
}

// adjustDeviceCounters handles the increment/decrement of device counters when a device status changes.
func (s *service) adjustDeviceCounters(ctx context.Context, tenant string, oldStatus, newStatus models.DeviceStatus) error {
	if err := s.store.NamespaceIncrementDeviceCount(ctx, tenant, oldStatus, -1); err != nil {
		return err
	}

	if err := s.store.NamespaceIncrementDeviceCount(ctx, tenant, newStatus, 1); err != nil { // nolint:revive
		return err
	}

	return nil
}

func contains(list []string, item string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}

	return false
}
