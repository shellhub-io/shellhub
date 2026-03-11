package migrate

import (
	"context"
	"sort"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/uptrace/bun"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *Migrator) deepValidateDevices(ctx context.Context, r *ValidationReport) error {
	cursor, err := m.mongo.Collection("devices").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx) //nolint:errcheck

	batch := make([]mongoDevice, 0, batchSize)

	for cursor.Next(ctx) {
		var doc mongoDevice
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		batch = append(batch, doc)
		if len(batch) >= batchSize {
			if err := m.compareDeviceBatch(ctx, r, batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	if len(batch) > 0 {
		return m.compareDeviceBatch(ctx, r, batch)
	}

	return nil
}

func (m *Migrator) compareDeviceBatch(ctx context.Context, r *ValidationReport, batch []mongoDevice) error {
	ids := make([]string, len(batch))
	expected := make(map[string]*entity.Device, len(batch))
	for i, doc := range batch {
		e := convertDevice(doc)
		ids[i] = e.ID
		expected[e.ID] = e
	}

	var actual []entity.Device
	if err := m.pg.NewSelect().Model(&actual).Where("id IN (?)", bun.List(ids)).Scan(ctx); err != nil {
		return err
	}

	pgMap := make(map[string]*entity.Device, len(actual))
	for i := range actual {
		pgMap[actual[i].ID] = &actual[i]
	}

	r.AddCompared("devices", int64(len(batch)))

	for _, id := range ids {
		exp := expected[id]
		act, ok := pgMap[id]
		if !ok {
			r.AddMissing("devices", id)

			continue
		}

		t := "devices"
		r.CheckField(t, id, "NamespaceID", exp.NamespaceID, act.NamespaceID)
		r.CheckTime(t, id, "CreatedAt", exp.CreatedAt, act.CreatedAt)
		r.CheckTimePtr(t, id, "RemovedAt", exp.RemovedAt, act.RemovedAt)
		r.CheckTime(t, id, "LastSeen", exp.LastSeen, act.LastSeen)
		r.CheckTime(t, id, "DisconnectedAt", exp.DisconnectedAt, act.DisconnectedAt)
		r.CheckField(t, id, "Status", exp.Status, act.Status)
		r.CheckTime(t, id, "StatusUpdatedAt", exp.StatusUpdatedAt, act.StatusUpdatedAt)
		r.CheckField(t, id, "Name", exp.Name, act.Name)
		r.CheckField(t, id, "MAC", exp.MAC, act.MAC)
		r.CheckField(t, id, "PublicKey", exp.PublicKey, act.PublicKey)
		r.CheckField(t, id, "Identifier", exp.Identifier, act.Identifier)
		r.CheckField(t, id, "PrettyName", exp.PrettyName, act.PrettyName)
		r.CheckField(t, id, "Version", exp.Version, act.Version)
		r.CheckField(t, id, "Arch", exp.Arch, act.Arch)
		r.CheckField(t, id, "Platform", exp.Platform, act.Platform)
		r.CheckFloat(t, id, "Longitude", exp.Longitude, act.Longitude)
		r.CheckFloat(t, id, "Latitude", exp.Latitude, act.Latitude)
	}

	return nil
}

func (m *Migrator) deepValidateDeviceTags(ctx context.Context, r *ValidationReport) error {
	cursor, err := m.mongo.Collection("devices").Find(ctx, bson.M{
		"tag_ids": bson.M{"$exists": true, "$ne": bson.A{}},
	})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx) //nolint:errcheck

	for cursor.Next(ctx) {
		var doc mongoDevice
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		expectedTags := make([]string, len(doc.TagIDs))
		for i, tagID := range doc.TagIDs {
			expectedTags[i] = ObjectIDToUUID(tagID)
		}
		sort.Strings(expectedTags)

		var actualTags []entity.DeviceTag
		if err := m.pg.NewSelect().Model(&actualTags).Where("device_id = ?", doc.UID).Scan(ctx); err != nil {
			return err
		}

		actualTagIDs := make([]string, len(actualTags))
		for i, dt := range actualTags {
			actualTagIDs[i] = dt.TagID
		}
		sort.Strings(actualTagIDs)

		r.AddCompared("device_tags", 1)
		r.CheckStrings("device_tags", doc.UID, "TagIDs", expectedTags, actualTagIDs)
	}

	return cursor.Err()
}
