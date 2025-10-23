package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
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
			"$addFields": bson.M{
				"active": bson.M{"$anyElementTrue": []any{"$active"}},
			},
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
			"$addFields": bson.M{
				"active": bson.M{"$anyElementTrue": []any{"$active"}},
			},
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
	r, err := s.db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": session.UID}, bson.M{"$set": session})
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

func (s *Store) ActiveSessionCreate(ctx context.Context, uid models.UID, session *models.Session) error {
	_, err := s.db.Collection("active_sessions").
		InsertOne(ctx, &models.ActiveSession{UID: uid, LastSeen: session.StartedAt, TenantID: session.TenantID})
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

func (s *Store) SessionEventsCreate(ctx context.Context, uid models.UID, event *models.SessionEvent) error {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return FromMongoError(err)
	}

	defer session.EndSession(ctx)

	txnOpts := options.Transaction().
		SetReadConcern(readconcern.Snapshot()).
		SetWriteConcern(writeconcern.Majority())

	if _, err := session.WithTransaction(ctx, func(ctx mongo.SessionContext) (any, error) {
		if _, err := s.db.Collection("sessions").UpdateOne(ctx,
			bson.M{"uid": uid},
			bson.M{
				"$addToSet": bson.M{
					"events.types": event.Type,
					"events.seats": event.Seat,
				},
			},
		); err != nil {
			return nil, err
		}

		if _, err := s.db.Collection("sessions_events").InsertOne(ctx, event); err != nil {
			return nil, err
		}

		return nil, nil
	}, txnOpts); err != nil {
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
	if err := cursosr.All(ctx, events); err != nil {
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
