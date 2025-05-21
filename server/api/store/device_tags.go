package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type DeviceTagsStore interface {
	// DevicePushTag adds a new tag to the list of tags for a device with the specified UID.
	// Returns an error if any issues occur during the tag addition or ErrNoDocuments when matching documents are found.
	DevicePushTag(ctx context.Context, uid models.UID, tag string) error

	// DevicePullTag removes a tag from the list of tags for a device with the specified UID.
	// Returns an error if any issues occur during the tag removal or ErrNoDocuments when matching documents are found.
	DevicePullTag(ctx context.Context, uid models.UID, tag string) error

	// DeviceSetTags sets the tags for a device with the specified UID.
	// It returns the number of matching documents, the number of modified documents, and any encountered errors.
	DeviceSetTags(ctx context.Context, uid models.UID, tags []string) (matchedCount int64, updatedCount int64, err error)

	// DeviceBulkRenameTag replaces all occurrences of the old tag with the new tag for all devices belonging to the specified tenant.
	// Returns the number of documents updated and an error if any issues occur during the tag renaming.
	DeviceBulkRenameTag(ctx context.Context, tenant, currentTag, newTag string) (updatedCount int64, err error)

	// DeviceBulkDeleteTag removes a tag from all devices belonging to the specified tenant.
	// Returns the number of documents updated and an error if any issues occur during the tag deletion.
	DeviceBulkDeleteTag(ctx context.Context, tenant, tag string) (deletedCount int64, err error)

	// DeviceGetTags retrieves all tags associated with the tenant.
	// Returns the tags, the number of tags, and an error if any issues occur.
	DeviceGetTags(ctx context.Context, tenant string) (tag []string, n int, err error)
}
