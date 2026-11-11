package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

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

	// WithAPIKeyID matches records whose api_key_id column equals the given key, which is
	// how the credentials an automation owns are told from a person's.
	WithAPIKeyID(apiKeyID string) QueryOption

	// WithoutAPIKeyOwner matches records no API key owns, leaving the ones a person does.
	WithoutAPIKeyOwner() QueryOption

	// ForUpdate locks the rows the query returns until the transaction in the context ends, so a
	// concurrent ForUpdate on the same rows waits for that transaction and then reads what it
	// committed. Outside a transaction the lock is released as soon as the query returns. The pg
	// store fails the query with ErrLockWithoutTableAlias when the query sets no table alias.
	ForUpdate() QueryOption

	// Match narrows the records to those the filter matches, as one condition joined with AND to the
	// query's other conditions, so no operator in the filter widens past them. Nodes join left to
	// right: an operator joins every condition after it until the next operator, conditions before
	// the first operator join with OR, AND binds tighter than OR, and a group is one bracketed
	// condition. A property the store does not know, or an operator it does not support, is
	// skipped. The pg store fails the query with the first value error a property raises, with
	// ErrFilterNodeInvalid for a node of an unknown type or with Params of another shape, and with
	// ErrQueryNotFound outside a store query.
	Match(fs *query.Filters) QueryOption

	// Paginate applies pagination to limit the number of records returned.
	// If paginator is nil, no pagination is applied.
	Paginate(paginator *query.Paginator) QueryOption

	// Sort applies sorting criteria to order the returned records.
	// If sorter is nil, no specific sorting is applied.
	Sort(sorter *query.Sorter) QueryOption
}
