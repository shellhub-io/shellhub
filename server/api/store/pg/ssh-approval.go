package pg

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store/pg/entity"
)

func (pg *Pg) SSHApprovalCreate(ctx context.Context, approval *models.SSHApproval) error {
	db := pg.GetConnection(ctx)

	if _, err := db.NewInsert().Model(entity.SSHApprovalFromModel(approval)).Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

func (pg *Pg) SSHApprovalGet(ctx context.Context, code string, now time.Time) (*models.SSHApproval, error) {
	db := pg.GetConnection(ctx)

	approval := new(entity.SSHApproval)
	if err := db.NewSelect().
		Model(approval).
		Where("code = ?", code).
		Where("expires_at > ?", now).
		Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.SSHApprovalToModel(approval), nil
}

func (pg *Pg) SSHApprovalDecide(ctx context.Context, code string, state models.SSHApprovalState, userID string, now time.Time) (bool, error) {
	db := pg.GetConnection(ctx)

	// Atomic claim: only the first caller moves the row out of pending, so the
	// affected-row count decides who wins. An already-decided or expired approval
	// matches no row and returns false, which the service reports as not found.
	res, err := db.NewUpdate().
		Model((*entity.SSHApproval)(nil)).
		Set("state = ?", string(state)).
		Set("decided_by = ?", userID).
		Where("code = ?", code).
		Where("state = ?", string(models.SSHApprovalPending)).
		Where("expires_at > ?", now).
		Exec(ctx)
	if err != nil {
		return false, fromSQLError(err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, fromSQLError(err)
	}

	return affected == 1, nil
}

func (pg *Pg) SSHApprovalCleanup(ctx context.Context, before time.Time) (int64, error) {
	db := pg.GetConnection(ctx)

	res, err := db.NewDelete().
		Model((*entity.SSHApproval)(nil)).
		Where("expires_at < ?", before).
		Exec(ctx)
	if err != nil {
		return 0, fromSQLError(err)
	}

	return res.RowsAffected()
}
