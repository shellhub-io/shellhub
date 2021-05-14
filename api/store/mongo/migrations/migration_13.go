package migrations

import (
	"context"

	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration_13 = migrate.Migration{
	Version: 13,
	Up: func(db *mongo.Database) error {
		mod := mongo.IndexModel{
			Keys:    bson.D{{"uid", 1}},
			Options: options.Index().SetName("uid").SetUnique(true),
		}
		_, err := db.Collection("devices").Indexes().CreateOne(context.TODO(), mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"last_seen", 1}},
			Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(30),
		}
		_, err = db.Collection("connected_devices").Indexes().CreateOne(context.TODO(), mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"uid", 1}},
			Options: options.Index().SetName("uid").SetUnique(false),
		}
		_, err = db.Collection("connected_devices").Indexes().CreateOne(context.TODO(), mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"uid", 1}},
			Options: options.Index().SetName("uid").SetUnique(true),
		}
		_, err = db.Collection("sessions").Indexes().CreateOne(context.TODO(), mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"last_seen", 1}},
			Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(30),
		}
		_, err = db.Collection("active_sessions").Indexes().CreateOne(context.TODO(), mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"uid", 1}},
			Options: options.Index().SetName("uid").SetUnique(false),
		}
		_, err = db.Collection("active_sessions").Indexes().CreateOne(context.TODO(), mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"username", 1}},
			Options: options.Index().SetName("username").SetUnique(true),
		}
		_, err = db.Collection("users").Indexes().CreateOne(context.TODO(), mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"tenant_id", 1}},
			Options: options.Index().SetName("tenant_id").SetUnique(true),
		}
		_, err = db.Collection("users").Indexes().CreateOne(context.TODO(), mod)
		if err != nil {
			return err
		}
		return nil
	},
	Down: func(db *mongo.Database) error {
		if _, err := db.Collection("devices").Indexes().DropOne(context.TODO(), "uid"); err != nil {
			return err
		}
		if _, err := db.Collection("connected_devices").Indexes().DropOne(context.TODO(), "last_seen"); err != nil {
			return err
		}
		if _, err := db.Collection("connected_devices").Indexes().DropOne(context.TODO(), "uid"); err != nil {
			return err
		}
		if _, err := db.Collection("sessions").Indexes().DropOne(context.TODO(), "uid"); err != nil {
			return err
		}
		if _, err := db.Collection("active_sessions").Indexes().DropOne(context.TODO(), "last_seen"); err != nil {
			return err
		}
		if _, err := db.Collection("users").Indexes().DropOne(context.TODO(), "username"); err != nil {
			return err
		}
		_, err := db.Collection("users").Indexes().DropOne(context.TODO(), "tenant_id")
		return err
	},
}
