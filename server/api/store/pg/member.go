package pg

import (
	"context"
	"slices"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg/entity"
	"github.com/uptrace/bun"
)

// NamespaceCreateMembership implements [store.MemberStore].
func (pg *Pg) NamespaceCreateMembership(ctx context.Context, sc scope.Scope, membership *models.Member) error {
	db := pg.GetConnection(ctx)

	tenantID, err := requireBounded(sc)
	if err != nil {
		return err
	}

	membership.AddedAt = clock.Now()
	entity := entity.MembershipFromModel(tenantID, membership)
	if _, err := db.NewInsert().Model(entity).Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

// NamespaceUpdateMembership implements [store.MemberStore].
func (pg *Pg) NamespaceUpdateMembership(ctx context.Context, sc scope.Scope, member *models.Member) error {
	db := pg.GetConnection(ctx)

	tenantID, err := requireBounded(sc)
	if err != nil {
		return err
	}

	e := entity.MembershipFromModel(tenantID, member)
	e.UpdatedAt = clock.Now()
	r, err := db.NewUpdate().Model(e).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if count, err := r.RowsAffected(); err != nil || count == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// NamespaceDeleteMembership implements [store.MemberStore].
func (pg *Pg) NamespaceDeleteMembership(ctx context.Context, sc scope.Scope, member *models.Member) error {
	db := pg.GetConnection(ctx)

	tenantID, err := requireBounded(sc)
	if err != nil {
		return err
	}

	e := entity.MembershipFromModel(tenantID, member)
	r, err := db.NewDelete().Model(e).WherePK().Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if count, err := r.RowsAffected(); err != nil || count == 0 {
		return store.ErrNoDocuments
	}

	if _, err := db.NewUpdate().
		Model((*entity.User)(nil)).
		Set("preferred_namespace_id = NULL").
		Where("id = ? AND preferred_namespace_id = ?", member.ID, tenantID).
		Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

// NamespaceDepartMember implements [store.MemberStore].
func (pg *Pg) NamespaceDepartMember(ctx context.Context, sc scope.Scope, memberID string, departure store.MemberDeparture) (*models.MemberDeparted, error) {
	tenantID, err := requireBounded(sc)
	if err != nil {
		return nil, err
	}

	departed := &models.MemberDeparted{TenantID: tenantID, MemberID: memberID}

	depart := func(ctx context.Context) error {
		if err := pg.keepDepartingDevices(ctx, tenantID, memberID, departure.KeepDevices); err != nil {
			return err
		}

		removed, err := pg.removeOwnedDevices(ctx, sc, tenantID, memberID)
		if err != nil {
			return err
		}

		departed.RemovedDevices = removed

		if departed.APIKeyDigests, err = pg.APIKeyDeleteAllByCreator(ctx, tenantID, memberID); err != nil {
			return err
		}

		if departure.KeepMembership {
			return nil
		}

		return pg.NamespaceDeleteMembership(ctx, sc, &models.Member{ID: memberID})
	}

	if _, ok := ctx.Value(txKey).(bun.Tx); ok {
		err = depart(ctx)
	} else {
		err = pg.WithTransaction(ctx, depart)
	}

	if err != nil {
		return nil, err
	}

	return departed, nil
}

func (pg *Pg) keepDepartingDevices(ctx context.Context, tenantID, memberID string, uids []string) error {
	if len(uids) == 0 {
		return nil
	}

	var kept []string
	if err := pg.GetConnection(ctx).NewUpdate().
		Model((*entity.Device)(nil)).
		Set("owner_id = NULL").
		Where("namespace_id = ?", tenantID).
		Where("owner_id = ?", memberID).
		Where("status = ?", models.DeviceStatusAccepted).
		Where("id IN (?)", bun.List(uids)).
		Returning("id").
		Scan(ctx, &kept); err != nil {
		return fromSQLError(err)
	}

	for _, uid := range uids {
		if !slices.Contains(kept, uid) {
			return store.ErrDeviceNotOwned
		}
	}

	return nil
}

func (pg *Pg) removeOwnedDevices(ctx context.Context, sc scope.Scope, tenantID, memberID string) ([]string, error) {
	now := clock.Now()

	removed := []string{}
	if err := pg.GetConnection(ctx).NewUpdate().
		Model((*entity.Device)(nil)).
		Set("status = ?", models.DeviceStatusRemoved).
		Set("removed_at = ?", now).
		Set("updated_at = ?", now).
		Set("owner_id = NULL").
		Where("namespace_id = ?", tenantID).
		Where("owner_id = ?", memberID).
		Where("status = ?", models.DeviceStatusAccepted).
		Returning("id").
		Scan(ctx, &removed); err != nil {
		return nil, fromSQLError(err)
	}

	if _, err := pg.GetConnection(ctx).NewUpdate().
		Model((*entity.Device)(nil)).
		Set("owner_id = NULL").
		Where("namespace_id = ?", tenantID).
		Where("owner_id = ?", memberID).
		Exec(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	if len(removed) == 0 {
		return removed, nil
	}

	count := int64(len(removed))
	if err := pg.NamespaceIncrementDeviceCount(ctx, sc, models.DeviceStatusAccepted, -count); err != nil {
		return nil, err
	}

	if err := pg.NamespaceIncrementDeviceCount(ctx, sc, models.DeviceStatusRemoved, count); err != nil {
		return nil, err
	}

	return removed, nil
}
