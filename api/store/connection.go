package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type ConnectionResolver uint

const (
	ConnectionIDResolver ConnectionResolver = iota + 1
)

type ConnectionStore interface {
	// ConnectionCreate creates a connection. Returns the inserted ID and an error if any.
	ConnectionCreate(ctx context.Context, connection *models.Connection) (insertedID string, err error)

	// ConnectionResolve fetches a connection using a specific resolver. Scope it to a
	// tenant by passing the InNamespace query option.
	ConnectionResolve(ctx context.Context, resolver ConnectionResolver, value string, opts ...QueryOption) (*models.Connection, error)

	// ConnectionList retrieves a list of connections. Returns the list, the total count
	// of matched documents, and an error if any.
	ConnectionList(ctx context.Context, opts ...QueryOption) (connections []models.Connection, count int, err error)

	// ConnectionUpdate updates a connection. Returns an error if any.
	ConnectionUpdate(ctx context.Context, connection *models.Connection) (err error)

	// ConnectionDelete deletes a connection. Returns an error if any.
	ConnectionDelete(ctx context.Context, connection *models.Connection) (err error)
}
