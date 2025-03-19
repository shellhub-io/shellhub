package pg

import "context"

// TODO: refactor tags entirely

func (pg *pg) TagsGet(ctx context.Context, tenant string) (tags []string, n int, err error) {
	return nil, 0, nil
}

func (pg *pg) TagsRename(ctx context.Context, tenant string, oldTag string, newTag string) (updatedCount int64, err error) {
	return 0, nil
}

func (pg *pg) TagsDelete(ctx context.Context, tenant string, tag string) (updatedCount int64, err error) {
	return 0, nil
}
