package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type KnownHostStore interface {
	// KnownHostResolve fetches the stored known host for a target, scoped by
	// owner. A non-empty ownerID resolves the caller's personal record; an empty
	// ownerID resolves the namespace-shared (team) record. Returns ErrNoDocuments
	// when none is stored.
	KnownHostResolve(ctx context.Context, tenantID, ownerID, host string, port int) (*models.KnownHost, error)

	// KnownHostUpsert creates or replaces the known host for its (tenant, owner,
	// host, port) scope.
	KnownHostUpsert(ctx context.Context, knownHost *models.KnownHost) error

	// KnownHostDelete removes the stored known host for a target scope.
	KnownHostDelete(ctx context.Context, tenantID, ownerID, host string, port int) error
}
