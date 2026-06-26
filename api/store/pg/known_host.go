package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

func (pg *Pg) KnownHostResolve(ctx context.Context, tenantID, ownerID, host string, port int) (*models.KnownHost, error) {
	db := pg.GetConnection(ctx)

	e := new(entity.KnownHost)
	q := db.NewSelect().
		Model(e).
		Where("namespace_id = ?", tenantID).
		Where("host = ?", host).
		Where("port = ?", port)

	if ownerID != "" {
		q = q.Where("owner_id = ?", ownerID)
	} else {
		q = q.Where("owner_id IS NULL")
	}

	if err := q.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.KnownHostToModel(e), nil
}

func (pg *Pg) KnownHostUpsert(ctx context.Context, knownHost *models.KnownHost) error {
	now := clock.Now()
	knownHost.ID = uuid.Generate()
	knownHost.CreatedAt = now
	knownHost.UpdatedAt = now

	// A single INSERT ... ON CONFLICT atomically inserts or replaces the stored
	// key, avoiding the resolve-then-write race on a concurrent first accept. The
	// scope picks which partial unique index arbitrates (personal vs team), since
	// only one applies to a given row.
	conflict := "CONFLICT (namespace_id, host, port) WHERE owner_id IS NULL"
	if knownHost.OwnerID != "" {
		conflict = "CONFLICT (namespace_id, owner_id, host, port) WHERE owner_id IS NOT NULL"
	}

	e := entity.KnownHostFromModel(knownHost)

	var result entity.KnownHost
	if err := pg.GetConnection(ctx).NewInsert().
		Model(e).
		On(conflict+" DO UPDATE").
		Set("key_type = EXCLUDED.key_type").
		Set("public_key = EXCLUDED.public_key").
		Set("fingerprint = EXCLUDED.fingerprint").
		Set("accepted_by = EXCLUDED.accepted_by").
		Set("updated_at = EXCLUDED.updated_at").
		Returning("id, created_at").
		Scan(ctx, &result); err != nil {
		return fromSQLError(err)
	}

	// On a conflict update the row keeps its original id/created_at; reflect the
	// stored values back onto the caller's model.
	knownHost.ID = result.ID
	knownHost.CreatedAt = result.CreatedAt

	return nil
}

func (pg *Pg) KnownHostDelete(ctx context.Context, tenantID, ownerID, host string, port int) error {
	db := pg.GetConnection(ctx)

	q := db.NewDelete().
		Model((*entity.KnownHost)(nil)).
		Where("namespace_id = ?", tenantID).
		Where("host = ?", host).
		Where("port = ?", port)

	if ownerID != "" {
		q = q.Where("owner_id = ?", ownerID)
	} else {
		q = q.Where("owner_id IS NULL")
	}

	r, err := q.Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if rows, err := r.RowsAffected(); err != nil || rows == 0 {
		return store.ErrNoDocuments
	}

	return nil
}
