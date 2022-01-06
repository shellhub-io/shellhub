package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type DeviceTagsStore interface {
	DeviceCreateTag(ctx context.Context, uid models.UID, tag string) error
	DeviceRemoveTag(ctx context.Context, uid models.UID, tag string) error
	DeviceRenameTag(ctx context.Context, tenantID string, currentTagName string, newTagName string) error
	DeviceUpdateTag(ctx context.Context, uid models.UID, tags []string) error
	DeviceGetTags(ctx context.Context, tenantID string) ([]string, int, error)
	DeviceDeleteTags(ctx context.Context, tenantID string, tagName string) error
}
