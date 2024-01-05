package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type DeviceTagsStore interface {
	DeviceCreateTag(ctx context.Context, uid models.UID, tag string) error
	DeviceRemoveTag(ctx context.Context, uid models.UID, tag string) error

	// DeviceUpdateTag sets the tags for a device with the specified UID.
	// It returns the number of matching documents, the number of modified documents, and any encountered errors.
	DeviceUpdateTag(ctx context.Context, uid models.UID, tags []string) (matchedCount int64, updatedCount int64, err error)

	DeviceRenameTag(ctx context.Context, tenant, currentTags, newTags string) (updatedCount int64, err error)
	DeviceDeleteTag(ctx context.Context, tenant, tag string) error
}
