package store

import "context"

type FirewallTagsStore interface {
	// FirewallRulePushTag adds a new tag to the list of tags for a device with the specified UID.
	// Returns an error if any issues occur during the tag addition or ErrNoDocuments when matching documents are found.
	//
	// The tag need to exist on a device. If it is not true, the action will fail.
	FirewallRulePushTag(ctx context.Context, id, tag string) error

	// FirewallRulePullTag removes a tag from the list of tags for a device with the specified UID.
	// Returns an error if any issues occur during the tag removal or ErrNoDocuments when matching documents are found.
	//
	// To remove a tag, that tag needs to exist on a device. If it is not, the action will fail.
	FirewallRulePullTag(ctx context.Context, id, tag string) error

	FirewallRuleSetTags(ctx context.Context, id string, tags []string) error

	// FirewallRuleBulkRenameTag replaces all occurrences of the old tag with the new tag for all firewall rules belonging to the specified tenant.
	// Returns the number of documents updated and an error if any issues occur during the tag renaming.
	FirewallRuleBulkRenameTag(ctx context.Context, tenant, currentTag, newTag string) (updatedCount int64, err error)

	// FirewallRuleBulkDeleteTag removes a tag from all firewall rules belonging to the specified tenant.
	// Returns the number of documents updated and an error if any issues occur during the tag deletion.
	FirewallRuleBulkDeleteTag(ctx context.Context, tenant, tag string) (updatedCount int64, err error)

	// FirewallRuleGetTags retrieves all tags associated with the tenant.
	// Returns the tags, the number of tags, and an error if any issues occur.
	FirewallRuleGetTags(ctx context.Context, tenant string) (tag []string, n int, err error)
}
