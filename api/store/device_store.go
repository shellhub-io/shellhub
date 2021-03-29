package store

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type DeviceStore interface {
	ListDevices(ctx context.Context, pagination paginator.Query, filters []models.Filter, status string, sort string, order string) ([]models.Device, int, error)
	GetDevice(ctx context.Context, uid models.UID) (*models.Device, error)
	DeleteDevice(ctx context.Context, uid models.UID) error
	AddDevice(ctx context.Context, d models.Device, hostname string) error
	RenameDevice(ctx context.Context, uid models.UID, name string) error
	LookupDevice(ctx context.Context, namespace, name string) (*models.Device, error)
	UpdateDeviceStatus(ctx context.Context, uid models.UID, online bool) error
	UpdatePendingStatus(ctx context.Context, uid models.UID, status string) error
	GetDeviceByMac(ctx context.Context, mac, tenant, status string) (*models.Device, error)
	GetDeviceByName(ctx context.Context, name, tenant string) (*models.Device, error)
	GetDeviceByUID(ctx context.Context, uid models.UID, tenant string) (*models.Device, error)
}
