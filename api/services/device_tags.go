package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
)

type DeviceTags interface {
	CreateDeviceTag(ctx context.Context, uid models.UID, name string) error
	RemoveDeviceTag(ctx context.Context, uid models.UID, name string) error
	UpdateDeviceTag(ctx context.Context, uid models.UID, tags []string) error
}

// DeviceMaxTags is the number of tags that a device can have.
const DeviceMaxTags = 3

func (s *service) CreateDeviceTag(ctx context.Context, uid models.UID, name string) error {
	if !validator.ValidateFieldTag(name) {
		return ErrInvalidFormat
	}

	device, err := s.store.DeviceGet(ctx, uid)
	if err != nil || device == nil {
		return ErrDeviceNotFound
	}

	if len(device.Tags) == DeviceMaxTags {
		return ErrMaxTagReached
	}

	if contains(device.Tags, name) {
		return ErrDuplicateTagName
	}

	return s.store.DeviceCreateTag(ctx, uid, name)
}

func (s *service) RemoveDeviceTag(ctx context.Context, uid models.UID, name string) error {
	device, err := s.store.DeviceGet(ctx, uid)
	if err != nil || device == nil {
		return ErrDeviceNotFound
	}

	if !contains(device.Tags, name) {
		return ErrTagNameNotFound
	}

	return s.store.DeviceRemoveTag(ctx, uid, name)
}

func (s *service) UpdateDeviceTag(ctx context.Context, uid models.UID, tags []string) error {
	listToSet := func(list []string) []string {
		s := make(map[string]bool)
		l := make([]string, 0)
		for _, o := range list {
			if _, ok := s[o]; !ok {
				s[o] = true
				l = append(l, o)
			}
		}

		return l
	}

	if len(tags) > DeviceMaxTags {
		return ErrMaxTagReached
	}

	tagSet := listToSet(tags)

	for _, tag := range tagSet {
		if !validator.ValidateFieldTag(tag) {
			return ErrInvalidFormat
		}
	}

	device, err := s.store.DeviceGet(ctx, uid)
	if err != nil || device == nil {
		return ErrDeviceNotFound
	}

	return s.store.DeviceUpdateTag(ctx, uid, tagSet)
}
