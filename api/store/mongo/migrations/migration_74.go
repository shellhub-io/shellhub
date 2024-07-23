package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const defaultAnnouncementMessage = `
******************************************************************
*                                                                *
*             Welcome to ShellHub Community Edition!             *
*                                                                *
* ShellHub is a next-generation SSH server, providing a          *
* seamless, secure, and user-friendly solution for remote        *
* access management. With ShellHub, you can manage all your      *
* devices effortlessly from a single platform, ensuring optimal  *
* security and productivity.                                     *
*                                                                *
* Want to learn more about ShellHub and explore other editions?  *
* Visit: https://shellhub.io                                     *
*                                                                *
* Join our community and contribute to our open-source project:  *
* https://github.com/shellhub-io/shellhub                        *
*                                                                *
* For assistance, please contact the system administrator.       *
*                                                                *
******************************************************************
`

var migration74 = migrate.Migration{
	Version:     74,
	Description: "Adding default message on announcement if is not set.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   74,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"settings.connection_announcement": "",
		}

		update := bson.M{
			"$set": bson.M{
				"settings.connection_announcement": defaultAnnouncementMessage,
			},
		}

		_, err := db.
			Collection("namespaces").
			UpdateMany(ctx, filter, update)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   74,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{
			"settings.connection_announcement": defaultAnnouncementMessage,
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
