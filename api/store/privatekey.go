package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type PrivateKeyStore interface {
	PrivateKeyCreate(ctx context.Context, key *models.PrivateKey) error
	PrivateKeyGet(ctx context.Context, fingerprint string) (*models.PrivateKey, error)
}
