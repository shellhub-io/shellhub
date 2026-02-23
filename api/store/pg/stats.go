package pg

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

func (pg *Pg) GetStats(ctx context.Context, tenantID string) (*models.Stats, error) {
	db := pg.GetConnection(ctx)

	onlineDevicesQuery := buildOnlineDevicesQuery(db, tenantID)
	onlineDevices, err := onlineDevicesQuery.Count(ctx)
	if err != nil {
		return nil, fromSQLError(err)
	}

	registeredDevicesQuery := buildRegisteredDevicesQuery(db, tenantID)
	registeredDevices, err := registeredDevicesQuery.Count(ctx)
	if err != nil {
		return nil, fromSQLError(err)
	}

	pendingDevicesQuery := buildPendingDevicesQuery(db, tenantID)
	pendingDevices, err := pendingDevicesQuery.Count(ctx)
	if err != nil {
		return nil, fromSQLError(err)
	}

	rejectedDevicesQuery := buildRejectedDevicesQuery(db, tenantID)
	rejectedDevices, err := rejectedDevicesQuery.Count(ctx)
	if err != nil {
		return nil, fromSQLError(err)
	}

	activeSessionsQuery := buildActiveSessionsQuery(db, tenantID)
	activeSessions, err := activeSessionsQuery.Count(ctx)
	if err != nil {
		return nil, fromSQLError(err)
	}

	stats := &models.Stats{
		RegisteredDevices: registeredDevices,
		OnlineDevices:     onlineDevices,
		PendingDevices:    pendingDevices,
		RejectedDevices:   rejectedDevices,
		ActiveSessions:    activeSessions,
	}

	return stats, nil
}

func buildOnlineDevicesQuery(db bun.IDB, tenantID string) *bun.SelectQuery {
	query := db.NewSelect().
		Model((*entity.Device)(nil)).
		Where("disconnected_at IS NULL").
		Where("last_seen > ?", time.Now().Add(-2*time.Minute)).
		Where("status = ?", "accepted")

	if tenantID != "" {
		query = query.Where("namespace_id = (SELECT id FROM namespaces WHERE id = ?)", tenantID)
	}

	return query
}

func buildRegisteredDevicesQuery(db bun.IDB, tenantID string) *bun.SelectQuery {
	query := db.NewSelect().
		Model((*entity.Device)(nil)).
		Where("status = ?", "accepted")

	if tenantID != "" {
		query = query.Where("namespace_id = (SELECT id FROM namespaces WHERE id = ?)", tenantID)
	}

	return query
}

func buildPendingDevicesQuery(db bun.IDB, tenantID string) *bun.SelectQuery {
	query := db.NewSelect().
		Model((*entity.Device)(nil)).
		Where("status = ?", "pending")

	if tenantID != "" {
		query = query.Where("namespace_id = (SELECT id FROM namespaces WHERE id = ?)", tenantID)
	}

	return query
}

func buildRejectedDevicesQuery(db bun.IDB, tenantID string) *bun.SelectQuery {
	query := db.NewSelect().
		Model((*entity.Device)(nil)).
		Where("status = ?", "rejected")

	if tenantID != "" {
		query = query.Where("namespace_id = (SELECT id FROM namespaces WHERE id = ?)", tenantID)
	}

	return query
}

func buildActiveSessionsQuery(db bun.IDB, tenantID string) *bun.SelectQuery {
	query := db.NewSelect().
		Model((*entity.ActiveSession)(nil)).
		Join("JOIN sessions ON active_session.session_id = sessions.id").
		Join("JOIN devices ON sessions.device_id = devices.id")

	if tenantID != "" {
		query = query.Where("devices.namespace_id = (SELECT id FROM namespaces WHERE id = ?)", tenantID)
	}

	return query
}
