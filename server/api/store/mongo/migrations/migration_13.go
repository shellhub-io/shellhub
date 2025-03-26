package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration13 = migrate.Migration{
	Version:     13,
	Description: "Change on several collections",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   13,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"uid", 1}},
			Options: options.Index().SetName("uid").SetUnique(true),
		}
		_, err := db.Collection("devices").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"last_seen", 1}},
			Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(30),
		}
		_, err = db.Collection("connected_devices").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"uid", 1}},
			Options: options.Index().SetName("uid").SetUnique(false),
		}
		_, err = db.Collection("connected_devices").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"uid", 1}},
			Options: options.Index().SetName("uid").SetUnique(true),
		}
		_, err = db.Collection("sessions").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"last_seen", 1}},
			Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(30),
		}
		_, err = db.Collection("active_sessions").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"uid", 1}},
			Options: options.Index().SetName("uid").SetUnique(false),
		}
		_, err = db.Collection("active_sessions").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"username", 1}},
			Options: options.Index().SetName("username").SetUnique(true),
		}
		_, err = db.Collection("users").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"tenant_id", 1}},
			Options: options.Index().SetName("tenant_id").SetUnique(true),
		}
		_, err = db.Collection("users").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   13,
			"action":    "Down",
		}).Info("Applying migration")
		if _, err := db.Collection("devices").Indexes().DropOne(ctx, "uid"); err != nil {
			return err
		}
		if _, err := db.Collection("connected_devices").Indexes().DropOne(ctx, "last_seen"); err != nil {
			return err
		}
		if _, err := db.Collection("connected_devices").Indexes().DropOne(ctx, "uid"); err != nil {
			return err
		}
		if _, err := db.Collection("sessions").Indexes().DropOne(ctx, "uid"); err != nil {
			return err
		}
		if _, err := db.Collection("active_sessions").Indexes().DropOne(ctx, "last_seen"); err != nil {
			return err
		}
		if _, err := db.Collection("users").Indexes().DropOne(ctx, "username"); err != nil {
			return err
		}
		_, err := db.Collection("users").Indexes().DropOne(ctx, "tenant_id")

		return err
	}),
}
