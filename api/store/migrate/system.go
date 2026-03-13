package migrate

import (
	"context"

	"github.com/google/uuid" //nolint:depguard // migration package generates UUIDs directly
	"github.com/shellhub-io/shellhub/api/store/pg/entity"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

type mongoSystem struct {
	Setup          bool             `bson:"setup"`
	Authentication *mongoSystemAuth `bson:"authentication"`
}

type mongoSystemAuth struct {
	Local *mongoSystemAuthLocal `bson:"local"`
}

type mongoSystemAuthLocal struct {
	Enabled bool `bson:"enabled"`
}

func convertSystem(doc mongoSystem) *entity.System {
	e := &entity.System{
		ID:    uuid.New().String(),
		Setup: doc.Setup,
	}

	if doc.Authentication != nil {
		if doc.Authentication.Local != nil {
			e.Authentication.Local.Enabled = doc.Authentication.Local.Enabled
		}
	}

	return e
}

func (m *Migrator) migrateSystems(ctx context.Context) error {
	cursor, err := m.mongo.Collection("system").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var batch []*entity.System
	for cursor.Next(ctx) {
		var doc mongoSystem
		if err := cursor.Decode(&doc); err != nil {
			return err
		}

		batch = append(batch, convertSystem(doc))
	}

	if err := cursor.Err(); err != nil {
		return err
	}

	if len(batch) > 0 {
		if _, err := m.pg.NewInsert().Model(&batch).Exec(ctx); err != nil {
			return err
		}
	}

	log.WithFields(log.Fields{"scope": "core", "count": len(batch)}).Info("Migrated systems")

	return nil
}
