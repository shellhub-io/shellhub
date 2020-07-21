package mongo

import (
	"context"

	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migrations = []migrate.Migration{
	// Version 1
	{
		Version: 1,
		Up: func(db *mongo.Database) error {
			return nil
		},
		Down: func(db *mongo.Database) error {
			return nil
		},
	},
	// Version 2
	{
		Version: 2,
		Up: func(db *mongo.Database) error {
			return renameField(db, "sessions", "device", "device_uid")
		},
		Down: func(db *mongo.Database) error {
			return renameField(db, "sessions", "device_uid", "device")
		},
	},
	// Version 3
	{
		Version: 3,
		Up: func(db *mongo.Database) error {
			return renameField(db, "devices", "attributes", "info")
		},
		Down: func(db *mongo.Database) error {
			return renameField(db, "devices", "info", "attributes")
		},
	},
	// Version 4
	{
		Version: 4,
		Up: func(db *mongo.Database) error {
			return renameField(db, "devices", "version", "info.version")
		},
		Down: func(db *mongo.Database) error {
			return renameField(db, "devices", "info.version", "version")
		},
	},
	{
		Version: 5,
		Up: func(db *mongo.Database) error {
			mod := mongo.IndexModel{
				Keys:    bson.D{{"email", 1}},
				Options: options.Index().SetName("email").SetUnique(true),
			}
			_, err := db.Collection("users").Indexes().CreateOne(context.TODO(), mod)

			return err
		},
		Down: func(db *mongo.Database) error {
			_, err := db.Collection("users").Indexes().DropOne(context.TODO(), "email")

			return err
		},
	},
	{
		Version: 6,
		Up: func(db *mongo.Database) error {
			mod := mongo.IndexModel{
				Keys:    bson.D{{"status", 1}},
				Options: options.Index().SetName("status").SetUnique(false),
			}
			if _, err := db.Collection("devices").Indexes().CreateOne(context.TODO(), mod); err != nil {
				return err
			}
			_, err := db.Collection("devices").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"status": "accepted"}})
			return err
		},
		Down: func(db *mongo.Database) error {
			if _, err := db.Collection("devices").UpdateMany(context.TODO(), bson.M{}, bson.M{"$unset": bson.M{"status": ""}}); err != nil {
				return err
			}
			_, err := db.Collection("status").Indexes().DropOne(context.TODO(), "status")
			return err
		},
	},
}

func ApplyMigrations(db *mongo.Database) error {
	m := migrate.NewMigrate(db, migrations...)

	if err := m.Up(migrate.AllAvailable); err != nil {
		return err
	}

	return EnsureIndexes(db)
}
