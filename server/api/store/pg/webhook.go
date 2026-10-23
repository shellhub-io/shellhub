package pg

import (
	"context"
	"encoding/json"
	"time"

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

// WebhookEventRecord implements [store.WebhookStore].
func (pg *Pg) WebhookEventRecord(ctx context.Context, event *models.WebhookEvent) (int64, error) {
	res, err := pg.GetConnection(ctx).NewRaw(`
		WITH targets AS (
			SELECT id FROM webhooks WHERE namespace_id = ?0 AND enabled AND ?1 = ANY(events)
		), recorded AS (
			INSERT INTO webhook_events (id, namespace_id, type, data, created_at)
			SELECT ?2, ?0, ?1, ?3, ?4 WHERE EXISTS (SELECT 1 FROM targets)
			RETURNING id
		)
		INSERT INTO webhook_deliveries (event_id, webhook_id, namespace_id, next_attempt_at, created_at, updated_at)
		SELECT recorded.id, targets.id, ?0, ?4, ?4, ?4 FROM recorded, targets`,
		event.TenantID, string(event.Type), event.ID, string(event.Data), event.CreatedAt,
	).Exec(ctx)
	if err != nil {
		return 0, fromSQLError(err)
	}

	return res.RowsAffected()
}

type webhookDispatchRow struct {
	entity.WebhookDelivery `bun:",extend"`

	Data            json.RawMessage `bun:"data"`
	EventCreatedAt  time.Time       `bun:"event_created_at"`
	URL             string          `bun:"url"`
	SecretEncrypted string          `bun:"secret_encrypted"`
}

// WebhookDeliveryClaim implements [store.WebhookStore].
func (pg *Pg) WebhookDeliveryClaim(ctx context.Context, now, leaseUntil time.Time, limit int) ([]store.WebhookDispatch, error) {
	rows := make([]webhookDispatchRow, 0, limit)

	if err := pg.GetConnection(ctx).NewRaw(`
		UPDATE webhook_deliveries AS delivery
		SET lease_until = ?1, updated_at = ?0
		FROM webhook_events AS event, webhooks AS webhook
		WHERE delivery.id IN (
			SELECT due.id FROM webhook_deliveries AS due
			JOIN webhooks AS subscriber ON subscriber.id = due.webhook_id
			WHERE due.status = 'pending' AND due.next_attempt_at <= ?0 AND (due.lease_until IS NULL OR due.lease_until <= ?0)
				AND subscriber.enabled
			ORDER BY due.next_attempt_at
			LIMIT ?2
			FOR UPDATE OF due SKIP LOCKED
		)
		AND event.id = delivery.event_id
		AND webhook.id = delivery.webhook_id
		RETURNING delivery.*, event.type AS event_type, event.data, event.created_at AS event_created_at,
			webhook.url, webhook.secret_encrypted`,
		now, leaseUntil, limit,
	).Scan(ctx, &rows); err != nil {
		return nil, fromSQLError(err)
	}

	dispatches := make([]store.WebhookDispatch, len(rows))
	for i := range rows {
		delivery := entity.WebhookDeliveryToModel(&rows[i].WebhookDelivery)
		dispatches[i] = store.WebhookDispatch{
			Delivery: *delivery,
			Event: models.WebhookEvent{
				ID:        delivery.EventID,
				TenantID:  delivery.TenantID,
				Type:      delivery.EventType,
				Data:      rows[i].Data,
				CreatedAt: rows[i].EventCreatedAt,
			},
			URL:             rows[i].URL,
			SecretEncrypted: rows[i].SecretEncrypted,
		}
	}

	return dispatches, nil
}

// WebhookDeliveryFinish implements [store.WebhookStore].
func (pg *Pg) WebhookDeliveryFinish(ctx context.Context, delivery *models.WebhookDelivery) error {
	delivery.UpdatedAt = pg.clock.Now()

	res, err := pg.GetConnection(ctx).NewUpdate().
		Model(entity.WebhookDeliveryFromModel(delivery)).
		Column("status", "attempts", "next_attempt_at", "last_status_code", "last_error", "delivered_at", "lease_until", "updated_at").
		WherePK().
		Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	if affected, err := res.RowsAffected(); err != nil || affected == 0 {
		return store.ErrNoDocuments
	}

	return nil
}

// WebhookDeliveryList implements [store.WebhookStore].
func (pg *Pg) WebhookDeliveryList(ctx context.Context, sc scope.Scope, webhookID string, opts ...store.QueryOption) ([]models.WebhookDelivery, int64, error) {
	if _, err := uuid.Parse(webhookID); err != nil {
		return []models.WebhookDelivery{}, 0, nil //nolint:nilerr // an ID that is not a UUID names no webhook, so it has no deliveries
	}

	entities := make([]entity.WebhookDelivery, 0)

	ctx = context.WithValue(ctx, CtxTableAlias, "webhook_delivery")
	query := pg.GetConnection(ctx).NewSelect().
		Model(&entities).
		ColumnExpr("webhook_delivery.*").
		ColumnExpr("webhook_event.type AS event_type").
		Join("JOIN webhook_events AS webhook_event ON webhook_event.id = webhook_delivery.event_id").
		Where("webhook_delivery.webhook_id = ?", webhookID).
		Order("webhook_delivery.created_at DESC", "webhook_delivery.id DESC")

	query, err := applyScopedOptions(ctx, query, sc, opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := ScanThenCount(ctx, query)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	deliveries := make([]models.WebhookDelivery, len(entities))
	for i := range entities {
		deliveries[i] = *entity.WebhookDeliveryToModel(&entities[i])
	}

	return deliveries, count, nil
}

// WebhookDeliveryPrune implements [store.WebhookStore].
func (pg *Pg) WebhookDeliveryPrune(ctx context.Context, before time.Time) (int64, error) {
	db := pg.GetConnection(ctx)

	res, err := db.NewDelete().
		Model((*entity.WebhookDelivery)(nil)).
		Where("status IN (?)", bun.List([]string{string(models.WebhookDeliveryDelivered), string(models.WebhookDeliveryFailed)})).
		Where("updated_at < ?", before).
		Exec(ctx)
	if err != nil {
		return 0, fromSQLError(err)
	}

	if _, err := db.NewDelete().
		Model((*entity.WebhookEvent)(nil)).
		Where("created_at < ?", before).
		Where("NOT EXISTS (SELECT 1 FROM webhook_deliveries WHERE webhook_deliveries.event_id = webhook_event.id)").
		Exec(ctx); err != nil {
		return 0, fromSQLError(err)
	}

	return res.RowsAffected()
}
