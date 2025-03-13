package pg

import (
	"context"
	"time"

	"github.com/lib/pq"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (s *Store) DeviceCreate(ctx context.Context, d models.Device, hostname string) error {
	if r := s.db.Save(&d); r.Error != nil {
		return r.Error
	}

	log.WithField("id", d.ID).Info("created device")

	return nil
}

func (s *Store) DeviceList(ctx context.Context, status models.DeviceStatus, pagination query.Paginator, filters query.Filters, sorter query.Sorter, acceptable store.DeviceAcceptable) ([]models.Device, int, error) {
	return nil, 0, nil
}

func (s *Store) DeviceListByUsage(ctx context.Context, tenantID string) ([]models.UID, error) {
	// TODO: this can be a filter in the DeviceList method
	return nil, nil
}

func (s *Store) DeviceGet(ctx context.Context, bar, foo, namespace string) (*models.Device, error) {
	d := new(models.Device)

	query := new(gorm.DB)
	switch bar {
	case "uid":
		query = s.db.Where("id = ?", foo)
	case "name":
		query = s.db.Where("devices.name = ?", foo)
	}

	if namespace != "" {
		query = query.Joins("JOIN namespaces ON devices.namespace_id = namespaces.id").Where("namespaces.name = ?", namespace)
	}

	r := query.First(d)

	return d, r.Error
}

func (s *Store) DeviceGetByMac(ctx context.Context, mac string, tenantID string, status models.DeviceStatus) (*models.Device, error) {
	// TODO: evaluate if this can be implemented in the DeviceGet method
	return nil, nil
}

func (s *Store) DeviceGetByName(ctx context.Context, name string, tenantID string, status models.DeviceStatus) (*models.Device, error) {
	// TODO: evaluate if this can be implemented in the DeviceGet method
	return nil, nil
}

func (s *Store) DeviceGetByUID(ctx context.Context, uid models.UID, tenantID string) (*models.Device, error) {
	// TODO: evaluate if this can be implemented in the DeviceGet method
	return nil, nil
}

func (s *Store) DeviceGetByPublicURLAddress(ctx context.Context, address string) (*models.Device, error) {
	// TODO: this is not used anymored and/or evaluate if this can be implemented in the DeviceGet metho
	return nil, nil
}

func (s *Store) DeviceLookup(ctx context.Context, namespace, hostname string) (*models.Device, error) {
	// TODO: evaluate if this can be implemented in the DeviceGet method
	return nil, nil
}

func (s *Store) DeviceUpdate(ctx context.Context, tenant string, uid models.UID, name *string, publicURL *bool) error {
	return nil
}

func (s *Store) DeviceSetOnline(ctx context.Context, connectedDevices []models.ConnectedDevice) error {
	query := `
	UPDATE "devices" 
	  SET last_seen = $2
	  FROM 
		(select unnest($1::bytea[]) as id) as data_table
	  WHERE "devices".id = data_table.id;
	`

	f := make([]string, 0)
	for _, b := range connectedDevices {
		f = append(f, b.UID)
	}

	return s.db.Exec(query, pq.Array(f), time.Now()).Error
}

func (s *Store) DeviceSetOffline(ctx context.Context, uid string) error {
	// TODO: this must be implemented in the DeviceUpdate method
	return nil
}

func (s *Store) DeviceRename(ctx context.Context, uid models.UID, hostname string) error {
	// TODO: this must be implemented in the DeviceUpdate method
	return nil
}

func (s *Store) DeviceUpdateOnline(ctx context.Context, uid models.UID, online bool) error {
	// TODO: this must be implemented in the DeviceUpdate method
	return nil
}

func (s *Store) DeviceUpdateLastSeen(ctx context.Context, uid models.UID, ts time.Time) error {
	// TODO: this must be implemented in the DeviceUpdate method
	return nil
}

func (s *Store) DeviceUpdateStatus(ctx context.Context, uid models.UID, status models.DeviceStatus) error {
	// TODO: this must be implemented in the DeviceUpdate method
	return nil
}

func (s *Store) DeviceCreatePublicURLAddress(ctx context.Context, uid models.UID) error {
	// TODO: this is not used anymored and/or must be implemented in the DeviceUpdate method
	return nil
}

func (s *Store) DeviceSetPosition(ctx context.Context, uid models.UID, position models.DevicePosition) error {
	// TODO: this must be implemented in the DeviceUpdate method
	return nil
}

func (s *Store) DeviceDelete(ctx context.Context, uid models.UID) error {
	return nil
}

func (s *Store) DeviceChooser(ctx context.Context, tenantID string, chosen []string) error {
	// TODO: this is not used anymore
	return nil
}

func (s *Store) DeviceRemovedCount(ctx context.Context, tenant string) (int64, error) {
	// TODO: removed devices are currently saved in another collection. we need to understand why and if it can be changed.
	return 0, nil
}

func (s *Store) DeviceRemovedList(ctx context.Context, tenant string, pagination query.Paginator, filters query.Filters, sorter query.Sorter) ([]models.DeviceRemoved, int, error) {
	// TODO: removed devices are currently saved in another collection. we need to understand why and if it can be changed.
	return nil, 0, nil
}

func (s *Store) DeviceRemovedGet(ctx context.Context, tenant string, uid models.UID) (*models.DeviceRemoved, error) {
	// TODO: removed devices are currently saved in another collection. we need to understand why and if it can be changed.
	return nil, nil
}

func (s *Store) DeviceRemovedInsert(ctx context.Context, tenant string, device *models.Device) error {
	// TODO: removed devices are currently saved in another collection. we need to understand why and if it can be changed.
	return nil
}

func (s *Store) DeviceRemovedDelete(ctx context.Context, tenant string, uid models.UID) error {
	// TODO: removed devices are currently saved in another collection. we need to understand why and if it can be changed.
	return nil
}

func (s *Store) DeviceGetTags(ctx context.Context, tenant string) (tag []string, n int, err error) {
	// TODO: with the gorm, this must be implemented as a polymorphism association
	return nil, 0, nil
}

func (s *Store) DevicePushTag(ctx context.Context, uid models.UID, tag string) error {
	// TODO: with the gorm, this must be implemented as a polymorphism association
	return nil
}

func (s *Store) DevicePullTag(ctx context.Context, uid models.UID, tag string) error {
	// TODO: with the gorm, this must be implemented as a polymorphism association
	return nil
}

func (s *Store) DeviceSetTags(ctx context.Context, uid models.UID, tags []string) (matchedCount int64, updatedCount int64, err error) {
	// TODO: with the gorm, this must be implemented as a polymorphism association
	return 0, 0, nil
}

func (s *Store) DeviceBulkRenameTag(ctx context.Context, tenant, currentTag, newTag string) (updatedCount int64, err error) {
	// TODO: with the gorm, this must be implemented as a polymorphism association
	return 0, nil
}

func (s *Store) DeviceBulkDeleteTag(ctx context.Context, tenant, tag string) (deletedCount int64, err error) {
	// TODO: with the gorm, this must be implemented as a polymorphism association
	return 0, nil
}
