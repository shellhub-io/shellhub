package store

import "context"

type TagsStore interface {
	TagsGet(ctx context.Context, tenant string) ([]string, int, error)
	TagRename(ctx context.Context, tenant string, tag string, newTag string) error
	TagDelete(ctx context.Context, tenant string, tag string) error
}
