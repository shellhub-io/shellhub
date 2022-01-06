package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

type DeviceTags interface {
	CreateTag(ctx context.Context, uid models.UID, name string) error
	RemoveTag(ctx context.Context, uid models.UID, name string) error
	RenameTag(ctx context.Context, tenantID string, currentName string, newName string) error
	UpdateTag(ctx context.Context, uid models.UID, tags []string) error
	GetTags(ctx context.Context, tenant string) ([]string, int, error)
	DeleteTags(ctx context.Context, tenant string, name string) error
}

func (s *service) CreateTag(ctx context.Context, uid models.UID, name string) error {
	if err := validateTagName(name); err != nil {
		return err
	}

	device, err := s.store.DeviceGet(ctx, uid)
	if err != nil || device == nil {
		return ErrDeviceNotFound
	}

	if len(device.Tags) == 5 {
		return ErrMaxTagReached
	}

	if contains(device.Tags, name) {
		return ErrDuplicateTagName
	}

	return s.store.DeviceCreateTag(ctx, uid, name)
}

func (s *service) RemoveTag(ctx context.Context, uid models.UID, name string) error {
	device, err := s.store.DeviceGet(ctx, uid)
	if err != nil || device == nil {
		return ErrDeviceNotFound
	}

	if !contains(device.Tags, name) {
		return ErrTagNameNotFound
	}

	return s.store.DeviceRemoveTag(ctx, uid, name)
}

func (s *service) RenameTag(ctx context.Context, tenantID string, currentName string, newName string) error {
	if err := validateTagName(newName); err != nil {
		return err
	}

	tags, count, err := s.store.DeviceGetTags(ctx, tenantID)
	if err != nil || count == 0 {
		return ErrNoTags
	}

	if !contains(tags, currentName) {
		return ErrTagNameNotFound
	}

	if contains(tags, newName) {
		return ErrDuplicateTagName
	}

	return s.store.DeviceRenameTag(ctx, tenantID, currentName, newName)
}

func (s *service) UpdateTag(ctx context.Context, uid models.UID, tags []string) error {
	for _, tag := range tags {
		if err := validateTagName(tag); err != nil {
			return err
		}
	}

	if len(tags) > 5 {
		return ErrMaxTagReached
	}

	device, err := s.store.DeviceGet(ctx, uid)
	if err != nil || device == nil {
		return ErrDeviceNotFound
	}

	return s.store.DeviceUpdateTag(ctx, uid, tags)
}

func (s *service) GetTags(ctx context.Context, tenant string) ([]string, int, error) {
	namespace, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil || namespace == nil {
		return nil, 0, ErrNamespaceNotFound
	}

	return s.store.DeviceGetTags(ctx, namespace.TenantID)
}

func (s *service) DeleteTags(ctx context.Context, tenant string, name string) error {
	namespace, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil || namespace == nil {
		return ErrNamespaceNotFound
	}

	return s.store.DeviceDeleteTags(ctx, namespace.TenantID, name)
}

func contains(tags []string, name string) bool {
	for _, tag := range tags {
		if tag == name {
			return true
		}
	}

	return false
}

func validateTagName(tag string) error {
	if _, err := validator.ValidateVar(tag, "required,min=3,max=255,alphanum,ascii,excludes=/@&:"); err != nil {
		return ErrInvalidFormat
	}

	return nil
}
