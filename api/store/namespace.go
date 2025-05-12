package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type NamespaceIdent string

const (
	NamespaceIdentTenantID NamespaceIdent = "id"
	NamespaceIdentName     NamespaceIdent = "name"
)

type NamespaceStore interface {
	// NamespaceCreate creates a new namespace with the provided data.
	// It returns the inserted ID or an error, if any.
	NamespaceCreate(ctx context.Context, namespace *models.Namespace) (string, error)

	// NamespaceCreateMemberships creates membership associations between users and a namespace.
	// It returns an error, if any.
	NamespaceCreateMemberships(ctx context.Context, tenantID string, memberships ...models.Member) error

	// NamespaceConflicts reports whether the target contains conflicting attributes with the database.
	// Pass zero values for attributes you do not wish to match on.
	// It returns an array of conflicting attribute fields and an error, if any.
	NamespaceConflicts(ctx context.Context, target *models.NamespaceConflicts) (conflicts []string, has bool, err error)

	// NamespaceList retrieves a list of namespaces based on the provided query options.
	// It returns a list of namespaces, the total count, and an error if any.
	NamespaceList(ctx context.Context, opts ...QueryOption) ([]models.Namespace, int, error)

	// NamespaceGet retrieves a namespace based on the provided NamespaceIdent.
	// It returns an error if no record was found.
	NamespaceGet(ctx context.Context, ident NamespaceIdent, val string, opts ...QueryOption) (*models.Namespace, error)

	// NamespaceSave updates the namespace.
	// It returns an error, if any.
	NamespaceSave(ctx context.Context, namespace *models.Namespace) (err error)

	NamespaceSaveMembership(ctx context.Context, tenantID string, member *models.Member) (err error)

	// NamespaceDelete deletes the namespace.
	// It returns an error, if any.
	NamespaceDelete(ctx context.Context, namespace *models.Namespace) (err error)

	NamespaceDeleteMembership(ctx context.Context, tenantID string, member *models.Member) (err error)
}
