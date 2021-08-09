package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration31 = migrate.Migration{
	Version:     31,
	Description: "add last_login field to collection users",
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 31 - Up")
		if _, err := db.Collection("namespaces").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"created_at": clock.Now()}}); err != nil {
			return err
		}

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 31 - Down")

		return nil
	},
}
