package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

type DeviceTags interface {
	CreateTag(ctx context.Context, uid models.UID, name string) error
	DeleteTag(ctx context.Context, uid models.UID, name string) error
	RenameTag(ctx context.Context, tenantID string, currentName string, newName string) error
	ListTag(ctx context.Context) ([]string, int, error)
	UpdateTag(ctx context.Context, uid models.UID, tags []string) error
	GetTags(ctx context.Context, tenant string) ([]string, int, error)
	DeleteAllTags(ctx context.Context, tenant string, name string) error
}

func (s *service) CreateTag(ctx context.Context, uid models.UID, name string) error {
	if err := validateTagName(name); err != nil {
		return err
	}

	device, err := s.store.DeviceGet(ctx, uid)
	if err != nil {
		return err
	}

	if device == nil {
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

func (s *service) DeleteTag(ctx context.Context, uid models.UID, name string) error {
	device, err := s.store.DeviceGet(ctx, uid)
	if err != nil {
		return err
	}

	if device == nil {
		return ErrDeviceNotFound
	}

	return s.store.DeviceDeleteTag(ctx, uid, name)
}

func (s *service) RenameTag(ctx context.Context, tenantID string, currentName string, newName string) error {
	if err := validateTagName(newName); err != nil {
		return err
	}

	return s.store.DeviceRenameTag(ctx, tenantID, currentName, newName)
}

func (s *service) ListTag(ctx context.Context) ([]string, int, error) {
	return s.store.DeviceListTag(ctx)
}

func (s *service) UpdateTag(ctx context.Context, uid models.UID, tags []string) error {
	for _, tag := range tags {
		if err := validateTagName(tag); err != nil {
			return err
		}
	}

	device, err := s.store.DeviceGet(ctx, uid)
	if err != nil {
		return err
	}

	if device == nil {
		return ErrDeviceNotFound
	}

	if len(device.Tags) == 5 {
		return ErrMaxTagReached
	}

	return s.store.DeviceUpdateTag(ctx, uid, tags)
}

func (s *service) GetTags(ctx context.Context, tenant string) ([]string, int, error) {
	ns, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil {
		return nil, 0, err
	}

	if ns == nil {
		return nil, 0, ErrNotFound
	}

	return s.store.DeviceGetTags(ctx, ns.TenantID)
}

func (s *service) DeleteAllTags(ctx context.Context, tenant string, name string) error {
	ns, err := s.store.NamespaceGet(ctx, tenant)
	if err != nil {
		return err
	}

	if ns == nil {
		return ErrNotFound
	}

	return s.store.DeviceDeleteAllTags(ctx, ns.TenantID, name)
}

func contains(s []string, name string) bool {
	for _, tag := range s {
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
