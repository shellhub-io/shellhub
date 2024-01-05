package store

import "context"

type TagsStore interface {
	TagsGet(ctx context.Context, tenant string) ([]string, int, error)
	TagsBulkRename(ctx context.Context, tenant string, tag string, newTag string) error
	TagsBulkDelete(ctx context.Context, tenant string, tag string) error
}
