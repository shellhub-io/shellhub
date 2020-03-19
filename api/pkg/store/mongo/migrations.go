package mongo

import (
	"context"

	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
			_, err := db.Collection("sessions").UpdateMany(context.Background(), bson.M{}, bson.M{"$rename": bson.M{"device": "device_uid"}})
			return err
		},
		Down: func(db *mongo.Database) error {
			_, err := db.Collection("sessions").UpdateMany(context.Background(), bson.M{}, bson.M{"$rename": bson.M{"device_uid": "device"}})
			return err
		},
	},
	// Version 3
	{
		Version: 3,
		Up: func(db *mongo.Database) error {
			_, err := db.Collection("devices").UpdateMany(context.Background(), bson.M{}, bson.M{"$rename": bson.M{"attributes": "info"}})
			return err
		},
		Down: func(db *mongo.Database) error {
			_, err := db.Collection("devices").UpdateMany(context.Background(), bson.M{}, bson.M{"$rename": bson.M{"info": "attributes"}})
			return err
		},
	},
	// Version 4
	{
		Version: 4,
		Up: func(db *mongo.Database) error {
			_, err := db.Collection("devices").UpdateMany(context.Background(), bson.M{}, bson.M{"$rename": bson.M{"version": "info.version"}})
			return err
		},
		Down: func(db *mongo.Database) error {
			_, err := db.Collection("devices").UpdateMany(context.Background(), bson.M{}, bson.M{"$rename": bson.M{"info.version": "version"}})
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
