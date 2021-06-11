package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration14 = migrate.Migration{
	Version: 14,
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 14 - Up")
		type user struct {
			Username      string `json:"username" bson:",omitempty"`
			TenantID      string `json:"tenant_id" bson:"tenant_id"`
			ID            string `json:"id,omitempty" bson:"_id,omitempty"`
			SessionRecord bool   `json:"session_record" bson:"session_record,omitempty"`
		}
		if _, err := db.Collection("users").Indexes().DropOne(context.TODO(), "tenant_id"); err != nil {
			return err
		}
		if _, err := db.Collection("users").Indexes().DropOne(context.TODO(), "session_record"); err != nil {
			return err
		}

		cursor, err := db.Collection("users").Find(context.TODO(), bson.D{})
		if err != nil {
			return err
		}
		defer cursor.Close(context.TODO())
		for cursor.Next(context.TODO()) {
			user := new(user)
			err := cursor.Decode(&user)
			if err != nil {
				return err
			}
			settings := &models.NamespaceSettings{SessionRecord: true}
			namespace := &models.Namespace{
				Owner:    user.ID,
				Members:  []interface{}{user.ID},
				TenantID: user.TenantID,
				Name:     user.Username,
				Settings: settings,
			}
			_, err = db.Collection("namespaces").InsertOne(context.TODO(), &namespace)
			if err != nil {
				return nil
			}

			if _, err := db.Collection("users").UpdateOne(context.TODO(), bson.M{"tenant_id": user.TenantID}, bson.M{"$unset": bson.M{"tenant_id": ""}}); err != nil {
				return err
			}
			if _, err := db.Collection("users").UpdateOne(context.TODO(), bson.M{"tenant_id": user.TenantID}, bson.M{"$unset": bson.M{"session_record": ""}}); err != nil {
				return err
			}
		}

		return cursor.Err()
	},
	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 14 - Down")
		cursor, err := db.Collection("namespaces").Find(context.TODO(), bson.D{})
		if err != nil {
			return err
		}
		defer cursor.Close(context.TODO())
		for cursor.Next(context.TODO()) {
			namespace := new(models.Namespace)
			err := cursor.Decode(&namespace)
			if err != nil {
				return err
			}

			_, err = db.Collection("users").UpdateOne(context.TODO(), bson.M{"_id": namespace.Owner}, bson.M{"$set": bson.M{"tenant": namespace.TenantID}})
			if err != nil {
				return err
			}
		}

		return err
	},
}
