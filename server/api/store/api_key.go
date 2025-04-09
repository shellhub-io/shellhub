package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type APIKeyStore interface {
	// APIKeyCreate creates an API key with the provided data. Returns the inserted ID and an error if any.
	APIKeyCreate(ctx context.Context, APIKey *models.APIKey) (insertedID string, err error)

	// APIKeyGet retrieves an API key based on its ID. Returns the API key and an error if any.
	APIKeyGet(ctx context.Context, id string) (apiKey *models.APIKey, err error)

	// APIKeyGetByName retrieves an API key based on its name and tenant ID. Returns the API key and an error if any.
	APIKeyGetByName(ctx context.Context, tenantID string, name string) (apiKey *models.APIKey, err error)

	// APIKeyConflicts reports whether the target contains conflicting attributes with the database. Pass zero values for
	// attributes you do not wish to match on.  It returns an array of conflicting attribute fields and an error, if any.
	//
	// API keys attributes can be duplicated in document level when the tenant id are different.
	APIKeyConflicts(ctx context.Context, tenantID string, target *models.APIKeyConflicts) (conflicts []string, has bool, err error)

	// APIKeyList retrieves a list of API keys for the specified tenant using the given paginator and sorter values.
	// Returns the list of API keys, the total count of matched documents, and an error if any.
	APIKeyList(ctx context.Context, tenantID string, paginator query.Paginator, sorter query.Sorter) (apiKeys []models.APIKey, count int, err error)

	// APIKeyUpdate updates an API key with the specified name and tenant ID using the given changes.
	// Any zero values in the changes (e.g., empty strings) will be ignored during the update.
	// Returns an error if any.
	APIKeyUpdate(ctx context.Context, tenantID, name string, changes *models.APIKeyChanges) (err error)

	// APIKeyDelete deletes an API key with the specified name and tenant ID. Returns an error if any.
	APIKeyDelete(ctx context.Context, tenantID, name string) (err error)
}
