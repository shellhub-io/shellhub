package migrate

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"github.com/uptrace/bun"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *Migrator) deepValidateUsers(ctx context.Context, r *ValidationReport) error {
	cursor, err := m.mongo.Collection("users").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx) //nolint:errcheck

	batch := make([]mongoUser, 0, batchSize)

	for cursor.Next(ctx) {
		var doc mongoUser
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		batch = append(batch, doc)
		if len(batch) >= batchSize {
			if err := m.compareUserBatch(ctx, r, batch); err != nil {
				return err
			}
			batch = batch[:0]
		}
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	if len(batch) > 0 {
		return m.compareUserBatch(ctx, r, batch)
	}

	return nil
}

func (m *Migrator) compareUserBatch(ctx context.Context, r *ValidationReport, batch []mongoUser) error {
	ids := make([]string, len(batch))
	expected := make(map[string]*entity.User, len(batch))
	for i, doc := range batch {
		e := convertUser(doc)
		ids[i] = e.ID
		expected[e.ID] = e
	}

	var actual []entity.User
	if err := m.pg.NewSelect().Model(&actual).Where("id IN (?)", bun.List(ids)).Scan(ctx); err != nil {
		return err
	}

	pgMap := make(map[string]*entity.User, len(actual))
	for i := range actual {
		pgMap[actual[i].ID] = &actual[i]
	}

	r.AddCompared("users", int64(len(batch)))

	for _, id := range ids {
		exp := expected[id]
		act, ok := pgMap[id]
		if !ok {
			r.AddMissing("users", id)

			continue
		}

		t := "users"
		r.CheckField(t, id, "Origin", exp.Origin, act.Origin)
		r.CheckField(t, id, "ExternalID", exp.ExternalID, act.ExternalID)
		r.CheckField(t, id, "Status", exp.Status, act.Status)
		r.CheckField(t, id, "Name", exp.Name, act.Name)
		r.CheckField(t, id, "Username", exp.Username, act.Username)
		r.CheckField(t, id, "Email", exp.Email, act.Email)
		r.CheckField(t, id, "PasswordDigest", exp.PasswordDigest, act.PasswordDigest)
		r.CheckField(t, id, "Admin", exp.Admin, act.Admin)
		r.CheckTime(t, id, "CreatedAt", exp.CreatedAt, act.CreatedAt)
		r.CheckTime(t, id, "LastLogin", exp.LastLogin, act.LastLogin)
		r.CheckField(t, id, "PreferredNamespace", exp.Preferences.PreferredNamespace, act.Preferences.PreferredNamespace)
		r.CheckStrings(t, id, "AuthMethods", exp.Preferences.AuthMethods, act.Preferences.AuthMethods)
		r.CheckField(t, id, "SecurityEmail", exp.Preferences.SecurityEmail, act.Preferences.SecurityEmail)
		r.CheckField(t, id, "MaxNamespaces", exp.Preferences.MaxNamespaces, act.Preferences.MaxNamespaces)
		r.CheckField(t, id, "EmailMarketing", exp.Preferences.EmailMarketing, act.Preferences.EmailMarketing)
	}

	return nil
}
