package services

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

// DeviceTags contains the service's function to manage device tags.
type DeviceTags interface {
	CreateDeviceTag(ctx context.Context, uid models.UID, tag string) error
	RemoveDeviceTag(ctx context.Context, uid models.UID, tag string) error
	UpdateDeviceTag(ctx context.Context, uid models.UID, tags []string) error
}

// DeviceMaxTags is the number of tags that a device can have.
const DeviceMaxTags = 3

// CreateDeviceTag creates a new tag to a device. UID is the device's UID and tag is the tag's name.
//
// If the device does not exist, a NewErrDeviceNotFound error will be returned.
// If the tag already exist, a NewErrTagDuplicated error will be returned.
// If the device already has the maximum number of tags, a NewErrTagLimit error will be returned.
// A unknown error will be returned if the tag is not created.
func (s *service) CreateDeviceTag(ctx context.Context, uid models.UID, tag string) error {
	return nil
}

// RemoveDeviceTag removes a tag from a device. UID is the device's UID and tag is the tag's name.
//
// If the device does not exist, a NewErrDeviceNotFound error will be returned.
// If the tag does not exist, a NewErrTagNotFound error will be returned.
// A unknown error will be returned if the tag is not removed.
func (s *service) RemoveDeviceTag(ctx context.Context, uid models.UID, tag string) error {
	return nil
}

// UpdateDeviceTag updates a device's tags. UID is the device's UID and tags is the new tags.
//
// If length of tags is greater than DeviceMaxTags, a NewErrTagLimit error will be returned.
// If tags' list contains a duplicated one, it is removed and the device's tag will be updated.
// If the device does not exist, a NewErrDeviceNotFound error will be returned.
func (s *service) UpdateDeviceTag(ctx context.Context, uid models.UID, tags []string) error {
	return nil
}
