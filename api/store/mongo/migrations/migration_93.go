package migrations

import (
	"context"
	"fmt"
	"runtime"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/semaphore"
)

var migration93 = migrate.Migration{
	Version:     93,
	Description: "Convert recorded terminal sessions to Asciinema format",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   93,
			"action":    "Up",
		}).Info("Applying migration")

		if !envs.IsCloud() && !envs.IsEnterprise() {
			return nil
		}

		cursor, err := db.Collection("recorded_sessions").Aggregate(ctx, []bson.M{
			{
				"$match": bson.M{},
			},
			{
				"$group": bson.M{
					"_id": "$uid",
				},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to query session UIDs: %w", err)
		}

		defer cursor.Close(ctx)

		var (
			maxWorkers = runtime.GOMAXPROCS(0)
			sem        = semaphore.NewWeighted(int64(maxWorkers))
		)

		for cursor.Next(ctx) {
			if err := sem.Acquire(ctx, 1); err != nil {
				log.Printf("Failed to acquire semaphore: %v", err)

				break
			}

			go func() {
				defer sem.Release(1)

				var result struct {
					UID string `bson:"_id"`
				}

				if err := cursor.Decode(&result); err != nil {
					log.WithError(err).Error("Failed to decode UID result")

					return
				}

				uid := result.UID
				log.WithField("uid", uid).Info("Processing session")

				query := []bson.M{
					{
						"$match": bson.M{
							"uid": uid,
						},
					},
					{
						"$sort": bson.M{
							"time": 1,
						},
					},
				}

				cursor, err := db.Collection("recorded_sessions").Aggregate(ctx, query)
				if err != nil {
					log.WithError(err).WithField("uid", uid).Error("Failed to query session records")

					return
				}

				record := &models.RecordedSession{}

				if cursor.Next(ctx) {
					if err := cursor.Decode(record); err != nil {
						log.WithError(err).WithField("uid", uid).Error("Failed to decode session record")

						return
					}
				}

				if _, err := db.Collection("sessions").UpdateOne(ctx,
					bson.M{"uid": uid},
					bson.M{
						"$addToSet": bson.M{
							"events.types": models.SessionEventTypePtyRequest,
							"events.seats": 0,
						},
					},
				); err != nil {
					return
				}

				if _, err := db.Collection("sessions_events").InsertOne(ctx, &models.SessionEvent{
					Session:   uid,
					Type:      models.SessionEventTypePtyRequest,
					Timestamp: record.Time,
					Data: &models.SSHPty{
						Term:     "",
						Columns:  uint32(record.Width),
						Rows:     uint32(record.Height),
						Width:    0,
						Height:   0,
						Modelist: []byte{},
					},
					Seat: 0,
				}); err != nil {
					return
				}

				lastWidth, lastHeight := record.Width, record.Height

				for cursor.Next(ctx) {
					if err := cursor.Decode(record); err != nil {
						log.WithError(err).WithField("uid", uid).Error("Failed to decode session record")

						return
					}

					if record.Width != lastWidth || record.Height != lastHeight {
						if _, err := db.Collection("sessions").UpdateOne(ctx,
							bson.M{"uid": uid},
							bson.M{
								"$addToSet": bson.M{
									"events.types": models.SessionEventTypeWindowChange,
									"events.seats": 0,
								},
							},
						); err != nil {
							return
						}

						if _, err := db.Collection("sessions_events").InsertOne(ctx, &models.SessionEvent{
							Session:   uid,
							Type:      models.SessionEventTypeWindowChange,
							Timestamp: record.Time,
							Data: &models.SSHWindowChange{
								Columns: uint32(record.Width),
								Rows:    uint32(record.Height),
								Width:   0,
								Height:  0,
							},
							Seat: 0,
						}); err != nil {
							return
						}

						lastWidth, lastHeight = record.Width, record.Height
					}

					if _, err := db.Collection("sessions").UpdateOne(ctx,
						bson.M{"uid": uid},
						bson.M{
							"$addToSet": bson.M{
								"events.types": models.SessionEventTypePtyOutput,
								"events.seats": 0,
							},
						},
					); err != nil {
						return
					}

					if _, err := db.Collection("sessions_events").InsertOne(ctx, &models.SessionEvent{
						Session:   uid,
						Type:      models.SessionEventTypePtyOutput,
						Timestamp: record.Time,
						Data: &models.SSHPtyOutput{
							Output: record.Message,
						},
						Seat: 0,
					}); err != nil {
						return
					}

				}

				log.WithField("uid", uid).Info("Successfully processed session")
			}()
		}

		if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
			log.Printf("Failed to acquire semaphore: %v", err)
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   93,
			"action":    "Down",
		}).Info("Reverting migration")

		if !envs.IsCloud() && !envs.IsEnterprise() {
			return nil
		}

		cursor, err := db.Collection("sessions").Find(ctx, bson.M{
			"events.types": bson.M{
				"$in": []models.SessionEventType{
					models.SessionEventTypePtyRequest,
					models.SessionEventTypePtyOutput,
					models.SessionEventTypeWindowChange,
				},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to query sessions: %w", err)
		}

		defer cursor.Close(ctx)

		var (
			maxWorkers = runtime.GOMAXPROCS(0)
			sem        = semaphore.NewWeighted(int64(maxWorkers))
		)

		for cursor.Next(ctx) {
			if err := sem.Acquire(ctx, 1); err != nil {
				log.Printf("Failed to acquire semaphore: %v", err)

				break
			}

			go func() {
				defer sem.Release(1)

				var session struct {
					UID string `bson:"uid"`
				}

				if err := cursor.Decode(&session); err != nil {
					log.WithError(err).Error("Failed to decode session")

					return
				}

				uid := session.UID
				log.WithField("uid", uid).Info("Reverting session")

				eventsCursor, err := db.Collection("sessions_events").Find(ctx, bson.M{
					"session": uid,
					"type": bson.M{
						"$in": []models.SessionEventType{
							models.SessionEventTypePtyRequest,
							models.SessionEventTypePtyOutput,
							models.SessionEventTypeWindowChange,
						},
					},
				}, options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}}))
				if err != nil {
					log.WithError(err).WithField("uid", uid).Error("Failed to query session events")

					return
				}

				defer eventsCursor.Close(ctx)

				var lastWidth, lastHeight uint32
				for eventsCursor.Next(ctx) {
					var event models.SessionEvent
					if err := eventsCursor.Decode(&event); err != nil {
						log.WithError(err).WithField("uid", uid).Error("Failed to decode event")

						continue
					}

					switch event.Type {
					case models.SessionEventTypePtyRequest:
						ptyReq := event.Data.(*models.SSHPty)

						lastWidth, lastHeight = ptyReq.Columns, ptyReq.Rows
					case models.SessionEventTypeWindowChange:
						winChange := event.Data.(*models.SSHWindowChange)

						lastWidth, lastHeight = winChange.Columns, winChange.Rows
					case models.SessionEventTypePtyOutput:
						ptyOutput := event.Data.(*models.SSHPtyOutput)

						_, err := db.Collection("recorded_sessions").InsertOne(ctx, bson.M{
							"uid":     uid,
							"message": ptyOutput.Output,
							"time":    event.Timestamp,
							"width":   lastWidth,
							"height":  lastHeight,
						})
						if err != nil {
							log.WithError(err).WithField("uid", uid).Error("Failed to insert recorded session")
						}
					}
				}

				_, err = db.Collection("sessions_events").DeleteMany(ctx, bson.M{
					"session": uid,
					"type": bson.M{
						"$in": []models.SessionEventType{
							models.SessionEventTypePtyRequest,
							models.SessionEventTypePtyOutput,
							models.SessionEventTypeWindowChange,
						},
					},
				})
				if err != nil {
					log.WithError(err).WithField("uid", uid).Error("Failed to delete session events")
				}

				_, err = db.Collection("sessions").UpdateOne(ctx,
					bson.M{"uid": uid},
					bson.M{
						"$pull": bson.M{
							"events.types": bson.M{
								"$in": []models.SessionEventType{
									models.SessionEventTypePtyRequest,
									models.SessionEventTypePtyOutput,
									models.SessionEventTypeWindowChange,
								},
							},
						},
					},
				)
				if err != nil {
					log.WithError(err).WithField("uid", uid).Error("Failed to update session")
				}

				log.WithField("uid", uid).Info("Successfully reverted session")
			}()
		}

		if err := sem.Acquire(ctx, int64(maxWorkers)); err != nil {
			log.WithError(err).Printf("Failed to acquire semaphore")
		}

		return nil
	}),
}
