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

var migration17 = migrate.Migration{
	Version:     17,
	Description: "Remove the namespaces, devices, session, connected_devices, firewall_rules and public_keys in the users",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   17,
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

			objID, err := primitive.ObjectIDFromHex(namespace.Owner)
			if err != nil {
				return err
			}

			user := new(models.User)
			if err := db.Collection("users").FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&user); err != nil {
				if err != mongo.ErrNoDocuments {
					return err
				}
				if _, err := db.Collection("namespaces").DeleteOne(context.TODO(), bson.M{"tenant_id": namespace.TenantID}); err != nil {
					return err
				}
			}
		}

		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(context.TODO())

		cursor, err = db.Collection("devices").Find(context.TODO(), bson.D{})
		if err != nil {
			return err
		}

		for cursor.Next(context.TODO()) {
			device := new(models.Device)
			err = cursor.Decode(&device)
			if err != nil {
				return err
			}

			namespace := Namespace{}
			if err := db.Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": device.TenantID}).Decode(&namespace); err != nil {
				if err != mongo.ErrNoDocuments {
					return err
				}
				if _, err := db.Collection("devices").DeleteOne(context.TODO(), bson.M{"uid": device.UID}); err != nil {
					return err
				}

				if _, err := db.Collection("sessions").DeleteMany(context.TODO(), bson.M{"device_uid": device.UID}); err != nil {
					return err
				}

				if _, err := db.Collection("connected_devices").DeleteMany(context.TODO(), bson.M{"uid": device.UID}); err != nil {
					return err
				}
			}
		}
		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(context.TODO())

		cursor, err = db.Collection("firewall_rules").Find(context.TODO(), bson.D{})
		if err != nil {
			return err
		}
		for cursor.Next(context.TODO()) {
			rule := new(models.FirewallRule)
			err := cursor.Decode(&rule)
			if err != nil {
				return err
			}

			namespace := Namespace{}
			if err := db.Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": rule.TenantID}).Decode(&namespace); err != nil {
				if err != mongo.ErrNoDocuments {
					return err
				}
				if _, err := db.Collection("firewall_rules").DeleteOne(context.TODO(), bson.M{"tenant_id": rule.TenantID}); err != nil {
					return err
				}
			}
		}
		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(context.TODO())

		cursor, err = db.Collection("public_keys").Find(context.TODO(), bson.D{})
		if err != nil {
			return err
		}

		for cursor.Next(context.TODO()) {
			key := new(models.PublicKey)
			err := cursor.Decode(&key)
			if err != nil {
				return err
			}
			namespace := Namespace{}
			if err := db.Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": key.TenantID}).Decode(&namespace); err != nil {
				if err != mongo.ErrNoDocuments {
					return err
				}
				if _, err := db.Collection("public_keys").DeleteOne(context.TODO(), bson.M{"tenant_id": key.TenantID}); err != nil {
					return err
				}
			}
		}
		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(context.TODO())

		return err
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   17,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	},
}
