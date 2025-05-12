package pg

import "context"

// TODO: refactor tags entirely

func (pg *Pg) TagsGet(ctx context.Context, tenant string) (tags []string, n int, err error) {
	return nil, 0, nil
}

func (pg *Pg) TagsRename(ctx context.Context, tenant string, oldTag string, newTag string) (updatedCount int64, err error) {
	return 0, nil
}

func (pg *Pg) TagsDelete(ctx context.Context, tenant string, tag string) (updatedCount int64, err error) {
	return 0, nil
}
