package migrate

import (
	"context"
	"strings"
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
	role := strings.ToLower(member.Role)
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
	validUsers, err := m.loadValidUsers(ctx)
	if err != nil {
		return err
	}

	cursor, err := m.mongo.Collection("namespaces").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	batch := make([]*entity.Namespace, 0, batchSize)
	total := 0
	skipped := 0

	for cursor.Next(ctx) {
		var doc mongoNamespace
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		ns := convertNamespace(doc)
		if _, ok := validUsers[ns.OwnerID]; !ok {
			log.WithFields(log.Fields{
				"scope":     "core",
				"owner":     ns.OwnerID,
				"namespace": doc.TenantID,
				"name":      doc.Name,
			}).Warn("Skipping namespace with orphaned owner")
			skipped++

			continue
		}

		batch = append(batch, ns)
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

	if skipped > 0 {
		m.addOrphans("namespaces", skipped)
	}

	log.WithFields(log.Fields{
		"scope":   "core",
		"count":   total,
		"skipped": skipped,
	}).Info("Migrated namespaces")

	return nil
}

func (m *Migrator) loadValidUsers(ctx context.Context) (map[string]struct{}, error) {
	var ids []struct {
		ID string `bun:"id"`
	}
	if err := m.pg.NewSelect().TableExpr("users").Column("id").Scan(ctx, &ids); err != nil {
		return nil, err
	}

	valid := make(map[string]struct{}, len(ids))
	for _, u := range ids {
		valid[u.ID] = struct{}{}
	}

	return valid, nil
}

func (m *Migrator) migrateMemberships(ctx context.Context) error {
	validUsers, err := m.loadValidUsers(ctx)
	if err != nil {
		return err
	}

	validNamespaces, err := m.loadValidNamespaces(ctx)
	if err != nil {
		return err
	}

	cursor, err := m.mongo.Collection("namespaces").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	batch := make([]*entity.Membership, 0, batchSize)
	total := 0
	skipped := 0

	for cursor.Next(ctx) {
		var doc mongoNamespace
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		if _, ok := validNamespaces[doc.TenantID]; !ok {
			skipped += len(doc.Members)

			continue
		}

		for _, member := range doc.Members {
			mb := convertMembership(doc.TenantID, member)
			if _, ok := validUsers[mb.UserID]; !ok {
				log.WithFields(log.Fields{
					"scope":     "core",
					"user":      mb.UserID,
					"namespace": doc.TenantID,
				}).Warn("Skipping membership with orphaned user")
				skipped++

				continue
			}

			batch = append(batch, mb)

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

	if skipped > 0 {
		m.addOrphans("memberships", skipped)
	}

	log.WithFields(log.Fields{
		"scope":   "core",
		"count":   total,
		"skipped": skipped,
	}).Info("Migrated memberships")

	return nil
}
