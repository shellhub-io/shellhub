package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg/entity"
	"github.com/uptrace/bun"
)

// WebhookCreate implements [store.WebhookStore].
func (pg *Pg) WebhookCreate(ctx context.Context, webhook *models.Webhook) (string, error) {
	db := pg.GetConnection(ctx)

	webhook.CreatedAt = pg.clock.Now()
	webhook.UpdatedAt = webhook.CreatedAt

	e := entity.WebhookFromModel(webhook)
	if _, err := db.NewInsert().Model(e).Returning("id").Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	webhook.ID = e.ID

	return e.ID, nil
}

// WebhookLockNamespace implements [store.WebhookStore].
func (pg *Pg) WebhookLockNamespace(ctx context.Context, sc scope.Scope) error {
	tenantID, err := requireBounded(sc)
	if err != nil {
		return err
	}

	tx, ok := pg.GetConnection(ctx).(bun.Tx)
	if !ok {
		return store.ErrLockOutsideTransaction
	}

	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtextextended(?, 0))", "webhooks:"+tenantID); err != nil {
		return fromSQLError(err)
	}

	return nil
}

// WebhookList implements [store.WebhookStore].
func (pg *Pg) WebhookList(ctx context.Context, sc scope.Scope, opts ...store.QueryOption) ([]models.Webhook, int64, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.Webhook, 0)

	query, err := applyScopedOptions(ctx, db.NewSelect().Model(&entities).Order("created_at ASC", "id ASC"), sc, opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := ScanThenCount(ctx, query)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	webhooks := make([]models.Webhook, len(entities))
	for i := range entities {
		webhooks[i] = *entity.WebhookToModel(&entities[i])
	}

	return webhooks, count, nil
}

// WebhookResolve implements [store.WebhookStore].
func (pg *Pg) WebhookResolve(ctx context.Context, sc scope.Scope, id string) (*models.Webhook, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, store.ErrNoDocuments
	}

	webhook, err := resolveUnique[entity.Webhook](ctx, pg.GetConnection(ctx), sc, "id", id, nil)
	if err != nil {
		return nil, err
	}

	return entity.WebhookToModel(webhook), nil
}

// WebhookUpdate implements [store.WebhookStore].
func (pg *Pg) WebhookUpdate(ctx context.Context, webhook *models.Webhook) error {
	db := pg.GetConnection(ctx)

	webhook.UpdatedAt = pg.clock.Now()

	res, err := db.NewUpdate().
		Model(entity.WebhookFromModel(webhook)).
		Column("url", "events", "enabled", "updated_at").
		WherePK().
		Where("namespace_id = ?", webhook.TenantID).
		Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if affected, err := res.RowsAffected(); err != nil || affected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// WebhookDelete implements [store.WebhookStore].
func (pg *Pg) WebhookDelete(ctx context.Context, sc scope.Scope, id string) error {
	tenantID, err := requireBounded(sc)
	if err != nil {
		return err
	}

	if _, err := uuid.Parse(id); err != nil {
		return store.ErrNoDocuments
	}

	res, err := pg.GetConnection(ctx).NewDelete().
		Model((*entity.Webhook)(nil)).
		Where("id = ?", id).
		Where("namespace_id = ?", tenantID).
		Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if affected, err := res.RowsAffected(); err != nil || affected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}
