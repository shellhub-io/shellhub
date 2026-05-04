package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration121 = migrate.Migration{
	Version:     121,
	Description: "Create indexes on oauth_clients collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   121,
			"action":    "Up",
		}).Info("Applying migration up")

		indexes := []struct {
			name  string
			model mongo.IndexModel
		}{
			{
				name: "client_id_unique",
				model: mongo.IndexModel{
					Keys:    bson.D{{Key: "client_id", Value: 1}},
					Options: options.Index().SetName("client_id_unique").SetUnique(true),
				},
			},
			{
				name: "tenant_id",
				model: mongo.IndexModel{
					Keys:    bson.D{{Key: "tenant_id", Value: 1}},
					Options: options.Index().SetName("tenant_id"),
				},
			},
			{
				name: "tenant_name_unique",
				model: mongo.IndexModel{
					Keys: bson.D{
						{Key: "tenant_id", Value: 1},
						{Key: "name", Value: 1},
					},
					// Mirror the PG partial index: tenantless dynamic clients
					// can share the same default name ("mcp-client") because the
					// uniqueness only applies when tenant_id is set. Without
					// this filter, the second dynamic registration call fails
					// with E11000.
					//
					// Mongo's partialFilterExpression does NOT accept $ne, so
					// we use $gt: "" — strings sort lexicographically, so any
					// non-empty string is > "". This excludes both empty-string
					// and missing-field documents from the unique constraint.
					Options: options.Index().
						SetName("tenant_name_unique").
						SetUnique(true).
						SetPartialFilterExpression(bson.M{"tenant_id": bson.M{"$gt": ""}}),
				},
			},
		}

		for _, ix := range indexes {
			if _, err := db.Collection("oauth_clients").Indexes().CreateOne(ctx, ix.model); err != nil {
				log.WithError(err).WithField("index", ix.name).Error("Failed to create index")

				return err
			}
		}

		log.Info("Successfully created indexes on oauth_clients collection")

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   121,
			"action":    "Down",
		}).Info("Applying migration down")

		indexes := []string{"client_id_unique", "tenant_id", "tenant_name_unique"}
		for _, ix := range indexes {
			if _, err := db.Collection("oauth_clients").Indexes().DropOne(ctx, ix); err != nil {
				log.WithError(err).WithField("index", ix).Error("Failed to drop index")

				return err
			}
		}

		log.Info("Successfully dropped indexes from oauth_clients collection")

		return nil
	}),
}
