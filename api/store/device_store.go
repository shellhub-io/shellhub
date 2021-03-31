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
	DeviceRename(ctx context.Context, uid models.UID, name string) error
	DeviceLookup(ctx context.Context, namespace, name string) (*models.Device, error)
	DeviceSetOnline(ctx context.Context, uid models.UID, online bool) error
	DeviceUpdateStatus(ctx context.Context, uid models.UID, status string) error
	DeviceGetByMac(ctx context.Context, mac, tenant, status string) (*models.Device, error)
	DeviceGetByName(ctx context.Context, name, tenant string) (*models.Device, error)
	DeviceGetByUID(ctx context.Context, uid models.UID, tenant string) (*models.Device, error)
}
