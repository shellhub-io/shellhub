package pg

import (
	"context"

	"github.com/google/uuid"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (pg *Pg) SessionList(ctx context.Context, opts ...store.QueryOption) ([]models.Session, int, error) {
	db := pg.getExecutor(ctx)

	entities := make([]entity.Session, 0)
	query := db.NewSelect().
		Model(&entities).
		Relation("Device").
		Relation("Device.Namespace")

	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, 0, fromSqlError(err)
	}

	count, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fromSqlError(err)
	}

	query = db.NewSelect().
		Model(&entities).
		Relation("Device").
		Relation("Device.Namespace").
		ColumnExpr("sessions.*, CASE WHEN active_sessions.session_id IS NOT NULL THEN true ELSE false END as active").
		Join("LEFT JOIN active_sessions ON sessions.id = active_sessions.session_id")

	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, 0, fromSqlError(err)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, 0, fromSqlError(err)
	}

	sessions := make([]models.Session, len(entities))
	for i, e := range entities {
		sessions[i] = *entity.SessionToModel(&e)
	}

	return sessions, count, nil
}

func (pg *Pg) SessionResolve(ctx context.Context, resolver store.SessionResolver, value string, opts ...store.QueryOption) (*models.Session, error) {
	db := pg.getExecutor(ctx)

	var sessionID string
	switch resolver {
	case store.SessionUIDResolver:
		sessionID = value
	default:
		return nil, store.ErrNoDocuments
	}

	e := &entity.Session{}
	query := db.NewSelect().
		Model(e).
		Relation("Device").
		Relation("Device.Namespace").
		ColumnExpr("sessions.*, CASE WHEN active_sessions.session_id IS NOT NULL THEN true ELSE false END as active").
		Join("LEFT JOIN active_sessions ON sessions.id = active_sessions.session_id").
		Where("sessions.id = ?", sessionID)

	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, fromSqlError(err)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return entity.SessionToModel(e), nil
}

func (pg *Pg) SessionCreate(ctx context.Context, session models.Session) (string, error) {
	db := pg.getExecutor(ctx)

	session.StartedAt = clock.Now()
	session.LastSeen = session.StartedAt
	session.Recorded = false

	device, err := pg.DeviceResolve(ctx, store.DeviceUIDResolver, string(session.DeviceUID))
	if err != nil {
		return "", fromSqlError(err)
	}

	session.TenantID = device.TenantID
	session.UID = uuid.New().String()

	e := entity.SessionFromModel(&session)
	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return "", fromSqlError(err)
	}

	return e.ID, nil
}

func (pg *Pg) SessionUpdate(ctx context.Context, session *models.Session) error {
	db := pg.getExecutor(ctx)

	e := entity.SessionFromModel(session)
	result, err := db.NewUpdate().Model(e).Where("id = ?", e.ID).Exec(ctx)
	if err != nil {
		return fromSqlError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fromSqlError(err)
	}

	if rowsAffected < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) ActiveSessionCreate(ctx context.Context, uid models.UID, session *models.Session) error {
	db := pg.getExecutor(ctx)

	activeSession := &models.ActiveSession{
		UID:      uid,
		LastSeen: session.StartedAt,
		TenantID: session.TenantID,
	}

	e := entity.ActiveSessionFromModel(activeSession)
	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return fromSqlError(err)
	}

	return nil
}

func (pg *Pg) ActiveSessionResolve(ctx context.Context, resolver store.SessionResolver, value string) (*models.ActiveSession, error) {
	db := pg.getExecutor(ctx)

	var sessionID string
	switch resolver {
	case store.SessionUIDResolver:
		sessionID = value
	default:
		return nil, store.ErrNoDocuments
	}

	e := &entity.ActiveSession{}
	if err := db.NewSelect().Model(e).Relation("Session").Relation("Session.Device").Where("session_id = ?", sessionID).Scan(ctx); err != nil {
		return nil, fromSqlError(err)
	}

	return entity.ActiveSessionToModel(e), nil
}

func (pg *Pg) ActiveSessionUpdate(ctx context.Context, activeSession *models.ActiveSession) error {
	db := pg.getExecutor(ctx)

	e := entity.ActiveSessionFromModel(activeSession)
	result, err := db.NewUpdate().Model(e).Where("session_id = ?", e.SessionID).Exec(ctx)
	if err != nil {
		return fromSqlError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fromSqlError(err)
	}

	if rowsAffected < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) ActiveSessionDelete(ctx context.Context, uid models.UID) error {
	db := pg.getExecutor(ctx)

	return pg.WithTransaction(ctx, func(ctx context.Context) error {
		db := pg.getExecutor(ctx)

		result, err := db.NewUpdate().
			Model((*entity.Session)(nil)).
			Set("seen_at = ?, closed = ?", clock.Now(), true).
			Where("id = ?", string(uid)).
			Exec(ctx)
		if err != nil {
			return fromSqlError(err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fromSqlError(err)
		}

		if rowsAffected < 1 {
			return store.ErrNoDocuments
		}

		if _, err := db.NewDelete().Model((*entity.ActiveSession)(nil)).Where("session_id = ?", string(uid)).Exec(ctx); err != nil {
			return fromSqlError(err)
		}

		return nil
	})
}

func (pg *Pg) SessionEventsCreate(ctx context.Context, uid models.UID, event *models.SessionEvent) error {
	db := pg.getExecutor(ctx)

	event.Session = string(uid)
	e := entity.SessionEventFromModel(event)
	e.ID = uuid.New().String()

	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return fromSqlError(err)
	}

	return nil
}

func (pg *Pg) SessionEventsList(ctx context.Context, uid models.UID, seat int, event models.SessionEventType, opts ...store.QueryOption) ([]models.SessionEvent, int, error) {
	db := pg.getExecutor(ctx)

	entities := make([]entity.SessionEvent, 0)
	query := db.NewSelect().
		Model(&entities).
		Where("session_id = ?", string(uid)).
		Where("seat = ?", seat).
		Where("type = ?", string(event)).
		Order("created_at ASC")

	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, 0, fromSqlError(err)
	}

	count, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fromSqlError(err)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, 0, fromSqlError(err)
	}

	events := make([]models.SessionEvent, len(entities))
	for i, e := range entities {
		events[i] = *entity.SessionEventToModel(&e)
	}

	return events, count, nil
}

func (pg *Pg) SessionEventsDelete(ctx context.Context, uid models.UID, seat int, event models.SessionEventType) error {
	db := pg.getExecutor(ctx)

	if _, err := db.NewDelete().
		Model((*entity.SessionEvent)(nil)).
		Where("session_id = ?", string(uid)).
		Where("seat = ?", seat).
		Where("type = ?", string(event)).
		Exec(ctx); err != nil {
		return fromSqlError(err)
	}

	return nil
}

func (pg *Pg) SessionUpdateDeviceUID(ctx context.Context, oldUID models.UID, newUID models.UID) error {
	db := pg.getExecutor(ctx)

	result, err := db.NewUpdate().
		Model((*entity.Session)(nil)).
		Set("device_id = ?", string(newUID)).
		Where("device_id = ?", string(oldUID)).
		Exec(ctx)
	if err != nil {
		return fromSqlError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fromSqlError(err)
	}

	if rowsAffected < 1 {
		return store.ErrNoDocuments
	}

	return nil
}
