package store

import "context"

type FirewallTagsStore interface {
	// FirewallRuleAddTag adds a new tag to the list of tags for a device with the specified UID.
	// Returns an error if any issues occur during the tag addition or ErrNoDocuments when matching documents are found.
	//
	// The tag need to exist on a device. If it is not true, the action will fail.
	FirewallRuleAddTag(ctx context.Context, id, tag string) error

	// FirewallRuleRemoveTag removes a tag from the list of tags for a device with the specified UID.
	// Returns an error if any issues occur during the tag removal or ErrNoDocuments when matching documents are found.
	//
	// To remove a tag, that tag needs to exist on a device. If it is not, the action will fail.
	FirewallRuleRemoveTag(ctx context.Context, id, tag string) error

	FirewallRuleUpdateTags(ctx context.Context, id string, tags []string) error

	// FirewallRuleRenameTag replaces all occurrences of the old tag with the new tag for all firewall rules belonging to the specified tenant.
	// Returns the number of documents updated and an error if any issues occur during the tag renaming.
	FirewallRuleRenameTag(ctx context.Context, tenant, currentTags, newTags string) (updatedCount int64, err error)

	// FirewallRuleDeleteTag removes a tag from all firewall rules belonging to the specified tenant.
	// Returns the number of documents updated and an error if any issues occur during the tag deletion.
	FirewallRuleDeleteTag(ctx context.Context, tenant, tag string) (updatedCount int64, err error)

	// FirewallRuleGetTags retrieves all tags associated with the tenant.
	// Returns the tags, the number of tags, and an error if any issues occur.
	FirewallRuleGetTags(ctx context.Context, tenant string) (tag []string, n int, err error)
}
