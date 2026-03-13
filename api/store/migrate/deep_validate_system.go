package migrate

import (
	"context"
	"database/sql"
	"errors"

	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	"go.mongodb.org/mongo-driver/bson"
)

func (m *Migrator) deepValidateSystems(ctx context.Context, r *ValidationReport) error {
	cursor, err := m.mongo.Collection("system").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx) //nolint:errcheck

	if !cursor.Next(ctx) {
		return cursor.Err()
	}

	var doc mongoSystem
	if err := cursor.Decode(&doc); err != nil {
		return err
	}

	expected := convertSystem(doc)

	var actual entity.System
	if err := m.pg.NewSelect().Model(&actual).Limit(1).Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.AddMissing("systems", "system_record")

			return nil
		}

		return err
	}

	r.AddCompared("systems", 1)

	t := "systems"
	id := actual.ID

	r.CheckField(t, id, "Setup", expected.Setup, actual.Setup)
	r.CheckField(t, id, "Auth.Local.Enabled", expected.Authentication.Local.Enabled, actual.Authentication.Local.Enabled)

	return nil
}
