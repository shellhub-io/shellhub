package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type APIKeyIdent string

const (
	APIKeyIdentID   APIKeyIdent = "key_digest"
	APIKeyIdentName APIKeyIdent = "name"
)

type APIKeyStore interface {
	// APIKeyCreate creates an API key with the provided data. Returns the inserted ID and an error if any.
	APIKeyCreate(ctx context.Context, APIKey *models.APIKey) (insertedID string, err error)

	// APIKeyConflicts reports whether the target contains conflicting attributes with the database. Pass zero values for
	// attributes you do not wish to match on.  It returns an array of conflicting attribute fields and an error, if any.
	//
	// API keys attributes can be duplicated in document level when the tenant id are different.
	APIKeyConflicts(ctx context.Context, tenantID string, target *models.APIKeyConflicts) (conflicts []string, has bool, err error)

	// APIKeyList retrieves a list of API keys for the specified tenant using the given paginator and sorter values.
	// Returns the list of API keys, the total count of matched documents, and an error if any.
	APIKeyList(ctx context.Context, opts ...QueryOption) (apiKeys []models.APIKey, count int, err error)

	APIKeyGet(ctx context.Context, ident APIKeyIdent, val string, tenantID string) (*models.APIKey, error)

	APIKeySave(ctx context.Context, apiKey *models.APIKey) error

	APIKeyDelete(ctx context.Context, apiKey *models.APIKey) error
}
