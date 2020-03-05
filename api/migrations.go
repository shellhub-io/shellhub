package main

import (
	migrate "github.com/xakep666/mongo-migrate"
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
}

func applyMigrations(db *mongo.Database) error {
	m := migrate.NewMigrate(db, migrations...)

	if err := m.Up(migrate.AllAvailable); err != nil {
		return err
	}

	return nil
}
