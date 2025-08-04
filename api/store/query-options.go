package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type NamespaceQueryOption func(ctx context.Context, ns *models.Namespace) error

type QueryOption func(ctx context.Context) error

type QueryOptions interface {
	// InNamespace matches a document that belongs to the provided namespace
	InNamespace(tenantID string) QueryOption

	// WithDeviceStatus matches a device with the provided status
	WithDeviceStatus(models.DeviceStatus) QueryOption

	// Match applies the provided query filters to match records
	Match(fs *query.Filters) QueryOption

	// Paginate applies pagination to limit the number of records returned.
	// If paginator is nil, no pagination is applied.
	Paginate(paginator *query.Paginator) QueryOption

	// Sort applies sorting criteria to order the returned records.
	// If sorter is nil, no specific sorting is applied.
	Sort(sorter *query.Sorter) QueryOption
}
