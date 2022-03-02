package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/validator"
)

type TagsService interface {
	GetTags(ctx context.Context, tenant string) ([]string, int, error)
	RenameTag(ctx context.Context, tenant string, oldTag string, newTag string) error
	DeleteTag(ctx context.Context, tenant string, tag string) error
}

func (s *service) GetTags(ctx context.Context, tenant string) ([]string, int, error) {
	namespace, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil || namespace == nil {
		return nil, 0, ErrNamespaceNotFound
	}

	return s.store.TagsGet(ctx, namespace.TenantID)
}

func (s *service) RenameTag(ctx context.Context, tenant string, oldTag string, newTag string) error {
	if !validator.ValidateFieldTag(newTag) {
		return ErrInvalidFormat
	}

	tags, count, err := s.store.TagsGet(ctx, tenant)
	if err != nil || count == 0 {
		return ErrNoTags
	}

	if !contains(tags, oldTag) {
		return ErrTagNameNotFound
	}

	if contains(tags, newTag) {
		return ErrDuplicateTagName
	}

	return s.store.TagRename(ctx, tenant, oldTag, newTag)
}

func (s *service) DeleteTag(ctx context.Context, tenant string, tag string) error {
	namespace, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil || namespace == nil {
		return ErrNamespaceNotFound
	}

	return s.store.TagDelete(ctx, namespace.TenantID, tag)
}
