package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"regexp"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"golang.org/x/crypto/ssh"
)

type Request struct {
	Namespace string
}

func (s *service) EvaluateKeyHostname(ctx context.Context, key *models.PublicKey, dev models.Device) (bool, error) {
	if key.Hostname == "" {
		return true, nil
	}

	ok, err := regexp.MatchString(key.Hostname, dev.Name)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (s *service) GetPublicKey(ctx context.Context, fingerprint, tenant string) (*models.PublicKey, error) {
	return s.store.PublicKeyGet(ctx, fingerprint, tenant)
}

func (s *service) CreatePublicKey(ctx context.Context, key *models.PublicKey, tenant string) error {
	key.CreatedAt = clock.Now()

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(key.Data) //nolint:dogsled
	if err != nil {
		return ErrInvalidFormat
	}

	key.Fingerprint = ssh.FingerprintLegacyMD5(pubKey)

	returnedKey, err := s.store.PublicKeyGet(ctx, key.Fingerprint, tenant)
	if err != nil && err != store.ErrNoDocuments {
		return err
	}

	if returnedKey != nil {
		return ErrDuplicateFingerprint
	}

	err = s.store.PublicKeyCreate(ctx, key)
	if err != nil {
		return err
	}

	return err
}

func (s *service) ListPublicKeys(ctx context.Context, pagination paginator.Query) ([]models.PublicKey, int, error) {
	return s.store.PublicKeyList(ctx, pagination)
}

func (s *service) UpdatePublicKey(ctx context.Context, fingerprint, tenant string, key *models.PublicKeyUpdate) (*models.PublicKey, error) {
	return s.store.PublicKeyUpdate(ctx, fingerprint, tenant, key)
}

func (s *service) DeletePublicKey(ctx context.Context, fingerprint, tenant string) error {
	return s.store.PublicKeyDelete(ctx, fingerprint, tenant)
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
		CreatedAt:   clock.Now(),
	}

	if err := s.store.PrivateKeyCreate(ctx, privateKey); err != nil {
		return nil, err
	}

	return privateKey, nil
}
