package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration119 = migrate.Migration{
	Version:     119,
	Description: "Create indexes on membership_invitations collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   119,
			"action":    "Up",
		}).Info("Applying migration up")

		indexes := []struct {
			name  string
			model mongo.IndexModel
		}{
			{
				name: "tenant_user_status_pending_unique",
				model: mongo.IndexModel{
					Keys: bson.D{
						{Key: "tenant_id", Value: 1},
						{Key: "user_id", Value: 1},
						{Key: "status", Value: 1},
					},
					Options: options.Index().
						SetName("tenant_user_status_pending_unique").
						SetUnique(true).
						SetPartialFilterExpression(bson.M{"status": "pending"}),
				},
			},
			{
				name: "tenant_user_created_at",
				model: mongo.IndexModel{
					Keys: bson.D{
						{Key: "tenant_id", Value: 1},
						{Key: "user_id", Value: 1},
					},
					Options: options.Index().SetName("tenant_user_created_at"),
				},
			},
			{
				name: "user_status",
				model: mongo.IndexModel{
					Keys: bson.D{
						{Key: "user_id", Value: 1},
						{Key: "status", Value: 1},
					},
					Options: options.Index().SetName("user_status"),
				},
			},
		}

		for _, ix := range indexes {
			if _, err := db.Collection("membership_invitations").Indexes().CreateOne(ctx, ix.model); err != nil {
				log.WithError(err).WithField("index", ix.name).Error("Failed to create index")

				return err
			}
		}

		log.Info("Successfully created indexes on membership_invitations collection")

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   119,
			"action":    "Down",
		}).Info("Applying migration down")

		indexes := []string{"tenant_user_status_pending_unique", "tenant_user_created_at", "user_status"}
		for _, ix := range indexes {
			if _, err := db.Collection("membership_invitations").Indexes().DropOne(ctx, ix); err != nil {
				log.WithError(err).WithField("index", ix).Error("Failed to drop index")

				return err
			}
		}

		log.Info("Successfully dropped indexes from membership_invitations collection")

		return nil
	}),
}
