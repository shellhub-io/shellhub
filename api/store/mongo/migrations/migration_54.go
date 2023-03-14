package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration54 = migrate.Migration{
	Version:     54,
	Description: "create index to devices' tenant_id and status",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   54,
			"action":    "Up",
		}).Info("Applying migration")
		fieldTenantID := "tenant_id"
		fieldStatus := "status"
		name := "tenant_id_1_status_1"

		if _, err := db.Collection("devices").Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys: bson.D{
				bson.E{Key: fieldTenantID, Value: 1},
				bson.E{Key: fieldStatus, Value: 1},
			},
			Options: &options.IndexOptions{ //nolint:exhaustruct
				Name: &name,
			},
		}); err != nil {
			return err
		}

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   54,
			"action":    "Down",
		}).Info("Applying migration")
		name := "tenant_id_1_status_1"

		if _, err := db.Collection("devices").Indexes().DropOne(context.Background(), name); err != nil {
			return err
		}

		return nil
	},
}
