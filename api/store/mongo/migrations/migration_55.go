package migrations

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration55 = migrate.Migration{
	Version:     55,
	Description: "create indexes on removed_devices for tenant_id, tenant_id and uid and timestamp",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   55,
			"action":    "Up",
		}).Info("Applying migration")
		fieldTenantID := "tenant_id"
		fieldUID := "uid"
		fieldTimestamp := "timestamp"

		expire, err := time.ParseDuration("720h")
		if err != nil {
			return err
		}
		expireSeconds := int32(expire.Seconds())

		fieldNameTenantID := "tenant_id_1"
		fieldNameTenantIDUID := "tenant_id_1_uid_1"
		fieldNameTimestamp := "timestamp_1"
		if _, err := db.Collection("removed_devices").Indexes().CreateMany(context.Background(), []mongo.IndexModel{
			{
				Keys: bson.D{
					bson.E{Key: fieldTenantID, Value: 1},
				},
				Options: &options.IndexOptions{
					Name: &fieldNameTenantID,
				},
			},
			{
				Keys: bson.D{
					bson.E{Key: fieldTenantID, Value: 1},
					bson.E{Key: fieldUID, Value: 1},
				},
				Options: &options.IndexOptions{
					Name: &fieldNameTenantIDUID,
				},
			},
			{
				Keys: bson.D{
					bson.E{Key: fieldTimestamp, Value: 1},
				},
				Options: &options.IndexOptions{
					Name:               &fieldNameTimestamp,
					ExpireAfterSeconds: &expireSeconds,
				},
			},
		}); err != nil {
			return err
		}

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   55,
			"action":    "Down",
		}).Info("Applying migration")
		fieldNameTenantID := "tenant_id_1"
		fieldNameTenantIDUID := "tenant_id_1_uid_1"
		fieldNameTimestamp := "timestamp_1"

		if _, err := db.Collection("removed_devices").Indexes().DropOne(context.Background(), fieldNameTenantID); err != nil {
			return err
		}

		if _, err := db.Collection("removed_devices").Indexes().DropOne(context.Background(), fieldNameTenantIDUID); err != nil {
			return err
		}

		if _, err := db.Collection("removed_devices").Indexes().DropOne(context.Background(), fieldNameTimestamp); err != nil {
			return err
		}

		return nil
	},
}
