package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type PublicKeyResolver int

const (
	PublicKeyFingerprintResolver PublicKeyResolver = iota
)

type PublicKeyStore interface {
	// PublicKeyList retrieves a list of public keys with optional filtering and pagination.
	PublicKeyList(ctx context.Context, opts ...QueryOption) ([]models.PublicKey, int, error)
	// PublicKeyResolve retrieves a public key by the given resolver type and value.
	PublicKeyResolve(ctx context.Context, resolver PublicKeyResolver, value string, opts ...QueryOption) (*models.PublicKey, error)
	// PublicKeyCreate creates a new public key and returns its fingerprint.
	PublicKeyCreate(ctx context.Context, key *models.PublicKey) (string, error)
	// PublicKeyUpdate updates an existing public key.
	PublicKeyUpdate(ctx context.Context, publicKey *models.PublicKey) error
	// PublicKeyDelete removes a public key.
	PublicKeyDelete(ctx context.Context, publicKey *models.PublicKey) error
}
