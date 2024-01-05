package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type DeviceTagsStore interface {
	// DeviceCreateTag adds a new tag to the list of tags for a device with the specified UID.
	// Returns an error if any issues occur during the tag addition or ErrNoDocuments when matching documents are found.
	DeviceCreateTag(ctx context.Context, uid models.UID, tag string) error

	// DeviceRemoveTag removes a tag from the list of tags for a device with the specified UID.
	// Returns an error if any issues occur during the tag removal or ErrNoDocuments when matching documents are found.
	DeviceRemoveTag(ctx context.Context, uid models.UID, tag string) error

	// DeviceUpdateTag sets the tags for a device with the specified UID.
	// It returns the number of matching documents, the number of modified documents, and any encountered errors.
	DeviceUpdateTag(ctx context.Context, uid models.UID, tags []string) (matchedCount int64, updatedCount int64, err error)

	// DeviceRenameTag replaces all occurrences of the old tag with the new tag for all devices belonging to the specified tenant.
	// Returns the number of documents updated and an error if any issues occur during the tag renaming.
	DeviceRenameTag(ctx context.Context, tenant, currentTags, newTags string) (updatedCount int64, err error)

	// DeviceDeleteTag removes a tag from all devices belonging to the specified tenant.
	// Returns the number of documents updated and an error if any issues occur during the tag deletion.
	DeviceDeleteTag(ctx context.Context, tenant, tag string) (deletedCount int64, err error)

	// DeviceGetTags retrieves all tags associated with the tenant.
	// Returns the tags, the number of tags, and an error if any issues occur.
	DeviceGetTags(ctx context.Context, tenant string) (tag []string, n int, err error)
}
