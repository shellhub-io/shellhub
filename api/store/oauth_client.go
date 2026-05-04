package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type OAuthClientResolver uint

const (
	OAuthClientIDResolver       OAuthClientResolver = iota + 1 // matches by internal ID
	OAuthClientClientIDResolver                                // matches by public client_id
)

type OAuthClientStore interface {
	// OAuthClientCreate persists a new OAuth client. Returns the inserted ID and an error, if any.
	OAuthClientCreate(ctx context.Context, client *models.OAuthClient) (insertedID string, err error)

	// OAuthClientResolve fetches an OAuth client using the specified resolver.
	// Returns the client and an error, if any.
	OAuthClientResolve(ctx context.Context, resolver OAuthClientResolver, value string, opts ...QueryOption) (*models.OAuthClient, error)

	// OAuthClientList retrieves OAuth clients.
	// Returns the list, total count of matched documents, and an error, if any.
	OAuthClientList(ctx context.Context, opts ...QueryOption) ([]models.OAuthClient, int, error)

	// OAuthClientDelete removes an OAuth client. Returns an error, if any.
	OAuthClientDelete(ctx context.Context, client *models.OAuthClient) error
}
