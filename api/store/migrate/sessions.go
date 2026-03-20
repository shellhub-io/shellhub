package migrate

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid" //nolint:depguard // migration package generates UUIDs directly
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type mongoSession struct {
	UID           string           `bson:"uid"`
	DeviceUID     string           `bson:"device_uid"`
	TenantID      string           `bson:"tenant_id"`
	Username      string           `bson:"username"`
	IPAddress     string           `bson:"ip_address"`
	StartedAt     time.Time        `bson:"started_at"`
	LastSeen      time.Time        `bson:"last_seen"`
	Closed        bool             `bson:"closed"`
	Authenticated bool             `bson:"authenticated"`
	Recorded      bool             `bson:"recorded"`
	Type          string           `bson:"type"`
	Term          string           `bson:"term"`
	Position      *mongoSessionPos `bson:"position"`
}

type mongoSessionPos struct {
	Longitude float64 `bson:"longitude"`
	Latitude  float64 `bson:"latitude"`
}

func convertSession(doc mongoSession) *entity.Session {
	sessionType := doc.Type
	if sessionType == "" {
		sessionType = "shell"
	}

	e := &entity.Session{
		ID:            doc.UID,
		NamespaceID:   doc.TenantID,
		DeviceID:      doc.DeviceUID,
		Username:      doc.Username,
		IPAddress:     doc.IPAddress,
		StartedAt:     doc.StartedAt,
		SeenAt:        doc.LastSeen,
		Closed:        doc.Closed,
		Authenticated: doc.Authenticated,
		Recorded:      doc.Recorded,
		Type:          sessionType,
		Term:          doc.Term,
		CreatedAt:     doc.StartedAt,
		UpdatedAt:     doc.LastSeen,
	}

	if doc.Position != nil {
		e.Longitude = doc.Position.Longitude
		e.Latitude = doc.Position.Latitude
	}

	return e
}

func convertSessionEvent(doc mongoSessionEvent) *entity.SessionEvent {
	var data string
	if doc.Data != nil {
		if dataBytes, err := json.Marshal(doc.Data); err == nil {
			data = string(dataBytes)
		}
	}

	return &entity.SessionEvent{
		ID:        uuid.New().String(),
		SessionID: doc.Session,
		Type:      doc.Type,
		Seat:      doc.Seat,
		Data:      data,
		CreatedAt: doc.Timestamp,
	}
}

func (m *Migrator) loadValidDevices(ctx context.Context) (map[string]struct{}, error) {
	var ids []struct {
		ID string `bun:"id"`
	}
	if err := m.pg.NewSelect().TableExpr("devices").Column("id").Scan(ctx, &ids); err != nil {
		return nil, err
	}

	valid := make(map[string]struct{}, len(ids))
	for _, d := range ids {
		valid[d.ID] = struct{}{}
	}

	return valid, nil
}

func (m *Migrator) migrateSessions(ctx context.Context) error {
	validDevices, err := m.loadValidDevices(ctx)
	if err != nil {
		return err
	}

	cursor, err := m.mongo.Collection("sessions").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	batch := make([]*entity.Session, 0, batchSize)
	total := 0
	skipped := 0

	for cursor.Next(ctx) {
		var doc mongoSession
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		if _, ok := validDevices[doc.DeviceUID]; !ok {
			log.WithFields(log.Fields{
				"scope":   "core",
				"session": doc.UID,
				"device":  doc.DeviceUID,
			}).Warn("Skipping session with orphaned device")
			skipped++

			continue
		}

		batch = append(batch, convertSession(doc))
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
		m.addOrphans("sessions", skipped)
	}

	log.WithFields(log.Fields{
		"scope":   "core",
		"count":   total,
		"skipped": skipped,
	}).Info("Migrated sessions")

	return nil
}

type mongoSessionEvent struct {
	Session   string    `bson:"session"`
	Type      string    `bson:"type"`
	Timestamp time.Time `bson:"timestamp"`
	Data      any       `bson:"data"`
	Seat      int       `bson:"seat"`
}

func (m *Migrator) loadValidSessions(ctx context.Context) (map[string]struct{}, error) {
	var ids []struct {
		ID string `bun:"id"`
	}
	if err := m.pg.NewSelect().TableExpr("sessions").Column("id").Scan(ctx, &ids); err != nil {
		return nil, err
	}

	valid := make(map[string]struct{}, len(ids))
	for _, s := range ids {
		valid[s.ID] = struct{}{}
	}

	return valid, nil
}

const sessionEventBatchSize = 5000

func (m *Migrator) migrateSessionEvents(ctx context.Context) error {
	validSessions, err := m.loadValidSessions(ctx)
	if err != nil {
		return err
	}

	// Disable triggers for bulk insert performance.
	if _, err := m.pg.ExecContext(ctx, "ALTER TABLE session_events DISABLE TRIGGER ALL"); err != nil {
		return err
	}

	defer func() {
		if _, err := m.pg.ExecContext(ctx, "ALTER TABLE session_events ENABLE TRIGGER ALL"); err != nil {
			log.WithError(err).WithField("scope", "core").Error("Failed to re-enable triggers on session_events")
		}
	}()

	cursor, err := m.mongo.Collection("sessions_events").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	batch := make([]*entity.SessionEvent, 0, sessionEventBatchSize)
	total := 0
	skipped := 0

	for cursor.Next(ctx) {
		var doc mongoSessionEvent
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		if _, ok := validSessions[doc.Session]; !ok {
			skipped++

			continue
		}

		batch = append(batch, convertSessionEvent(doc))
		if len(batch) >= sessionEventBatchSize {
			if _, err := m.pg.NewInsert().Model(&batch).Exec(ctx); err != nil {
				return err
			}
			total += len(batch)

			if total%10000 == 0 {
				log.WithFields(log.Fields{"scope": "core", "count": total}).Info("Session events migration progress")
			}

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
		m.addOrphans("session_events", skipped)
	}

	log.WithFields(log.Fields{
		"scope":   "core",
		"count":   total,
		"skipped": skipped,
	}).Info("Migrated session_events")

	return nil
}
