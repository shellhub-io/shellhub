package pg

import (
	"context" //nolint:gosec

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *pg) DeviceList(ctx context.Context, status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter, acceptable store.DeviceAcceptable) ([]models.Device, int, error) {
	return nil, 0, nil
}

func (pg *pg) DeviceGet(ctx context.Context, uid models.UID) (*models.Device, error) {
	return nil, nil
}

func (pg *pg) DeviceDelete(ctx context.Context, uid models.UID) error {
	return nil
}

func (pg *pg) DeviceCreate(ctx context.Context, d models.Device, hostname string) error {
	return nil
}

func (pg *pg) DeviceRename(ctx context.Context, uid models.UID, hostname string) error {
	return nil
}

func (pg *pg) DeviceLookup(ctx context.Context, namespace, hostname string) (*models.Device, error) {
	return nil, nil
}

// DeviceUpdateStatus updates the status of a specific device in the devices collection
func (pg *pg) DeviceUpdateStatus(ctx context.Context, uid models.UID, status models.DeviceStatus) error {
	return nil
}

func (pg *pg) DeviceListByUsage(ctx context.Context, tenant string) ([]models.UID, error) {
	return nil, nil
}

func (pg *pg) DeviceGetByMac(ctx context.Context, mac string, tenantID string, status models.DeviceStatus) (*models.Device, error) {
	return nil, nil
}

func (pg *pg) DeviceGetByName(ctx context.Context, name string, tenantID string, status models.DeviceStatus) (*models.Device, error) {
	return nil, nil
}

func (pg *pg) DeviceGetByUID(ctx context.Context, uid models.UID, tenantID string) (*models.Device, error) {
	return nil, nil
}

func (pg *pg) DeviceSetPosition(ctx context.Context, uid models.UID, position models.DevicePosition) error {
	return nil
}

func (pg *pg) DeviceChooser(ctx context.Context, tenantID string, chosen []string) error {
	return nil
}

func (pg *pg) DeviceConflicts(ctx context.Context, target *models.DeviceConflicts) ([]string, bool, error) {
	return nil, false, nil
}

func (pg *pg) DeviceUpdate(ctx context.Context, tenantID, uid string, changepg *models.DeviceChanges) error {
	return nil
}

func (pg *pg) DeviceBulkUpdate(ctx context.Context, uids []string, changepg *models.DeviceChanges) (int64, error) {
	return int64(0), nil
}

func (pg *pg) DeviceRemovedCount(ctx context.Context, tenant string) (int64, error) {
	return int64(0), nil
}

func (pg *pg) DeviceRemovedGet(ctx context.Context, tenant string, uid models.UID) (*models.DeviceRemoved, error) {
	return nil, nil
}

func (pg *pg) DeviceRemovedInsert(ctx context.Context, tenant string, device *models.Device) error { //nolint:revive
	return nil
}

func (pg *pg) DeviceRemovedDelete(ctx context.Context, tenant string, uid models.UID) error {
	return nil
}

func (pg *pg) DeviceRemovedList(ctx context.Context, tenant string, paginator query.Paginator, filters query.Filters, sorter query.Sorter) ([]models.DeviceRemoved, int, error) {
	return nil, 0, nil
}

func (pg *pg) DevicePushTag(ctx context.Context, uid models.UID, tag string) error {
	return nil
}

func (pg *pg) DevicePullTag(ctx context.Context, uid models.UID, tag string) error {
	return nil
}

func (pg *pg) DeviceSetTags(ctx context.Context, uid models.UID, tags []string) (int64, int64, error) {
	return int64(0), int64(0), nil
}

func (pg *pg) DeviceBulkRenameTag(ctx context.Context, tenant, currentTag, newTag string) (int64, error) {
	return int64(0), nil
}

func (pg *pg) DeviceBulkDeleteTag(ctx context.Context, tenant, tag string) (int64, error) {
	return int64(0), nil
}

func (pg *pg) DeviceGetTags(ctx context.Context, tenant string) ([]string, int, error) {
	return nil, 0, nil
}
