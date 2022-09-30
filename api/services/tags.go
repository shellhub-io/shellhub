package services

import (
	"context"

	"github.com/shellhub-io/shellhub/api/businesses"
)

type TagsService interface {
	GetTags(ctx context.Context, tenant string) ([]string, int, error)
	RenameTag(ctx context.Context, tenant string, oldTag string, newTag string) error
	DeleteTag(ctx context.Context, tenant string, tag string) error
}

func (s *service) GetTags(ctx context.Context, tenant string) ([]string, int, error) {
	return businesses.Tag(ctx, s.store).
		FromTenant(tenant).
		Get()
}

func (s *service) RenameTag(ctx context.Context, tenant string, oldTag string, newTag string) error {
	return businesses.Tag(ctx, s.store).
		FromTenant(tenant).
		FromTag(oldTag).
		ToTag(newTag).
		Rename()
}

func (s *service) DeleteTag(ctx context.Context, tenant string, tag string) error {
	return businesses.Tag(ctx, s.store).
		FromTenant(tenant).
		FromTag(tag).
		Delete()
}
