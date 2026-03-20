package migrate

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	log "github.com/sirupsen/logrus"
	"github.com/uptrace/bun"
	"go.mongodb.org/mongo-driver/bson"
)

const sessionChunkSize = 500

func (m *Migrator) deepValidateSessionEvents(ctx context.Context, r *ValidationReport) error {
	var sessionIDs []string
	if err := m.pg.NewSelect().
		TableExpr("sessions").
		Column("id").
		OrderExpr("id ASC").
		Scan(ctx, &sessionIDs); err != nil {
		return err
	}

	for i := 0; i < len(sessionIDs); i += sessionChunkSize {
		end := i + sessionChunkSize
		if end > len(sessionIDs) {
			end = len(sessionIDs)
		}

		if err := m.compareSessionEventChunk(ctx, r, sessionIDs[i:end]); err != nil {
			return err
		}

		if end%5000 == 0 || end == len(sessionIDs) {
			log.WithFields(log.Fields{
				"scope":    "core",
				"progress": fmt.Sprintf("%d/%d", end, len(sessionIDs)),
			}).Info("Session events deep validation progress")
		}
	}

	return nil
}

// normalizedEvent holds the comparable fields of a session event.
type normalizedEvent struct {
	Type      string
	Seat      int
	Data      string
	CreatedAt string
}

func normalizeMongoEvent(doc mongoSessionEvent) normalizedEvent {
	return normalizedEvent{
		Type:      doc.Type,
		Seat:      doc.Seat,
		Data:      normalizeJSON(convertSessionEventData(doc.Data)),
		CreatedAt: doc.Timestamp.Truncate(1e6).UTC().String(),
	}
}

func normalizePGEvent(e *entity.SessionEvent) normalizedEvent {
	return normalizedEvent{
		Type:      e.Type,
		Seat:      e.Seat,
		Data:      normalizeJSON(e.Data),
		CreatedAt: e.CreatedAt.Truncate(1e6).UTC().String(),
	}
}

func (m *Migrator) compareSessionEventChunk(ctx context.Context, r *ValidationReport, sessionIDs []string) error {
	cursor, err := m.mongo.Collection("sessions_events").Find(ctx, bson.M{
		"session": bson.M{"$in": sessionIDs},
	})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx) //nolint:errcheck

	mongoBySession := make(map[string][]mongoSessionEvent)
	for cursor.Next(ctx) {
		var doc mongoSessionEvent
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		mongoBySession[doc.Session] = append(mongoBySession[doc.Session], doc)
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	var pgEvents []entity.SessionEvent
	if err := m.pg.NewSelect().
		Model(&pgEvents).
		Where("session_id IN (?)", bun.List(sessionIDs)).
		Scan(ctx); err != nil {
		return err
	}

	pgBySession := make(map[string][]entity.SessionEvent)
	for _, e := range pgEvents {
		pgBySession[e.SessionID] = append(pgBySession[e.SessionID], e)
	}

	for _, sid := range sessionIDs {
		mongoDocs := mongoBySession[sid]
		pgDocs := pgBySession[sid]

		r.AddCompared("session_events", int64(len(mongoDocs)))

		if len(mongoDocs) != len(pgDocs) {
			r.AddMismatch(FieldMismatch{
				Table:    "session_events",
				RecordID: fmt.Sprintf("session:%s", sid),
				Field:    "count",
				Expected: fmt.Sprintf("%d", len(mongoDocs)),
				Actual:   fmt.Sprintf("%d", len(pgDocs)),
			})

			continue
		}

		// Build a multiset from Mongo events and match against PG.
		expected := make(map[string]int)
		for _, doc := range mongoDocs {
			key := eventKey(normalizeMongoEvent(doc))
			expected[key]++
		}

		for _, e := range pgDocs {
			key := eventKey(normalizePGEvent(&e))
			if expected[key] > 0 {
				expected[key]--
			} else {
				r.AddMismatch(FieldMismatch{
					Table:    "session_events",
					RecordID: fmt.Sprintf("session:%s", sid),
					Field:    "event",
					Expected: "(not in mongo)",
					Actual:   key,
				})
			}
		}

		for key, count := range expected {
			if count > 0 {
				r.AddMismatch(FieldMismatch{
					Table:    "session_events",
					RecordID: fmt.Sprintf("session:%s", sid),
					Field:    "event",
					Expected: key,
					Actual:   fmt.Sprintf("(missing %d in pg)", count),
				})
			}
		}
	}

	return nil
}

func eventKey(e normalizedEvent) string {
	return fmt.Sprintf("%s|%d|%s|%s", e.Type, e.Seat, e.CreatedAt, e.Data)
}

func convertSessionEventData(data any) string {
	if data == nil {
		return ""
	}

	b, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	return string(b)
}

// normalizeJSON unmarshals and re-marshals JSON to ensure consistent
// key ordering for comparison.
func normalizeJSON(s string) string {
	if s == "" {
		return ""
	}

	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return s
	}

	b, err := json.Marshal(v)
	if err != nil {
		return s
	}

	return string(b)
}
