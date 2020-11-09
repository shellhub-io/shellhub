package sshkeys

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
	"golang.org/x/crypto/ssh"
)

type Service interface {
	ListPublicKeys(ctx context.Context, pagination paginator.Query) ([]models.PublicKey, int, error)
	GetPublicKey(ctx context.Context, fingerprint string) (*models.PublicKey, error)
	CreatePublicKey(ctx context.Context, key *models.PublicKey) error
	UpdatePublicKey(ctx context.Context, fingerprint string, key *models.PublicKeyUpdate) (*models.PublicKey, error)
	DeletePublicKey(ctx context.Context, fingerprint string) error
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

func (s *service) CreatePublicKey(ctx context.Context, key *models.PublicKey) error {
	key.CreatedAt = time.Now()

	// Assign current tenant from ctx
	if tenant := c.Tenant(); tenant != nil {
		key.TenantID = tenant.ID
	}

	return s.store.CreatePublicKey(ctx, key)
}

func (s *service) ListPublicKeys(ctx context.Context, pagination paginator.Query) ([]models.PublicKey, int, error) {
	return s.store.ListPublicKeys(ctx, pagination)
}

func (s *service) UpdatePublicKey(ctx context.Context, fingerprint string, key *models.PublicKeyUpdate) (*models.PublicKey, error) {
	return s.store.UpdatePublicKey(ctx, fingerprint, key)
}

func (s *service) DeletePublicKey(ctx context.Context, fingerprint string) error {
	return s.store.DeletePublicKey(ctx, fingerprint)
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
