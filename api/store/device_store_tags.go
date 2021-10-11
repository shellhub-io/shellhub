package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
)

type DeviceTagsStore interface {
	DeviceCreateTag(ctx context.Context, uid models.UID, tag string) error
	DeviceDeleteTag(ctx context.Context, uid models.UID, tag string) error
	DeviceRenameTag(ctx context.Context, tenantID string, currentTagName string, newTagName string) error
	DeviceListTag(ctx context.Context) ([]string, int, error)
	DeviceUpdateTag(ctx context.Context, uid models.UID, tags []string) error
	DeviceGetTags(ctx context.Context, tenantID string) ([]string, int, error)
	DeviceDeleteAllTags(ctx context.Context, tenantID string, tagName string) error
}
