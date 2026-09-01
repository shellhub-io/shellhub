package pg

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg/entity"
	"github.com/uptrace/bun"
)

// GetStats implements [store.StatsStore].
func (pg *Pg) GetStats(ctx context.Context, sc scope.Scope) (*models.Stats, error) {
	db := pg.GetConnection(ctx)

	if !sc.IsValid() {
		return nil, store.ErrInvalidScope
	}

	onlineDevices, err := countInScope(ctx, buildOnlineDevicesQuery(db), sc)
	if err != nil {
		return nil, err
	}

	registeredDevices, err := countInScope(ctx, buildRegisteredDevicesQuery(db), sc)
	if err != nil {
		return nil, err
	}

	pendingDevices, err := countInScope(ctx, buildPendingDevicesQuery(db), sc)
	if err != nil {
		return nil, err
	}

	rejectedDevices, err := countInScope(ctx, buildRejectedDevicesQuery(db), sc)
	if err != nil {
		return nil, err
	}

	activeSessions, err := countInScope(context.WithValue(ctx, CtxTableAlias, "devices"), buildActiveSessionsQuery(db), sc)
	if err != nil {
		return nil, err
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

// CountRegisteredDevices implements [store.StatsStore].
func (pg *Pg) CountRegisteredDevices(ctx context.Context, sc scope.Scope) (int, error) {
	db := pg.GetConnection(ctx)

	if !sc.IsValid() {
		return 0, store.ErrInvalidScope
	}

	return countInScope(ctx, buildRegisteredDevicesQuery(db), sc)
}

func countInScope(ctx context.Context, query *bun.SelectQuery, sc scope.Scope) (int, error) {
	query, err := applyScopedOptions(ctx, query, sc)
	if err != nil {
		return 0, err
	}

	count, err := query.Count(ctx)
	if err != nil {
		return 0, fromSQLError(err)
	}

	return count, nil
}

func buildOnlineDevicesQuery(db bun.IDB) *bun.SelectQuery {
	return db.NewSelect().
		Model((*entity.Device)(nil)).
		Where("disconnected_at IS NULL").
		Where("last_seen > ?", clock.Now().Add(-2*time.Minute)).
		Where("status = ?", "accepted")
}

func buildRegisteredDevicesQuery(db bun.IDB) *bun.SelectQuery {
	return db.NewSelect().
		Model((*entity.Device)(nil)).
		Where("status = ?", "accepted")
}

func buildPendingDevicesQuery(db bun.IDB) *bun.SelectQuery {
	return db.NewSelect().
		Model((*entity.Device)(nil)).
		Where("status = ?", "pending")
}

func buildRejectedDevicesQuery(db bun.IDB) *bun.SelectQuery {
	return db.NewSelect().
		Model((*entity.Device)(nil)).
		Where("status = ?", "rejected")
}

func buildActiveSessionsQuery(db bun.IDB) *bun.SelectQuery {
	return db.NewSelect().
		Model((*entity.ActiveSession)(nil)).
		Join("JOIN sessions ON active_session.session_id = sessions.id").
		Join("JOIN devices ON sessions.device_id = devices.id")
}
