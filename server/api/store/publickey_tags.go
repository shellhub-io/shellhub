package store

import "context"

type PublicKeyTagsStore interface {
	// PublicKeyPushTag adds a new tag to the list of tags for a device with the specified UID.
	// Returns an error if any issues occur during the tag addition or ErrNoDocuments when matching documents are found.
	//
	// The tag need to exist on a device. If it is not true, the action will fail.
	PublicKeyPushTag(ctx context.Context, tenant, fingerprint, tag string) error

	// PublicKeyPullTag removes a tag from the list of tags for a device with the specified UID.
	// Returns an error if any issues occur during the tag removal or ErrNoDocuments when matching documents are found.
	//
	// To remove a tag, that tag needs to exist on a device. If it is not, the action will fail.
	PublicKeyPullTag(ctx context.Context, tenant, fingerprint, tag string) error

	// PublicKeySetTags sets the tags for a public key with the specified fingerprint and tenant.
	// It returns the number of matching documents, the number of modified documents, and any encountered errors.
	//
	// All tags need to exist on a device. If it is not true, the update action will fail.
	PublicKeySetTags(ctx context.Context, tenant, fingerprint string, tags []string) (matchedCount int64, updatedCount int64, err error)

	// PublicKeyBulkRenameTag replaces all occurrences of the old tag with the new tag for all public keys to the specified tenant.
	// Returns the number of documents updated and an error if any issues occur during the tag renaming.
	PublicKeyBulkRenameTag(ctx context.Context, tenant, currentTag, newTag string) (updatedCount int64, err error)

	// PublicKeyBulkDeleteTag removes a tag from all public keys belonging to the specified tenant.
	// Returns the number of documents updated and an error if any issues occur during the tag deletion.
	PublicKeyBulkDeleteTag(ctx context.Context, tenant, tag string) (updatedCount int64, err error)

	// PublicKeyGetTags retrieves all tags associated with the tenant.
	// Returns the tags, the number of tags, and an error if any issues occur.
	PublicKeyGetTags(ctx context.Context, tenant string) (tag []string, size int, err error)
}
