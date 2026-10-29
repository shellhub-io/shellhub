package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// NamespaceResolver names the field a namespace is looked up by.
type NamespaceResolver uint

// The fields a namespace can be resolved by. The zero value is not one, so an unset resolver
// cannot silently mean the first.
const (
	NamespaceTenantIDResolver NamespaceResolver = iota + 1
	NamespaceNameResolver
)

// NamespaceStore persists namespaces, their settings, and their device counts.
type NamespaceStore interface {
	// NamespaceList retrieves a list of namespaces based on the provided filters and pagination settings.
	// If the user ID is available in the context, it will only match namespaces that the user is a member
	// of and does not have a pending membership status. A list of options can be passed to inject
	// additional data into each namespace in the list.
	//
	// It returns the list of namespaces, the total count of matching documents (ignoring pagination), and
	// an error if any.
	NamespaceList(ctx context.Context, opts ...QueryOption) ([]models.Namespace, int64, error)

	// NamespaceResolve fetches a namespace using a specific resolver.
	//
	// It returns the resolved namespace if found and an error, if any.
	NamespaceResolve(ctx context.Context, resolver NamespaceResolver, value string) (*models.Namespace, error)

	// NamespaceGetDeviceLimit returns the namespace's device ceiling and its accepted-device
	// count. It exists so the device-limit check does not have to go through NamespaceResolve,
	// which carries every membership and its user for callers that never read them.
	//
	// It returns ErrNoDocuments when no namespace matches the tenant ID.
	NamespaceGetDeviceLimit(ctx context.Context, tenantID string) (models.NamespaceDeviceLimit, error)

	// NamespaceGetMembers returns the namespace's members as enriched MemberView rows (name,
	// username, email, role and a flattened status), joining the users table so the caller gets
	// full identity rather than the thin Member embedded in a namespace. A list of options can be
	// passed for sorting and pagination.
	//
	// It returns the members, the total count (ignoring pagination), and an error if any.
	NamespaceGetMembers(ctx context.Context, sc scope.Scope, opts ...QueryOption) ([]models.MemberView, int64, error)

	// NamespaceGetPreferred retrieves the user's preferred namespace. If the user has no preferred namespace it returns
	// the first namespace where the user is a member (typically the first one the user was added to). A list of options
	// can be passed via `opts` to inject additional data into the namespace.
	//
	// It returns the namespace or an error if any.
	NamespaceGetPreferred(ctx context.Context, userID string) (*models.Namespace, error)

	// NamespaceCreate stores a namespace and returns its tenant ID. It always stores it as new,
	// clearing namespace.Settings.Legacy on the caller's struct; only NamespaceUpgrade writes the
	// flag. It seeds the owner's access policy, enrolling the owner as a member when
	// namespace.Members omits them, appending to the slice the caller passed, because that policy
	// has a foreign key to the membership row. It joins the transaction ctx carries, so the
	// caller's rollback also removes the namespace.
	NamespaceCreate(ctx context.Context, namespace *models.Namespace) (string, error)

	// NamespaceLockOwner blocks until no other transaction holds the namespace ownership lock of
	// ownerID, then holds it until the transaction in ctx ends. Locks on different owners do not
	// wait on each other. A caller enforcing a namespace limit takes it before counting the
	// owner's namespaces, so the count sees what the previous holder committed. It returns
	// ErrLockOutsideTransaction when ctx carries no transaction, ctx's error when ctx ends while
	// waiting, and an error wrapping ErrInternal when the lock query fails otherwise.
	NamespaceLockOwner(ctx context.Context, ownerID string) error

	NamespaceConflicts(ctx context.Context, target *models.NamespaceConflicts) (conflicts []string, has bool, err error)

	// NamespaceUpdate updates a namespace. It returns an error, if any, or store.ErrNoDocuments if the
	// 	namespace does not exist. It never writes the legacy flag; only NamespaceUpgrade does.
	NamespaceUpdate(ctx context.Context, namespace *models.Namespace) error

	// NamespaceUpgrade clears the legacy flag of a namespace in one conditional update, so two
	// concurrent calls cannot both succeed. It returns [ErrNamespaceNotLegacy] when the namespace is
	// already new or does not exist, and the database's error otherwise.
	NamespaceUpgrade(ctx context.Context, tenantID string) error

	// NamespaceIncrementDeviceCount atomically increments or decrements the device count for a specific status within a namespace.
	// Returns [ErrNoDocuments] if the namespace is not found.
	NamespaceIncrementDeviceCount(ctx context.Context, sc scope.Scope, status models.DeviceStatus, count int64) error

	NamespaceDelete(ctx context.Context, namespace *models.Namespace) error
	NamespaceDeleteMany(ctx context.Context, tenantIDs []string) (int64, error)

	// NamespaceSyncDeviceCounts recalculates and sets the device counter cache fields
	// for all namespaces based on actual device data.
	NamespaceSyncDeviceCounts(ctx context.Context) error
}
