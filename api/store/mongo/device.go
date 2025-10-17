package mongo

import (
	"context" //nolint:gosec
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// DeviceList returns a list of devices based on the given filters, pagination and sorting.
func (s *Store) DeviceList(ctx context.Context, acceptable store.DeviceAcceptable, opts ...store.QueryOption) ([]models.Device, int, error) {
	query := []bson.M{
		{
			"$match": bson.M{
				"uid": bson.M{
					"$ne": nil,
				},
			},
		},
		{
			"$addFields": bson.M{
				"online": bson.M{
					"$cond": bson.M{
						"if": bson.M{
							"$and": bson.A{
								bson.M{"$eq": bson.A{"$disconnected_at", nil}},
								bson.M{"$gt": bson.A{"$last_seen", primitive.NewDateTimeFromTime(time.Now().Add(-2 * time.Minute))}},
							},
						},
						"then": true,
						"else": false,
					},
				},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "tags",
				"localField":   "tag_ids",
				"foreignField": "_id",
				"as":           "tags",
			},
		},
	}

	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "query", &query)); err != nil {
			return nil, 0, err
		}
	}

	// When the listing mode is [store.DeviceListModeMaxDeviceReached], we should evaluate the `removed_devices`
	// collection to check its `accetable` status.
	switch acceptable {
	case store.DeviceAcceptableFromRemoved:
		query = append(query, bson.M{
			"$addFields": bson.M{
				"acceptable": bson.M{
					"$eq": bson.A{"$status", models.DeviceStatusRemoved},
				},
			},
		})
	case store.DeviceAcceptableAsFalse:
		query = append(query, bson.M{
			"$addFields": bson.M{
				"acceptable": false,
			},
		})
	case store.DeviceAcceptableIfNotAccepted:
		query = append(query, bson.M{
			"$addFields": bson.M{
				"acceptable": bson.M{
					"$cond": bson.M{
						"if":   bson.M{"$ne": bson.A{"$status", models.DeviceStatusAccepted}},
						"then": true,
						"else": false,
					},
				},
			},
		})
	}

	count, err := CountAllMatchingDocuments(ctx, s.db.Collection("devices"), query)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	query = append(query, []bson.M{
		{
			"$lookup": bson.M{
				"from":         "namespaces",
				"localField":   "tenant_id",
				"foreignField": "tenant_id",
				"as":           "namespace",
			},
		},
		{
			"$addFields": bson.M{
				"namespace": "$namespace.name",
			},
		},
		{
			"$unwind": "$namespace",
		},
	}...)

	devices := make([]models.Device, 0)

	cursor, err := s.db.Collection("devices").Aggregate(ctx, query)
	if err != nil {
		return devices, count, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		device := new(models.Device)

		if err = cursor.Decode(&device); err != nil {
			return devices, count, err
		}

		devices = append(devices, *device)
	}

	return devices, count, FromMongoError(err)
}

func (s *Store) DeviceResolve(ctx context.Context, resolver store.DeviceResolver, value string, opts ...store.QueryOption) (*models.Device, error) {
	matchStage := bson.M{}
	switch resolver {
	case store.DeviceUIDResolver:
		matchStage["uid"] = value
	case store.DeviceHostnameResolver:
		matchStage["name"] = value
	case store.DeviceMACResolver:
		matchStage["identity"] = bson.M{"mac": value}
	}

	query := []bson.M{
		{
			"$match": matchStage,
		},
		{
			"$addFields": bson.M{
				"online": bson.M{
					"$cond": bson.M{
						"if": bson.M{
							"$and": bson.A{
								bson.M{"$eq": bson.A{"$disconnected_at", nil}},
								bson.M{"$gt": bson.A{"$last_seen", primitive.NewDateTimeFromTime(time.Now().Add(-2 * time.Minute))}},
							},
						},
						"then": true,
						"else": false,
					},
				},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "namespaces",
				"localField":   "tenant_id",
				"foreignField": "tenant_id",
				"as":           "namespace",
			},
		},
		{
			"$addFields": bson.M{
				"namespace": "$namespace.name",
			},
		},
		{
			"$unwind": "$namespace",
		},
		{
			"$lookup": bson.M{
				"from":         "tags",
				"localField":   "tag_ids",
				"foreignField": "_id",
				"as":           "tags",
			},
		},
	}

	for _, opt := range opts {
		if err := opt(context.WithValue(ctx, "query", &query)); err != nil {
			return nil, err
		}
	}

	cursor, err := s.db.Collection("devices").Aggregate(ctx, query)
	if err != nil {
		return nil, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	cursor.Next(ctx)

	device := new(models.Device)
	if err := cursor.Decode(&device); err != nil {
		return nil, FromMongoError(err)
	}

	return device, nil
}

func (s *Store) DeviceCreate(ctx context.Context, device *models.Device) (string, error) {
	if _, err := s.db.Collection("devices").InsertOne(ctx, device); err != nil {
		return "", FromMongoError(err)
	}

	return device.UID, nil
}

func (s *Store) DeviceConflicts(ctx context.Context, target *models.DeviceConflicts) ([]string, bool, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"$or": []bson.M{
					{"name": target.Name},
				},
			},
		},
	}

	cursor, err := s.db.Collection("devices").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, false, FromMongoError(err)
	}
	defer cursor.Close(ctx)

	conflicts := make([]string, 0)
	for cursor.Next(ctx) {
		device := new(models.DeviceConflicts)
		if err := cursor.Decode(&device); err != nil {
			return nil, false, FromMongoError(err)
		}

		if device.Name == target.Name {
			conflicts = append(conflicts, "name")
		}
	}

	return conflicts, len(conflicts) > 0, nil
}

func (s *Store) DeviceUpdate(ctx context.Context, device *models.Device) error {
	bsonBytes, err := bson.Marshal(device)
	if err != nil {
		return FromMongoError(err)
	}

	doc := make(bson.M)
	if err := bson.Unmarshal(bsonBytes, &doc); err != nil {
		return FromMongoError(err)
	}

	// Convert string TagIDs to MongoDB ObjectIDs for referential integrity
	delete(doc, "tags")
	if tagIDs, ok := doc["tag_ids"].(bson.A); ok && len(tagIDs) > 0 {
		for i, id := range tagIDs {
			if idStr, ok := id.(string); ok {
				objID, _ := primitive.ObjectIDFromHex(idStr)
				tagIDs[i] = objID
			}
		}
	}

	filter := bson.M{"uid": device.UID, "tenant_id": device.TenantID}
	r, err := s.db.Collection("devices").UpdateOne(ctx, filter, bson.M{"$set": doc})
	if err != nil {
		return FromMongoError(err)
	}

	if r.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	if err := s.cache.Delete(ctx, "device"+"/"+device.UID); err != nil {
		logrus.WithError(err).WithField("uid", device.UID).Error("cannot delete device from cache")
	}

	return nil
}

func (s *Store) DeviceHeartbeat(ctx context.Context, uids []string, lastSeen time.Time) (int64, error) {
	filter := bson.M{"uid": bson.M{"$in": uids}}
	update := bson.M{"$set": bson.M{"last_seen": lastSeen, "disconnected_at": nil}}
	r, err := s.db.Collection("devices").UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, FromMongoError(err)
	}

	for _, uid := range uids {
		if err := s.cache.Delete(ctx, "device"+"/"+uid); err != nil {
			logrus.WithError(err).WithField("uid", uid).Error("cannot delete device from cache")
		}
	}

	return r.ModifiedCount, nil
}

func (s *Store) DeviceDelete(ctx context.Context, device *models.Device) error {
	deletedCount, err := s.DeviceDeleteMany(ctx, []string{device.UID})
	switch {
	case err != nil:
		return err
	case deletedCount < 1:
		return store.ErrNoDocuments
	default:
		return nil
	}
}

func (s *Store) DeviceDeleteMany(ctx context.Context, uids []string) (int64, error) {
	fn := s.deviceDeleteManyFn(uids)

	// Check if already inside a MongoDB transaction to avoid nested transactions.
	// Nested transactions cause WriteConflict errors.
	if mctx, ok := ctx.(mongo.SessionContext); ok {
		logrus.Debug("reusing existing MongoDB session from context")

		deletedCount, err := fn(mctx)
		if _, ok := deletedCount.(int64); !ok || err != nil {
			logrus.WithError(err).WithField("uids", uids).Error("device deletion failed in existing session")

			return 0, err
		}

		return deletedCount.(int64), nil
	} else { // nolint:revive
		logrus.WithField("uids", uids).Debug("creating new MongoDB session")

		mongoSession, err := s.db.Client().StartSession()
		if err != nil {
			return 0, FromMongoError(err)
		}
		defer mongoSession.EndSession(ctx)

		deletedCount, err := mongoSession.WithTransaction(ctx, fn)
		if _, ok := deletedCount.(int64); !ok || err != nil {
			logrus.WithError(err).Error("device deletion transaction failed")

			return 0, err
		}

		return deletedCount.(int64), nil
	}
}

func (s *Store) deviceDeleteManyFn(uids []string) func(ctx mongo.SessionContext) (any, error) {
	return func(mctx mongo.SessionContext) (any, error) {
		r, err := s.db.Collection("devices").DeleteMany(mctx, bson.M{"uid": bson.M{"$in": uids}})
		if err != nil {
			return int64(0), FromMongoError(err)
		}

		if _, err := s.db.Collection("sessions").DeleteMany(mctx, bson.M{"device_uid": bson.M{"$in": uids}}); err != nil {
			return int64(0), FromMongoError(err)
		}

		if _, err := s.db.Collection("tunnels").DeleteMany(mctx, bson.M{"device": bson.M{"$in": uids}}); err != nil {
			return int64(0), FromMongoError(err)
		}

		for _, uid := range uids {
			if err := s.cache.Delete(mctx, strings.Join([]string{"device", uid}, "/")); err != nil {
				logrus.WithError(err).WithField("uid", uid).Warn("deviceDeleteManyFn: cannot delete device from cache")
			}
		}

		return r.DeletedCount, nil
	}
}
