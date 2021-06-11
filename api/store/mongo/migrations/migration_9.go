package migrations

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration9 = migrate.Migration{
	Version: 9,
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 9 - Up")
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
		logrus.Info("Applying migration 9 - Down")

		return nil
	},
}
