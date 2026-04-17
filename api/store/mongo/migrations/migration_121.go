package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration121 = migrate.Migration{
	Version:     121,
	Description: "Add namespace and device SSH settings",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 121, "action": "Up"}).Info("Applying migration")

		if _, err := db.Collection("namespaces").UpdateMany(ctx, bson.M{}, bson.M{
			"$set": bson.M{
				"settings.allow_password":         true,
				"settings.allow_public_key":       true,
				"settings.allow_root":             true,
				"settings.allow_empty_passwords":  true,
				"settings.allow_tty":              true,
				"settings.allow_tcp_forwarding":   true,
				"settings.allow_web_endpoints":    true,
				"settings.allow_sftp":             true,
				"settings.allow_agent_forwarding": true,
			},
		}); err != nil {
			log.WithError(err).Error("Failed to add allow_* settings to namespace settings")

			return err
		}

		if _, err := db.Collection("devices").UpdateMany(ctx, bson.M{}, bson.M{
			"$set": bson.M{
				"ssh": bson.M{
					"allow_password":         true,
					"allow_public_key":       true,
					"allow_root":             true,
					"allow_empty_passwords":  true,
					"allow_tty":              true,
					"allow_tcp_forwarding":   true,
					"allow_web_endpoints":    true,
					"allow_sftp":             true,
					"allow_agent_forwarding": true,
				},
			},
		}); err != nil {
			log.WithError(err).Error("Failed to add ssh settings to devices")

			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 121, "action": "Down"}).Info("Reverting migration")

		if _, err := db.Collection("namespaces").UpdateMany(ctx, bson.M{}, bson.M{
			"$unset": bson.M{
				"settings.allow_password":         "",
				"settings.allow_public_key":       "",
				"settings.allow_root":             "",
				"settings.allow_empty_passwords":  "",
				"settings.allow_tty":              "",
				"settings.allow_tcp_forwarding":   "",
				"settings.allow_web_endpoints":    "",
				"settings.allow_sftp":             "",
				"settings.allow_agent_forwarding": "",
			},
		}); err != nil {
			log.WithError(err).Error("Failed to remove allow_* settings from namespace settings")

			return err
		}

		if _, err := db.Collection("devices").UpdateMany(ctx, bson.M{}, bson.M{
			"$unset": bson.M{
				"ssh": "",
			},
		}); err != nil {
			log.WithError(err).Error("Failed to remove ssh settings from devices")

			return err
		}

		return nil
	}),
}
