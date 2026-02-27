package migrate

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type mongoDevice struct {
	UID             string           `bson:"uid"`
	CreatedAt       time.Time        `bson:"created_at"`
	RemovedAt       *time.Time       `bson:"removed_at"`
	Name            string           `bson:"name"`
	Identity        *mongoDeviceID   `bson:"identity"`
	Info            *mongoDeviceInfo `bson:"info"`
	PublicKey       string           `bson:"public_key"`
	TenantID        string           `bson:"tenant_id"`
	LastSeen        time.Time        `bson:"last_seen"`
	DisconnectedAt  *time.Time       `bson:"disconnected_at"`
	Status          string           `bson:"status"`
	StatusUpdatedAt time.Time        `bson:"status_updated_at"`
	Position        *mongoDevicePos  `bson:"position"`
	TagIDs          []string         `bson:"tag_ids"`
}

type mongoDeviceID struct {
	MAC string `bson:"mac"`
}

type mongoDeviceInfo struct {
	ID         string `bson:"id"`
	PrettyName string `bson:"pretty_name"`
	Version    string `bson:"version"`
	Arch       string `bson:"arch"`
	Platform   string `bson:"platform"`
}

type mongoDevicePos struct {
	Latitude  float64 `bson:"latitude"`
	Longitude float64 `bson:"longitude"`
}

func convertDevice(doc mongoDevice) *entity.Device {
	status := doc.Status
	if status == "" {
		status = "pending"
	}

	e := &entity.Device{
		ID:              doc.UID,
		NamespaceID:     doc.TenantID,
		CreatedAt:       doc.CreatedAt,
		UpdatedAt:       time.Time{},
		RemovedAt:       doc.RemovedAt,
		LastSeen:        doc.LastSeen,
		Status:          status,
		StatusUpdatedAt: doc.StatusUpdatedAt,
		Name:            doc.Name,
		PublicKey:       doc.PublicKey,
	}

	if doc.DisconnectedAt != nil {
		e.DisconnectedAt = *doc.DisconnectedAt
	}
	if doc.Identity != nil {
		e.MAC = doc.Identity.MAC
	}
	if doc.Info != nil {
		e.Identifier = doc.Info.ID
		e.PrettyName = doc.Info.PrettyName
		e.Version = doc.Info.Version
		e.Arch = doc.Info.Arch
		e.Platform = doc.Info.Platform
	}
	if doc.Position != nil {
		e.Longitude = doc.Position.Longitude
		e.Latitude = doc.Position.Latitude
	}

	return e
}

func (m *Migrator) migrateDevices(ctx context.Context) error {
	cursor, err := m.mongo.Collection("devices").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	batch := make([]*entity.Device, 0, batchSize)
	total := 0

	for cursor.Next(ctx) {
		var doc mongoDevice
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		batch = append(batch, convertDevice(doc))
		if len(batch) >= batchSize {
			if _, err := m.pg.NewInsert().Model(&batch).Exec(ctx); err != nil {
				return err
			}
			total += len(batch)
			batch = batch[:0]
		}
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	if len(batch) > 0 {
		if _, err := m.pg.NewInsert().Model(&batch).Exec(ctx); err != nil {
			return err
		}
		total += len(batch)
	}

	log.WithField("count", total).Info("Migrated devices")

	return nil
}

func (m *Migrator) migrateDeviceTags(ctx context.Context) error {
	cursor, err := m.mongo.Collection("devices").Find(ctx, bson.M{"tag_ids": bson.M{"$exists": true, "$ne": bson.A{}}})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	batch := make([]*entity.DeviceTag, 0, batchSize)
	total := 0

	for cursor.Next(ctx) {
		var doc mongoDevice
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		for _, tagID := range doc.TagIDs {
			e := &entity.DeviceTag{
				DeviceID:  doc.UID,
				TagID:     ObjectIDToUUID(tagID),
				CreatedAt: doc.CreatedAt,
			}
			batch = append(batch, e)

			if len(batch) >= batchSize {
				if _, err := m.pg.NewInsert().Model(&batch).Exec(ctx); err != nil {
					return err
				}
				total += len(batch)
				batch = batch[:0]
			}
		}
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	if len(batch) > 0 {
		if _, err := m.pg.NewInsert().Model(&batch).Exec(ctx); err != nil {
			return err
		}
		total += len(batch)
	}

	log.WithField("count", total).Info("Migrated device_tags")

	return nil
}
