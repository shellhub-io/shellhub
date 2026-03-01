package migrate

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type mongoNamespace struct {
	TenantID             string           `bson:"tenant_id"`
	Name                 string           `bson:"name"`
	Owner                string           `bson:"owner"`
	Type                 string           `bson:"type"`
	MaxDevices           int              `bson:"max_devices"`
	CreatedAt            time.Time        `bson:"created_at"`
	DevicesAcceptedCount int64            `bson:"devices_accepted_count"`
	DevicesPendingCount  int64            `bson:"devices_pending_count"`
	DevicesRejectedCount int64            `bson:"devices_rejected_count"`
	DevicesRemovedCount  int64            `bson:"devices_removed_count"`
	Settings             *mongoNSSettings `bson:"settings"`
	Members              []mongoMember    `bson:"members"`
}

type mongoNSSettings struct {
	SessionRecord          bool   `bson:"session_record"`
	ConnectionAnnouncement string `bson:"connection_announcement"`
}

type mongoMember struct {
	ID      string    `bson:"id"`
	AddedAt time.Time `bson:"added_at"`
	Role    string    `bson:"role"`
}

func convertNamespace(doc mongoNamespace) *entity.Namespace {
	nsType := doc.Type
	if nsType == "" {
		nsType = "personal"
	}

	e := &entity.Namespace{
		ID:                   doc.TenantID,
		CreatedAt:            doc.CreatedAt,
		UpdatedAt:            time.Time{},
		Type:                 nsType,
		Name:                 doc.Name,
		OwnerID:              ObjectIDToUUID(doc.Owner),
		DevicesAcceptedCount: doc.DevicesAcceptedCount,
		DevicesPendingCount:  doc.DevicesPendingCount,
		DevicesRejectedCount: doc.DevicesRejectedCount,
		DevicesRemovedCount:  doc.DevicesRemovedCount,
		Settings: entity.NamespaceSettings{
			MaxDevices: doc.MaxDevices,
		},
	}

	if doc.Settings != nil {
		e.Settings.SessionRecord = doc.Settings.SessionRecord
		e.Settings.ConnectionAnnouncement = doc.Settings.ConnectionAnnouncement
	}

	return e
}

func convertMembership(tenantID string, member mongoMember) *entity.Membership {
	role := member.Role
	if role == "" {
		role = "observer"
	}

	return &entity.Membership{
		UserID:      ObjectIDToUUID(member.ID),
		NamespaceID: tenantID,
		CreatedAt:   member.AddedAt,
		UpdatedAt:   time.Time{},
		Role:        role,
	}
}

func (m *Migrator) migrateNamespaces(ctx context.Context) error {
	cursor, err := m.mongo.Collection("namespaces").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	batch := make([]*entity.Namespace, 0, batchSize)
	total := 0

	for cursor.Next(ctx) {
		var doc mongoNamespace
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		batch = append(batch, convertNamespace(doc))
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

	log.WithField("count", total).Info("Migrated namespaces")

	return nil
}

func (m *Migrator) migrateMemberships(ctx context.Context) error {
	cursor, err := m.mongo.Collection("namespaces").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	batch := make([]*entity.Membership, 0, batchSize)
	total := 0

	for cursor.Next(ctx) {
		var doc mongoNamespace
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		for _, member := range doc.Members {
			batch = append(batch, convertMembership(doc.TenantID, member))

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

	log.WithField("count", total).Info("Migrated memberships")

	return nil
}
