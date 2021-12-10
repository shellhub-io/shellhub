package migrations

import (
	"context"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/shellhub-io/shellhub/pkg/authorizer"
)

var migration37 = migrate.Migration{
	Version:     37,
	Description: "Change member's role from array of ID to a list of members' object",
	Up: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   37,
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
			Billing      interface{}        `json:"billing" bson:"billing,omitempty"`
		}

		for cursor.Next(context.TODO()) {
			namespace := new(Namespace)
			err = cursor.Decode(&namespace)
			if err != nil {
				return err
			}

			owner := namespace.Owner
			members := namespace.Members
			memberList := []models.Member{}

			for _, member := range members {
				if owner != member {
					m := models.Member{
						ID:   member.(string),
						Role: authorizer.MemberRoleObserver,
					}

					memberList = append(memberList, m)
				} else if owner == member {
					m := models.Member{
						ID:   member.(string),
						Role: authorizer.MemberRoleOwner,
					}

					memberList = append(memberList, m)
				}
			}

			if _, err := db.Collection("namespaces").UpdateOne(context.TODO(), bson.M{"tenant_id": namespace.TenantID}, bson.M{"$set": bson.M{"members": memberList}}); err != nil {
				return err
			}
		}

		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(context.TODO())

		return nil
	},
	Down: func(db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   37,
			"action":    "Down",
		}).Info("Applying migration")
		cursor, err := db.Collection("namespaces").Find(context.TODO(), bson.D{})
		if err != nil {
			return err
		}

		for cursor.Next(context.TODO()) {
			namespace := new(models.Namespace)
			err = cursor.Decode(&namespace)
			if err != nil {
				return err
			}

			var membersList []interface{}
			for _, member := range namespace.Members {
				membersList = append(membersList, member.ID)
			}

			if _, err := db.Collection("namespaces").UpdateOne(context.TODO(), bson.M{"tenant_id": namespace.TenantID}, bson.M{"$set": bson.M{"members": membersList}}); err != nil {
				return err
			}
		}

		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(context.TODO())

		return nil
	},
}
