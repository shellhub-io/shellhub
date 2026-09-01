package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// NamespaceQueryOption shapes a query that has already resolved its namespace, so it can
// read the namespace's own settings — a device limit, for instance.
type NamespaceQueryOption func(ctx context.Context, ns *models.Namespace) error

// QueryOption shapes a query. Options are applied in order and may fail, which fails the
// query rather than silently returning unshaped results.
type QueryOption func(ctx context.Context) error

// QueryOptions carries the optional shaping of a query — filtering, pagination, sorting. The
// namespace an operation is bounded to is deliberately not here: it is a required scope.Scope
// parameter of every namespace-bound operation, rather than something a caller may forget to pass.
type QueryOptions interface {
	// WithDeviceStatus matches a device with the provided status
	WithDeviceStatus(models.DeviceStatus) QueryOption

	// WithMember filters namespaces where the given user is a member.
	WithMember(userID string) QueryOption

	// WithUserID matches records whose user_id column equals the given user.
	WithUserID(userID string) QueryOption

	// Match applies the provided query filters to match records
	Match(fs *query.Filters) QueryOption

	// Paginate applies pagination to limit the number of records returned.
	// If paginator is nil, no pagination is applied.
	Paginate(paginator *query.Paginator) QueryOption

	// Sort applies sorting criteria to order the returned records.
	// If sorter is nil, no specific sorting is applied.
	Sort(sorter *query.Sorter) QueryOption
}
