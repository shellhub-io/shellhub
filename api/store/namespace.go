package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type NamespaceResolver uint

const (
	NamespaceTenantIDResolver NamespaceResolver = iota + 1
	NamespaceNameResolver
)

type NamespaceStore interface {
	// NamespaceList retrieves a list of namespaces based on the provided filters and pagination settings.
	// If the user ID is available in the context, it will only match namespaces that the user is a member
	// of and does not have a pending membership status. A list of options can be passed to inject
	// additional data into each namespace in the list.
	//
	// It returns the list of namespaces, the total count of matching documents (ignoring pagination), and
	// an error if any.
	NamespaceList(ctx context.Context, opts ...QueryOption) ([]models.Namespace, int, error)

	// NamespaceResolve fetches a namespace using a specific resolver.
	//
	// It returns the resolved namespace if found and an error, if any.
	NamespaceResolve(ctx context.Context, resolver NamespaceResolver, value string) (*models.Namespace, error)

	// NamespaceGetPreferred retrieves the user's preferred namespace. If the user has no preferred namespace it returns
	// the first namespace where the user is a member (typically the first one the user was added to). A list of options
	// can be passed via `opts` to inject additional data into the namespace.
	//
	// It returns the namespace or an error if any.
	NamespaceGetPreferred(ctx context.Context, userID string) (*models.Namespace, error)

	NamespaceCreate(ctx context.Context, namespace *models.Namespace) (string, error)

	NamespaceConflicts(ctx context.Context, target *models.NamespaceConflicts) (conflicts []string, has bool, err error)

	// NamespaceUpdate updates a namespace. It returns an error, if any, or store.ErrNoDocuments if the
	// 	namespace does not exist.
	NamespaceUpdate(ctx context.Context, namespace *models.Namespace) error

	// NamespaceIncrementDeviceCount atomically increments or decrements the device count for a specific status within a namespace.
	// Returns [ErrNoDocuments] if the namespace is not found.
	NamespaceIncrementDeviceCount(ctx context.Context, tenantID string, status models.DeviceStatus, count int64) error

	NamespaceDelete(ctx context.Context, namespace *models.Namespace) error
	NamespaceDeleteMany(ctx context.Context, tenantIDs []string) (int64, error)

	// NamespaceSyncDeviceCounts recalculates and sets the device counter cache fields
	// for all namespaces based on actual device data.
	NamespaceSyncDeviceCounts(ctx context.Context) error
}
