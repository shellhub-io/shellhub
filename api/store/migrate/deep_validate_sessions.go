package migrate

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/uptrace/bun"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *Migrator) deepValidateSessions(ctx context.Context, r *ValidationReport) error {
	validDevices, err := m.loadValidDevices(ctx)
	if err != nil {
		return err
	}

	cursor, err := m.mongo.Collection("sessions").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx) //nolint:errcheck

	batch := make([]mongoSession, 0, batchSize)

	for cursor.Next(ctx) {
		var doc mongoSession
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		// Skip sessions whose device was not migrated (orphaned).
		if _, ok := validDevices[doc.DeviceUID]; !ok {
			continue
		}

		batch = append(batch, doc)
		if len(batch) >= batchSize {
			if err := m.compareSessionBatch(ctx, r, batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	if len(batch) > 0 {
		return m.compareSessionBatch(ctx, r, batch)
	}

	return nil
}

func (m *Migrator) compareSessionBatch(ctx context.Context, r *ValidationReport, batch []mongoSession) error {
	ids := make([]string, len(batch))
	expected := make(map[string]*entity.Session, len(batch))
	for i, doc := range batch {
		e := convertSession(doc)
		ids[i] = e.ID
		expected[e.ID] = e
	}

	var actual []entity.Session
	if err := m.pg.NewSelect().Model(&actual).Where("id IN (?)", bun.List(ids)).Scan(ctx); err != nil {
		return err
	}

	pgMap := make(map[string]*entity.Session, len(actual))
	for i := range actual {
		pgMap[actual[i].ID] = &actual[i]
	}

	r.AddCompared("sessions", int64(len(batch)))

	for _, id := range ids {
		exp := expected[id]
		act, ok := pgMap[id]
		if !ok {
			r.AddMissing("sessions", id)

			continue
		}

		t := "sessions"
		r.CheckField(t, id, "DeviceID", exp.DeviceID, act.DeviceID)
		r.CheckField(t, id, "Username", exp.Username, act.Username)
		r.CheckField(t, id, "IPAddress", exp.IPAddress, act.IPAddress)
		r.CheckTime(t, id, "StartedAt", exp.StartedAt, act.StartedAt)
		r.CheckTime(t, id, "SeenAt", exp.SeenAt, act.SeenAt)
		r.CheckField(t, id, "Closed", exp.Closed, act.Closed)
		r.CheckField(t, id, "Authenticated", exp.Authenticated, act.Authenticated)
		r.CheckField(t, id, "Recorded", exp.Recorded, act.Recorded)
		r.CheckField(t, id, "Type", exp.Type, act.Type)
		r.CheckField(t, id, "Term", exp.Term, act.Term)
		r.CheckFloat(t, id, "Longitude", exp.Longitude, act.Longitude)
		r.CheckFloat(t, id, "Latitude", exp.Latitude, act.Latitude)
		r.CheckTime(t, id, "CreatedAt", exp.CreatedAt, act.CreatedAt)
		r.CheckTime(t, id, "UpdatedAt", exp.UpdatedAt, act.UpdatedAt)
	}

	return nil
}
