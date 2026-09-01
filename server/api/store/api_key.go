package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// APIKeyResolver names the field an API key is looked up by.
type APIKeyResolver uint

// The fields an API key can be resolved by. The zero value is not one, so an unset resolver
// cannot silently mean the first.
const (
	APIKeyIDResolver APIKeyResolver = iota + 1
	APIKeyNameResolver
)

// APIKeyStore persists API keys. Only a key's hash is stored, so a lookup by key means a
// lookup by that hash.
type APIKeyStore interface {
	// APIKeyCreate creates an API key with the provided data. Returns the inserted ID and an error if any.
	APIKeyCreate(ctx context.Context, APIKey *models.APIKey) (insertedID string, err error)

	// APIKeyResolve fetches an API key using a specific resolver within the given namespace scope.
	//
	// It returns the resolved API key if found and an error, if any.
	APIKeyResolve(ctx context.Context, sc scope.Scope, resolver APIKeyResolver, value string, opts ...QueryOption) (*models.APIKey, error)

	// APIKeyConflicts reports whether the target contains conflicting attributes with the database. Pass zero values for
	// attributes you do not wish to match on.  It returns an array of conflicting attribute fields and an error, if any.
	//
	// API keys attributes can be duplicated in document level when the tenant id are different.
	APIKeyConflicts(ctx context.Context, sc scope.Scope, target *models.APIKeyConflicts) (conflicts []string, has bool, err error)

	// APIKeyList retrieves a list of API keys within the given namespace scope.
	// Returns the list of API keys, the total count of matched documents, and an error if any.
	APIKeyList(ctx context.Context, sc scope.Scope, opts ...QueryOption) (apiKeys []models.APIKey, count int, err error)

	// APIKeyUpdate updates an API key. It returns an error if any.
	APIKeyUpdate(ctx context.Context, apiKey *models.APIKey) (err error)

	// APIKeyDelete deletes an API key. It returns an error if any.
	APIKeyDelete(ctx context.Context, apiKey *models.APIKey) (err error)

	// APIKeyDeleteAllByCreator deletes every API key created by creatorID within the given tenant.
	// It is used to revoke a member's keys when they leave or are removed from the namespace.
	// Deleting no keys is not an error. It returns an error if any.
	APIKeyDeleteAllByCreator(ctx context.Context, tenantID, creatorID string) (err error)
}
