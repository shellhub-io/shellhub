package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration39 = migrate.Migration{
	Version:     39,
	Description: "Create a new field on namespaces to store the API tokens",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   39,
			"action":    "Up",
		}).Info("Applying migration")
		_, err := db.Collection("namespaces").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"api_tokens": []models.Token{}}})

		return err
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   39,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	},
}
