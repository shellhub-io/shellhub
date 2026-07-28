package pg

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg/entity"
	"github.com/uptrace/bun"
)

func (pg *Pg) GetStats(ctx context.Context, sc scope.Scope) (*models.Stats, error) {
	db := pg.GetConnection(ctx)

	if !sc.IsValid() {
		return nil, store.ErrInvalidScope
	}

	onlineDevicesQuery := buildOnlineDevicesQuery(db, sc)
	onlineDevices, err := onlineDevicesQuery.Count(ctx)
	if err != nil {
		return nil, fromSQLError(err)
	}

	registeredDevicesQuery := buildRegisteredDevicesQuery(db, sc)
	registeredDevices, err := registeredDevicesQuery.Count(ctx)
	if err != nil {
		return nil, fromSQLError(err)
	}

	pendingDevicesQuery := buildPendingDevicesQuery(db, sc)
	pendingDevices, err := pendingDevicesQuery.Count(ctx)
	if err != nil {
		return nil, fromSQLError(err)
	}

	rejectedDevicesQuery := buildRejectedDevicesQuery(db, sc)
	rejectedDevices, err := rejectedDevicesQuery.Count(ctx)
	if err != nil {
		return nil, fromSQLError(err)
	}

	activeSessionsQuery := buildActiveSessionsQuery(db, sc)
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

func buildOnlineDevicesQuery(db bun.IDB, sc scope.Scope) *bun.SelectQuery {
	query := db.NewSelect().
		Model((*entity.Device)(nil)).
		Where("disconnected_at IS NULL").
		Where("last_seen > ?", time.Now().Add(-2*time.Minute)).
		Where("status = ?", "accepted")

	if sc.IsBounded() {
		query = query.Where("namespace_id = (SELECT id FROM namespaces WHERE id = ?)", sc.TenantID())
	}

	return query
}

func buildRegisteredDevicesQuery(db bun.IDB, sc scope.Scope) *bun.SelectQuery {
	query := db.NewSelect().
		Model((*entity.Device)(nil)).
		Where("status = ?", "accepted")

	if sc.IsBounded() {
		query = query.Where("namespace_id = (SELECT id FROM namespaces WHERE id = ?)", sc.TenantID())
	}

	return query
}

func buildPendingDevicesQuery(db bun.IDB, sc scope.Scope) *bun.SelectQuery {
	query := db.NewSelect().
		Model((*entity.Device)(nil)).
		Where("status = ?", "pending")

	if sc.IsBounded() {
		query = query.Where("namespace_id = (SELECT id FROM namespaces WHERE id = ?)", sc.TenantID())
	}

	return query
}

func buildRejectedDevicesQuery(db bun.IDB, sc scope.Scope) *bun.SelectQuery {
	query := db.NewSelect().
		Model((*entity.Device)(nil)).
		Where("status = ?", "rejected")

	if sc.IsBounded() {
		query = query.Where("namespace_id = (SELECT id FROM namespaces WHERE id = ?)", sc.TenantID())
	}

	return query
}

func buildActiveSessionsQuery(db bun.IDB, sc scope.Scope) *bun.SelectQuery {
	query := db.NewSelect().
		Model((*entity.ActiveSession)(nil)).
		Join("JOIN sessions ON active_session.session_id = sessions.id").
		Join("JOIN devices ON sessions.device_id = devices.id")

	if sc.IsBounded() {
		query = query.Where("devices.namespace_id = (SELECT id FROM namespaces WHERE id = ?)", sc.TenantID())
	}

	return query
}
