package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type DeviceStore interface {
	DeviceList(ctx context.Context, pagination paginator.Query, filters []models.Filter, status string, sort string, order string) ([]models.Device, int, error)
	DeviceGet(ctx context.Context, uid models.UID) (*models.Device, error)
	DeviceDelete(ctx context.Context, uid models.UID) error
	DeviceCreate(ctx context.Context, d models.Device, hostname string) error
	DeviceRename(ctx context.Context, uid models.UID, hostname string) error
	DeviceLookup(ctx context.Context, namespace, hostname string) (*models.Device, error)
	DeviceSetOnline(ctx context.Context, uid models.UID, online bool) error
	DeviceUpdateStatus(ctx context.Context, uid models.UID, status string) error
	DeviceGetByMac(ctx context.Context, mac string, tenantID string, status string) (*models.Device, error)
	DeviceGetByName(ctx context.Context, name string, tenantID string) (*models.Device, error)
	DeviceGetByUID(ctx context.Context, uid models.UID, tenantID string) (*models.Device, error)
	DeviceSetPosition(ctx context.Context, uid models.UID, position models.DevicePosition) error
	DeviceListByUsage(ctx context.Context, tenantID string) ([]models.Device, error)
	DeviceChoice(ctx context.Context, tenantID string, chosen []string) error
	DeviceCreateTag(ctx context.Context, uid models.UID, tag string) error
	DeviceDeleteTag(ctx context.Context, uid models.UID, tag string) error
	DeviceRenameTag(ctx context.Context, tenantID string, currentTagName string, newTagName string) error
	DeviceListTag(ctx context.Context) ([]string, int, error)
	DeviceUpdateTag(ctx context.Context, uid models.UID, tags []string) error
	DeviceGetTags(ctx context.Context, tenantID string) ([]string, int, error)
	DeviceDeleteAllTags(ctx context.Context, tenantID string, tagName string) error
}
