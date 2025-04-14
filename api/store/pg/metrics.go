package pg

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/store/pg/internal/entity"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *pg) GetStats(ctx context.Context) (*models.Stats, error) {
	var tenantID string
	if tenant := gateway.TenantFromContext(ctx); tenant != nil {
		tenantID = tenant.ID
	}

	onlineDevices := 0
	query := pg.driver.NewSelect().
		Model((*entity.Device)(nil)).
		Where("disconnected_at IS NULL").
		Where("seen_at > ?", time.Now().Add(-2*time.Minute)).
		Where("status = ?", "accepted")

	if tenantID != "" {
		query = query.Where("namespace_id = (SELECT id FROM namespaces WHERE id = ?)", tenantID)
	}

	count, err := query.Count(ctx)
	if err != nil {
		return nil, fromSqlError(err)
	}
	onlineDevices = count

	registeredDevices := 0
	query = pg.driver.NewSelect().
		Model((*entity.Device)(nil)).
		Where("status = ?", "accepted")

	if tenantID != "" {
		query = query.Where("namespace_id = (SELECT id FROM namespaces WHERE id = ?)", tenantID)
	}

	count, err = query.Count(ctx)
	if err != nil {
		return nil, fromSqlError(err)
	}
	registeredDevices = count

	pendingDevices := 0
	query = pg.driver.NewSelect().
		Model((*entity.Device)(nil)).
		Where("status = ?", "pending")

	if tenantID != "" {
		query = query.Where("namespace_id = (SELECT id FROM namespaces WHERE id = ?)", tenantID)
	}

	count, err = query.Count(ctx)
	if err != nil {
		return nil, fromSqlError(err)
	}
	pendingDevices = count

	rejectedDevices := 0
	query = pg.driver.NewSelect().
		Model((*entity.Device)(nil)).
		Where("status = ?", "rejected")

	if tenantID != "" {
		query = query.Where("namespace_id = (SELECT id FROM namespaces WHERE id = ?)", tenantID)
	}

	count, err = query.Count(ctx)
	if err != nil {
		return nil, fromSqlError(err)
	}
	rejectedDevices = count

	// activeSessions := 0
	// query = pg.driver.NewSelect().
	// 	Model((*entity.ActiveSession)(nil))
	//
	// if tenantID != "" {
	// 	query = query.Where("tenant_id = ?", tenantID)
	// }
	//
	// count, err = query.Count(ctx)
	// if err != nil {
	// 	return nil, fromSqlError(err)
	// }
	// activeSessions = count

	return &models.Stats{
		RegisteredDevices: registeredDevices,
		OnlineDevices:     onlineDevices,
		PendingDevices:    pendingDevices,
		RejectedDevices:   rejectedDevices,
		ActiveSessions:    0,
	}, nil
}
