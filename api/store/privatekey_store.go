package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type PrivateKeyStore interface {
	CreatePrivateKey(ctx context.Context, key *models.PrivateKey) error
	GetPrivateKey(ctx context.Context, fingerprint string) (*models.PrivateKey, error)
}
