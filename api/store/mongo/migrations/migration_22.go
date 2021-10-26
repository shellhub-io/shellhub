package migrations

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration22 = migrate.Migration{
	Version:     22,
	Description: "Insert the user on the members group for the namespace",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   22,
			"action":    "Up",
		}).Info("Applying migration")
		cursor, err := db.Collection("namespaces").Find(context.TODO(), bson.D{})
		if err != nil {
			return err
		}

		type NamespaceSettings struct {
			SessionRecord bool `json:"session_record" bson:"session_record,omitempty"`
		}

		type Namespace struct {
			Name         string             `json:"name"  validate:"required,hostname_rfc1123,excludes=."`
			Owner        string             `json:"owner"`
			TenantID     string             `json:"tenant_id" bson:"tenant_id,omitempty"`
			Members      []interface{}      `json:"members" bson:"members"`
			Settings     *NamespaceSettings `json:"settings"`
			Devices      int                `json:"devices" bson:",omitempty"`
			Sessions     int                `json:"sessions" bson:",omitempty"`
			MaxDevices   int                `json:"max_devices" bson:"max_devices"`
			DevicesCount int                `json:"devices_count" bson:"devices_count,omitempty"`
			CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
		}

		for cursor.Next(context.TODO()) {

			namespace := Namespace{}

			err = cursor.Decode(&namespace)
			if err != nil {
				return err
			}

			for _, memberID := range namespace.Members {
				user := new(models.User)
				objID, err := primitive.ObjectIDFromHex(memberID.(string))
				if err != nil {
					return err
				}
				if err := db.Collection("users").FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&user); err != nil {
					if _, err := db.Collection("namespaces").UpdateOne(context.TODO(), bson.M{"tenant_id": namespace.TenantID}, bson.M{"$pull": bson.M{"members": memberID}}); err != nil {
						return err
					}
				}
			}
		}

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   22,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	},
}
