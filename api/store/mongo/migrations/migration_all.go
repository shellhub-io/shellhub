package migrations

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"os"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/geoip/geolite2"
	"github.com/shellhub-io/shellhub/pkg/hash"
	"github.com/shellhub-io/shellhub/pkg/models"
	// "github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

var migration1 = migrate.Migration{
	Version:     1,
	Description: "Create the database for the system",
	Up: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   1,
			"action":    "Up",
		}).Info("Applying migration")

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   1,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration2 = migrate.Migration{
	Version:     2,
	Description: "Rename the column device to device_uid",
	Up: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   2,
			"action":    "Up",
		}).Info("Applying migration")

		return renameField(db, "sessions", "device", "device_uid")
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   2,
			"action":    "Down",
		}).Info("Applying migration")

		return renameField(db, "sessions", "device_uid", "device")
	}),
}

var migration3 = migrate.Migration{
	Version:     3,
	Description: "Rename the column attributes to info",
	Up: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   3,
			"action":    "Up",
		}).Info("Applying migration")

		return renameField(db, "devices", "attributes", "info")
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   3,
			"action":    "Down",
		}).Info("Applying migration")

		return renameField(db, "devices", "info", "attributes")
	}),
}

var migration4 = migrate.Migration{
	Version:     4,
	Description: "Rename the column version to info.version",
	Up: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   4,
			"action":    "Up",
		}).Info("Applying migration")

		return renameField(db, "devices", "version", "info.version")
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   4,
			"action":    "Down",
		}).Info("Applying migration")

		return renameField(db, "devices", "info.version", "version")
	}),
}

var migration5 = migrate.Migration{
	Version:     5,
	Description: "Set the email as unique on users collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   5,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"email", 1}},
			Options: options.Index().SetName("email").SetUnique(true),
		}
		_, err := db.Collection("users").Indexes().CreateOne(ctx, mod)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   5,
			"action":    "Down",
		}).Info("Applying migration")
		_, err := db.Collection("users").Indexes().DropOne(ctx, "email")

		return err
	}),
}

var migration6 = migrate.Migration{
	Version:     6,
	Description: "Unset unique on status in the devices collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   6,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"status", 1}},
			Options: options.Index().SetName("status").SetUnique(false),
		}
		if _, err := db.Collection("devices").Indexes().CreateOne(ctx, mod); err != nil {
			return err
		}
		_, err := db.Collection("devices").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"status": "accepted"}})

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   6,
			"action":    "Down",
		}).Info("Applying migration")
		if _, err := db.Collection("devices").UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"status": ""}}); err != nil {
			return err
		}
		_, err := db.Collection("status").Indexes().DropOne(ctx, "status")

		return err
	}),
}

var migration7 = migrate.Migration{
	Version:     7,
	Description: "Unset unique on uid and message in the recoded_sessions collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   7,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"uid", 1}},
			Options: options.Index().SetName("uid").SetUnique(false),
		}
		if _, err := db.Collection("recorded_sessions").Indexes().CreateOne(ctx, mod); err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"message", 1}},
			Options: options.Index().SetName("message").SetUnique(false),
		}
		_, err := db.Collection("recorded_sessions").Indexes().CreateOne(ctx, mod)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   7,
			"action":    "Down",
		}).Info("Applying migration")
		if _, err := db.Collection("recorded_sessions").UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"uid": ""}}); err != nil {
			return err
		}
		if _, err := db.Collection("recorded_sessions").UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"message": ""}}); err != nil {
			return err
		}
		if _, err := db.Collection("recorded_sessions").Indexes().DropOne(ctx, "uid"); err != nil {
			return err
		}
		_, err := db.Collection("recorded_sessions").Indexes().DropOne(ctx, "message")

		return err
	}),
}

var migration8 = migrate.Migration{
	Version:     8,
	Description: "Unset unique on recorded in the sessions collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   8,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"recorded", 1}},
			Options: options.Index().SetName("recorded").SetUnique(false),
		}
		if _, err := db.Collection("sessions").Indexes().CreateOne(ctx, mod); err != nil {
			return err
		}
		_, err := db.Collection("sessions").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"recorded": false}})

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   8,
			"action":    "Down",
		}).Info("Applying migration")
		if _, err := db.Collection("sessions").UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"recorded": ""}}); err != nil {
			return err
		}
		_, err := db.Collection("sessions").Indexes().DropOne(ctx, "recorded")

		return err
	}),
}

var migration9 = migrate.Migration{
	Version:     9,
	Description: "Set all devices names to lowercase in the devices collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   9,
			"action":    "Up",
		}).Info("Applying migration")
		cursor, err := db.Collection("devices").Find(ctx, bson.D{})
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			device := new(models.Device)
			err := cursor.Decode(&device)
			if err != nil {
				return err
			}

			device.Name = strings.ToLower(device.Name)
			if _, err = db.Collection("devices").UpdateOne(ctx, bson.M{"uid": device.UID}, bson.M{"$set": bson.M{"name": strings.ToLower(device.Name)}}); err != nil {
				return err
			}
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   9,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration10 = migrate.Migration{
	Version:     10,
	Description: "Unset unique on session_record in the users collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   10,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"session_record", 1}},
			Options: options.Index().SetName("session_record").SetUnique(false),
		}
		if _, err := db.Collection("users").Indexes().CreateOne(ctx, mod); err != nil {
			return err
		}
		_, err := db.Collection("users").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"session_record": true}})

		return err
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   10,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration11 = migrate.Migration{
	Version:     11,
	Description: "Create a ttl for the private_keys collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   11,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"created_at", 1}},
			Options: options.Index().SetName("ttl").SetExpireAfterSeconds(60),
		}
		_, err := db.Collection("private_keys").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   11,
			"action":    "Down",
		}).Info("Applying migration")
		_, err := db.Collection("private_keys").Indexes().DropOne(ctx, "ttl")

		return err
	}),
}

var migration12 = migrate.Migration{
	Version:     12,
	Description: "Set the tenant_id as unique in the namespaces collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   12,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"tenant_id", 1}},
			Options: options.Index().SetName("tenant_id").SetUnique(true),
		}
		if _, err := db.Collection("namespaces").Indexes().CreateOne(ctx, mod); err != nil {
			return err
		}
		mod = mongo.IndexModel{
			Keys:    bson.D{{"name", 1}},
			Options: options.Index().SetName("name").SetUnique(true),
		}
		_, err := db.Collection("namespaces").Indexes().CreateOne(ctx, mod)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   12,
			"action":    "Down",
		}).Info("Applying migration")
		_, err := db.Collection("namespaces").Indexes().DropOne(ctx, "tenant_id")

		return err
	}),
}

var migration13 = migrate.Migration{
	Version:     13,
	Description: "Change on several collections",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   13,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"uid", 1}},
			Options: options.Index().SetName("uid").SetUnique(true),
		}
		_, err := db.Collection("devices").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"last_seen", 1}},
			Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(30),
		}
		_, err = db.Collection("connected_devices").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"uid", 1}},
			Options: options.Index().SetName("uid").SetUnique(false),
		}
		_, err = db.Collection("connected_devices").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"uid", 1}},
			Options: options.Index().SetName("uid").SetUnique(true),
		}
		_, err = db.Collection("sessions").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"last_seen", 1}},
			Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(30),
		}
		_, err = db.Collection("active_sessions").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"uid", 1}},
			Options: options.Index().SetName("uid").SetUnique(false),
		}
		_, err = db.Collection("active_sessions").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"username", 1}},
			Options: options.Index().SetName("username").SetUnique(true),
		}
		_, err = db.Collection("users").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		mod = mongo.IndexModel{
			Keys:    bson.D{{"tenant_id", 1}},
			Options: options.Index().SetName("tenant_id").SetUnique(true),
		}
		_, err = db.Collection("users").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   13,
			"action":    "Down",
		}).Info("Applying migration")
		if _, err := db.Collection("devices").Indexes().DropOne(ctx, "uid"); err != nil {
			return err
		}
		if _, err := db.Collection("connected_devices").Indexes().DropOne(ctx, "last_seen"); err != nil {
			return err
		}
		if _, err := db.Collection("connected_devices").Indexes().DropOne(ctx, "uid"); err != nil {
			return err
		}
		if _, err := db.Collection("sessions").Indexes().DropOne(ctx, "uid"); err != nil {
			return err
		}
		if _, err := db.Collection("active_sessions").Indexes().DropOne(ctx, "last_seen"); err != nil {
			return err
		}
		if _, err := db.Collection("users").Indexes().DropOne(ctx, "username"); err != nil {
			return err
		}
		_, err := db.Collection("users").Indexes().DropOne(ctx, "tenant_id")

		return err
	}),
}

var migration14 = migrate.Migration{
	Version:     14,
	Description: "Set the right tenant_id in the users collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   14,
			"action":    "Up",
		}).Info("Applying migration")
		type user struct {
			Username      string `json:"username" bson:",omitempty"`
			TenantID      string `json:"tenant_id" bson:"tenant_id"`
			ID            string `json:"id,omitempty" bson:"_id,omitempty"`
			SessionRecord bool   `json:"session_record" bson:"session_record,omitempty"`
		}
		if _, err := db.Collection("users").Indexes().DropOne(ctx, "tenant_id"); err != nil {
			return err
		}
		if _, err := db.Collection("users").Indexes().DropOne(ctx, "session_record"); err != nil {
			return err
		}

		cursor, err := db.Collection("users").Find(ctx, bson.D{})
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			user := new(user)
			err := cursor.Decode(&user)
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

			settings := NamespaceSettings{
				SessionRecord: true,
			}

			namespace := Namespace{
				Owner:    user.ID,
				Members:  []interface{}{user.ID},
				TenantID: user.TenantID,
				Name:     user.Username,
				Settings: &settings,
			}

			_, err = db.Collection("namespaces").InsertOne(ctx, &namespace)
			if err != nil {
				return nil
			}

			if _, err := db.Collection("users").UpdateOne(ctx, bson.M{"tenant_id": user.TenantID}, bson.M{"$unset": bson.M{"tenant_id": ""}}); err != nil {
				return err
			}
			if _, err := db.Collection("users").UpdateOne(ctx, bson.M{"tenant_id": user.TenantID}, bson.M{"$unset": bson.M{"session_record": ""}}); err != nil {
				return err
			}
		}

		return cursor.Err()
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   14,
			"action":    "Down",
		}).Info("Applying migration")
		cursor, err := db.Collection("namespaces").Find(ctx, bson.D{})
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			namespace := new(models.Namespace)
			err := cursor.Decode(&namespace)
			if err != nil {
				return err
			}

			_, err = db.Collection("users").UpdateOne(ctx, bson.M{"_id": namespace.Owner}, bson.M{"$set": bson.M{"tenant": namespace.TenantID}})
			if err != nil {
				return err
			}
		}

		return err
	}),
}

var migration15 = migrate.Migration{
	Version:     15,
	Description: "Set all names to lowercase in the namespaces",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   15,
			"action":    "Up",
		}).Info("Applying migration")
		_, err := db.Collection("namespaces").UpdateMany(ctx, bson.M{}, []bson.M{
			{
				"$set": bson.M{
					"name": bson.M{"$toLower": "$name"},
				},
			},
		})

		return err
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   15,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration16 = migrate.Migration{
	Version:     16,
	Description: "Set the fingerprint as unique on public_keys collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   16,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"fingerprint", 1}},
			Options: options.Index().SetName("fingerprint").SetUnique(true),
		}
		_, err := db.Collection("public_keys").Indexes().CreateOne(ctx, mod)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   16,
			"action":    "Down",
		}).Info("Applying migration")
		_, err := db.Collection("public_keys").Indexes().DropOne(ctx, "fingerprint")

		return err
	}),
}

var migration17 = migrate.Migration{
	Version:     17,
	Description: "Remove the namespaces, devices, session, connected_devices, firewall_rules and public_keys in the users",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   17,
			"action":    "Up",
		}).Info("Applying migration")
		cursor, err := db.Collection("namespaces").Find(ctx, bson.D{})
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

		for cursor.Next(ctx) {
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
			if err := db.Collection("users").FindOne(ctx, bson.M{"_id": objID}).Decode(&user); err != nil {
				if err != mongo.ErrNoDocuments {
					return err
				}
				if _, err := db.Collection("namespaces").DeleteOne(ctx, bson.M{"tenant_id": namespace.TenantID}); err != nil {
					return err
				}
			}
		}

		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(ctx)

		cursor, err = db.Collection("devices").Find(ctx, bson.D{})
		if err != nil {
			return err
		}

		for cursor.Next(ctx) {
			device := new(models.Device)
			err = cursor.Decode(&device)
			if err != nil {
				return err
			}

			namespace := Namespace{}
			if err := db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": device.TenantID}).Decode(&namespace); err != nil {
				if err != mongo.ErrNoDocuments {
					return err
				}
				if _, err := db.Collection("devices").DeleteOne(ctx, bson.M{"uid": device.UID}); err != nil {
					return err
				}

				if _, err := db.Collection("sessions").DeleteMany(ctx, bson.M{"device_uid": device.UID}); err != nil {
					return err
				}

				if _, err := db.Collection("connected_devices").DeleteMany(ctx, bson.M{"uid": device.UID}); err != nil {
					return err
				}
			}
		}
		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(ctx)

		cursor, err = db.Collection("firewall_rules").Find(ctx, bson.D{})
		if err != nil {
			return err
		}
		for cursor.Next(ctx) {
			rule := new(models.FirewallRule)
			err := cursor.Decode(&rule)
			if err != nil {
				return err
			}

			namespace := Namespace{}
			if err := db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": rule.TenantID}).Decode(&namespace); err != nil {
				if err != mongo.ErrNoDocuments {
					return err
				}
				if _, err := db.Collection("firewall_rules").DeleteOne(ctx, bson.M{"tenant_id": rule.TenantID}); err != nil {
					return err
				}
			}
		}
		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(ctx)

		cursor, err = db.Collection("public_keys").Find(ctx, bson.D{})
		if err != nil {
			return err
		}

		for cursor.Next(ctx) {
			key := new(models.PublicKey)
			err := cursor.Decode(&key)
			if err != nil {
				return err
			}
			namespace := Namespace{}
			if err := db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": key.TenantID}).Decode(&namespace); err != nil {
				if err != mongo.ErrNoDocuments {
					return err
				}
				if _, err := db.Collection("public_keys").DeleteOne(ctx, bson.M{"tenant_id": key.TenantID}); err != nil {
					return err
				}
			}
		}
		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(ctx)

		return err
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   17,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration18 = migrate.Migration{
	Version:     18,
	Description: "Set the max_devices value in the namespaces collection to 3 on enterprise",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   18,
			"action":    "Up",
		}).Info("Applying migration")
		if envs.IsEnterprise() {
			_, err := db.Collection("namespaces").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"max_devices": 3}})

			return err
		}
		_, err := db.Collection("namespaces").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"max_devices": -1}})

		return err
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   18,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration19 = migrate.Migration{
	Version:     19,
	Description: "Remove all fingerprint associated with a public keys collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   19,
			"action":    "Up",
		}).Info("Applying migration")
		_, err := db.Collection("public_keys").Indexes().DropOne(ctx, "fingerprint")

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   19,
			"action":    "Down",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"fingerprint", 1}},
			Options: options.Index().SetName("fingerprint").SetUnique(true),
		}
		_, err := db.Collection("public_keys").Indexes().CreateOne(ctx, mod)

		return err
	}),
}

var migration20 = migrate.Migration{
	Version:     20,
	Description: "Change the model on db for firewall_rules collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   20,
			"action":    "Up",
		}).Info("Applying migration")
		cursor, err := db.Collection("firewall_rules").Find(ctx, bson.D{})
		if err != nil {
			return err
		}

		type firewallRule struct {
			ID                        primitive.ObjectID `json:"id" bson:"_id"`
			TenantID                  string             `json:"tenant_id" bson:"tenant_id"`
			models.FirewallRuleFields `bson:",inline"`
		}

		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			firewall := new(models.FirewallRule)
			err := cursor.Decode(&firewall)
			if err != nil {
				return err
			}
			objID, err := primitive.ObjectIDFromHex(firewall.ID)
			replacedRule := firewallRule{
				TenantID:           firewall.TenantID,
				ID:                 objID,
				FirewallRuleFields: firewall.FirewallRuleFields,
			}

			if err == nil {
				if errDelete := db.Collection("firewall_rules").FindOneAndDelete(ctx, bson.M{"_id": firewall.ID}); errDelete.Err() != nil {
					continue
				}

				if _, err := db.Collection("firewall_rules").InsertOne(ctx, replacedRule); err != nil {
					return err
				}
			}
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   20,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration21 = migrate.Migration{
	Version:     21,
	Description: "Remove all sessions, recorded_sessions for the devices",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   21,
			"action":    "Up",
		}).Info("Applying migration")
		cursor, err := db.Collection("sessions").Find(ctx, bson.D{})
		if err != nil {
			return err
		}

		for cursor.Next(ctx) {
			session := new(models.Session)
			err = cursor.Decode(&session)
			if err != nil {
				return err
			}

			device := new(models.Device)
			if err := db.Collection("devices").FindOne(ctx, bson.M{"uid": session.DeviceUID}).Decode(&device); err != nil {
				if err != mongo.ErrNoDocuments {
					return err
				}

				if _, err := db.Collection("sessions").DeleteMany(ctx, bson.M{"device_uid": session.DeviceUID}); err != nil {
					return err
				}
			}
		}

		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(ctx)

		cursor, err = db.Collection("recorded_sessions").Find(ctx, bson.D{})
		if err != nil {
			return err
		}

		for cursor.Next(ctx) {
			record := new(models.RecordedSession)
			err = cursor.Decode(&record)
			if err != nil {
				return err
			}

			namespace := new(models.Namespace)
			if err := db.Collection("namespaces").FindOne(ctx, bson.M{"tenant_id": record.TenantID}).Decode(&namespace); err != nil {
				if err != mongo.ErrNoDocuments {
					return err
				}

				if _, err := db.Collection("recorded_sessions").DeleteMany(ctx, bson.M{"tenant_id": record.TenantID}); err != nil {
					return err
				}
			}
			session := new(models.Session)
			if err := db.Collection("sessions").FindOne(ctx, bson.M{"uid": record.UID}).Decode(&session); err != nil {
				if err != mongo.ErrNoDocuments {
					return err
				}

				if _, err := db.Collection("recorded_sessions").DeleteMany(ctx, bson.M{"uid": record.UID}); err != nil {
					return err
				}
			}

		}
		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(ctx)

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   21,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration22 = migrate.Migration{
	Version:     22,
	Description: "Insert the user on the members group for the namespace",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   22,
			"action":    "Up",
		}).Info("Applying migration")
		cursor, err := db.Collection("namespaces").Find(ctx, bson.D{})
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

		for cursor.Next(ctx) {

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
				if err := db.Collection("users").FindOne(ctx, bson.M{"_id": objID}).Decode(&user); err != nil {
					if _, err := db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": namespace.TenantID}, bson.M{"$pull": bson.M{"members": memberID}}); err != nil {
						return err
					}
				}
			}
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   22,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration23 = migrate.Migration{
	Version:     23,
	Description: "change dot in namespace name and hostname to -",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   23,
			"action":    "Up",
		}).Info("Applying migration")
		if _, err := db.Collection("namespaces").UpdateMany(ctx, bson.D{}, []bson.M{
			{
				"$set": bson.M{
					"name": bson.M{
						"$replaceAll": bson.M{"input": "$name", "find": ".", "replacement": "-"},
					},
				},
			},
		}); err != nil {
			return err
		}

		if _, err := db.Collection("devices").UpdateMany(ctx, bson.D{}, []bson.M{
			{
				"$set": bson.M{
					"name": bson.M{
						"$replaceAll": bson.M{"input": "$name", "find": ".", "replacement": "-"},
					},
				},
			},
		}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   23,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration24 = migrate.Migration{
	Version:     24,
	Description: "convert names and emails to lowercase",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   24,
			"action":    "Up",
		}).Info("Applying migration")
		if _, err := db.Collection("users").UpdateMany(ctx, bson.D{}, []bson.M{
			{
				"$set": bson.M{
					"username": bson.M{"$toLower": "$username"},
					"email":    bson.M{"$toLower": "$email"},
				},
			},
		}); err != nil {
			return err
		}

		_, err := db.Collection("namespaces").UpdateMany(ctx, bson.D{}, []bson.M{
			{
				"$set": bson.M{"name": bson.M{"$toLower": "$name"}},
			},
		})

		return err
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   24,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration25 = migrate.Migration{
	Version:     25,
	Description: "remove devices with no namespaces related",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   25,
			"action":    "Up",
		}).Info("Applying migration")
		query := []bson.M{
			{
				"$lookup": bson.M{
					"from":         "namespaces",
					"localField":   "tenant_id",
					"foreignField": "tenant_id",
					"as":           "namespace",
				},
			},
			{
				"$addFields": bson.M{
					"namespace": bson.M{"$anyElementTrue": []interface{}{"$namespace"}},
				},
			},

			{
				"$match": bson.M{
					"namespace": bson.M{"$eq": true},
				},
			},

			{
				"$unset": "namespace",
			},

			{
				"$out": "devices",
			},
		}

		_, err := db.Collection("devices").Aggregate(ctx, query)

		return err
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   25,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration26 = migrate.Migration{
	Version:     26,
	Description: "Create collection used to recover password and activate account",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   26,
			"action":    "Up",
		}).Info("Applying migration")
		indexModel := mongo.IndexModel{
			Keys:    bson.D{{"created_at", 1}},
			Options: options.Index().SetName("ttl").SetExpireAfterSeconds(86400),
		}
		_, err := db.Collection("recovery_tokens").Indexes().CreateOne(ctx, indexModel)
		if err != nil {
			return err
		}

		indexModel = mongo.IndexModel{
			Keys:    bson.D{{"token", 1}},
			Options: options.Index().SetName("token").SetUnique(false),
		}
		if _, err := db.Collection("recovery_tokens").Indexes().CreateOne(ctx, indexModel); err != nil {
			return err
		}

		indexModel = mongo.IndexModel{
			Keys:    bson.D{{"user", 1}},
			Options: options.Index().SetName("user").SetUnique(false),
		}
		if _, err := db.Collection("recovery_tokens").Indexes().CreateOne(ctx, indexModel); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   26,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration27 = migrate.Migration{
	Version:     27,
	Description: "Set closed field in the sessions",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   27,
			"action":    "Up",
		}).Info("Applying migration")
		_, err := db.Collection("sessions").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"closed": true}})

		return err
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   27,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration28 = migrate.Migration{
	Version:     28,
	Description: "add timestamps fields to collections users and devices",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   28,
			"action":    "Up",
		}).Info("Applying migration")
		if _, err := db.Collection("users").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"created_at": clock.Now()}}); err != nil {
			return err
		}

		if _, err := db.Collection("devices").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"created_at": clock.Now()}}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   28,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration29 = migrate.Migration{
	Version:     29,
	Description: "add last_login field to collection users",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   29,
			"action":    "Up",
		}).Info("Applying migration")
		if _, err := db.Collection("users").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"last_login": nil}}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   29,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration30 = migrate.Migration{
	Version:     30,
	Description: "add remote_addr field to collection devices",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   30,
			"action":    "Up",
		}).Info("Applying migration")
		if _, err := db.Collection("devices").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"remote_addr": ""}}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   30,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration31 = migrate.Migration{
	Version:     31,
	Description: "add last_login field to collection namespaces",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   31,
			"action":    "Up",
		}).Info("Applying migration")
		if _, err := db.Collection("namespaces").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"created_at": clock.Now()}}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   31,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration32 = migrate.Migration{
	Version:     32,
	Description: "add authenticated field to collection users",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   32,
			"action":    "Up",
		}).Info("Applying migration")
		if _, err := db.Collection("users").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"authenticated": true}}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   32,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration33 = migrate.Migration{
	Version:     33,
	Description: "add tags field to collection devices",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   33,
			"action":    "Up",
		}).Info("Applying migration")
		mod := mongo.IndexModel{
			Keys:    bson.D{{"tags", 1}},
			Options: options.Index().SetName("tags").SetUnique(false),
		}
		_, err := db.Collection("devices").Indexes().CreateOne(ctx, mod)
		if err != nil {
			return err
		}

		if _, err := db.Collection("devices").UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"tags": []string{}}}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   33,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration34 = migrate.Migration{
	Version:     34,
	Description: "create online index in devices collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   34,
			"action":    "Up",
		}).Info("Applying migration")

		indexModel := mongo.IndexModel{
			Keys: bson.D{{"online", 1}},
		}

		_, err := db.Collection("devices").Indexes().CreateOne(ctx, indexModel)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   34,
			"action":    "Down",
		}).Info("Applying migration")

		_, err := db.Collection("devices").Indexes().DropOne(ctx, "online")

		return err
	}),
}

var migration35 = migrate.Migration{
	Version:     35,
	Description: "Rename the column authenticated to confirmed",
	Up: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   35,
			"action":    "Up",
		}).Info("Applying migration")

		return renameField(db, "users", "authenticated", "confirmed")
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   35,
			"action":    "Down",
		}).Info("Applying migration")

		return renameField(db, "users", "confirmed", "authenticated")
	}),
}

var migration36 = migrate.Migration{
	Version:     36,
	Description: "update max_devices to 3",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   36,
			"action":    "Up",
		}).Info("Applying migration")

		if envs.IsCloud() {
			if _, err := db.Collection("namespaces").UpdateMany(ctx, bson.M{"billing": nil}, bson.M{"$set": bson.M{"max_devices": 3}}); err != nil {
				return err
			}
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   36,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration37 = migrate.Migration{
	Version:     37,
	Description: "Change member's role from array of ID to a list of members' object",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   37,
			"action":    "Up",
		}).Info("Applying migration")
		cursor, err := db.Collection("namespaces").Find(ctx, bson.D{})
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

		for cursor.Next(ctx) {
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
						Role: authorizer.RoleObserver,
					}

					memberList = append(memberList, m)
				} else if owner == member {
					m := models.Member{
						ID:   member.(string),
						Role: authorizer.RoleOwner,
					}

					memberList = append(memberList, m)
				}
			}

			if _, err := db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": namespace.TenantID}, bson.M{"$set": bson.M{"members": memberList}}); err != nil {
				return err
			}
		}

		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(ctx)

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   37,
			"action":    "Down",
		}).Info("Applying migration")
		cursor, err := db.Collection("namespaces").Find(ctx, bson.D{})
		if err != nil {
			return err
		}

		for cursor.Next(ctx) {
			namespace := new(models.Namespace)
			err = cursor.Decode(&namespace)
			if err != nil {
				return err
			}

			var membersList []interface{}
			for _, member := range namespace.Members {
				membersList = append(membersList, member.ID)
			}

			if _, err := db.Collection("namespaces").UpdateOne(ctx, bson.M{"tenant_id": namespace.TenantID}, bson.M{"$set": bson.M{"members": membersList}}); err != nil {
				return err
			}
		}

		if err := cursor.Err(); err != nil {
			return err
		}

		cursor.Close(ctx)

		return nil
	}),
}

var migration38 = migrate.Migration{
	Version:     38,
	Description: "Set last_login to created_at, when created_at is a zero value",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   38,
			"action":    "Up",
		}).Info("Applying migration")
		zeroTime := time.Time{}.UTC()
		_, err := db.Collection("users").Aggregate(ctx,
			mongo.Pipeline{
				{
					{"$match", bson.D{
						{"$and", bson.A{
							bson.M{"$or": bson.A{
								bson.M{"created_at": bson.M{"$eq": zeroTime}},
								bson.M{"created_at": bson.M{"$eq": nil}},
							}},
							bson.M{"last_login": bson.M{"$ne": zeroTime}},
						}},
					}},
				},
				{
					{"$replaceRoot", bson.D{{"newRoot", bson.M{"$mergeObjects": bson.A{"$$ROOT", bson.M{"created_at": "$last_login"}}}}}},
				},
				{
					{"$merge", bson.M{"into": "users"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   38,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration39 = migrate.Migration{
	Version:     39,
	Description: "remove online index from devices collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   39,
			"action":    "Up",
		}).Info("Applying migration")

		if _, err := db.Collection("devices").Indexes().DropOne(ctx, "online_1"); err != nil {
			return err
		}

		_, err := db.Collection("devices").UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"online": nil}})

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   39,
			"action":    "Down",
		}).Info("Applying migration")

		indexModel := mongo.IndexModel{
			Keys: bson.D{{"online", 1}},
		}

		_, err := db.Collection("devices").Indexes().CreateOne(ctx, indexModel)

		return err
	}),
}

var migration40 = migrate.Migration{
	Version:     40,
	Description: "remove online index from devices collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   40,
			"action":    "Up",
		}).Info("Applying migration")

		if _, err := db.Collection("connected_devices").Indexes().DropOne(ctx, "last_seen"); err != nil {
			return err
		}

		mod := mongo.IndexModel{
			Keys:    bson.D{{"last_seen", 1}},
			Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(60),
		}
		if _, err := db.Collection("connected_devices").Indexes().CreateOne(ctx, mod); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   40,
			"action":    "Down",
		}).Info("Applying migration")

		if _, err := db.Collection("connected_devices").Indexes().DropOne(ctx, "last_seen"); err != nil {
			return err
		}

		mod := mongo.IndexModel{
			Keys:    bson.D{{"last_seen", 1}},
			Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(30),
		}
		if _, err := db.Collection("connected_devices").Indexes().CreateOne(ctx, mod); err != nil {
			return err
		}

		return nil
	}),
}

var migration41 = migrate.Migration{
	Version:     41,
	Description: "update online index from devices collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   41,
			"action":    "Up",
		}).Info("Applying migration")

		if _, err := db.Collection("connected_devices").Indexes().DropOne(ctx, "last_seen"); err != nil {
			return err
		}

		mod := mongo.IndexModel{
			Keys:    bson.D{{"last_seen", 1}},
			Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(120),
		}
		if _, err := db.Collection("connected_devices").Indexes().CreateOne(ctx, mod); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   41,
			"action":    "Down",
		}).Info("Applying migration")

		if _, err := db.Collection("connected_devices").Indexes().DropOne(ctx, "last_seen"); err != nil {
			return err
		}

		mod := mongo.IndexModel{
			Keys:    bson.D{{"last_seen", 1}},
			Options: options.Index().SetName("last_seen").SetExpireAfterSeconds(60),
		}
		if _, err := db.Collection("connected_devices").Indexes().CreateOne(ctx, mod); err != nil {
			return err
		}

		return nil
	}),
}

var migration42 = migrate.Migration{
	Version:     42,
	Description: "change hostname to filter hostname",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   42,
			"action":    "Up",
		}).Info("Applying migration")

		_, err := db.Collection("public_keys").Aggregate(ctx,
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$set", bson.M{"filter.hostname": "$hostname"}},
				},
				{
					{"$unset", "hostname"},
				},
				{
					{"$merge", bson.M{"into": "public_keys", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   42,
			"action":    "Down",
		}).Info("Applying migration")

		_, err := db.Collection("public_keys").Aggregate(ctx,
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$set", bson.M{"hostname": "$filter.hostname"}},
				},
				{
					{"$unset", "filter"},
				},
				{
					{"$merge", bson.M{"into": "public_keys", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
}

var migration43 = migrate.Migration{
	Version:     43,
	Description: "add tags field to firewall_rules collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   43,
			"action":    "Up",
		}).Info("Applying migration")

		_, err := db.Collection("firewall_rules").Aggregate(ctx,
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$set", bson.M{"filter.hostname": "$hostname"}},
				},
				{
					{"$unset", "hostname"},
				},
				{
					{"$merge", bson.M{"into": "firewall_rules", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   43,
			"action":    "Down",
		}).Info("Applying migration")

		_, err := db.Collection("firewall_rules").Aggregate(ctx,
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$set", bson.M{"hostname": "$filter.hostname"}},
				},
				{
					{"$unset", "filter"},
				},
				{
					{"$merge", bson.M{"into": "firewall_rules", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
}

var migration44 = migrate.Migration{
	Version:     44,
	Description: "remove duplicated tags on public keys",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   44,
			"action":    "Up",
		}).Info("Applying migration")

		_, err := db.Collection("public_keys").Aggregate(ctx,
			mongo.Pipeline{
				{
					{"$match", bson.M{"filter.tags": bson.M{"$exists": true}}},
				},
				{
					{"$unwind", "$filter.tags"},
				},
				{
					{"$group", bson.M{
						"_id":  "$_id",
						"body": bson.M{"$push": "$$ROOT"},
						"tags": bson.M{
							"$addToSet": "$filter.tags",
						},
						"count": bson.M{
							"$sum": 1,
						},
					}},
				},
				{
					{"$replaceRoot", bson.D{{"newRoot", bson.M{"$mergeObjects": bson.A{bson.M{"$arrayElemAt": bson.A{"$body", 0}}, bson.M{"filter": bson.M{"tags": "$tags"}}, bson.M{"_id": "$_id"}}}}}},
				},
				{
					{"$merge", bson.M{"into": "public_keys", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   44,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration45 = migrate.Migration{
	Version:     45,
	Description: "remove duplicated tags on firewall rules",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   45,
			"action":    "Up",
		}).Info("Applying migration")

		_, err := db.Collection("firewall_rules").Aggregate(ctx,
			mongo.Pipeline{
				{
					{"$match", bson.M{"filter.tags": bson.M{"$exists": true}}},
				},
				{
					{"$unwind", "$filter.tags"},
				},
				{
					{"$group", bson.M{
						"_id":  "$_id",
						"body": bson.M{"$push": "$$ROOT"},
						"tags": bson.M{
							"$addToSet": "$filter.tags",
						},
						"count": bson.M{
							"$sum": 1,
						},
					}},
				},
				{
					{"$replaceRoot", bson.D{{"newRoot", bson.M{"$mergeObjects": bson.A{bson.M{"$arrayElemAt": bson.A{"$body", 0}}, bson.M{"filter": bson.M{"tags": "$tags"}}, bson.M{"_id": "$_id"}}}}}},
				},
				{
					{"$merge", bson.M{"into": "firewall_rules", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   45,
			"action":    "Down",
		}).Info("Applying migration")

		return nil
	}),
}

var migration46 = migrate.Migration{
	Version:     46,
	Description: "change public keys with empty username in favor of .* regexp",
	Up: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   46,
			"action":    "Up",
		}).Info("Applying migration")

		_, err := db.Collection("public_keys").Aggregate(context.Background(),
			mongo.Pipeline{
				{
					{"$match", bson.M{"username": ""}},
				},
				{
					{"$set", bson.M{"username": ".*"}},
				},
				{
					{"$merge", bson.M{"into": "public_keys", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   46,
			"action":    "Down",
		}).Info("Applying migration")

		_, err := db.Collection("public_keys").Aggregate(context.Background(),
			mongo.Pipeline{
				{
					{"$match", bson.M{"username": ".*"}},
				},
				{
					{"$set", bson.M{"username": ""}},
				},
				{
					{"$merge", bson.M{"into": "public_keys", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
}

var migration47 = migrate.Migration{
	Version:     47,
	Description: "",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   47,
			"action":    "Up",
		}).Info("Applying migration up")

		var locator geoip.Locator
		if os.Getenv("GEOIP") == "true" {
			locator, _ = geolite2.NewLocator(ctx, geolite2.FetchFromLicenseKey(os.Getenv("MAXMIND_LICENSE")))
		} else {
			locator = geoip.NewNullGeoLite()
		}

		cursor, err := db.Collection("sessions").Find(ctx, bson.D{})
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			session := new(models.Session)
			if err := cursor.Decode(session); err != nil {
				return err
			}

			position, err := locator.GetPosition(net.ParseIP(session.IPAddress))
			if err != nil {
				return err
			}

			if _, err := db.Collection("sessions").UpdateOne(ctx, bson.M{"uid": session.UID}, bson.M{"$set": bson.M{"position": position}}); err != nil {
				return err
			}
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   47,
			"action":    "Down",
		}).Info("Applying migration down")

		_, err := db.Collection("sessions").Aggregate(context.Background(),
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$unset", "position"},
				},
				{
					{"$merge", bson.M{"into": "sessions", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
}

// invertFirewallRulePriority inverts the priority of the firewall rules.
//
// The priority of the firewall rules is inverted to follow a common pattern in the industry.
//
// If any error occurs, the migration is aborted.
func invertFirewallRulePriority(db *mongo.Database) error {
	ctx := context.Background()

	type Properties struct {
		ID       string
		Priority int
	}

	options := new(options.FindOptions)
	options.SetSort(bson.D{{"priority", 1}}) // Sort by priority in ascending order.

	namespaces, err := db.Collection("namespaces").Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer namespaces.Close(ctx)

	for namespaces.Next(ctx) {
		var properties []Properties

		var namespace models.Namespace
		if err := namespaces.Decode(&namespace); err != nil {
			return err
		}

		rules, err := db.Collection("firewall_rules").Find(ctx, bson.M{"tenant_id": namespace.TenantID}, options)
		if err != nil {
			return err
		}
		defer rules.Close(ctx)

		for rules.Next(ctx) {
			rule := new(models.FirewallRule)
			if err := rules.Decode(rule); err != nil {
				return err
			}

			properties = append(properties, Properties{
				ID:       rule.ID,
				Priority: rule.Priority,
			})
		}

		for index := 0; index <= len(properties)-1; index++ {
			id, _ := primitive.ObjectIDFromHex(properties[index].ID)

			_, err := db.Collection("firewall_rules").UpdateMany(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"priority": properties[len(properties)-1-index].Priority}})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

var migration48 = migrate.Migration{
	Version:     48,
	Description: "invert Firewall priority",
	Up: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   48,
			"action":    "Up",
		}).Info("Applying migration up")

		return invertFirewallRulePriority(db)
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   48,
			"action":    "Down",
		}).Info("Applying migration down")

		return invertFirewallRulePriority(db)
	}),
}

var migration49 = migrate.Migration{
	Version:     49,
	Description: "set the number of namespaces owned by each user",
	Up: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   49,
			"action":    "Up",
		}).Info("Applying migration up")

		_, err := db.Collection("users").Aggregate(context.Background(),
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{
						"$set", bson.M{"tmp_id": bson.M{"$toString": "$_id"}}, // FIXME: I guess we could eliminate the "$_id" conversion from objectID to string.
					},
				},
				{
					{
						"$lookup", bson.M{
							"from":         "namespaces",
							"foreignField": "owner",
							"localField":   "tmp_id",
							"as":           "tmp",
						},
					},
				},
				{
					{
						"$set", bson.M{"namespaces": bson.M{"$size": "$tmp"}},
					},
				},
				{
					{
						"$unset", bson.A{"tmp_id", "tmp"},
					},
				},
				{
					{"$merge", bson.M{"into": "users", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   49,
			"action":    "Down",
		}).Info("Applying migration down")

		_, err := db.Collection("users").Aggregate(context.Background(),
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$unset", "namespaces"},
				},
				{
					{"$merge", bson.M{"into": "users", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
}

var migration50 = migrate.Migration{
	Version:     50,
	Description: "set max number of namespaces per user",
	Up: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   50,
			"action":    "Up",
		}).Info("Applying migration up")

		var err error
		if envs.IsCloud() {
			_, err = db.Collection("users").Aggregate(context.Background(),
				mongo.Pipeline{
					{
						{"$match", bson.M{}},
					},
					{
						{"$set", bson.M{"tmp": bson.M{"$toString": "$_id"}}},
					},
					{
						{
							"$lookup", bson.M{
								"from": "namespaces",
								"let":  bson.M{"owner": "$tmp"},
								"pipeline": mongo.Pipeline{
									{
										{"$match", bson.M{
											"$expr": bson.M{
												"$and": bson.A{
													bson.M{"$eq": bson.A{"$owner", "$$owner"}},
													bson.M{"$eq": bson.A{"$billing.active", true}},
												},
											},
										}},
									},
								},
								"as": "list",
							},
						},
					},
					{
						{"$set", bson.M{"max_namespaces": bson.M{"$add": bson.A{bson.M{"$size": "$list"}, 1}}}},
					},
					{
						{"$unset", bson.A{"tmp", "list"}},
					},
					{
						{"$merge", bson.M{"into": "users", "whenMatched": "replace"}},
					},
				},
			)
		} else {
			_, err = db.Collection("users").Aggregate(context.Background(),
				mongo.Pipeline{
					{
						{"$match", bson.M{}},
					},
					{
						{"$set", bson.M{"max_namespaces": -1}},
					},
					{
						{"$merge", bson.M{"into": "users", "whenMatched": "replace"}},
					},
				},
			)
		}
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   50,
			"action":    "Down",
		}).Info("Applying migration down")

		_, err := db.Collection("users").Aggregate(context.Background(),
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$unset", "max_namespaces"},
				},
				{
					{"$merge", bson.M{"into": "users", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
}

var migration51 = migrate.Migration{
	Version:     51,
	Description: "create index for name on devices",
	Up: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   51,
			"action":    "Up",
		}).Info("Applying migration up")
		Name := "name"

		if _, err := db.Collection("devices").Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys: bson.M{
				Name: 1,
			},
			Options: &options.IndexOptions{ //nolint:exhaustruct
				Name: &Name,
			},
		}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   51,
			"action":    "Down",
		}).Info("Applying migration down")
		Name := "name"

		if _, err := db.Collection("devices").Indexes().DropOne(context.Background(), Name); err != nil {
			return err
		}

		return nil
	}),
}

var migration52 = migrate.Migration{
	Version:     52,
	Description: "add marketing field to users",
	Up: migrate.MigrationFunc(func(_ context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   52,
			"action":    "Up",
		}).Info("Applying migration")

		_, err := db.Collection("users").Aggregate(context.Background(),
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$set", bson.M{"email_marketing": true}},
				},
				{
					{"$merge", bson.M{"into": "users", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   52,
			"action":    "Down",
		}).Info("Applying migration")

		_, err := db.Collection("users").Aggregate(context.Background(),
			mongo.Pipeline{
				{
					{"$match", bson.M{}},
				},
				{
					{"$set", bson.M{"email_marketing": false}},
				},
				{
					{"$merge", bson.M{"into": "users", "whenMatched": "replace"}},
				},
			},
		)
		if err != nil {
			return err
		}

		return nil
	}),
}

var migration53 = migrate.Migration{
	Version:     53,
	Description: "create index to announcement ID",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   53,
			"action":    "Up",
		}).Info("Applying migration")
		field := "uuid"
		collection := "announcements"
		unique := true

		if _, err := db.Collection(collection).Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys: bson.M{
				field: 1,
			},
			Options: &options.IndexOptions{ //nolint:exhaustruct
				Name:   &field,
				Unique: &unique,
			},
		}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   53,
			"action":    "Down",
		}).Info("Applying migration")
		index := "uuid"
		collection := "announcements"

		if _, err := db.Collection(collection).Indexes().DropOne(context.Background(), index); err != nil {
			return err
		}

		return nil
	}),
}

var migration54 = migrate.Migration{
	Version:     54,
	Description: "create index to devices' tenant_id and status",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   54,
			"action":    "Up",
		}).Info("Applying migration")
		fieldTenantID := "tenant_id"
		fieldStatus := "status"
		name := "tenant_id_1_status_1"

		if _, err := db.Collection("devices").Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys: bson.D{
				bson.E{Key: fieldTenantID, Value: 1},
				bson.E{Key: fieldStatus, Value: 1},
			},
			Options: &options.IndexOptions{ //nolint:exhaustruct
				Name: &name,
			},
		}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   54,
			"action":    "Down",
		}).Info("Applying migration")
		name := "tenant_id_1_status_1"

		if _, err := db.Collection("devices").Indexes().DropOne(context.Background(), name); err != nil {
			return err
		}

		return nil
	}),
}

var migration55 = migrate.Migration{
	Version:     55,
	Description: "create indexes on removed_devices for tenant_id, tenant_id and uid and timestamp",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   55,
			"action":    "Up",
		}).Info("Applying migration")
		fieldTenantID := "tenant_id"
		fieldUID := "uid"
		fieldTimestamp := "timestamp"

		expire, err := time.ParseDuration("720h")
		if err != nil {
			return err
		}
		expireSeconds := int32(expire.Seconds())

		fieldNameTenantID := "tenant_id_1"
		fieldNameTenantIDUID := "tenant_id_1_uid_1"
		fieldNameTimestamp := "timestamp_1"
		if _, err := db.Collection("removed_devices").Indexes().CreateMany(context.Background(), []mongo.IndexModel{
			{
				Keys: bson.D{
					bson.E{Key: fieldTenantID, Value: 1},
				},
				Options: &options.IndexOptions{
					Name: &fieldNameTenantID,
				},
			},
			{
				Keys: bson.D{
					bson.E{Key: fieldTenantID, Value: 1},
					bson.E{Key: fieldUID, Value: 1},
				},
				Options: &options.IndexOptions{
					Name: &fieldNameTenantIDUID,
				},
			},
			{
				Keys: bson.D{
					bson.E{Key: fieldTimestamp, Value: 1},
				},
				Options: &options.IndexOptions{
					Name:               &fieldNameTimestamp,
					ExpireAfterSeconds: &expireSeconds,
				},
			},
		}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   55,
			"action":    "Down",
		}).Info("Applying migration")
		fieldNameTenantID := "tenant_id_1"
		fieldNameTenantIDUID := "tenant_id_1_uid_1"
		fieldNameTimestamp := "timestamp_1"

		if _, err := db.Collection("removed_devices").Indexes().DropOne(context.Background(), fieldNameTenantID); err != nil {
			return err
		}

		if _, err := db.Collection("removed_devices").Indexes().DropOne(context.Background(), fieldNameTenantIDUID); err != nil {
			return err
		}

		if _, err := db.Collection("removed_devices").Indexes().DropOne(context.Background(), fieldNameTimestamp); err != nil {
			return err
		}

		return nil
	}),
}

var migration56 = migrate.Migration{
	Version:     56,
	Description: "create index for public url address on devices",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   56,
			"action":    "Up",
		}).Info("Applying migration up")
		field := "public_url_address"
		unique := true
		sparse := true

		if _, err := db.Collection("devices").Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys: bson.M{
				field: 1,
			},
			Options: &options.IndexOptions{ //nolint:exhaustruct
				Unique: &unique,
				Sparse: &sparse,
				Name:   &field,
			},
		}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   56,
			"action":    "Down",
		}).Info("Applying migration down")
		field := "public_url_address"

		if _, err := db.Collection("devices").Indexes().DropOne(context.Background(), field); err != nil {
			return err
		}

		return nil
	}),
}

var migration57 = migrate.Migration{
	Version:     57,
	Description: "update billing state to status and its values",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   57,
			"action":    "Up",
		}).Info("Applying migration up")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"billing": bson.M{
						"$ne": nil,
					},
				},
			},
			{
				"$addFields": bson.M{
					"billing.status": bson.M{
						"$switch": bson.M{
							"branches": []bson.M{
								{
									"case": bson.M{
										"$eq": []string{"$billing.state", "processed"},
									},
									"then": "active",
								},
								{
									"case": bson.M{
										"$eq": []string{"$billing.state", "past_due"},
									},
									"then": "past_due",
								},
								{
									"case": bson.M{
										"$eq": []string{"$billing.state", "pending"},
									},
									"then": "canceled",
								},
								{
									"case": bson.M{
										"$eq": []string{"$billing.state", "inactive"},
									},
									"then": "inactive",
								},
								{
									"case": bson.M{
										"$eq": []string{"$billing.state", "canceled"},
									},
									"then": "canceled",
								},
							},
							"default": "canceled",
						},
					},
				},
			},
			{
				"$unset": "billing.state",
			},
			{
				"$merge": bson.M{
					"into":        "namespaces",
					"whenMatched": "replace",
				},
			},
		}

		_, err := db.Collection("namespaces").Aggregate(context.Background(), pipeline)
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   57,
			"action":    "Down",
		}).Info("Applying migration down")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"billing": bson.M{
						"$ne": nil,
					},
				},
			},
			{
				"$addFields": bson.M{
					"billing.state": bson.M{
						"$switch": bson.M{
							"branches": []bson.M{
								{
									"case": bson.M{
										"$eq": []string{"$billing.status", "active"},
									},
									"then": "processed",
								},
								{
									"case": bson.M{
										"$eq": []string{"$billing.status", "past_due"},
									},
									"then": "past_due",
								},
								{
									"case": bson.M{
										"$eq": []string{"$billing.status", "inactive"},
									},
									"then": "inactive",
								},
								{
									"case": bson.M{
										"$eq": []string{"$billing.status", "canceled"},
									},
									"then": "canceled",
								},
							},
							"default": "canceled",
						},
					},
				},
			},
			{
				"$unset": "billing.status",
			},
			{
				"$merge": bson.M{
					"into":        "namespaces",
					"whenMatched": "replace",
				},
			},
		}

		_, err := db.Collection("namespaces").Aggregate(context.Background(), pipeline)
		if err != nil {
			return err
		}

		return nil
	}),
}

var migration58 = migrate.Migration{
	Version:     58,
	Description: "",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   58,
			"action":    "Up",
		}).Info("Applying migration up")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"billing": bson.M{
						"$ne": nil,
					},
				},
			},
			{
				"$set": bson.M{
					"billing.current_period_end": bson.M{
						"$convert": bson.M{
							"input": "$billing.current_period_end",
							"to":    "long",
						},
					},
				},
			},
			{
				"$merge": bson.M{
					"into":        "namespaces",
					"whenMatched": "replace",
				},
			},
		}

		_, err := db.Collection("namespaces").Aggregate(context.Background(), pipeline)
		if err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   58,
			"action":    "Down",
		}).Info("Applying migration down")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"billing": bson.M{
						"$ne": nil,
					},
				},
			},
			{
				"$set": bson.M{
					"billing.current_period_end": bson.M{
						"$convert": bson.M{
							"input": "$billing.current_period_end",
							"to":    "date",
						},
					},
				},
			},
			{
				"$merge": bson.M{
					"into":        "namespaces",
					"whenMatched": "replace",
				},
			},
		}

		_, err := db.Collection("namespaces").Aggregate(context.Background(), pipeline)
		if err != nil {
			return err
		}

		return nil
	}),
}

var migration59 = migrate.Migration{
	Version:     59,
	Description: "Converts all 'name' field values in the 'users' collection to lowercase.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   59,
			"action":    "Up",
		}).Info("Starting migration Up action.")

		_, err := db.Collection("users").UpdateMany(
			ctx,
			bson.M{},
			[]bson.M{
				{
					"$set": bson.M{
						"username": bson.M{
							"$toLower": "$username",
						},
						"email": bson.M{
							"$toLower": "$email",
						},
					},
				},
			},
		)
		if err != nil {
			log.WithFields(log.Fields{
				"component": "migration",
				"version":   59,
				"action":    "Up",
				"error":     err.Error(),
			}).Error("Failed to execute Up migration.")

			return err
		}

		log.WithFields(log.Fields{
			"component": "migration",
			"version":   59,
			"action":    "Up",
		}).Info("Completed migration Up action successfully.")

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   59,
			"action":    "Down",
		}).Info("Starting migration Down action.")

		log.WithFields(log.Fields{
			"component": "migration",
			"version":   59,
			"action":    "Down",
		}).Info("Completed migration Down action successfully.")

		return nil
	}),
}

var migration60 = migrate.Migration{
	Version:     60,
	Description: "create index for tenant_id on active_sessions",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   60,
			"action":    "Up",
		}).Info("Applying migration up")
		indexName := "tenant_id"
		if _, err := db.Collection("active_sessions").Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys: bson.M{
				"tenant_id": 1,
			},
			Options: &options.IndexOptions{ //nolint:exhaustruct
				Name: &indexName,
			},
		}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   60,
			"action":    "Down",
		}).Info("Applying migration down")
		if _, err := db.Collection("active_sessions").Indexes().DropOne(context.Background(), "tenant_id"); err != nil {
			return err
		}

		return nil
	}),
}

var migration61 = migrate.Migration{
	Version:     61,
	Description: "delete devices with empty name",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   61,
			"action":    "Up",
		}).Info("Applying migration up")
		if _, err := db.Collection("devices").DeleteMany(context.Background(), bson.M{"$or": bson.A{
			bson.M{"name": ""},
			bson.M{"name": bson.M{"$exists": false}},
		}}); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		// This migration is not reversible.
		return nil
	}),
}

var migration62 = migrate.Migration{
	Version:     62,
	Description: "create index for tenant_id on recorded_sessions",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   62,
			"action":    "Up",
		}).Info("Applying migration up")

		indexName := "tenant_id"
		_, err := db.Collection("recorded_sessions").Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys: bson.M{
				"tenant_id": 1,
			},
			Options: &options.IndexOptions{ //nolint:exhaustruct
				Name: &indexName,
			},
		})
		if err != nil {
			log.WithFields(log.Fields{
				"component": "migration",
				"version":   62,
				"action":    "Up",
			}).WithError(err).Info("Error while trying to apply migration 62")

			return err
		}

		log.WithFields(log.Fields{
			"component": "migration",
			"version":   62,
			"action":    "Up",
		}).Info("Succeeds to to apply migration 62")

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   62,
			"action":    "Down",
		}).Info("Applying migration down")
		if _, err := db.Collection("recorded_sessions").Indexes().DropOne(context.Background(), "tenant_id"); err != nil {
			return err
		}

		return nil
	}),
}

var migration63 = migrate.Migration{
	Version:     63,
	Description: "add MFA fields to collection users",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   63,
			"action":    "Up",
		}).Info("Applying migration")

		update := bson.M{
			"$set": bson.M{
				"status_mfa": false,
				"secret":     "",
				"codes":      []string{},
			},
		}

		if _, err := db.Collection("users").UpdateMany(ctx, bson.M{}, update); err != nil {
			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   63,
			"action":    "Down",
		}).Info("Reverting migration")

		update := bson.M{
			"$unset": bson.M{
				"status_mfa": "",
				"secret":     "",
				"codes":      "",
			},
		}

		if _, err := db.Collection("users").UpdateMany(ctx, bson.M{}, update); err != nil {
			return err
		}

		return nil
	}),
}

var migration64 = migrate.Migration{
	Version:     64,
	Description: "Adding the 'settings.connection_announcement' attribute to the namespace if it does not already exist.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   64,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"settings.connection_announcement": bson.M{"$in": []interface{}{nil, ""}},
		}

		update := bson.M{
			"$set": bson.M{
				"settings.connection_announcement": "",
			},
		}

		_, err := db.
			Collection("namespaces").
			UpdateMany(ctx, filter, update)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   64,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"settings.connection_announcement": bson.M{"$in": []interface{}{nil, ""}},
		}

		update := bson.M{
			"$unset": bson.M{
				"settings.connection_announcement": "",
			},
		}

		_, err := db.
			Collection("namespaces").
			UpdateMany(ctx, filter, update)

		return err
	}),
}

var migration65 = migrate.Migration{
	Version:     65,
	Description: "Adding the 'recovery_email' attribute to the user if it does not already exist.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   65,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"recovery_email": bson.M{"$in": []interface{}{nil, ""}},
		}

		update := bson.M{
			"$set": bson.M{
				"recovery_email": "",
			},
		}

		_, err := db.
			Collection("users").
			UpdateMany(ctx, filter, update)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   65,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"_id": bson.M{"$ne": nil},
		}

		update := bson.M{
			"$unset": bson.M{
				"recovery_email": "",
			},
		}

		_, err := db.
			Collection("users").
			UpdateMany(ctx, filter, update)

		return err
	}),
}

var migration66 = migrate.Migration{
	Version:     66,
	Description: "Replace the user's MFA attributes.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.
			WithFields(log.Fields{
				"component": "migration",
				"version":   66,
				"action":    "Up",
			}).
			Info("Applying migration")

		filter := bson.M{
			"_id": bson.M{
				"$ne": nil,
			},
		}

		rename := bson.M{
			"$rename": bson.M{
				"status_mfa": "mfa.enabled",
				"secret":     "mfa.secret",
				"codes":      "mfa.recovery_codes",
			},
		}

		if _, err := db.Collection("users").UpdateMany(ctx, filter, rename); err != nil {
			return err
		}

		unset := bson.M{
			"$unset": bson.M{
				"status_mfa": "",
				"secret":     "",
				"codes":      "",
			},
		}

		_, err := db.Collection("users").UpdateMany(ctx, filter, unset)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.
			WithFields(log.Fields{
				"component": "migration",
				"version":   66,
				"action":    "Up",
			}).
			Info("Applying migration")

		log.Info("Unable to undo the MFA object")

		return nil
	}),
}

var migration67 = migrate.Migration{
	Version:     67,
	Description: "Hash the user's MFA recovery code before storing it as a plain string.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.
			WithFields(log.Fields{
				"component": "migration",
				"version":   67,
				"action":    "Up",
			}).
			Info("Applying migration")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"mfa.enabled": true,
					"mfa.recovery_codes.0": bson.M{
						"$not": bson.M{
							"$regex": "^\\$",
						},
					},
				},
			},
		}

		cursor, err := db.Collection("users").Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		updateModels := make([]mongo.WriteModel, 0)

		for cursor.Next(ctx) {
			user := new(models.User)
			if err := cursor.Decode(user); err != nil {
				return err
			}

			recoveryCodes := make([]string, 0)
			for _, c := range user.MFA.RecoveryCodes {
				hash, err := hash.Do(c)
				if err != nil {
					return err
				}

				recoveryCodes = append(recoveryCodes, hash)

			}

			filter := bson.M{"username": user.Username}
			update := bson.M{"$set": bson.M{"mfa.recovery_codes": recoveryCodes}}

			updateModels = append(updateModels, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(false))
		}

		if len(updateModels) > 0 {
			if _, err := db.Collection("users").BulkWrite(ctx, updateModels); err != nil {
				return err
			}
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.
			WithFields(log.Fields{
				"component": "migration",
				"version":   67,
				"action":    "Down",
			}).
			Info("Applying migration")

		log.Info("Unable to undo the recovery code hash")

		return nil
	}),
}

var migration68 = migrate.Migration{
	Version:     68,
	Description: "Rename `api_keys.user_id` to `api_keys.created_by`.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.
			WithFields(log.Fields{
				"component": "migration",
				"version":   68,
				"action":    "Up",
			}).
			Info("Applying migration")

		filter := bson.M{
			"user_id": bson.M{"$nin": []interface{}{nil, ""}},
		}

		rename := bson.M{
			"$rename": bson.M{
				"user_id": "created_by",
			},
		}

		if _, err := db.Collection("api_keys").UpdateMany(ctx, filter, rename); err != nil {
			return err
		}

		unset := bson.M{
			"$unset": bson.M{
				"user_id": "",
			},
		}

		_, err := db.Collection("api_keys").UpdateMany(ctx, filter, unset)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.
			WithFields(log.Fields{
				"component": "migration",
				"version":   68,
				"action":    "Down",
			}).
			Info("Applying migration")

		filter := bson.M{
			"created_by": bson.M{"$nin": []interface{}{nil, ""}},
		}

		rename := bson.M{
			"$rename": bson.M{
				"created_by": "user_id",
			},
		}

		if _, err := db.Collection("api_keys").UpdateMany(ctx, filter, rename); err != nil {
			return err
		}

		unset := bson.M{
			"$unset": bson.M{
				"created_by": "",
			},
		}

		_, err := db.Collection("api_keys").UpdateMany(ctx, filter, unset)

		return err
	}),
}

var migration69 = migrate.Migration{
	Version:     69,
	Description: "Hash API key ID. It will delete the old document and create a new one with the hashed ID.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.
			WithFields(log.Fields{
				"component": "migration",
				"version":   69,
				"action":    "Up",
			}).
			Info("Applying migration")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"_id": bson.M{
						"$regex": "^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$",
					},
				},
			},
		}

		cursor, err := db.Collection("api_keys").Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		updateModels := make([]mongo.WriteModel, 0)
		deleteModels := make([]mongo.WriteModel, 0)

		for cursor.Next(ctx) {
			apiKey := new(models.APIKey)
			if err := cursor.Decode(apiKey); err != nil {
				return err
			}

			idSum := sha256.Sum256([]byte(apiKey.ID))
			hashedID := hex.EncodeToString(idSum[:])

			doc := &models.APIKey{
				ID:        hashedID,
				Name:      apiKey.Name,
				CreatedBy: apiKey.CreatedBy,
				TenantID:  apiKey.TenantID,
				Role:      apiKey.Role,
				CreatedAt: apiKey.CreatedAt,
				UpdatedAt: apiKey.UpdatedAt,
				ExpiresIn: apiKey.ExpiresIn,
			}

			deleteModels = append(deleteModels, mongo.NewDeleteOneModel().SetFilter(bson.M{"_id": apiKey.ID}))
			updateModels = append(updateModels, mongo.NewInsertOneModel().SetDocument(doc))
		}

		if len(updateModels) > 0 || len(deleteModels) > 0 {
			mongoSession, err := db.Client().StartSession()
			if err != nil {
				return err
			}
			defer mongoSession.EndSession(ctx)

			_, err = mongoSession.WithTransaction(ctx, func(mongoctx mongo.SessionContext) (interface{}, error) {
				if _, err := db.Collection("api_keys").BulkWrite(ctx, updateModels); err != nil {
					return nil, err
				}

				_, err := db.Collection("api_keys").BulkWrite(ctx, deleteModels)

				return nil, err
			})
			if err != nil {
				return err
			}
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.
			WithFields(log.Fields{
				"component": "migration",
				"version":   69,
				"action":    "Down",
			}).
			Info("Applying migration")

		log.Info("Unable to undo the api key hash")

		return nil
	}),
}

var migration70 = migrate.Migration{
	Version:     70,
	Description: "Adding the 'preferences' attribute to the user if it does not already exist.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   70,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"preferences": bson.M{"$exists": false},
		}

		update := bson.M{
			"$set": bson.M{
				"preferences": bson.M{},
			},
		}

		_, err := db.
			Collection("users").
			UpdateMany(ctx, filter, update)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   70,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"preferences": bson.M{"$exists": true},
		}

		update := bson.M{
			"$unset": bson.M{
				"preferences": "",
			},
		}

		_, err := db.
			Collection("users").
			UpdateMany(ctx, filter, update)

		return err
	}),
}

var migration71 = migrate.Migration{
	Version:     71,
	Description: "Adding the 'preferences.preferred_namespace' attribute to the user if it does not already exist.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   71,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"preferences":                     bson.M{"$exists": true},
			"preferences.preferred_namespace": bson.M{"$exists": false},
		}

		update := bson.M{
			"$set": bson.M{
				"preferences.preferred_namespace": "",
			},
		}

		_, err := db.
			Collection("users").
			UpdateMany(ctx, filter, update)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   71,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"preferences":                     bson.M{"$exists": true},
			"preferences.preferred_namespace": bson.M{"$exists": true},
		}

		update := bson.M{
			"$unset": bson.M{
				"preferences.preferred_namespace": "",
			},
		}

		_, err := db.
			Collection("users").
			UpdateMany(ctx, filter, update)

		return err
	}),
}

var migration72 = migrate.Migration{
	Version:     72,
	Description: "Adding the 'members.$.status' attribute to the namespace if it does not already exist.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   72,
			"action":    "Up",
		}).Info("Applying migration")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"tenant_id": bson.M{
						"$exists": true,
					},
				},
			},
		}

		cursor, err := db.Collection("namespaces").Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		updateModels := make([]mongo.WriteModel, 0)

		for cursor.Next(ctx) {
			namespace := new(models.Namespace)
			if err := cursor.Decode(namespace); err != nil {
				return err
			}

			for _, m := range namespace.Members {
				if m.Status == "" {
					updateModel := mongo.
						NewUpdateOneModel().
						SetFilter(bson.M{"tenant_id": namespace.TenantID, "members": bson.M{"$elemMatch": bson.M{"id": m.ID}}}).
						SetUpdate(bson.M{"$set": bson.M{"members.$.status": models.DeviceStatusAccepted}})

					updateModels = append(updateModels, updateModel)
				}
			}
		}

		if len(updateModels) == 0 {
			return nil
		}

		_, err = db.Collection("namespaces").BulkWrite(ctx, updateModels)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   72,
			"action":    "Down",
		}).Info("Reverting migration")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"tenant_id": bson.M{
						"$exists": true,
					},
				},
			},
		}

		cursor, err := db.Collection("namespaces").Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		updateModels := make([]mongo.WriteModel, 0)

		for cursor.Next(ctx) {
			namespace := new(models.Namespace)
			if err := cursor.Decode(namespace); err != nil {
				return err
			}

			for _, m := range namespace.Members {
				if m.Status != "" {
					updateModel := mongo.
						NewUpdateOneModel().
						SetFilter(bson.M{"tenant_id": namespace.TenantID, "members": bson.M{"$elemMatch": bson.M{"id": m.ID}}}).
						SetUpdate(bson.M{"$unset": bson.M{"members.$.status": ""}})

					updateModels = append(updateModels, updateModel)
				}
			}
		}

		if len(updateModels) == 0 {
			return nil
		}

		_, err = db.Collection("namespaces").BulkWrite(ctx, updateModels)

		return err
	}),
}

var migration73 = migrate.Migration{
	Version:     73,
	Description: "Adding the 'members.$.added_at' attribute to the namespace if it does not already exist. The value is the Go time.Time zeroer",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   73,
			"action":    "Up",
		}).Info("Applying migration")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"tenant_id": bson.M{
						"$exists": true,
					},
				},
			},
		}

		cursor, err := db.Collection("namespaces").Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		updateModels := make([]mongo.WriteModel, 0)

		for cursor.Next(ctx) {
			namespace := new(models.Namespace)
			if err := cursor.Decode(namespace); err != nil {
				return err
			}

			for _, m := range namespace.Members {
				if m.AddedAt == (time.Time{}) {
					updateModel := mongo.
						NewUpdateOneModel().
						SetFilter(bson.M{"tenant_id": namespace.TenantID, "members": bson.M{"$elemMatch": bson.M{"id": m.ID}}}).
						// We update the added_at field to the same value as in the if statement
						// because when the attribute is null in MongoDB, it will be converted
						// to the zero value of time.Time.
						SetUpdate(bson.M{"$set": bson.M{"members.$.added_at": time.Time{}}})

					updateModels = append(updateModels, updateModel)
				}
			}
		}

		if len(updateModels) == 0 {
			return nil
		}

		_, err = db.Collection("namespaces").BulkWrite(ctx, updateModels)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   73,
			"action":    "Down",
		}).Info("Reverting migration")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"tenant_id": bson.M{
						"$exists": true,
					},
				},
			},
		}

		cursor, err := db.Collection("namespaces").Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		updateModels := make([]mongo.WriteModel, 0)

		for cursor.Next(ctx) {
			namespace := new(models.Namespace)
			if err := cursor.Decode(namespace); err != nil {
				return err
			}

			for _, m := range namespace.Members {
				updateModel := mongo.
					NewUpdateOneModel().
					SetFilter(bson.M{"tenant_id": namespace.TenantID, "members": bson.M{"$elemMatch": bson.M{"id": m.ID}}}).
					SetUpdate(bson.M{"$unset": bson.M{"members.$.added_at": ""}})

				updateModels = append(updateModels, updateModel)
			}
		}

		if len(updateModels) == 0 {
			return nil
		}

		_, err = db.Collection("namespaces").BulkWrite(ctx, updateModels)

		return err
	}),
}

var migration74 = migrate.Migration{
	Version:     74,
	Description: "Adding default message on announcement if is not set.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   74,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"settings.connection_announcement": "",
		}

		annoucementMsg := ""
		if envs.IsCommunity() {
			annoucementMsg = models.DefaultAnnouncementMessage
		}

		update := bson.M{
			"$set": bson.M{
				"settings.connection_announcement": annoucementMsg,
			},
		}

		_, err := db.
			Collection("namespaces").
			UpdateMany(ctx, filter, update)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   74,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"settings.connection_announcement": models.DefaultAnnouncementMessage,
		}

		update := bson.M{
			"$set": bson.M{
				"settings.connection_announcement": "",
			},
		}

		_, err := db.
			Collection("namespaces").
			UpdateMany(ctx, filter, update)

		return err
	}),
}

var migration75 = migrate.Migration{
	Version:     75,
	Description: "Convert user.confirmed to user.status",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   75,
			"action":    "Up",
		}).Info("Applying migration")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"confirmed": bson.M{
						"$exists": true,
					},
					"status": bson.M{
						"$exists": false,
					},
				},
			},
		}

		cursor, err := db.Collection("users").Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		updateModels := make([]mongo.WriteModel, 0)

		for cursor.Next(ctx) {
			user := make(map[string]interface{})
			if err := cursor.Decode(&user); err != nil {
				return err
			}

			updateModel := mongo.
				NewUpdateOneModel().
				SetFilter(bson.M{"_id": user["_id"]})

			if confirmed := user["confirmed"]; confirmed == true {
				updateModel.SetUpdate(bson.M{"$set": bson.M{"status": models.UserStatusConfirmed.String()}, "$unset": bson.M{"confirmed": ""}})
			} else {
				updateModel.SetUpdate(bson.M{"$set": bson.M{"status": models.UserStatusNotConfirmed.String()}, "$unset": bson.M{"confirmed": ""}})
			}

			updateModels = append(updateModels, updateModel)
		}

		if len(updateModels) == 0 {
			return nil
		}

		_, err = db.Collection("users").BulkWrite(ctx, updateModels)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   75,
			"action":    "Down",
		}).Info("Reverting migration")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"confirmed": bson.M{
						"$exists": false,
					},
					"status": bson.M{
						"$exists": true,
					},
				},
			},
		}

		cursor, err := db.Collection("users").Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		updateModels := make([]mongo.WriteModel, 0)

		for cursor.Next(ctx) {
			user := make(map[string]interface{})
			if err := cursor.Decode(&user); err != nil {
				return err
			}

			updateModel := mongo.
				NewUpdateOneModel().
				SetFilter(bson.M{"_id": user["_id"]})

			if status := user["status"].(string); status == models.UserStatusConfirmed.String() {
				updateModel.SetUpdate(bson.M{"$set": bson.M{"confirmed": true}, "$unset": bson.M{"status": ""}})
			} else {
				updateModel.SetUpdate(bson.M{"$set": bson.M{"confirmed": false}, "$unset": bson.M{"status": ""}})
			}

			updateModels = append(updateModels, updateModel)
		}

		if len(updateModels) == 0 {
			return nil
		}

		_, err = db.Collection("users").BulkWrite(ctx, updateModels)

		return err
	}),
}

var migration76 = migrate.Migration{
	Version:     76,
	Description: "Remove user.namespace from users collection.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 76).
			WithField("action", " Up").
			Info("Applying migration")

		filter := bson.M{"namespaces": bson.M{"$exists": true}}
		update := bson.M{"$unset": bson.M{"namespaces": ""}}

		_, err := db.Collection("users").UpdateMany(ctx, filter, update)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 76).
			WithField("action", "Down").
			Info("Applying migration")

		filter := []bson.M{
			{
				"$match": bson.M{
					"namespaces": bson.M{
						"$exists": false,
					},
				},
			},
		}

		cursor, err := db.Collection("users").Aggregate(ctx, filter)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		updateModels := make([]mongo.WriteModel, 0)
		for cursor.Next(ctx) {
			user := new(models.User)
			if err := cursor.Decode(user); err != nil {
				return err
			}

			cursor, err := db.Collection("namespaces").Find(ctx, bson.M{"members": bson.M{"$elemMatch": bson.M{"id": user.ID, "role": "owner"}}})
			if err != nil {
				return err
			}
			defer cursor.Close(ctx)

			namespaces := make([]models.Namespace, 0)
			if err := cursor.All(ctx, &namespaces); err != nil {
				continue
			}

			userID, _ := primitive.ObjectIDFromHex(user.ID)
			updateModel := mongo.
				NewUpdateOneModel().
				SetFilter(bson.M{"_id": userID}).
				SetUpdate(bson.M{"$set": bson.M{"namespaces": len(namespaces)}})

			updateModels = append(updateModels, updateModel)
		}

		if len(updateModels) == 0 {
			return nil
		}

		_, err = db.Collection("users").BulkWrite(ctx, updateModels)

		return err
	}),
}

var migration77 = migrate.Migration{
	Version:     77,
	Description: "Recreate the unique index on the 'username' field in the 'users' collection with a partial filter for documents where the 'username' field is a string.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 77).
			WithField("action", " Up").
			Info("Applying migration")

		_, _ = db.Collection("users").Indexes().DropOne(ctx, "username")

		indexModel := mongo.IndexModel{
			Keys:    bson.M{"username": 1},
			Options: options.Index().SetName("username").SetUnique(true).SetPartialFilterExpression(bson.M{"username": bson.M{"$type": "string"}}),
		}

		_, err := db.Collection("users").Indexes().CreateOne(ctx, indexModel)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 77).
			WithField("action", "Down").
			Info("Reverting migration")

		_, err := db.Collection("users").Indexes().DropOne(ctx, "username")

		return err
	}),
}

var migration78 = migrate.Migration{
	Version:     78,
	Description: "Recreate the unique index on the 'email' field in the 'users' collection with a partial filter for documents where the 'email' field is a string.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 78).
			WithField("action", " Up").
			Info("Applying migration")

		_, _ = db.Collection("users").Indexes().DropOne(ctx, "email")

		indexModel := mongo.IndexModel{
			Keys:    bson.M{"email": 1},
			Options: options.Index().SetName("email").SetUnique(true).SetPartialFilterExpression(bson.M{"email": bson.M{"$type": "string"}}),
		}

		_, err := db.Collection("users").Indexes().CreateOne(ctx, indexModel)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 78).
			WithField("action", "Down").
			Info("Reverting migration")

		_, err := db.Collection("users").Indexes().DropOne(ctx, "email")

		return err
	}),
}

var migration79 = migrate.Migration{
	Version:     79,
	Description: "create and populate the system collection",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 79).
			WithField("action", " Up").
			Info("Applying migration")

		if err := db.CreateCollection(ctx, "system"); err != nil {
			return err
		}

		if envs.IsCommunity() {
			users, err := db.Collection("users").CountDocuments(ctx, bson.M{})
			if err != nil {
				return err
			}

			if _, err := db.Collection("system").InsertOne(ctx, bson.M{
				"setup": users > 0,
			}); err != nil {
				return err
			}
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 79).
			WithField("action", "Down").
			Info("Reverting migration")

		return db.Collection("system").Drop(ctx)
	}),
}

var migration80 = migrate.Migration{
	Version:     80,
	Description: "Remove the 'message' index from the 'recorded_sessions' collection.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 80).
			WithField("action", " Up").
			Info("Applying migration")

		_, err := db.Collection("recorded_sessions").Indexes().DropOne(ctx, "message")

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 80).
			WithField("action", "Down").
			Info("Applying migration")

		index := mongo.IndexModel{
			Keys:    bson.D{{Key: "message", Value: 1}},
			Options: options.Index().SetName("message").SetUnique(false),
		}

		_, err := db.Collection("recorded_sessions").Indexes().CreateOne(ctx, index)

		return err
	}),
}

var migration81 = migrate.Migration{
	Version:     81,
	Description: "Create a 'time' index in the 'recorded_sessions' collection.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 81).
			WithField("action", " Up").
			Info("Applying migration")

		index := mongo.IndexModel{
			Keys:    bson.D{{Key: "time", Value: 1}},
			Options: options.Index().SetName("time").SetUnique(false),
		}

		_, err := db.Collection("recorded_sessions").Indexes().CreateOne(ctx, index)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithField("component", "migration").
			WithField("version", 81).
			WithField("action", "Down").
			Info("Applying migration")

		_, err := db.Collection("recorded_sessions").Indexes().DropOne(ctx, "time")

		return err
	}),
}

var migration82 = migrate.Migration{
	Version:     82,
	Description: "Adding the 'namespaces.type' attribute to the namespaces if it does not already exist.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   82,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"type": bson.M{"$in": []interface{}{nil, ""}},
		}

		update := bson.M{
			"$set": bson.M{
				"type": models.TypeTeam,
			},
		}

		_, err := db.
			Collection("namespaces",
				options.Collection().SetWriteConcern(writeconcern.Majority()),
			).
			UpdateMany(ctx, filter, update)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   82,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"type": bson.M{"$in": []interface{}{nil, ""}},
		}

		update := bson.M{
			"$unset": bson.M{
				"type": models.TypeTeam,
			},
		}

		_, err := db.
			Collection("namespaces",
				options.Collection().SetWriteConcern(writeconcern.Majority()),
			).
			UpdateMany(ctx, filter, update)

		return err
	}),
}
