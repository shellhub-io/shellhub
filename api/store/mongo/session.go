package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo/queries"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

func (s *Store) SessionList(ctx context.Context, paginator query.Paginator) ([]models.Session, int, error) {
	query := []bson.M{
		{
			"$match": bson.M{
				"uid": bson.M{
					"$ne": nil,
				},
			},
		},
	}

	// Only match for the respective tenant if requested
	if tenant := gateway.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	queryCount := query
	queryCount = append(queryCount, bson.M{"$count": "count"})
	count, err := AggregateCount(ctx, s.db.Collection("sessions"), queryCount)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	query = append(query, bson.M{
		"$sort": bson.M{
			"started_at": -1,
		},
	})

	query = append(query, queries.FromPaginator(&paginator)...)
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

		device, err := s.DeviceResolve(ctx, store.DeviceUIDResolver, string(session.DeviceUID))
		if err != nil {
			return sessions, count, err
		}

		session.Device = device
		sessions = append(sessions, *session)
	}

	return sessions, count, err
}

func (s *Store) SessionGet(ctx context.Context, uid models.UID) (*models.Session, error) {
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

	// Only match for the respective tenant if requested
	if tenant := gateway.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
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

func (s *Store) SessionUpdate(ctx context.Context, uid models.UID, sess *models.Session, update *models.SessionUpdate) error {
	clientSession, err := s.db.Client().StartSession()
	if err != nil {
		return err
	}
	defer clientSession.EndSession(ctx)

	if update.Authenticated != nil && !sess.Authenticated {
		if err := s.SessionCreateActive(ctx, uid, sess); err != nil {
			return err
		}
	}

	fields := bson.M{}
	if update.Authenticated != nil {
		fields["authenticated"] = *update.Authenticated
	}
	if update.Type != nil {
		fields["type"] = *update.Type
	}
	if update.Recorded != nil {
		fields["recorded"] = *update.Recorded
	}

	if len(fields) > 0 {
		res, err := s.db.Collection("sessions").
			UpdateOne(ctx,
				bson.M{"uid": uid},
				bson.M{"$set": fields},
			)
		if err != nil {
			return FromMongoError(err)
		}
		if res.MatchedCount < 1 {
			return store.ErrNoDocuments
		}
	}

	return nil
}

func (s *Store) SessionSetRecorded(ctx context.Context, uid models.UID, recorded bool) error {
	session, err := s.db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"recorded": recorded}})
	if err != nil {
		return FromMongoError(err)
	}

	if session.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) SessionSetType(ctx context.Context, uid models.UID, kind string) error {
	session, err := s.db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"type": kind}})
	if err != nil {
		return FromMongoError(err)
	}

	if session.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) SessionCreate(ctx context.Context, session models.Session) (*models.Session, error) {
	session.StartedAt = clock.Now()
	session.LastSeen = session.StartedAt
	session.Recorded = false

	device, err := s.DeviceResolve(ctx, store.DeviceUIDResolver, string(session.DeviceUID))
	if err != nil {
		return nil, FromMongoError(err)
	}

	session.TenantID = device.TenantID

	if _, err := s.db.Collection("sessions").InsertOne(ctx, &session); err != nil {
		return nil, FromMongoError(err)
	}

	return &session, nil
}

func (s *Store) SessionSetLastSeen(ctx context.Context, uid models.UID) error {
	session := models.Session{}

	err := s.db.Collection("sessions").FindOne(ctx, bson.M{"uid": uid}).Decode(&session)
	if err != nil {
		return FromMongoError(err)
	}

	if session.Closed {
		return nil
	}

	session.LastSeen = clock.Now()

	opts := options.Update().SetUpsert(true)
	_, err = s.db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": session.UID}, bson.M{"$set": session}, opts)
	if err != nil {
		return FromMongoError(err)
	}

	if _, err := s.db.Collection("active_sessions").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"last_seen": clock.Now()}}); err != nil {
		return FromMongoError(err)
	}

	return nil
}

// SessionDeleteActives sets a session's "closed" status to true and deletes all related active_sessions.
func (s *Store) SessionDeleteActives(ctx context.Context, uid models.UID) error {
	mongoSession, err := s.db.Client().StartSession()
	if err != nil {
		return FromMongoError(err)
	}
	defer mongoSession.EndSession(ctx)

	_, err = mongoSession.WithTransaction(ctx, func(_ mongo.SessionContext) (any, error) {
		session := new(models.Session)

		query := bson.M{"uid": uid}
		update := bson.M{"$set": bson.M{"last_seen": clock.Now(), "closed": true}}

		if err := s.db.Collection("sessions").FindOneAndUpdate(ctx, query, update).Decode(&session); err != nil {
			return nil, FromMongoError(err)
		}

		_, err := s.db.Collection("active_sessions").DeleteMany(ctx, bson.M{"uid": session.UID})

		return nil, FromMongoError(err)
	})

	return err
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

func (s *Store) SessionCreateActive(ctx context.Context, uid models.UID, session *models.Session) error {
	_, err := s.db.Collection("active_sessions").InsertOne(ctx, &models.ActiveSession{
		UID:      uid,
		LastSeen: session.StartedAt,
		TenantID: session.TenantID,
	})
	if err != nil {
		return FromMongoError(err)
	}

	return nil
}

// SessionEvent saves a [models.SessionEvent] into the database.
//
// It pushes the event into events type array, and the event type into a separated set. The set is used to improve the
// performance of indexing when looking for sessions.
func (s *Store) SessionEvent(ctx context.Context, uid models.UID, event *models.SessionEvent) error {
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

func (s *Store) SessionListEvents(ctx context.Context, uid models.UID, paginator query.Paginator, filters query.Filters, sorter query.Sorter) ([]models.SessionEvent, int, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"session": uid,
			},
		},
	}

	queryMatch, err := queries.FromFilters(&filters)
	if err != nil {
		log.WithError(err).Error("failed to create filters")

		return nil, 0, FromMongoError(err)
	}

	pipeline = append(pipeline, queryMatch...)

	countPipeline := append(pipeline, bson.M{"$count": "count"})
	count, err := AggregateCount(ctx, s.db.Collection("sessions_events"), countPipeline)
	if err != nil {
		log.WithError(err).Error("failed to count sessions_events")

		return nil, 0, FromMongoError(err)
	}

	if sorter.By == "" {
		sorter.By = "timestamp"
	}

	pipeline = append(pipeline, queries.FromSorter(&sorter)...)
	pipeline = append(pipeline, queries.FromPaginator(&paginator)...)

	opts := options.Aggregate().SetAllowDiskUse(true)
	cursor, err := s.db.Collection("sessions_events").Aggregate(ctx, pipeline, opts)
	if err != nil {
		log.WithError(err).Error("failed to run aggregation against sessions_events collection")

		return nil, 0, FromMongoError(err)
	}

	defer cursor.Close(ctx)

	events := make([]models.SessionEvent, 0)
	for cursor.Next(ctx) {
		var event models.SessionEvent
		if err := cursor.Decode(&event); err != nil {
			log.WithError(err).Error("failed to decode the event from the cursor")

			return nil, 0, err
		}

		switch event.Type {
		case models.SessionEventTypeWindowChange:
			prim := event.Data.(primitive.D)

			data, err := bson.Marshal(prim)
			if err != nil {
				return nil, 0, err
			}

			model := models.SSHWindowChange{}
			if err := bson.Unmarshal(data, &model); err != nil {
				return nil, 0, err
			}

			event.Data = model
		case models.SessionEventTypePtyRequest:
			// NOTE: We're converting the data returned by MongoDB when the field is a [any] to out structure.
			prim := event.Data.(primitive.D)

			data, err := bson.Marshal(prim)
			if err != nil {
				return nil, 0, err
			}

			model := models.SSHPty{}
			if err := bson.Unmarshal(data, &model); err != nil {
				return nil, 0, err
			}

			event.Data = model
		}

		events = append(events, event)
	}

	return events, count, nil
}

func (s *Store) SessionDeleteEvents(ctx context.Context, uid models.UID, seat int, event models.SessionEventType) error {
	filter := bson.M{
		"session": uid,
		"seat":    seat,
		"type":    event,
	}

	if _, err := s.db.Collection("sessions_events").DeleteMany(ctx, filter); err != nil {
		return FromMongoError(err)
	}

	return nil
}
