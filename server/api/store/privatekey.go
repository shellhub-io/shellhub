package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// PrivateKeyStore persists the short-lived keys the server mints to reach a device on a
// client's behalf.
type PrivateKeyStore interface {
	PrivateKeyCreate(ctx context.Context, key *models.PrivateKey) error
	PrivateKeyGet(ctx context.Context, fingerprint string) (*models.PrivateKey, error)
}
