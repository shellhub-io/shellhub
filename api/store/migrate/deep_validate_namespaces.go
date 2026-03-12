package migrate

import (
	"context"
	"database/sql"
	"errors"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/uptrace/bun"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *Migrator) deepValidateNamespaces(ctx context.Context, r *ValidationReport) error {
	cursor, err := m.mongo.Collection("namespaces").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx) //nolint:errcheck

	batch := make([]mongoNamespace, 0, batchSize)

	for cursor.Next(ctx) {
		var doc mongoNamespace
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		batch = append(batch, doc)
		if len(batch) >= batchSize {
			if err := m.compareNamespaceBatch(ctx, r, batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	if len(batch) > 0 {
		return m.compareNamespaceBatch(ctx, r, batch)
	}

	return nil
}

func (m *Migrator) compareNamespaceBatch(ctx context.Context, r *ValidationReport, batch []mongoNamespace) error {
	ids := make([]string, len(batch))
	expected := make(map[string]*entity.Namespace, len(batch))
	for i, doc := range batch {
		e := convertNamespace(doc)
		ids[i] = e.ID
		expected[e.ID] = e
	}

	var actual []entity.Namespace
	if err := m.pg.NewSelect().Model(&actual).Where("id IN (?)", bun.List(ids)).Scan(ctx); err != nil {
		return err
	}

	pgMap := make(map[string]*entity.Namespace, len(actual))
	for i := range actual {
		pgMap[actual[i].ID] = &actual[i]
	}

	r.AddCompared("namespaces", int64(len(batch)))

	for _, id := range ids {
		exp := expected[id]
		act, ok := pgMap[id]
		if !ok {
			r.AddMissing("namespaces", id)

			continue
		}

		t := "namespaces"
		r.CheckTime(t, id, "CreatedAt", exp.CreatedAt, act.CreatedAt)
		r.CheckField(t, id, "Type", exp.Type, act.Type)
		r.CheckField(t, id, "Name", exp.Name, act.Name)
		r.CheckField(t, id, "OwnerID", exp.OwnerID, act.OwnerID)
		r.CheckField(t, id, "DevicesAcceptedCount", exp.DevicesAcceptedCount, act.DevicesAcceptedCount)
		r.CheckField(t, id, "DevicesPendingCount", exp.DevicesPendingCount, act.DevicesPendingCount)
		r.CheckField(t, id, "DevicesRejectedCount", exp.DevicesRejectedCount, act.DevicesRejectedCount)
		r.CheckField(t, id, "DevicesRemovedCount", exp.DevicesRemovedCount, act.DevicesRemovedCount)
		r.CheckField(t, id, "Settings.MaxDevices", exp.Settings.MaxDevices, act.Settings.MaxDevices)
		r.CheckField(t, id, "Settings.SessionRecord", exp.Settings.SessionRecord, act.Settings.SessionRecord)
		r.CheckField(t, id, "Settings.ConnectionAnnouncement", exp.Settings.ConnectionAnnouncement, act.Settings.ConnectionAnnouncement)
	}

	return nil
}

func (m *Migrator) deepValidateMemberships(ctx context.Context, r *ValidationReport) error {
	validUsers, err := m.loadValidUsers(ctx)
	if err != nil {
		return err
	}

	cursor, err := m.mongo.Collection("namespaces").Find(ctx, bson.M{
		"members": bson.M{"$exists": true, "$ne": bson.A{}},
	})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx) //nolint:errcheck

	for cursor.Next(ctx) {
		var doc mongoNamespace
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		for _, member := range doc.Members {
			exp := convertMembership(doc.TenantID, member)

			// Skip memberships whose user was not migrated (orphaned).
			if _, ok := validUsers[exp.UserID]; !ok {
				continue
			}

			var act entity.Membership
			err := m.pg.NewSelect().Model(&act).
				Where("user_id = ?", exp.UserID).
				Where("namespace_id = ?", exp.NamespaceID).
				Scan(ctx)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					r.AddMissing("memberships", exp.UserID+"/"+exp.NamespaceID)
					r.AddCompared("memberships", 1)

					continue
				}

				return err
			}

			r.AddCompared("memberships", 1)

			id := exp.UserID + "/" + exp.NamespaceID
			t := "memberships"
			r.CheckField(t, id, "Role", exp.Role, act.Role)
			r.CheckTime(t, id, "CreatedAt", exp.CreatedAt, act.CreatedAt)
		}
	}

	return cursor.Err()
}
