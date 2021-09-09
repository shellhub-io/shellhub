package mongo

import (
	"context"

	"github.com/pkg/errors"
	"github.com/shellhub-io/shellhub/api/store/mongo/migrations"
	"github.com/sirupsen/logrus"
	lock "github.com/square/mongo-lock"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

func ApplyMigrations(db *mongo.Database) error {
	logrus.Info("Creating lock for the resource migrations")

	lockClient := lock.NewClient(db.Collection("locks", options.Collection().SetWriteConcern(writeconcern.New(writeconcern.WMajority()))))
	if err := lockClient.CreateIndexes(context.TODO()); err != nil {
		logrus.WithError(err).Fatal("Failed to create a lock for the database")
	}

	logrus.Info("Locking the resource migrations")

	lockID := "0"

	if err := lockClient.XLock(context.TODO(), "migrations", lockID, lock.LockDetails{}); err != nil {
		logrus.WithError(err).Fatal("Failed to lock the migrations")
	}

	defer func() {
		logrus.Info("Unlocking the resource migrations")

		if _, err := lockClient.Unlock(context.TODO(), lockID); err != nil {
			logrus.WithError(err).Fatal("Failed to unlock the migrations")
		}
	}()

	if err := fixMigrations072(db); err != nil {
		logrus.WithError(err).Fatal("Failed to fix the migrations lock bug")
	}

	list := migrations.GenerateMigrations()
	migration := migrate.NewMigrate(db, list...)

	current, _, err := migration.Version()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get current migration version")
	}

	latest := list[len(list)-1] //nolint:ifshort

	if current == latest.Version {
		logrus.Info("No migrations to apply")

		return nil
	}

	logrus.WithFields(logrus.Fields{
		"from": current,
		"to":   latest.Version,
	}).Info("Migrating database")

	return migration.Up(migrate.AllAvailable)
}

// This function is necessary due the lock bug on v0.7.2.
func fixMigrations072(db *mongo.Database) error {
	// Search for lock in migrations collection.
	if _, err := db.Collection("migrations").Find(context.TODO(),
		bson.M{"resource": "migrations"},
	); err != nil && err == mongo.ErrNoDocuments {
		// No documents found, nothing to do.
		return nil
	} else if err != nil {
		return errors.Wrap(err, "Failed to find a lock for the migrations")
	}

	// Creates a temporary collection containing unique migration documents.
	if _, err := db.Collection("migrations").Aggregate(context.TODO(), []bson.M{
		{"$match": bson.M{"version": bson.M{"$ne": nil}}},
		{"$sort": bson.M{"_id": 1}},
		{"$group": bson.M{"_id": "$version", "doc": bson.M{"$first": "$$ROOT"}}},
		{"$replaceRoot": bson.M{"newRoot": "$doc"}},
		{"$out": "migrations_tmp"},
	}); err != nil {
		return errors.Wrap(err, "Failed to create a temporary collection")
	}

	// Cleanup migrations collection.
	if _, err := db.Collection("migrations").DeleteMany(context.TODO(), bson.M{}); err != nil {
		return errors.Wrap(err, "Failed to cleanup the migrations collection")
	}

	// Copy documents from temporary collection to migrations collection.
	if _, err := db.Collection("migrations_tmp").Aggregate(context.TODO(), []bson.M{{"$out": "migrations"}}); err != nil {
		return errors.Wrap(err, "Failed to copy the documents to a new migration collection")
	}

	// Drop temporary collection.
	return db.Collection("migrations_tmp").Drop(context.TODO())
}
