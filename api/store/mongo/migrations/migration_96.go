package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/envs"
	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration96 = migrate.Migration{
	Version:     MigrationVersion96,
	Description: "Drops the recorded_session collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   MigrationVersion96,
			"action":    "Up",
		}).Info("Applying migration")

		if !envs.IsEnterprise() {
			return nil
		}

		if err := db.Collection("recorded_sessions").Drop(ctx); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   MigrationVersion96,
			"action":    "Down",
		}).Info("Reverting migration")

		if !envs.IsEnterprise() {
			return nil
		}

		if err := db.CreateCollection(ctx, "recorded_sessions"); err != nil {
			return err
		}

		return nil
	}),
}
