package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

func (s *Store) PrivateKeyCreate(ctx context.Context, key *models.PrivateKey) error {
	// TODO: private keys are now saved only in the frontend and this can be removed
	return nil
}

func (s *Store) PrivateKeyGet(ctx context.Context, fingerprint string) (*models.PrivateKey, error) {
	// TODO: private keys are now saved only in the frontend and this can be removedV
	return nil, nil
}
