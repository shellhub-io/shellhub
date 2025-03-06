package services

import (
	"context"
)

type TagsService interface {
	GetTags(ctx context.Context, tenant string) ([]string, int, error)
	RenameTag(ctx context.Context, tenant string, oldTag string, newTag string) error
	DeleteTag(ctx context.Context, tenant string, tag string) error
}

func (s *service) GetTags(ctx context.Context, tenant string) ([]string, int, error) {
	return nil, 0, nil
}

func (s *service) RenameTag(ctx context.Context, tenant string, oldTag string, newTag string) error {
	return nil
}

func (s *service) DeleteTag(ctx context.Context, tenant string, tag string) error {
	return nil
}
