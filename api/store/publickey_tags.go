package store

import "context"

type PublicKeyTagsStore interface {
	PublicKeyAddTag(ctx context.Context, tenant, fingerprint, tag string) error
	PublicKeyRemoveTag(ctx context.Context, tenant, fingerprint, tag string) error

	// PublicKeyUpdateTags sets the tags for a public key with the specified fingerprint and tenant.
	// It returns the number of matching documents, the number of modified documents, and any encountered errors.
	//
	// All tags need to exist on a device. If it is not true, the update action will fail.
	PublicKeyUpdateTags(ctx context.Context, tenant, fingerprint string, tags []string) (matchedCount int64, updatedCount int64, err error)

	PublicKeyRenameTag(ctx context.Context, tenant, currentTags, newTags string) (updatedCount int64, err error)
	PublicKeyDeleteTag(ctx context.Context, tenant, tag string) (updatedCount int64, err error)
	PublicKeyGetTags(ctx context.Context, tenant string) ([]string, int, error)
}
