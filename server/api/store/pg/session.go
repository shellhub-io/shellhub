package pg

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/api/store/pg/entity"
	"github.com/uptrace/bun"
)

func (pg *Pg) SessionList(ctx context.Context, sc scope.Scope, opts ...store.QueryOption) ([]models.Session, int, error) {
	db := pg.GetConnection(ctx)

	// Sessions and devices both have namespace_id; qualify the column to
	// avoid ambiguity when SessionSelectQuery JOINs the devices table.
	ctx = context.WithValue(ctx, CtxTableAlias, "session")

	entities := make([]entity.Session, 0)
	query := db.NewSelect().
		Model(&entities)

	var err error
	query, err = applyScopedOptions(ctx, query, sc, opts...)
	if err != nil {
		return nil, 0, err
	}

	count, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fromSQLError(err)
	}

	query = SessionSelectQuery(db.NewSelect().Model(&entities))

	query, err = applyScopedOptions(ctx, query, sc, opts...)
	if err != nil {
		return nil, 0, err
	}

	if err = query.Scan(ctx); err != nil {
		return nil, 0, fromSQLError(err)
	}

	threshold := clock.Now().Add(-2 * DeviceOnlineThreshold)
	sessions := make([]models.Session, len(entities))
	for i, e := range entities {
		if e.Device != nil {
			e.Device.Online = DeviceIsOnline(e.Device, threshold)
		}
		sessions[i] = *entity.SessionToModel(&e)
	}

	return sessions, count, nil
}

func (pg *Pg) SessionResolve(ctx context.Context, sc scope.Scope, resolver store.SessionResolver, value string, opts ...store.QueryOption) (*models.Session, error) {
	db := pg.GetConnection(ctx)

	// Sessions and devices both have namespace_id; qualify the column so the
	// scope predicate doesn't hit an ambiguous column when SessionSelectQuery
	// JOINs the devices table.
	ctx = context.WithValue(ctx, CtxTableAlias, "session")

	var sessionID string
	switch resolver {
	case store.SessionUIDResolver:
		sessionID = value
	default:
		return nil, store.ErrNoDocuments
	}

	e := &entity.Session{}
	query := SessionSelectQuery(db.NewSelect().Model(e)).
		Where("session.id = ?", sessionID)

	var err error
	query, err = applyScopedOptions(ctx, query, sc, opts...)
	if err != nil {
		return nil, err
	}

	if err = query.Scan(ctx); err != nil {
		return nil, fromSQLError(err)
	}

	if e.Device != nil {
		e.Device.Online = DeviceIsOnline(e.Device, clock.Now().Add(-2*DeviceOnlineThreshold))
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

	// The device is what establishes the session's namespace, so this lookup cannot be bounded by
	// the namespace it is about to discover. The device UID is a hash the SSH server has already
	// authenticated the connection against.
	sc := scope.NewUnbounded("resolving the device that determines the new session's namespace")

	device, err := pg.DeviceResolve(ctx, sc, store.DeviceUIDResolver, string(session.DeviceUID))
	if err != nil {
		return "", fromSQLError(err)
	}

	session.TenantID = device.TenantID

	e := entity.SessionFromModel(&session)
	e.CreatedAt = clock.Now()
	if _, err := db.NewInsert().Model(e).Exec(ctx); err != nil {
		return "", fromSQLError(err)
	}

	return e.ID, nil
}

func (pg *Pg) SessionUpdate(ctx context.Context, session *models.Session) error {
	db := pg.GetConnection(ctx)

	e := entity.SessionFromModel(session)
	result, err := db.NewUpdate().
		Model(e).
		OmitZero().
		ExcludeColumn("closed", "authenticated", "recorded").
		Set("closed = ?", e.Closed).
		Set("authenticated = ?", e.Authenticated).
		Set("recorded = ?", e.Recorded).
		Where("id = ?", e.ID).
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

func (pg *Pg) ActiveSessionCreate(ctx context.Context, session *models.Session) error {
	db := pg.GetConnection(ctx)

	activeSession := &models.ActiveSession{UID: models.UID(session.UID), LastSeen: session.StartedAt, TenantID: session.TenantID}
	e := entity.ActiveSessionFromModel(activeSession)
	if _, err := db.NewInsert().Model(e).
		On("CONFLICT (session_id) DO UPDATE").
		Set("seen_at = NOW()").
		Exec(ctx); err != nil {
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

func (pg *Pg) SessionEventsCreateMany(ctx context.Context, events []models.SessionEvent) error {
	if len(events) == 0 {
		return nil
	}

	db := pg.GetConnection(ctx)

	entities := make([]*entity.SessionEvent, 0, len(events))
	for i := range events {
		e := entity.SessionEventFromModel(&events[i])
		e.ID = uuid.Generate()
		entities = append(entities, e)
	}

	if _, err := db.NewInsert().Model(&entities).Exec(ctx); err != nil {
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

func (pg *Pg) SessionListExpired(ctx context.Context, before time.Time, limit int) ([]string, error) {
	if limit <= 0 {
		return []string{}, nil
	}

	db := pg.GetConnection(ctx)

	// Ordering by started_at walks sessions_started_at_idx and takes the oldest rows first, so
	// successive batches drain the backlog from its far end instead of re-reading it.
	//
	// The anti-join is the safety rail: closed is not reliable for this (a session whose server
	// died never gets closed and would otherwise be immortal), whereas a row in active_sessions
	// is what a live session actually holds. Deleting one would cascade that row away underneath
	// the connection still using it.
	uids := make([]string, 0, limit)
	if err := db.NewSelect().
		Model((*entity.Session)(nil)).
		Column("id").
		Where("started_at < ?", before).
		Where("NOT EXISTS (SELECT 1 FROM active_sessions WHERE active_sessions.session_id = session.id)").
		Order("started_at ASC").
		Limit(limit).
		Scan(ctx, &uids); err != nil {
		return nil, fromSQLError(err)
	}

	return uids, nil
}

func (pg *Pg) SessionDeleteMany(ctx context.Context, uids []string) (int64, error) {
	if len(uids) == 0 {
		return 0, nil
	}

	db := pg.GetConnection(ctx)

	res, err := db.NewDelete().
		Model((*entity.Session)(nil)).
		Where("id IN (?)", bun.List(uids)).
		Exec(ctx)
	if err != nil {
		return 0, fromSQLError(err)
	}

	return res.RowsAffected()
}

// SessionSelectQuery applies the standard session SELECT decorations: relations,
// computed columns (active, event_types, event_seats), and the active_sessions JOIN.
// The caller provides the base query with the desired model (core or cloud entity).
//
// event_types and event_seats are computed with correlated subqueries scoped to each
// session row instead of aggregating the whole session_events table. The previous
// derived-table form (string_agg ... GROUP BY session_id with no filter) forced
// Postgres to scan and group every row in session_events on each call, even for the
// list path that only returns a page. As session_events grows, that turns into a
// multi-second full scan whose cost scales with the total event count rather than the
// page size. The correlated form runs only for the page's sessions and uses
// session_events_session_id_created_at_idx, keeping the query bounded.
func SessionSelectQuery(q *bun.SelectQuery) *bun.SelectQuery {
	return q.
		Relation("Device").
		Relation("Device.Namespace").
		ColumnExpr("session.*").
		ColumnExpr("CASE WHEN active_session.session_id IS NOT NULL THEN true ELSE false END AS active").
		ColumnExpr(`COALESCE((
			SELECT string_agg(DISTINCT type::text, ',')
			FROM session_events
			WHERE session_id = session.id
		), '') AS event_types`).
		ColumnExpr(`COALESCE((
			SELECT string_agg(DISTINCT seat::text, ',')
			FROM session_events
			WHERE session_id = session.id
		), '') AS event_seats`).
		Join("LEFT JOIN active_sessions AS active_session ON session.id = active_session.session_id")
}
