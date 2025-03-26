package migrations

import (
	"context"
	"net"
	"os"

	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/geoip/geolite2"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration47 = migrate.Migration{
	Version:     47,
	Description: "",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
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
		logrus.WithFields(logrus.Fields{
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
