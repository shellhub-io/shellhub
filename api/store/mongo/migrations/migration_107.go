package migrations

import (
	"context"

	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration107 = migrate.Migration{
	Version:     107,
	Description: "Restructure SAML signon_url to signon_urls object with post and redirect fields",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 107, "action": "Up"}).Info("Applying migration")

		pipeline := []bson.M{
			{
				"$set": bson.M{
					"authentication.saml.idp.signon_urls": bson.M{
						"post":              bson.M{"$ifNull": []any{"$authentication.saml.idp.signon_url", ""}},
						"redirect":          "",
						"preferred_binding": "post",
					},
				},
			},
			{
				"$unset": "authentication.saml.idp.signon_url",
			},
		}

		if _, err := db.Collection("system").UpdateOne(ctx, bson.M{}, pipeline); err != nil {
			log.WithError(err).Error("Failed to update system document")

			return err
		}

		log.Info("Successfully restructured SAML signon_url to signon_urls")

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{"component": "migration", "version": 107, "action": "Down"}).Info("Reverting migration")

		pipeline := []bson.M{
			{
				"$set": bson.M{
					"authentication.saml.idp.signon_url": "$authentication.saml.idp.signon_urls.post",
				},
			},
			{
				"$unset": "authentication.saml.idp.signon_urls",
			},
		}

		if _, err := db.Collection("system").UpdateOne(ctx, bson.M{}, pipeline); err != nil {
			log.WithError(err).Error("Failed to revert system document")

			return err
		}

		log.Info("Successfully reverted SAML signon_urls to signon_url")

		return nil
	}),
}
