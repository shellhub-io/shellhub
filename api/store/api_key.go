package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type APIKeyResolver uint

const (
	APIKeyIDResolver APIKeyResolver = iota + 1
	APIKeyNameResolver
)

type APIKeyStore interface {
	// APIKeyCreate creates an API key with the provided data. Returns the inserted ID and an error if any.
	APIKeyCreate(ctx context.Context, APIKey *models.APIKey) (insertedID string, err error)

	// APIKeyResolve fetches an API key using a specific resolver within a given tenant ID.
	//
	// It returns the resolved API key if found and an error, if any.
	APIKeyResolve(ctx context.Context, resolver APIKeyResolver, value string, opts ...QueryOption) (*models.APIKey, error)

	// APIKeyConflicts reports whether the target contains conflicting attributes with the database. Pass zero values for
	// attributes you do not wish to match on.  It returns an array of conflicting attribute fields and an error, if any.
	//
	// API keys attributes can be duplicated in document level when the tenant id are different.
	APIKeyConflicts(ctx context.Context, tenantID string, target *models.APIKeyConflicts) (conflicts []string, has bool, err error)

	// APIKeyList retrieves a list of API keys.
	// Returns the list of API keys, the total count of matched documents, and an error if any.
	APIKeyList(ctx context.Context, opts ...QueryOption) (apiKeys []models.APIKey, count int, err error)

	// APIKeyUpdate updates an API key. It returns an error if any.
	APIKeyUpdate(ctx context.Context, apiKey *models.APIKey) (err error)

	// APIKeyDelete deletes an API key. It returns an error if any.
	APIKeyDelete(ctx context.Context, apiKey *models.APIKey) (err error)
}
