package pg

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/uptrace/bun"
)

func (pg *Pg) SessionList(ctx context.Context, opts ...store.QueryOption) ([]models.Session, int, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.Session, 0)
	query := db.NewSelect().
		Model(&entities).
		Relation("Device").
		Relation("Device.Namespace")

	var err error
	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	query = SessionSelectQuery(db.NewSelect().Model(&entities), "")

	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, 0, err
	}

	if err = query.Scan(ctx); err != nil {
		return nil, 0, fromSQLError(err)
	}

	sessions := make([]models.Session, len(entities))
	for i, e := range entities {
		sessions[i] = *entity.SessionToModel(&e)
	}

	return sessions, count, nil
}

func (pg *Pg) SessionResolve(ctx context.Context, resolver store.SessionResolver, value string, opts ...store.QueryOption) (*models.Session, error) {
	db := pg.GetConnection(ctx)

	var sessionID string
	switch resolver {
	case store.SessionUIDResolver:
		sessionID = value
	default:
		return nil, store.ErrNoDocuments
	}

	e := &entity.Session{}
	query := SessionSelectQuery(db.NewSelect().Model(e), sessionID).
		Where("session.id = ?", sessionID)

	var err error
	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, err
	}

	if err = query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	return entity.SessionToModel(e), nil
}

func (pg *Pg) SessionCreate(ctx context.Context, session models.Session) (string, error) {
	db := pg.GetConnection(ctx)

	session.StartedAt = clock.Now()
	session.LastSeen = session.StartedAt
	session.Recorded = false

	if session.UID == "" {
		session.UID = uuid.Generate()
	}

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
	db := pg.GetConnection(ctx)

	e := entity.SessionFromModel(session)
	result, err := db.NewUpdate().Model(e).OmitZero().Where("id = ?", e.ID).Exec(ctx)
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
	db := pg.GetConnection(ctx)

	activeSession := &models.ActiveSession{UID: models.UID(session.UID), LastSeen: session.StartedAt, TenantID: session.TenantID}
	e := entity.ActiveSessionFromModel(activeSession)
	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

func (pg *Pg) ActiveSessionResolve(ctx context.Context, resolver store.SessionResolver, value string) (*models.ActiveSession, error) {
	db := pg.GetConnection(ctx)

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
	db := pg.GetConnection(ctx)

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
		db := pg.GetConnection(ctx)

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
	db := pg.GetConnection(ctx)

	e := entity.SessionEventFromModel(event)
	e.ID = uuid.Generate()

	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return fromSQLError(err)
	}

	return nil
}

func (pg *Pg) SessionEventsList(ctx context.Context, uid models.UID, seat int, event models.SessionEventType, opts ...store.QueryOption) ([]models.SessionEvent, int, error) {
	db := pg.GetConnection(ctx)

	entities := make([]entity.SessionEvent, 0)
	query := db.NewSelect().
		Model(&entities).
		Where("session_id = ?", string(uid)).
		Where("seat = ?", seat).
		Where("type = ?", string(event)).
		Order("created_at ASC")

	var err error
	query, err = applyOptions(ctx, query, opts...)
	if err != nil {
		return nil, 0, err
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
	db := pg.GetConnection(ctx)

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
	db := pg.GetConnection(ctx)

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

// SessionSelectQuery applies the standard session SELECT decorations: relations,
// computed columns (active, event_types, event_seats), and LEFT JOINs.
// If sessionID is non-empty, event subqueries are filtered for performance.
// The caller provides the base query with the desired model (core or cloud entity).
func SessionSelectQuery(q *bun.SelectQuery, sessionID string) *bun.SelectQuery {
	q = q.
		Relation("Device").
		Relation("Device.Namespace").
		ColumnExpr("session.*").
		ColumnExpr("CASE WHEN active_session.session_id IS NOT NULL THEN true ELSE false END AS active").
		ColumnExpr("COALESCE(event_types.types, '') AS event_types").
		ColumnExpr("COALESCE(event_seats.seats, '') AS event_seats").
		Join("LEFT JOIN active_sessions AS active_session ON session.id = active_session.session_id")

	if sessionID != "" {
		q = q.
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
			) event_seats ON session.id = event_seats.session_id`, sessionID)
	} else {
		q = q.
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
	}

	return q
}
