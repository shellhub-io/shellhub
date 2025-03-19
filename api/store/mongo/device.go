package mongo

import (
	"context" //nolint:gosec
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo/queries"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DeviceList returns a list of devices based on the given filters, pagination and sorting.
func (s *Store) DeviceList(ctx context.Context, status models.DeviceStatus, paginator query.Paginator, filters query.Filters, sorter query.Sorter, acceptable store.DeviceAcceptable) ([]models.Device, int, error) {
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
	}

	// Only match for the respective tenant if requested
	if tenant := gateway.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	if status != "" {
		query = append([]bson.M{{
			"$match": bson.M{
				"status": status,
			},
		}}, query...)
	}

	// When the listing mode is [store.DeviceListModeMaxDeviceReached], we should evaluate the `removed_devices`
	// collection to check its `accetable` status.
	switch acceptable {
	case store.DeviceAcceptableFromRemoved:
		query = append(query, []bson.M{
			{
				"$lookup": bson.M{
					"from":         "removed_devices",
					"localField":   "uid",
					"foreignField": "device.uid",
					"as":           "removed",
				},
			},
			{
				"$addFields": bson.M{
					"acceptable": bson.M{
						"$cond": bson.M{
							"if": bson.M{
								"$and": bson.A{
									bson.M{"$ne": bson.A{"$status", models.DeviceStatusAccepted}},
									bson.M{"$anyElementTrue": []interface{}{"$removed"}},
								},
							},
							"then": true,
							"else": false,
						},
					},
				},
			},
			{
				"$unset": "removed",
			},
		}...)
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

	queryMatch, err := queries.FromFilters(&filters)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}
	query = append(query, queryMatch...)

	queryCount := query
	queryCount = append(queryCount, bson.M{"$count": "count"})
	count, err := AggregateCount(ctx, s.db.Collection("devices"), queryCount)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	if sorter.By == "" {
		sorter.By = "last_seen"
	}

	query = append(query, queries.FromSorter(&sorter)...)
	query = append(query, queries.FromPaginator(&paginator)...)

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

func (s *Store) DeviceGet(ctx context.Context, uid models.UID) (*models.Device, error) {
	query := []bson.M{
		{
			"$match": bson.M{"uid": uid},
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
	}

	// Only match for the respective tenant if requested
	if tenant := gateway.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	device := new(models.Device)

	cursor, err := s.db.Collection("devices").Aggregate(ctx, query)
	if err != nil {
		return nil, FromMongoError(err)
	}
	defer cursor.Close(ctx)
	cursor.Next(ctx)

	err = cursor.Decode(&device)
	if err != nil {
		return nil, FromMongoError(err)
	}

	return device, nil
}

func (s *Store) DeviceDelete(ctx context.Context, uid models.UID) error {
	mongoSession, err := s.db.Client().StartSession()
	if err != nil {
		return FromMongoError(err)
	}
	defer mongoSession.EndSession(ctx)

	_, err = mongoSession.WithTransaction(ctx, func(_ mongo.SessionContext) (interface{}, error) {
		dev, err := s.db.Collection("devices").DeleteOne(ctx, bson.M{"uid": uid})
		if err != nil {
			return nil, FromMongoError(err)
		}

		if dev.DeletedCount < 1 {
			return nil, store.ErrNoDocuments
		}

		if err := s.cache.Delete(ctx, strings.Join([]string{"device", string(uid)}, "/")); err != nil {
			logrus.Error(err)
		}

		if _, err := s.db.Collection("sessions").DeleteMany(ctx, bson.M{"device_uid": uid}); err != nil {
			return nil, FromMongoError(err)
		}

		if _, err := s.db.Collection("tunnels").DeleteMany(ctx, bson.M{"device": uid}); err != nil {
			return nil, FromMongoError(err)
		}

		return nil, nil
	})

	return err
}

func (s *Store) DeviceCreate(ctx context.Context, d models.Device, hostname string) error {
	if hostname == "" {
		hostname = strings.ReplaceAll(d.Identity.MAC, ":", "-")
	}

	var dev *models.Device
	if err := s.cache.Get(ctx, strings.Join([]string{"device", d.UID}, "/"), &dev); err != nil {
		logrus.Error(err)
	}

	q := bson.M{
		"$setOnInsert": bson.M{
			"name":              hostname,
			"status":            "pending",
			"status_updated_at": time.Now(),
			"created_at":        clock.Now(),
			"tags":              []string{},
		},
		"$set": d,
	}
	opts := options.Update().SetUpsert(true)
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": d.UID}, q, opts)

	return FromMongoError(err)
}

func (s *Store) DeviceRename(ctx context.Context, uid models.UID, hostname string) error {
	dev, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"name": hostname}})
	if err != nil {
		return FromMongoError(err)
	}

	if dev.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) DeviceLookup(ctx context.Context, namespace, hostname string) (*models.Device, error) {
	ns := new(models.Namespace)
	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"name": namespace}).Decode(&ns); err != nil {
		return nil, FromMongoError(err)
	}

	device := new(models.Device)
	if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": ns.TenantID, "name": hostname, "status": "accepted"}).Decode(&device); err != nil {
		return nil, FromMongoError(err)
	}

	return device, nil
}

// DeviceUpdateStatus updates the status of a specific device in the devices collection
func (s *Store) DeviceUpdateStatus(ctx context.Context, uid models.UID, status models.DeviceStatus) error {
	updateOptions := options.FindOneAndUpdate().SetReturnDocument(options.After)
	result := s.db.Collection("devices", options.Collection()).
		FindOneAndUpdate(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"status": status, "status_updated_at": clock.Now()}}, updateOptions)

	if result.Err() != nil {
		return FromMongoError(result.Err())
	}

	device := new(models.Device)
	if err := result.Decode(&device); err != nil {
		return FromMongoError(err)
	}

	return nil
}

func (s *Store) DeviceListByUsage(ctx context.Context, tenant string) ([]models.UID, error) {
	query := []bson.M{
		{
			"$match": bson.M{
				"tenant_id": tenant,
			},
		},
		{
			"$group": bson.M{
				"_id": "$device_uid",
				"count": bson.M{
					"$sum": 1,
				},
			},
		},
		{
			"$sort": bson.M{
				"count": -1,
			},
		},
		{
			"$limit": 3,
		},
	}

	uids := make([]models.UID, 0)

	cursor, err := s.db.Collection("sessions").Aggregate(ctx, query)
	if err != nil {
		return uids, FromMongoError(err)
	}

	for cursor.Next(ctx) {
		var dev map[string]interface{}

		err = cursor.Decode(&dev)
		if err != nil {
			return uids, err
		}

		uids = append(uids, models.UID(dev["_id"].(string)))
	}

	return uids, nil
}

func (s *Store) DeviceGetByMac(ctx context.Context, mac string, tenantID string, status models.DeviceStatus) (*models.Device, error) {
	device := new(models.Device)

	switch status {
	case "":
		if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": tenantID, "identity": bson.M{"mac": mac}}).Decode(&device); err != nil {
			return nil, FromMongoError(err)
		}
	default:
		if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": tenantID, "status": status, "identity": bson.M{"mac": mac}}).Decode(&device); err != nil {
			return nil, FromMongoError(err)
		}
	}

	return device, nil
}

func (s *Store) DeviceGetByName(ctx context.Context, name string, tenantID string, status models.DeviceStatus) (*models.Device, error) {
	device := new(models.Device)

	if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": tenantID, "name": name, "status": string(status)}).Decode(&device); err != nil {
		return nil, FromMongoError(err)
	}

	return device, nil
}

func (s *Store) DeviceGetByUID(ctx context.Context, uid models.UID, tenantID string) (*models.Device, error) {
	var device *models.Device
	if err := s.cache.Get(ctx, strings.Join([]string{"device", string(uid)}, "/"), &device); err != nil {
		logrus.Error(err)
	}

	if device != nil {
		return device, nil
	}

	if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": tenantID, "uid": uid}).Decode(&device); err != nil {
		return nil, FromMongoError(err)
	}

	if err := s.cache.Set(ctx, strings.Join([]string{"device", string(uid)}, "/"), device, time.Minute); err != nil {
		logrus.Error(err)
	}

	return device, nil
}

func (s *Store) DeviceSetPosition(ctx context.Context, uid models.UID, position models.DevicePosition) error {
	dev, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"position": position}})
	if err != nil {
		return FromMongoError(err)
	}

	if dev.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	return nil
}

func (s *Store) DeviceChooser(ctx context.Context, tenantID string, chosen []string) error {
	filter := bson.M{
		"status":    "accepted",
		"tenant_id": tenantID,
		"uid": bson.M{
			"$nin": chosen,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"status": "pending",
		},
	}

	_, err := s.db.Collection("devices").UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
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

func (s *Store) DeviceUpdate(ctx context.Context, tenantID, uid string, changes *models.DeviceChanges) error {
	filter := bson.M{"uid": uid}
	if tenantID != "" {
		filter["tenant_id"] = tenantID
	}

	r, err := s.db.Collection("devices").UpdateMany(ctx, filter, bson.M{"$set": changes})
	if err != nil {
		return FromMongoError(err)
	}

	if r.MatchedCount < 1 {
		return store.ErrNoDocuments
	}

	if err := s.cache.Delete(ctx, "device"+uid+"/"); err != nil {
		logrus.WithError(err).WithField("uid", uid).Error("cannot delete device from cache")
	}

	return nil
}

func (s *Store) DeviceBulkUpdate(ctx context.Context, uids []string, changes *models.DeviceChanges) (int64, error) {
	res, err := s.db.Collection("devices").UpdateMany(ctx, bson.M{"uid": bson.M{"$in": uids}}, bson.M{"$set": changes})
	if err != nil {
		return 0, FromMongoError(err)
	}

	return res.ModifiedCount, nil
}

func (s *Store) DeviceRemovedCount(ctx context.Context, tenant string) (int64, error) {
	count, err := s.db.Collection("removed_devices").CountDocuments(ctx, bson.M{"device.tenant_id": tenant})
	if err != nil {
		return 0, FromMongoError(err)
	}

	return count, nil
}

func (s *Store) DeviceRemovedGet(ctx context.Context, tenant string, uid models.UID) (*models.DeviceRemoved, error) {
	var slot models.DeviceRemoved
	err := s.db.Collection("removed_devices").FindOne(ctx, bson.M{"device.tenant_id": tenant, "device.uid": uid}).Decode(&slot)
	if err != nil {
		return nil, FromMongoError(err)
	}

	return &slot, nil
}

func (s *Store) DeviceRemovedInsert(ctx context.Context, tenant string, device *models.Device) error { //nolint:revive
	now := time.Now()

	device.Status = models.DeviceStatusRemoved
	device.StatusUpdatedAt = now

	_, err := s.db.Collection("removed_devices").InsertOne(ctx, models.DeviceRemoved{
		Timestamp: now,
		Device:    device,
	})
	if err != nil {
		return FromMongoError(err)
	}

	return nil
}

func (s *Store) DeviceRemovedDelete(ctx context.Context, tenant string, uid models.UID) error {
	_, err := s.db.Collection("removed_devices").DeleteOne(ctx, bson.M{"device.tenant_id": tenant, "device.uid": uid})
	if err != nil {
		return FromMongoError(err)
	}

	return nil
}

func (s *Store) DeviceRemovedList(ctx context.Context, tenant string, paginator query.Paginator, filters query.Filters, sorter query.Sorter) ([]models.DeviceRemoved, int, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"device.tenant_id": tenant,
			},
		},
	}

	pipeline = append(pipeline, queries.FromPaginator(&paginator)...)

	queryFilter, err := queries.FromFilters(&filters)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	pipeline = append(pipeline, queryFilter...)

	if sorter.By == "" {
		sorter.By = "timestamp"
	}
	if sorter.Order == "" {
		sorter.Order = query.OrderDesc
	}
	pipeline = append(pipeline, queries.FromSorter(&sorter)...)

	aggregation, err := s.db.Collection("removed_devices").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, FromMongoError(err)
	}

	var devices []models.DeviceRemoved
	if err := aggregation.All(ctx, &devices); err != nil {
		return nil, 0, FromMongoError(err)
	}

	return devices, len(devices), nil
}
