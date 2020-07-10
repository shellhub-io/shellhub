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
			_, err := db.Collection("devices").Indexes().CreateOne(context.TODO(), mod)
			_, err = db.Collection("devices").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"status": "allow"}})

			return err
		},
		Down: func(db *mongo.Database) error {
			_, err := db.Collection("status").Indexes().DropOne(context.TODO(), "status")

			return err
		},
	},
	{
		Version: 7,
		Up: func(db *mongo.Database) error {
			mod := mongo.IndexModel{
				Keys:    bson.D{{"uid", 1}},
				Options: options.Index().SetName("uid").SetUnique(false),
			}
			_, err := db.Collection("recorded_sessions").Indexes().CreateOne(context.TODO(), mod)

			mod = mongo.IndexModel{
				Keys:    bson.D{{"uid", 1}},
				Options: options.Index().SetName("message").SetUnique(false),
			}
			_, err = db.Collection("recorded_sessions").Indexes().CreateOne(context.TODO(), mod)

			return err
		},
		Down: func(db *mongo.Database) error {
			_, err := db.Collection("recorded_sessions").Indexes().DropOne(context.TODO(), "uid")
			_, err = db.Collection("recorded_sessions").Indexes().DropOne(context.TODO(), "message")

			return err
		},
	},
	{
		Version: 7,
		Up: func(db *mongo.Database) error {
			mod := mongo.IndexModel{
				Keys:    bson.D{{"recorded", 1}},
				Options: options.Index().SetName("recorded").SetUnique(false),
			}
			_, err := db.Collection("sessions").Indexes().CreateOne(context.TODO(), mod)
			_, err = db.Collection("sessions").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"recorded": false}})

			return err
		},
		Down: func(db *mongo.Database) error {
			_, err := db.Collection("sessions").UpdateMany(context.TODO(), bson.M{}, bson.M{"$unset": bson.M{"recorded": ""}})
			_, err = db.Collection("sessions").Indexes().DropOne(context.TODO(), "recorded")

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
