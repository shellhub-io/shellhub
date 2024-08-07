package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type NamespaceStore interface {
	// NamespaceList retrieves a list of namespaces. When the user's ID is available in the context, it will exclude
	// namespaces where the user has a pending membership status.
	NamespaceList(ctx context.Context, paginator query.Paginator, filters query.Filters, export bool) ([]models.Namespace, int, error)

	// NamespaceGet retrieves a namespace identified by the given tenantID.
	// If countDevices is set to true, it populates the [github.com/shellhub-io/shellhub/pkg/models.Namespace.DevicesCount].
	// Otherwise, the value will always be 0.
	//
	// It returns the namespace or an error if any.
	NamespaceGet(ctx context.Context, tenantID string, countDevices bool) (*models.Namespace, error)
	// NamespaceGetPreferred retrieves a namespace identified by the given tenantID where the user is a member.
	// If the tenantID is an empty string, it defaults to the first namespace found where the user is a member
	// (usually the first one to it was added).
	//
	// It returns the namespace or an error if any.
	NamespaceGetPreferred(ctx context.Context, tenantID, userID string) (*models.Namespace, error)

	NamespaceGetByName(ctx context.Context, name string) (*models.Namespace, error)
	NamespaceCreate(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error)

	// NamespaceEdit updates a namespace with the specified tenant.
	// It returns an error, if any, or store.ErrNoDocuments if the namespace does not exist.
	NamespaceEdit(ctx context.Context, tenant string, changes *models.NamespaceChanges) error

	NamespaceUpdate(ctx context.Context, tenantID string, namespace *models.Namespace) error
	NamespaceDelete(ctx context.Context, tenantID string) error

	// NamespaceAddMember adds a new member to the namespace with the specified tenantID.
	// It returns an error if any.
	NamespaceAddMember(ctx context.Context, tenantID string, member *models.Member) error
	// NamespaceUpdateMember updates a member with the specified memberID in the namespace with the specified tenantID with
	// the changes. It returns an error if any.
	NamespaceUpdateMember(ctx context.Context, tenantID string, memberID string, changes *models.MemberChanges) error
	// NamespaceRemoveMember removes a member with the specified memberID in the namespace with the specified tenantID.
	// If the namespace's tenant ID is the member's preffered tenant ID, it will set the value to an empty string.
	// It returns an error if any.
	NamespaceRemoveMember(ctx context.Context, tenantID string, memberID string) error

	NamespaceSetSessionRecord(ctx context.Context, sessionRecord bool, tenantID string) error
	NamespaceGetSessionRecord(ctx context.Context, tenantID string) (bool, error)
}
