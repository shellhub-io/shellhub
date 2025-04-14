package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/query"
)

type QueryOption func(ctx context.Context) error

type QueryOptions interface {
	Paginate(query.Paginator) QueryOption
	Order(query.Sorter) QueryOption
	Filter(query.Filters) QueryOption
	WithMember(string) QueryOption
	InNamespace(string) QueryOption
	WithStatus(string) QueryOption
}
