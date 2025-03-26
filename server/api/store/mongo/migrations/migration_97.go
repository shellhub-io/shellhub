package migrations

import (
	"context"
	"errors"
	"runtime"

	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/semaphore"
)

var ErrMigration97StatusNot200 = errors.New("failed to save the session as asciinema file")

var migration97 = migrate.Migration{
	Version:     97,
	Description: "Save session's events as Asciinema file on Object Storage",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   97,
			"action":    "Up",
		}).Info("Applying migration")

		if !envs.IsEnterprise() {
			return nil
		}

		cursor, err := db.Collection("sessions").Aggregate(ctx, []bson.M{
			{
				"$match": bson.M{
					"recorded": true,
					"events.types": bson.M{
						"$all": []models.SessionEventType{
							models.SessionEventTypePtyRequest,
							models.SessionEventTypePtyOutput,
						},
					},
				},
			},
		})
		if err != nil {
			log.WithError(err).Error("Failed to find recorded sessions")

			return err
		}

		defer cursor.Close(ctx)

		cli, err := internalclient.NewClient()
		if err != nil {
			log.WithError(err).Error("Failed to find recorded sessions")

			return err
		}

		var (
			maxWorkers = runtime.GOMAXPROCS(0)
			sem        = semaphore.NewWeighted(int64(maxWorkers))
		)

		session := &models.Session{}

		for cursor.Next(ctx) {
			if err := sem.Acquire(ctx, 1); err != nil {
				log.Printf("Failed to acquire semaphore: %v", err)

				break
			}

			if err := cursor.Decode(&session); err != nil {
				log.WithError(err).Error("Failed to decode UID result")
				sem.Release(1)

				return err
			}

			go func(session models.Session) {
				defer sem.Release(1)

				for s := range session.Events.Seats {
					uid, seat := session.UID, s

					log.WithFields(log.Fields{
						"uid":  uid,
						"seat": seat,
					}).Debug("Processing session as Asciinema file")

					if err := cli.SaveSession(uid, seat); err != nil {
						log.WithError(err).Error("Error on saving session a session")

						return
					}

					log.WithFields(log.Fields{
						"uid":  uid,
						"seat": seat,
					}).Debug("Session saved as Asciinema file")
				}
			}(*session)
		}

		if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
			log.Printf("Failed to acquire semaphore: %v", err)
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   93,
			"action":    "Down",
		}).Info("Cannot undo migration")

		return nil
	}),
}
