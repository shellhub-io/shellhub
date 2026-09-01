package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// PublicKeyResolver names the field a public key is looked up by.
type PublicKeyResolver int

// The fields a public key can be resolved by.
const (
	PublicKeyFingerprintResolver PublicKeyResolver = iota
)

// PublicKeyStore persists a namespace's public keys and the rules restricting each.
type PublicKeyStore interface {
	// PublicKeyList retrieves a list of public keys within the given namespace scope, with optional
	// filtering and pagination.
	PublicKeyList(ctx context.Context, sc scope.Scope, opts ...QueryOption) ([]models.PublicKey, int, error)
	// PublicKeyResolve retrieves a public key by the given resolver type and value, within the given
	// namespace scope.
	PublicKeyResolve(ctx context.Context, sc scope.Scope, resolver PublicKeyResolver, value string, opts ...QueryOption) (*models.PublicKey, error)
	// PublicKeyCreate creates a new public key and returns its fingerprint.
	PublicKeyCreate(ctx context.Context, key *models.PublicKey) (string, error)
	// PublicKeyUpdate updates an existing public key.
	PublicKeyUpdate(ctx context.Context, publicKey *models.PublicKey) error
	// PublicKeyDelete removes a public key.
	PublicKeyDelete(ctx context.Context, publicKey *models.PublicKey) error
}
