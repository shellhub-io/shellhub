package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type SSHKeysService interface {
	EvaluateKeyFilter(ctx context.Context, key *models.PublicKey, dev models.Device) (bool, error)
	EvaluateKeyUsername(ctx context.Context, key *models.PublicKey, username string) (bool, error)
	ListPublicKeys(ctx context.Context, paginator query.Paginator) ([]models.PublicKey, int, error)
	GetPublicKey(ctx context.Context, fingerprint, tenant string) (*models.PublicKey, error)
	CreatePublicKey(ctx context.Context, req requests.PublicKeyCreate, tenant string) (*responses.PublicKeyCreate, error)
	UpdatePublicKey(ctx context.Context, fingerprint, tenant string, key requests.PublicKeyUpdate) (*models.PublicKey, error)
	DeletePublicKey(ctx context.Context, fingerprint, tenant string) error
	CreatePrivateKey(ctx context.Context) (*models.PrivateKey, error)
}

type Request struct {
	Namespace string
}

func (s *service) EvaluateKeyFilter(_ context.Context, key *models.PublicKey, dev models.Device) (bool, error) {
	return false, nil
}

func (s *service) EvaluateKeyUsername(_ context.Context, key *models.PublicKey, username string) (bool, error) {
	return false, nil
}

func (s *service) GetPublicKey(ctx context.Context, fingerprint, tenant string) (*models.PublicKey, error) {
	return nil, nil
}

func (s *service) CreatePublicKey(ctx context.Context, req requests.PublicKeyCreate, tenant string) (*responses.PublicKeyCreate, error) {
	return nil, nil
}

func (s *service) ListPublicKeys(ctx context.Context, paginator query.Paginator) ([]models.PublicKey, int, error) {
	return nil, 0, nil
}

func (s *service) UpdatePublicKey(ctx context.Context, fingerprint, tenant string, key requests.PublicKeyUpdate) (*models.PublicKey, error) {
	return nil, nil
}

func (s *service) DeletePublicKey(ctx context.Context, fingerprint, tenant string) error {
	return nil
}

func (s *service) CreatePrivateKey(ctx context.Context) (*models.PrivateKey, error) {
	return nil, nil
}
