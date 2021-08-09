package migrations

import (
	"context"

	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GenerateMigrations() []migrate.Migration {
	return []migrate.Migration{
		migration1,
		migration2,
		migration3,
		migration4,
		migration5,
		migration6,
		migration7,
		migration8,
		migration9,
		migration10,
		migration11,
		migration12,
		migration13,
		migration14,
		migration15,
		migration16,
		migration17,
		migration18,
		migration19,
		migration20,
		migration21,
		migration22,
		migration23,
		migration24,
		migration25,
		migration26,
		migration27,
		migration28,
		migration29,
		migration30,
		migration31,
	}
}

func renameField(db *mongo.Database, coll, from, to string) error {
	_, err := db.Collection(coll).UpdateMany(context.Background(), bson.M{}, bson.M{"$rename": bson.M{from: to}})

	return err
}
