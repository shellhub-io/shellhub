package mongo

import (
	"context"

	"github.com/shellhub-io/shellhub/api/store/mongo/migrations"
	"github.com/sirupsen/logrus"
	"github.com/square/mongo-lock"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

func ApplyMigrations(db *mongo.Database) error {
	logrus.Info("Creating lock for the resource migrations")

	lockClient := lock.NewClient(db.Collection("migrations", options.Collection().SetWriteConcern(writeconcern.New(writeconcern.WMajority()))))
	if err := lockClient.CreateIndexes(context.TODO()); err != nil {
		logrus.WithError(err).Fatal("Failed to create a lock for the database")
	}

	logrus.Info("Locking the resource migrations")

	lockId := "0"

	if err := lockClient.XLock(context.TODO(), "migrations", lockId, lock.LockDetails{}); err != nil {
		logrus.WithError(err).Fatal("Failed to lock the migrations")
	}

	err := migrate.NewMigrate(db, migrations.GenerateMigrations()...).Up(migrate.AllAvailable)

	logrus.Info("Unlocking the resource migrations")

	if _, err := lockClient.Unlock(context.TODO(), lockId); err != nil {
		logrus.WithError(err).Fatal("Failed to unlock the migrations")
	}

	return err
}
