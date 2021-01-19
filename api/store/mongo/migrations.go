package mongo

import (
	"context"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/models"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
			return renameField(db, "sessions", "device", "device_uid")
		},
		Down: func(db *mongo.Database) error {
			return renameField(db, "sessions", "device_uid", "device")
		},
	},
	// Version 3
	{
		Version: 3,
		Up: func(db *mongo.Database) error {
			return renameField(db, "devices", "attributes", "info")
		},
		Down: func(db *mongo.Database) error {
			return renameField(db, "devices", "info", "attributes")
		},
	},
	// Version 4
	{
		Version: 4,
		Up: func(db *mongo.Database) error {
			return renameField(db, "devices", "version", "info.version")
		},
		Down: func(db *mongo.Database) error {
			return renameField(db, "devices", "info.version", "version")
		},
	},
	{
		Version: 5,
		Up: func(db *mongo.Database) error {
			mod := mongo.IndexModel{
				Keys:    bson.D{{"email", 1}},
				Options: options.Index().SetName("email").SetUnique(true),
			}
			_, err := db.Collection("users").Indexes().CreateOne(context.TODO(), mod)

			return err
		},
		Down: func(db *mongo.Database) error {
			_, err := db.Collection("users").Indexes().DropOne(context.TODO(), "email")

			return err
		},
	},
	{
		Version: 6,
		Up: func(db *mongo.Database) error {
			mod := mongo.IndexModel{
				Keys:    bson.D{{"status", 1}},
				Options: options.Index().SetName("status").SetUnique(false),
			}
			if _, err := db.Collection("devices").Indexes().CreateOne(context.TODO(), mod); err != nil {
				return err
			}
			_, err := db.Collection("devices").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"status": "accepted"}})
			return err
		},
		Down: func(db *mongo.Database) error {
			if _, err := db.Collection("devices").UpdateMany(context.TODO(), bson.M{}, bson.M{"$unset": bson.M{"status": ""}}); err != nil {
				return err
			}
			_, err := db.Collection("status").Indexes().DropOne(context.TODO(), "status")
			return err
		},
	},
	{
		Version: 7,
		Up: func(db *mongo.Database) error {
			mod := mongo.IndexModel{
				Keys:    bson.D{{"uid", 1}},
				Options: options.Index().SetName("uid").SetUnique(false),
			}
			if _, err := db.Collection("recorded_sessions").Indexes().CreateOne(context.TODO(), mod); err != nil {
				return err
			}

			mod = mongo.IndexModel{
				Keys:    bson.D{{"message", 1}},
				Options: options.Index().SetName("message").SetUnique(false),
			}
			_, err := db.Collection("recorded_sessions").Indexes().CreateOne(context.TODO(), mod)
			return err
		},
		Down: func(db *mongo.Database) error {
			if _, err := db.Collection("recorded_sessions").UpdateMany(context.TODO(), bson.M{}, bson.M{"$unset": bson.M{"uid": ""}}); err != nil {
				return err
			}
			if _, err := db.Collection("recorded_sessions").UpdateMany(context.TODO(), bson.M{}, bson.M{"$unset": bson.M{"message": ""}}); err != nil {
				return err
			}
			if _, err := db.Collection("recorded_sessions").Indexes().DropOne(context.TODO(), "uid"); err != nil {
				return err
			}
			_, err := db.Collection("recorded_sessions").Indexes().DropOne(context.TODO(), "message")
			return err
		},
	},
	{
		Version: 8,
		Up: func(db *mongo.Database) error {
			mod := mongo.IndexModel{
				Keys:    bson.D{{"recorded", 1}},
				Options: options.Index().SetName("recorded").SetUnique(false),
			}
			if _, err := db.Collection("sessions").Indexes().CreateOne(context.TODO(), mod); err != nil {
				return err
			}
			_, err := db.Collection("sessions").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"recorded": false}})
			return err
		},
		Down: func(db *mongo.Database) error {
			if _, err := db.Collection("sessions").UpdateMany(context.TODO(), bson.M{}, bson.M{"$unset": bson.M{"recorded": ""}}); err != nil {
				return err
			}
			_, err := db.Collection("sessions").Indexes().DropOne(context.TODO(), "recorded")
			return err
		},
	},
	{
		Version: 9,
		Up: func(db *mongo.Database) error {
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
			return nil
		},
	},
	{
		Version: 10,
		Up: func(db *mongo.Database) error {
			mod := mongo.IndexModel{
				Keys:    bson.D{{"session_record", 1}},
				Options: options.Index().SetName("session_record").SetUnique(false),
			}
			if _, err := db.Collection("users").Indexes().CreateOne(context.TODO(), mod); err != nil {
				return err
			}
			_, err := db.Collection("users").UpdateMany(context.TODO(), bson.M{}, bson.M{"$set": bson.M{"session_record": true}})
			return err
		},
		Down: func(db *mongo.Database) error {
			return nil
		},
	},
	{
		Version: 11,
		Up: func(db *mongo.Database) error {
			mod := mongo.IndexModel{
				Keys:    bson.D{{"created_at", 1}},
				Options: options.Index().SetName("ttl").SetExpireAfterSeconds(60),
			}
			_, err := db.Collection("private_keys").Indexes().CreateOne(context.TODO(), mod)
			if err != nil {
				return err
			}

			return nil
		},
		Down: func(db *mongo.Database) error {
			_, err := db.Collection("private_keys").Indexes().DropOne(context.TODO(), "ttl")
			return err
		},
	},
	{
		Version: 12,
		Up: func(db *mongo.Database) error {
			mod := mongo.IndexModel{
				Keys:    bson.D{{"tenant_id", 1}},
				Options: options.Index().SetName("tenant_id").SetUnique(true),
			}
			if _, err := db.Collection("namespaces").Indexes().CreateOne(context.TODO(), mod); err != nil {
				return err
			}
			mod = mongo.IndexModel{
				Keys:    bson.D{{"name", 1}},
				Options: options.Index().SetName("name").SetUnique(true),
			}
			_, err := db.Collection("namespaces").Indexes().CreateOne(context.TODO(), mod)
			return err
		},
		Down: func(db *mongo.Database) error {
			_, err := db.Collection("namespaces").Indexes().DropOne(context.TODO(), "tenant_id")
			return err
		},
	},
	{
		Version: 13,
		Up: func(db *mongo.Database) error {
			mod := mongo.IndexModel{
				Keys:    bson.D{{"uid", 1}},
				Options: options.Index().SetName("uid").SetUnique(true),
			}
			_, err := db.Collection("devices").Indexes().CreateOne(context.TODO(), mod)
			if err != nil {
				return err
			}

			mod = mongo.IndexModel{
				Keys:    bson.D{{"last_seen", 1}},
				Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(30),
			}
			_, err = db.Collection("connected_devices").Indexes().CreateOne(context.TODO(), mod)
			if err != nil {
				return err
			}

			mod = mongo.IndexModel{
				Keys:    bson.D{{"uid", 1}},
				Options: options.Index().SetName("uid").SetUnique(false),
			}
			_, err = db.Collection("connected_devices").Indexes().CreateOne(context.TODO(), mod)
			if err != nil {
				return err
			}

			mod = mongo.IndexModel{
				Keys:    bson.D{{"uid", 1}},
				Options: options.Index().SetName("uid").SetUnique(true),
			}
			_, err = db.Collection("sessions").Indexes().CreateOne(context.TODO(), mod)
			if err != nil {
				return err
			}

			mod = mongo.IndexModel{
				Keys:    bson.D{{"last_seen", 1}},
				Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(30),
			}
			_, err = db.Collection("active_sessions").Indexes().CreateOne(context.TODO(), mod)
			if err != nil {
				return err
			}

			mod = mongo.IndexModel{
				Keys:    bson.D{{"uid", 1}},
				Options: options.Index().SetName("uid").SetUnique(false),
			}
			_, err = db.Collection("active_sessions").Indexes().CreateOne(context.TODO(), mod)
			if err != nil {
				return err
			}

			mod = mongo.IndexModel{
				Keys:    bson.D{{"username", 1}},
				Options: options.Index().SetName("username").SetUnique(true),
			}
			_, err = db.Collection("users").Indexes().CreateOne(context.TODO(), mod)
			if err != nil {
				return err
			}

			mod = mongo.IndexModel{
				Keys:    bson.D{{"tenant_id", 1}},
				Options: options.Index().SetName("tenant_id").SetUnique(true),
			}
			_, err = db.Collection("users").Indexes().CreateOne(context.TODO(), mod)
			if err != nil {
				return err
			}
			return nil
		},
		Down: func(db *mongo.Database) error {
			if _, err := db.Collection("devices").Indexes().DropOne(context.TODO(), "uid"); err != nil {
				return err
			}
			if _, err := db.Collection("connected_devices").Indexes().DropOne(context.TODO(), "last_seen"); err != nil {
				return err
			}
			if _, err := db.Collection("connected_devices").Indexes().DropOne(context.TODO(), "uid"); err != nil {
				return err
			}
			if _, err := db.Collection("sessions").Indexes().DropOne(context.TODO(), "uid"); err != nil {
				return err
			}
			if _, err := db.Collection("active_sessions").Indexes().DropOne(context.TODO(), "last_seen"); err != nil {
				return err
			}
			if _, err := db.Collection("users").Indexes().DropOne(context.TODO(), "username"); err != nil {
				return err
			}
			_, err := db.Collection("users").Indexes().DropOne(context.TODO(), "tenant_id")
			return err
		},
	},
	{
		Version: 14,
		Up: func(db *mongo.Database) error {
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
					Members:  []string{user.ID},
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
			}

			return err
		},
	},
	{
		Version: 15,
		Up: func(db *mongo.Database) error {
			mod := mongo.IndexModel{
				Keys:    bson.D{{"fingerprint", 1}},
				Options: options.Index().SetName("fingerprint").SetUnique(true),
			}
			_, err := db.Collection("public_keys").Indexes().CreateOne(context.TODO(), mod)
			return err
		},

		Down: func(db *mongo.Database) error {
			_, err := db.Collection("public_keys").Indexes().DropOne(context.TODO(), "fingerprint")
			return err
		},
	},
}

func ApplyMigrations(db *mongo.Database) error {
	m := migrate.NewMigrate(db, migrations...)

	return m.Up(migrate.AllAvailable)
}
