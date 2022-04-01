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
		return NewErrTagInvalid(name, nil)
	}

	device, err := s.store.DeviceGet(ctx, uid)
	if err != nil || device == nil {
		return NewErrDeviceNotFound(uid, err)
	}

	if len(device.Tags) == DeviceMaxTags {
		return NewErrTagLimit(DeviceMaxTags, nil)
	}

	if contains(device.Tags, name) {
		return NewErrTagDuplicated(name, nil)
	}

	return s.store.DeviceCreateTag(ctx, uid, name)
}

func (s *service) RemoveDeviceTag(ctx context.Context, uid models.UID, name string) error {
	device, err := s.store.DeviceGet(ctx, uid)
	if err != nil || device == nil {
		return NewErrDeviceNotFound(uid, err)
	}

	if !contains(device.Tags, name) {
		return NewErrTagNotFound(name, nil)
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
		return NewErrTagLimit(DeviceMaxTags, nil)
	}

	tagSet := listToSet(tags)

	for _, tag := range tagSet {
		if !validator.ValidateFieldTag(tag) {
			return NewErrTagInvalid(tag, nil)
		}
	}

	device, err := s.store.DeviceGet(ctx, uid)
	if err != nil || device == nil {
		return NewErrDeviceNotFound(uid, err)
	}

	return s.store.DeviceUpdateTag(ctx, uid, tagSet)
}
