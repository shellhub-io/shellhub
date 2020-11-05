package mongo

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/models"
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
	{
		Version: 7,
		Up: func(db *mongo.Database) error {
			mod := mongo.IndexModel{
				Keys:    bson.D{{"uid", 1}},
				Options: options.Index().SetName("uid").SetUnique(false),
			}
			if _, err := db.Collection("recorded_sessions").Indexes().CreateOne(context.TODO(), mod); err != nil {
				return err
			}

			mod = mongo.IndexModel{
				Keys:    bson.D{{"message", 1}},
				Options: options.Index().SetName("message").SetUnique(false),
			}
			_, err := db.Collection("recorded_sessions").Indexes().CreateOne(context.TODO(), mod)
			return err
		},
		Down: func(db *mongo.Database) error {
			if _, err := db.Collection("recorded_sessions").UpdateMany(context.TODO(), bson.M{}, bson.M{"$unset": bson.M{"uid": ""}}); err != nil {
				return err
			}
			if _, err := db.Collection("recorded_sessions").UpdateMany(context.TODO(), bson.M{}, bson.M{"$unset": bson.M{"message": ""}}); err != nil {
				return err
			}
			if _, err := db.Collection("recorded_sessions").Indexes().DropOne(context.TODO(), "uid"); err != nil {
				return err
			}
			_, err := db.Collection("recorded_sessions").Indexes().DropOne(context.TODO(), "message")
			return err
		},
	},
	{
		Version: 8,
		Up: func(db *mongo.Database) error {
			mod := mongo.IndexModel{
				Keys:    bson.D{{"recorded", 1}},
				Options: options.Index().SetName("recorded").SetUnique(false),
			}
			if _, err := db.Collection("sessions").Indexes().CreateOne(context.TODO(), mod); err != nil {
				return err
			}
			_, err := db.Collection("sessions").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"recorded": false}})
			return err
		},
		Down: func(db *mongo.Database) error {
			if _, err := db.Collection("sessions").UpdateMany(context.TODO(), bson.M{}, bson.M{"$unset": bson.M{"recorded": ""}}); err != nil {
				return err
			}
			_, err := db.Collection("sessions").Indexes().DropOne(context.TODO(), "recorded")
			return err
		},
	},
	{
		Version: 9,
		Up: func(db *mongo.Database) error {
			cursor, err := db.Collection("devices").Find(context.TODO(), bson.D{})
			if err != nil {
				return err
			}
			defer cursor.Close(context.TODO())
			for cursor.Next(context.TODO()) {
				device := new(models.Device)
				err := cursor.Decode(&device)
				if err != nil {
					return err
				}

				device.Name = strings.ToLower(device.Name)
				if _, err = db.Collection("devices").UpdateOne(context.TODO(), bson.M{"uid": device.UID}, bson.M{"$set": bson.M{"name": strings.ToLower(device.Name)}}); err != nil {
					return err
				}
			}

			return nil
		},

		Down: func(db *mongo.Database) error {
			return nil
		},
	},
	{
		Version: 10,
		Up: func(db *mongo.Database) error {
			mod := mongo.IndexModel{
				Keys:    bson.D{{"session_record", 1}},
				Options: options.Index().SetName("session_record").SetUnique(false),
			}
			if _, err := db.Collection("users").Indexes().CreateOne(context.TODO(), mod); err != nil {
				return err
			}
			_, err := db.Collection("users").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"session_record": true}})
			return err
		},
		Down: func(db *mongo.Database) error {
			return nil
		},
	},
	{
		Version: 11,
		Up: func(db *mongo.Database) error {
			mod := mongo.IndexModel{
				Keys:    bson.D{{"created_at", 1}},
				Options: options.Index().SetName("ttl").SetExpireAfterSeconds(60),
			}
			_, err := db.Collection("private_keys").Indexes().CreateOne(context.TODO(), mod)
			if err != nil {
				return err
			}

			return nil
		},
		Down: func(db *mongo.Database) error {
			_, err := db.Collection("private_keys").Indexes().DropOne(context.TODO(), "ttl")
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
