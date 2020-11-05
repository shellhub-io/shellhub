package sshkeys

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"golang.org/x/crypto/ssh"
)

type Service interface {
	GetPublicKey(ctx context.Context, fingerprint string) (*models.PublicKey, error)
	CreatePrivateKey(ctx context.Context) (*models.PrivateKey, error)
}

type service struct {
	store store.Store
}

func NewService(store store.Store) Service {
	return &service{store}
}

func (s *service) GetPublicKey(ctx context.Context, fingerprint string) (*models.PublicKey, error) {
	return s.store.GetPublicKey(ctx, fingerprint)
}

func (s *service) CreatePrivateKey(ctx context.Context) (*models.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	pubKey, err := ssh.NewPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}

	privateKey := &models.PrivateKey{
		Data: pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		}),
		Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
		CreatedAt:   time.Now(),
	}

	if err := s.store.CreatePrivateKey(ctx, privateKey); err != nil {
		return nil, err
	}

	return privateKey, nil
}
