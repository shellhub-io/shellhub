package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type NamespaceStore interface {
	// NamespaceList retrieves a list of namespaces based on the provided filters and pagination settings.
	// If the user ID is available in the context, it will only match namespaces that the user is a member
	// of and does not have a pending membership status. A list of options can be passed to inject
	// additional data into each namespace in the list.
	//
	// It returns the list of namespaces, the total count of matching documents (ignoring pagination), and
	// an error if any.
	NamespaceList(ctx context.Context, paginator query.Paginator, filters query.Filters, opts ...NamespaceQueryOption) ([]models.Namespace, int, error)

	// NamespaceGet retrieves a namespace identified by the given tenantID. A list of options can be
	// passed to inject additional data into the namespace.
	//
	// It returns the namespace or an error if any.
	NamespaceGet(ctx context.Context, tenantID string, opts ...NamespaceQueryOption) (*models.Namespace, error)

	// NamespaceGetByName retrieves a namespace by its name, similar to [Store.NamespaceGet], but matches by name instead
	// of tenantID.
	NamespaceGetByName(ctx context.Context, name string, opts ...NamespaceQueryOption) (*models.Namespace, error)

	// NamespaceGetPreferred retrieves the user's preferred namespace. If the user has no preferred namespace it returns
	// the first namespace where the user is a member (typically the first one the user was added to). A list of options
	// can be passed via `opts` to inject additional data into the namespace.
	//
	// It returns the namespace or an error if any.
	NamespaceGetPreferred(ctx context.Context, userID string, opts ...NamespaceQueryOption) (*models.Namespace, error)

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

	// NamespaceIncrementDeviceCount atomically increments or decrements the device count for a specific status within a namespace.
	// Returns [ErrNoDocuments] if the namespace is not found.
	NamespaceIncrementDeviceCount(ctx context.Context, tenantID string, status models.DeviceStatus, count int64) error
}
