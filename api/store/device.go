package store

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type DeviceListMode uint

const (
	DeviceListModeDefault DeviceListMode = iota + 1
	// DeviceListModeMaxDeviceReached is used to indicate to the DeviceList method that the namepsace's device maxium
	// number of devices has been reached and should set the "acceptable" value to true for devices that were recently
	// removed.
	DeviceListModeMaxDeviceReached
)

type DeviceStore interface {
	DeviceList(ctx context.Context, pagination paginator.Query, filters []models.Filter, status models.DeviceStatus, sort string, order string, mode DeviceListMode) ([]models.Device, int, error)
	DeviceGet(ctx context.Context, uid models.UID) (*models.Device, error)
	DeviceUpdate(ctx context.Context, tenant string, uid models.UID, name *string, publicURL *bool) error
	DeviceDelete(ctx context.Context, uid models.UID) error
	DeviceCreate(ctx context.Context, d models.Device, hostname string) error
	DeviceRename(ctx context.Context, uid models.UID, hostname string) error
	DeviceLookup(ctx context.Context, namespace, hostname string) (*models.Device, error)
	DeviceSetOnline(ctx context.Context, uid models.UID, timestamp time.Time, online bool) error
	DeviceUpdateOnline(ctx context.Context, uid models.UID, online bool) error
	DeviceUpdateLastSeen(ctx context.Context, uid models.UID, ts time.Time) error
	DeviceUpdateStatus(ctx context.Context, uid models.UID, status models.DeviceStatus) error
	DeviceGetByMac(ctx context.Context, mac string, tenantID string, status models.DeviceStatus) (*models.Device, error)
	DeviceGetByName(ctx context.Context, name string, tenantID string, status models.DeviceStatus) (*models.Device, error)
	DeviceGetByUID(ctx context.Context, uid models.UID, tenantID string) (*models.Device, error)
	DeviceSetPosition(ctx context.Context, uid models.UID, position models.DevicePosition) error
	DeviceListByUsage(ctx context.Context, tenantID string) ([]models.UID, error)
	DeviceChooser(ctx context.Context, tenantID string, chosen []string) error
	DeviceRemovedCount(ctx context.Context, tenant string) (int64, error)
	DeviceRemovedGet(ctx context.Context, tenant string, uid models.UID) (*models.DeviceRemoved, error)
	DeviceRemovedInsert(ctx context.Context, tenant string, device *models.Device) error
	DeviceRemovedDelete(ctx context.Context, tenant string, uid models.UID) error
	DeviceRemovedList(ctx context.Context, tenant string, pagination paginator.Query, filters []models.Filter, sort string, order string) ([]models.DeviceRemoved, int, error)
	DeviceCreatePublicURLAddress(ctx context.Context, uid models.UID) error
	DeviceGetByPublicURLAddress(ctx context.Context, address string) (*models.Device, error)

	// DeviceSwapAccepted swaps the old accepted device with the device identified by UID.
	// This function updates the status, transfers sessions, and deletes the old device.
	DeviceSwapAccepted(ctx context.Context, uid models.UID, oldDevice *models.Device) error
}
