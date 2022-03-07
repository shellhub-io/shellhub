package store

import "context"

type FirewallTagsStore interface {
	FirewallRuleAddTag(ctx context.Context, id, tag string) error
	FirewallRuleRemoveTag(ctx context.Context, id, tag string) error
	FirewallRuleUpdateTags(ctx context.Context, id string, tags []string) error
	FirewallRuleRenameTag(ctx context.Context, tenant, tagCurrent, tagNew string) error
	FirewallRuleDeleteTag(ctx context.Context, tenant, tag string) error
	FirewallRuleGetTags(ctx context.Context, tenant string) ([]string, int, error)
}
