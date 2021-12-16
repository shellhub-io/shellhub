package mongo

import (
	"context"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *Store) DeviceList(ctx context.Context, pagination paginator.Query, filters []models.Filter, status string, sort string, order string) ([]models.Device, int, error) {
	queryMatch, err := buildFilterQuery(filters)
	if err != nil {
		return nil, 0, fromMongoError(err)
	}

	query := []bson.M{
		{

			"$lookup": bson.M{
				"from":         "connected_devices",
				"localField":   "uid",
				"foreignField": "uid",
				"as":           "online",
			},
		},
		{
			"$addFields": bson.M{
				"online": bson.M{"$anyElementTrue": []interface{}{"$online"}},
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

	if status != "" {
		query = append([]bson.M{{
			"$match": bson.M{
				"status": status,
			},
		}}, query...)
	}

	orderVal := map[string]int{
		"asc":  1,
		"desc": -1,
	}

	if sort != "" {
		query = append(query, bson.M{
			"$sort": bson.M{sort: orderVal[order]},
		})
	} else {
		query = append(query, bson.M{
			"$sort": bson.M{"last_seen": -1},
		})
	}

	// Apply filters if any
	if len(queryMatch) > 0 {
		query = append(query, queryMatch...)
	}

	// Only match for the respective tenant if requested
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	queryCount := query
	queryCount = append(queryCount, bson.M{"$count": "count"})
	count, err := aggregateCount(ctx, s.db.Collection("devices"), queryCount)
	if err != nil {
		return nil, 0, fromMongoError(err)
	}

	query = append(query, buildPaginationQuery(pagination)...)

	devices := make([]models.Device, 0)

	cursor, err := s.db.Collection("devices").Aggregate(ctx, query)
	if err != nil {
		return devices, count, fromMongoError(err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		device := new(models.Device)
		err = cursor.Decode(&device)
		if err != nil {
			return devices, count, err
		}
		devices = append(devices, *device)
	}

	return devices, count, fromMongoError(err)
}

func (s *Store) DeviceGet(ctx context.Context, uid models.UID) (*models.Device, error) {
	query := []bson.M{
		{
			"$match": bson.M{"uid": uid},
		},
		{
			"$lookup": bson.M{
				"from":         "connected_devices",
				"localField":   "uid",
				"foreignField": "uid",
				"as":           "online",
			},
		},
		{
			"$addFields": bson.M{
				"online": bson.M{"$anyElementTrue": []interface{}{"$online"}},
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
	if tenant := apicontext.TenantFromContext(ctx); tenant != nil {
		query = append(query, bson.M{
			"$match": bson.M{
				"tenant_id": tenant.ID,
			},
		})
	}

	device := new(models.Device)

	cursor, err := s.db.Collection("devices").Aggregate(ctx, query)
	if err != nil {
		return nil, fromMongoError(err)
	}
	defer cursor.Close(ctx)
	cursor.Next(ctx)

	err = cursor.Decode(&device)
	if err != nil {
		return nil, fromMongoError(err)
	}

	return device, nil
}

func (s *Store) DeviceDelete(ctx context.Context, uid models.UID) error {
	if _, err := s.db.Collection("devices").DeleteOne(ctx, bson.M{"uid": uid}); err != nil {
		return fromMongoError(err)
	}

	if err := s.cache.Delete(ctx, string(uid)); err != nil {
		logrus.Error(err)
	}

	if _, err := s.db.Collection("sessions").DeleteMany(ctx, bson.M{"device_uid": uid}); err != nil {
		return fromMongoError(err)
	}

	_, err := s.db.Collection("connected_devices").DeleteMany(ctx, bson.M{"uid": uid})

	return fromMongoError(err)
}

func (s *Store) DeviceCreate(ctx context.Context, d models.Device, hostname string) error {
	mac := strings.ReplaceAll(d.Identity.MAC, ":", "-")
	if hostname == "" {
		hostname = mac
	}

	var dev *models.Device
	if err := s.cache.Get(ctx, strings.Join([]string{"device", d.UID}, "/"), &dev); err != nil {
		logrus.Error(err)
	}

	q := bson.M{
		"$setOnInsert": bson.M{
			"name":       hostname,
			"status":     "pending",
			"created_at": clock.Now(),
			"tags":       []string{},
		},
		"$set": d,
	}
	opts := options.Update().SetUpsert(true)
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": d.UID}, q, opts)

	return fromMongoError(err)
}

func (s *Store) DeviceRename(ctx context.Context, uid models.UID, hostname string) error {
	if _, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"name": hostname}}); err != nil {
		return fromMongoError(err)
	}

	return nil
}

func (s *Store) DeviceLookup(ctx context.Context, namespace, hostname string) (*models.Device, error) {
	ns := new(models.Namespace)
	if err := s.db.Collection("namespaces").FindOne(ctx, bson.M{"name": namespace}).Decode(&ns); err != nil {
		return nil, fromMongoError(err)
	}

	device := new(models.Device)
	if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": ns.TenantID, "name": hostname, "status": "accepted"}).Decode(&device); err != nil {
		return nil, fromMongoError(err)
	}

	return device, nil
}

func (s *Store) DeviceSetOnline(ctx context.Context, uid models.UID, online bool) error {
	device := new(models.Device)
	if err := s.db.Collection("devices").FindOne(ctx, bson.M{"uid": uid}).Decode(&device); err != nil {
		return fromMongoError(err)
	}

	if !online {
		_, err := s.db.Collection("connected_devices").DeleteMany(ctx, bson.M{"uid": uid})

		return fromMongoError(err)
	}

	device.LastSeen = clock.Now()
	opts := options.Update().SetUpsert(true)
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": device.UID}, bson.M{"$set": bson.M{"last_seen": device.LastSeen}}, opts)
	if err != nil {
		return fromMongoError(err)
	}

	cd := &models.ConnectedDevice{
		UID:      device.UID,
		TenantID: device.TenantID,
		LastSeen: clock.Now(),
		Status:   device.Status,
	}

	if _, err := s.db.Collection("connected_devices").InsertOne(ctx, &cd); err != nil {
		return fromMongoError(err)
	}

	return nil
}

func (s *Store) DeviceUpdateOnline(ctx context.Context, uid models.UID, online bool) error {
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"online": online}})

	return fromMongoError(err)
}

func (s *Store) DeviceUpdateLastSeen(ctx context.Context, uid models.UID, ts time.Time) error {
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"last_seen": ts}})

	return fromMongoError(err)
}

func (s *Store) DeviceUpdateStatus(ctx context.Context, uid models.UID, status string) error {
	device := new(models.Device)
	if err := s.db.Collection("devices").FindOne(ctx, bson.M{"uid": uid}).Decode(&device); err != nil {
		return fromMongoError(err)
	}

	opts := options.Update().SetUpsert(true)
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": device.UID}, bson.M{"$set": bson.M{"status": status}}, opts)
	if err != nil {
		return fromMongoError(err)
	}

	cd := &models.ConnectedDevice{
		UID:      device.UID,
		TenantID: device.TenantID,
		LastSeen: clock.Now(),
		Status:   status,
	}

	if _, err := s.db.Collection("connected_devices").InsertOne(ctx, &cd); err != nil {
		return fromMongoError(err)
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
		return uids, fromMongoError(err)
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

func (s *Store) DeviceGetByMac(ctx context.Context, mac string, tenantID string, status string) (*models.Device, error) {
	device := new(models.Device)
	if status != "" {
		if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": tenantID, "identity": bson.M{"mac": mac}, "status": status}).Decode(&device); err != nil {
			return nil, fromMongoError(err)
		}
	} else {
		if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": tenantID, "identity": bson.M{"mac": mac}}).Decode(&device); err != nil {
			return nil, fromMongoError(err)
		}
	}

	return device, nil
}

func (s *Store) DeviceGetByName(ctx context.Context, name string, tenantID string) (*models.Device, error) {
	device := new(models.Device)
	if err := s.db.Collection("devices").FindOne(ctx, bson.M{"tenant_id": tenantID, "name": name}).Decode(&device); err != nil {
		return nil, fromMongoError(err)
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
		return nil, fromMongoError(err)
	}

	if err := s.cache.Set(ctx, strings.Join([]string{"device", string(uid)}, "/"), device, time.Minute); err != nil {
		logrus.Error(err)
	}

	return device, nil
}

func (s *Store) DeviceSetPosition(ctx context.Context, uid models.UID, position models.DevicePosition) error {
	_, err := s.db.Collection("devices").UpdateOne(ctx, bson.M{"uid": uid}, bson.M{"$set": bson.M{"position": position}})

	return err
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
