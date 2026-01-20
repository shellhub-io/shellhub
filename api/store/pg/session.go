package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
)

func (pg *Pg) SessionList(ctx context.Context, opts ...store.QueryOption) ([]models.Session, int, error) {
	db := pg.getConnection(ctx)

	entities := make([]entity.Session, 0)
	query := db.NewSelect().
		Model(&entities).
		Relation("Device").
		Relation("Device.Namespace")

	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, 0, fromSQLError(err)
	}

	count, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	query = db.NewSelect().
		Model(&entities).
		Relation("Device").
		Relation("Device.Namespace").
		ColumnExpr("session.*").
		ColumnExpr(sessionExprActive()).
		ColumnExpr(sessionExprEventTypes()).
		ColumnExpr(sessionExprEventSeats()).
		Join("LEFT JOIN active_sessions AS active_session ON session.id = active_session.session_id").
		Join(`LEFT JOIN (
			SELECT session_id, string_agg(DISTINCT type::text, ',') as types
			FROM session_events
			GROUP BY session_id
		) event_types ON session.id = event_types.session_id`).
		Join(`LEFT JOIN (
			SELECT session_id, string_agg(DISTINCT seat::text, ',') as seats
			FROM session_events
			GROUP BY session_id
		) event_seats ON session.id = event_seats.session_id`)

	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, 0, fromSQLError(err)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, 0, fromSQLError(err)
	}

	sessions := make([]models.Session, len(entities))
	for i, e := range entities {
		sessions[i] = *entity.SessionToModel(&e)
	}

	return sessions, count, nil
}

func (pg *Pg) SessionResolve(ctx context.Context, resolver store.SessionResolver, value string, opts ...store.QueryOption) (*models.Session, error) {
	db := pg.getConnection(ctx)

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
		ColumnExpr("session.*").
		ColumnExpr(sessionExprActive()).
		ColumnExpr(sessionExprEventTypes()).
		ColumnExpr(sessionExprEventSeats()).
		Join("LEFT JOIN active_sessions AS active_session ON session.id = active_session.session_id").
		Join(`LEFT JOIN (
			SELECT session_id, string_agg(DISTINCT type::text, ',') as types
			FROM session_events
			WHERE session_id = ?
			GROUP BY session_id
		) event_types ON session.id = event_types.session_id`, sessionID).
		Join(`LEFT JOIN (
			SELECT session_id, string_agg(DISTINCT seat::text, ',') as seats
			FROM session_events
			WHERE session_id = ?
			GROUP BY session_id
		) event_seats ON session.id = event_seats.session_id`, sessionID).
		Where("session.id = ?", sessionID)

	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, fromSQLError(err)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.SessionToModel(e), nil
}

func (pg *Pg) SessionCreate(ctx context.Context, session models.Session) (string, error) {
	db := pg.getConnection(ctx)

	session.StartedAt = clock.Now()
	session.LastSeen = session.StartedAt
	session.Recorded = false

	device, err := pg.DeviceResolve(ctx, store.DeviceUIDResolver, string(session.DeviceUID))
	if err != nil {
		return "", fromSQLError(err)
	}

	session.TenantID = device.TenantID

	e := entity.SessionFromModel(&session)
	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	return e.ID, nil
}

func (pg *Pg) SessionUpdate(ctx context.Context, session *models.Session) error {
	db := pg.getConnection(ctx)

	e := entity.SessionFromModel(session)
	result, err := db.NewUpdate().Model(e).Where("id = ?", e.ID).Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) ActiveSessionCreate(ctx context.Context, session *models.Session) error {
	db := pg.getConnection(ctx)

	activeSession := &models.ActiveSession{UID: models.UID(session.UID), LastSeen: session.StartedAt, TenantID: session.TenantID}
	e := entity.ActiveSessionFromModel(activeSession)
	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

func (pg *Pg) ActiveSessionResolve(ctx context.Context, resolver store.SessionResolver, value string) (*models.ActiveSession, error) {
	db := pg.getConnection(ctx)

	var sessionID string
	switch resolver {
	case store.SessionUIDResolver:
		sessionID = value
	default:
		return nil, store.ErrNoDocuments
	}

	e := &entity.ActiveSession{}
	if err := db.NewSelect().Model(e).Relation("Session").Relation("Session.Device").Where("session_id = ?", sessionID).Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.ActiveSessionToModel(e), nil
}

func (pg *Pg) ActiveSessionUpdate(ctx context.Context, activeSession *models.ActiveSession) error {
	db := pg.getConnection(ctx)

	e := entity.ActiveSessionFromModel(activeSession)
	result, err := db.NewUpdate().Model(e).Where("session_id = ?", e.SessionID).Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (pg *Pg) ActiveSessionDelete(ctx context.Context, uid models.UID) error {
	return pg.WithTransaction(ctx, func(ctx context.Context) error {
		db := pg.getConnection(ctx)

		result, err := db.NewUpdate().
			Model((*entity.Session)(nil)).
			Set("closed = ?", true).
			Set("seen_at = ?", clock.Now()).
			Where("id = ?", string(uid)).
			Exec(ctx)
		if err != nil {
			return fromSQLError(err)
		}

		if rowsAffected, _ := result.RowsAffected(); rowsAffected < 1 {
			return store.ErrNoDocuments
		}

		if _, err := db.NewDelete().
			Model((*entity.ActiveSession)(nil)).
			Where("session_id = ?", string(uid)).
			Exec(ctx); err != nil {
			return fromSQLError(err)
		}

		return nil
	})
}

func (pg *Pg) SessionEventsCreate(ctx context.Context, event *models.SessionEvent) error {
	db := pg.getConnection(ctx)

	e := entity.SessionEventFromModel(event)
	e.ID = uuid.Generate()

	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

func (pg *Pg) SessionEventsList(ctx context.Context, uid models.UID, seat int, event models.SessionEventType, opts ...store.QueryOption) ([]models.SessionEvent, int, error) {
	db := pg.getConnection(ctx)

	entities := make([]entity.SessionEvent, 0)
	query := db.NewSelect().
		Model(&entities).
		Where("session_id = ?", string(uid)).
		Where("seat = ?", seat).
		Where("type = ?", string(event)).
		Order("created_at ASC")

	if err := applyOptions(ctx, query, opts...); err != nil {
		return nil, 0, fromSQLError(err)
	}

	count, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	if err := query.Scan(ctx); err != nil {
		return nil, 0, fromSQLError(err)
	}

	events := make([]models.SessionEvent, len(entities))
	for i, e := range entities {
		events[i] = *entity.SessionEventToModel(&e)
	}

	return events, count, nil
}

func (pg *Pg) SessionEventsDelete(ctx context.Context, uid models.UID, seat int, event models.SessionEventType) error {
	db := pg.getConnection(ctx)

	if _, err := db.NewDelete().
		Model((*entity.SessionEvent)(nil)).
		Where("session_id = ?", string(uid)).
		Where("seat = ?", seat).
		Where("type = ?", string(event)).
		Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

func (pg *Pg) SessionUpdateDeviceUID(ctx context.Context, oldUID models.UID, newUID models.UID) error {
	db := pg.getConnection(ctx)

	result, err := db.NewUpdate().
		Model((*entity.Session)(nil)).
		Set("device_id = ?", string(newUID)).
		Where("device_id = ?", string(oldUID)).
		Exec(ctx)
	if err != nil {
		return fromSQLError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fromSQLError(err)
	}

	if rowsAffected < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

// sessionExprActive returns the SQL expression for the "active" field.
func sessionExprActive() string {
	return "CASE WHEN active_session.session_id IS NOT NULL THEN true ELSE false END AS active"
}

// sessionExprEventTypes returns the SQL expression for aggregated event types.
func sessionExprEventTypes() string {
	return "COALESCE(event_types.types, '') AS event_types"
}

// sessionExprEventSeats returns the SQL expression for aggregated event seats.
func sessionExprEventSeats() string {
	return "COALESCE(event_seats.seats, '') AS event_seats"
}
