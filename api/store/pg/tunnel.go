package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
)

func (pg *Pg) TunnelUpdateDeviceUID(ctx context.Context, tenantID, oldUID, newUID string) error {
	db := pg.GetConnection(ctx)

	_, err := db.NewUpdate().
		Model((*entity.Tunnel)(nil)).
		Set("device_id = ?", newUID).
		Set("updated_at = ?", clock.Now()).
		Where("namespace_id = ?", tenantID).
		Where("device_id = ?", oldUID).
		Exec(ctx)

	return fromSQLError(err)
}
