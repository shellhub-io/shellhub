package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// InstanceAPIKeyResolver names the field an instance API key is looked up by.
type InstanceAPIKeyResolver uint

// The fields an instance API key can be resolved by. The zero value is not one, so an unset
// resolver cannot silently mean the first.
const (
	InstanceAPIKeyIDResolver InstanceAPIKeyResolver = iota + 1
	InstanceAPIKeyNameResolver
)

// InstanceAPIKeyStore persists the API keys that authenticate as an instance administrator.
//
// No method takes a [scope.Scope]: an instance API key belongs to the instance rather than to a
// namespace, so there is nothing to bound a query to. That also keeps instance keys unreachable
// from [APIKeyStore], which is what stops a namespace key from ever resolving to an administrator.
type InstanceAPIKeyStore interface {
	// InstanceAPIKeyCreate creates an instance API key with the provided data. It returns the
	// inserted ID, and ErrDuplicate when the name or the digest is already taken.
	InstanceAPIKeyCreate(ctx context.Context, apiKey *models.InstanceAPIKey) (insertedID string, err error)

	// InstanceAPIKeyResolve fetches an instance API key using a specific resolver. It returns
	// ErrNoDocuments when no key matches, and ErrResolverNotFound for an unknown resolver.
	InstanceAPIKeyResolve(ctx context.Context, resolver InstanceAPIKeyResolver, value string, opts ...QueryOption) (*models.InstanceAPIKey, error)

	// InstanceAPIKeyList retrieves every instance API key, together with the total count before
	// pagination.
	InstanceAPIKeyList(ctx context.Context, opts ...QueryOption) (apiKeys []models.InstanceAPIKey, count int, err error)

	// InstanceAPIKeyDelete deletes the instance API key with the given name. It returns
	// ErrNoDocuments when no key matches.
	InstanceAPIKeyDelete(ctx context.Context, name string) error
}
