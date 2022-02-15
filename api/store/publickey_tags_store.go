package store

import "context"

type PublicKeyTagsStore interface {
	PublicKeyAddTag(ctx context.Context, tenant, fingerprint, tag string) error
	PublicKeyRemoveTag(ctx context.Context, tenant, fingerprint, tag string) error
	PublicKeyUpdateTags(ctx context.Context, tenant, fingerprint string, tags []string) error
}
