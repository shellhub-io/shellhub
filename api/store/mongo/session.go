package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (s *Store) SessionList(ctx context.Context, opts ...store.QueryOption) ([]models.Session, int, error) {
	query := []bson.M{{"$match": bson.M{"uid": bson.M{"$ne": nil}}}}
	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "query", &query)); err != nil {
			return nil, 0, err
		}
	}

	count, err := CountAllMatchingDocuments(ctx, s.db.Collection("sessions"), query)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	query = append(query, []bson.M{
		{
			"$lookup": bson.M{
				"from":         "active_sessions",
				"localField":   "uid",
				"foreignField": "uid",
				"as":           "active",
			},
		},
		{
			"$lookup": bson.M{
				"from": "sessions_events",
				"let":  bson.M{"sessionUID": "$uid"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{"$eq": []string{"$session", "$$sessionUID"}},
						},
					},
					{
						"$group": bson.M{
							"_id":   nil,
							"types": bson.M{"$addToSet": "$type"},
							"seats": bson.M{"$addToSet": "$seat"},
						},
					},
				},
				"as": "eventData",
			},
		},
		{
			"$addFields": bson.M{
				"active": bson.M{"$anyElementTrue": []any{"$active"}},
				"events": bson.M{
					"$cond": bson.M{
						"if": bson.M{"$gt": []any{bson.M{"$size": "$eventData"}, 0}},
						"then": bson.M{
							"types": bson.M{"$arrayElemAt": []any{"$eventData.types", 0}},
							"seats": bson.M{"$arrayElemAt": []any{"$eventData.seats", 0}},
						},
						"else": bson.M{
							"types": []string{},
							"seats": []int{},
						},
					},
				},
			},
		},
		{
			"$unset": "eventData",
		},
	}...)

	sessions := make([]models.Session, 0)
	cursor, err := s.db.Collection("sessions").Aggregate(ctx, query)
	if err != nil {
		return sessions, count, FromMongoError(err)
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		session := new(models.Session)
		err = cursor.Decode(&session)
		if err != nil {
			return sessions, count, err
		}

		// WARNING: N+1 query problem - DeviceResolve makes a separate database call
		// for each session in the result set. For large result sets, consider using
		// a $lookup stage in the aggregation pipeline or batch-loading devices.
		device, err := s.DeviceResolve(ctx, store.DeviceUIDResolver, string(session.DeviceUID))
		if err != nil {
			return sessions, count, err
		}

		session.Device = device
		sessions = append(sessions, *session)
	}

	return sessions, count, err
}

func (s *Store) SessionResolve(ctx context.Context, resolver store.SessionResolver, value string, opts ...store.QueryOption) (*models.Session, error) {
	var uid models.UID
	switch resolver {
	case store.SessionUIDResolver:
		uid = models.UID(value)
	default:
		return nil, store.ErrNoDocuments
	}

	query := []bson.M{
		{
			"$match": bson.M{"uid": uid},
		},
		{
			"$lookup": bson.M{
				"from":         "active_sessions",
				"localField":   "uid",
				"foreignField": "uid",
				"as":           "active",
			},
		},
		{
			"$lookup": bson.M{
				"from": "sessions_events",
				"let":  bson.M{"sessionUID": "$uid"},
				"pipeline": []bson.M{
					{
						"$match": bson.M{
							"$expr": bson.M{"$eq": []string{"$session", "$$sessionUID"}},
						},
					},
					{
						"$group": bson.M{
							"_id":   nil,
							"types": bson.M{"$addToSet": "$type"},
							"seats": bson.M{"$addToSet": "$seat"},
						},
					},
				},
				"as": "eventData",
			},
		},
		{
			"$addFields": bson.M{
				"active": bson.M{"$anyElementTrue": []any{"$active"}},
				"events": bson.M{
					"$cond": bson.M{
						"if": bson.M{"$gt": []any{bson.M{"$size": "$eventData"}, 0}},
						"then": bson.M{
							"types": bson.M{"$arrayElemAt": []any{"$eventData.types", 0}},
							"seats": bson.M{"$arrayElemAt": []any{"$eventData.seats", 0}},
						},
						"else": bson.M{
							"types": []string{},
							"seats": []int{},
						},
					},
				},
			},
		},
		{
			"$unset": "eventData",
		},
	}

	cursor, err := s.db.Collection("sessions").Aggregate(ctx, query)
	if err != nil {
		return nil, FromMongoError(err)
	}

	defer cursor.Close(ctx)
	cursor.Next(ctx)

	session := new(models.Session)
	if err = cursor.Decode(&session); err != nil {
		return nil, FromMongoError(err)
	}

	device, err := s.DeviceResolve(ctx, store.DeviceUIDResolver, string(session.DeviceUID))
	if err != nil {
		return nil, FromMongoError(err)
	}

	session.Device = device

	return session, nil
}

func (s *Store) SessionUpdate(ctx context.Context, session *models.Session) error {
	// Convert to bson.M omitting zero values (mimics PostgreSQL's OmitZero)
	update := toBSONOmitZero(session)

	// Remove UID from update as it's used in the filter
	delete(update, "uid")

	// Special handling for booleans: they should always be included even if false
	// because false is a valid intentional value, not a zero-value to omit
	update["closed"] = session.Closed
	update["authenticated"] = session.Authenticated
	update["recorded"] = session.Recorded

	r, err := s.db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": session.UID}, bson.M{"$set": update})
	if err != nil {
		return FromMongoError(err)
	}

	if r.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) SessionCreate(ctx context.Context, session models.Session) (string, error) {
	session.StartedAt = clock.Now()
	session.LastSeen = session.StartedAt
	session.Recorded = false

	if session.UID == "" {
		session.UID = uuid.Generate()
	}

	device, err := s.DeviceResolve(ctx, store.DeviceUIDResolver, string(session.DeviceUID))
	if err != nil {
		return "", FromMongoError(err)
	}

	session.TenantID = device.TenantID

	if _, err := s.db.Collection("sessions").InsertOne(ctx, &session); err != nil {
		return "", FromMongoError(err)
	}

	return session.UID, nil
}

func (s *Store) SessionUpdateDeviceUID(ctx context.Context, oldUID models.UID, newUID models.UID) error {
	session, err := s.db.Collection("sessions").UpdateMany(ctx, bson.M{"device_uid": oldUID}, bson.M{"$set": bson.M{"device_uid": newUID}})
	if err != nil {
		return FromMongoError(err)
	}

	if session.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) ActiveSessionResolve(ctx context.Context, resolver store.SessionResolver, value string) (*models.ActiveSession, error) {
	var uid models.UID
	switch resolver {
	case store.SessionUIDResolver:
		uid = models.UID(value)
	default:
		return nil, store.ErrNoDocuments
	}

	activeSession := new(models.ActiveSession)
	if err := s.db.Collection("active_sessions").FindOne(ctx, bson.M{"uid": uid}).Decode(activeSession); err != nil {
		return nil, FromMongoError(err)
	}

	return activeSession, nil
}

func (s *Store) ActiveSessionCreate(ctx context.Context, session *models.Session) error {
	_, err := s.db.Collection("active_sessions").
		InsertOne(ctx, &models.ActiveSession{UID: models.UID(session.UID), LastSeen: session.StartedAt, TenantID: session.TenantID})
	if err != nil {
		return FromMongoError(err)
	}

	return nil
}

func (s *Store) ActiveSessionUpdate(ctx context.Context, activeSession *models.ActiveSession) error {
	r, err := s.db.Collection("active_sessions").UpdateOne(ctx, bson.M{"uid": activeSession.UID}, bson.M{"$set": activeSession})
	if err != nil {
		return FromMongoError(err)
	}

	if r.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) ActiveSessionDelete(ctx context.Context, uid models.UID) error {
	mongoSession, err := s.db.Client().StartSession()
	if err != nil {
		return FromMongoError(err)
	}
	defer mongoSession.EndSession(ctx)

	_, err = mongoSession.WithTransaction(ctx, func(_ mongo.SessionContext) (any, error) {
		r, err := s.db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"last_seen": clock.Now(), "closed": true}})
		if err != nil {
			return nil, FromMongoError(err)
		}

		if r.MatchedCount < 1 {
			return nil, store.ErrNoDocuments
		}

		if _, err := s.db.Collection("active_sessions").DeleteMany(ctx, bson.M{"uid": uid}); err != nil {
			return nil, FromMongoError(err)
		}

		return nil, nil
	})

	return err
}

func (s *Store) SessionEventsCreate(ctx context.Context, event *models.SessionEvent) error {
	if _, err := s.db.Collection("sessions_events").InsertOne(ctx, event); err != nil {
		return FromMongoError(err)
	}

	return nil
}

func (s *Store) SessionEventsList(ctx context.Context, uid models.UID, seat int, event models.SessionEventType, opts ...store.QueryOption) ([]models.SessionEvent, int, error) {
	query := []bson.M{
		{
			"$match": bson.M{
				"session": uid,
				"seat":    seat,
				"type":    event,
			},
		},
		{
			"$sort": bson.M{
				"timestamp": 1,
			},
		},
	}

	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "query", &query)); err != nil {
			return nil, 0, err
		}
	}

	count, err := CountAllMatchingDocuments(ctx, s.db.Collection("sessions_events"), query)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	cursosr, err := s.db.Collection("sessions_events").Aggregate(ctx, query)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	events := make([]models.SessionEvent, 0)
	if err := cursosr.All(ctx, &events); err != nil {
		return nil, 0, FromMongoError(err)
	}

	return events, count, nil
}

func (s *Store) SessionEventsDelete(ctx context.Context, uid models.UID, seat int, event models.SessionEventType) error {
	if _, err := s.db.Collection("sessions_events").DeleteMany(ctx, bson.M{"session": uid, "seat": seat, "type": event}); err != nil {
		return FromMongoError(err)
	}

	return nil
}
