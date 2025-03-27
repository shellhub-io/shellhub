package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type NamespaceIdent string

const (
	NamespaceIdentID   NamespaceIdent = "id"
	NamespaceIdentName NamespaceIdent = "name"
)

type NamespaceStore interface {
	NamespaceCreate(ctx context.Context, namespace *models.Namespace) (string, error)

	NamespaceCreateMemberships(ctx context.Context, memberships []models.Membership) error

	NamespaceConflicts(ctx context.Context, target *models.NamespaceConflicts) (conflicts []string, has bool, err error)

	// NamespaceList retrieves a list of namespaces based on the provided filters and pagination settings.
	// If the user ID is available in the context, it will only match namespaces that the user is a member
	// of and does not have a pending membership status. A list of options can be passed to inject
	// additional data into each namespace in the list.
	//
	// It returns the list of namespaces, the total count of matching documents (ignoring pagination), and
	// an error if any.
	NamespaceList(ctx context.Context, opts ...QueryOption) ([]models.Namespace, int, error)

	NamespaceGet(ctx context.Context, ident NamespaceIdent, val string, opts ...QueryOption) (*models.Namespace, error)

	NamespaceSetSessionRecord(ctx context.Context, sessionRecord bool, tenantID string) error
	NamespaceGetSessionRecord(ctx context.Context, tenantID string) (bool, error)
}
