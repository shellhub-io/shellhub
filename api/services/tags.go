package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type TagsService interface {
	GetTags(ctx context.Context, tenant string) ([]string, int, error)
	RenameTag(ctx context.Context, tenant string, oldTag string, newTag string) error
	DeleteTag(ctx context.Context, tenant string, tag string) error
}

func (s *service) GetTags(ctx context.Context, tenant string) ([]string, int, error) {
	namespace, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil || namespace == nil {
		return nil, 0, NewErrNamespaceNotFound(tenant, err)
	}

	return s.store.TagsGet(ctx, namespace.TenantID)
}

func (s *service) RenameTag(ctx context.Context, tenant string, oldTag string, newTag string) error {
	if ok, err := s.validator.Struct(models.NewDeviceTag(newTag)); !ok || err != nil {
		return NewErrTagInvalid(newTag, err)
	}

	tags, count, err := s.store.TagsGet(ctx, tenant)
	if err != nil || count == 0 {
		return NewErrTagEmpty(tenant, err)
	}

	if !contains(tags, oldTag) {
		return NewErrTagNotFound(oldTag, nil)
	}

	if contains(tags, newTag) {
		return NewErrTagDuplicated(newTag, nil)
	}

	return s.store.TagRename(ctx, tenant, oldTag, newTag)
}

func (s *service) DeleteTag(ctx context.Context, tenant string, tag string) error {
	if ok, err := s.validator.Struct(models.NewDeviceTag(tag)); !ok || err != nil {
		return NewErrTagInvalid(tag, err)
	}

	namespace, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil || namespace == nil {
		return NewErrNamespaceNotFound(tenant, err)
	}

	tags, count, err := s.store.TagsGet(ctx, namespace.TenantID)
	if err != nil || count == 0 {
		return NewErrTagEmpty(tenant, err)
	}

	if !contains(tags, tag) {
		return NewErrTagNotFound(tag, nil)
	}

	return s.store.TagDelete(ctx, namespace.TenantID, tag)
}
