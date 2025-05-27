package migrations

import (
	"context"
	"fmt"
	"slices"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration95 = migrate.Migration{
	Version:     95,
	Description: "Convert recorded sessions into session's events",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   95,
			"action":    "Up",
		}).Info("Applying migration")

		if !envs.IsEnterprise() {
			log.Info("skipping migration as the ShellHub instance isn't enterprise")

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

		for cursor.Next(ctx) {
			var result struct {
				UID string `bson:"_id"`
			}

			if err := cursor.Decode(&result); err != nil {
				log.WithError(err).Error("Failed to decode UID result")

				return err
			}

			uid := result.UID

			log.WithField("uid", uid).Debug("Processing session")

			logger := log.WithFields(log.Fields{
				"uid": uid,
			})

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
				logger.WithError(err).Error("Failed to query session records")

				return err
			}

			defer cursor.Close(ctx)

			s := db.Collection("sessions").FindOne(ctx, bson.M{
				"uid": uid,
			})
			if err != nil {
				logger.WithError(err).Error("Failed to query session records")

				return err
			}

			if s.Err() != nil {
				if _, err := db.Collection("recorded_sessions").DeleteMany(ctx, bson.M{
					"uid": uid,
				}); err != nil {
					logger.WithError(err).Error("failed to delete the recorded session when session isn't found")

					return err
				}

				log.WithField("uid", uid).Debug("Deleted recorded session for a not found session")

				continue
			}

			session := &models.Session{}
			if err := s.Decode(session); err != nil {
				logger.WithError(err).Error("failed to decode the session")

				return err
			}

			record := &models.RecordedSession{}

			if cursor.Next(ctx) {
				if err := cursor.Decode(record); err != nil {
					logger.WithError(err).Error("Failed to decode session record")

					return err
				}
			}

			if !slices.Contains(session.Events.Types, string(models.SessionEventTypePtyRequest)) {
				if _, err := db.Collection("sessions").UpdateOne(ctx,
					bson.M{"uid": uid},
					bson.M{
						"$addToSet": bson.M{
							"events.types": models.SessionEventTypePtyRequest,
							"events.seats": 0,
						},
					},
				); err != nil {
					logger.WithError(err).Error("Failed to update session events types to pty-req")

					return err
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
					logger.WithError(err).Error("Failed to insert session event pty-req")

					return err
				}
			}

			lastWidth, lastHeight := record.Width, record.Height

			if _, err := db.Collection("sessions").UpdateOne(ctx,
				bson.M{"uid": uid},
				bson.M{
					"$addToSet": bson.M{
						"events.types": models.SessionEventTypePtyOutput,
						"events.seats": 0,
					},
				},
			); err != nil {
				logger.WithError(err).Error("Failed to update session events types to pty-output")

				return err
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
				logger.WithError(err).Error("Failed to insert session event pty-output")

				return err
			}

			for cursor.Next(ctx) {
				if err := cursor.Decode(record); err != nil {
					logger.WithError(err).Error("Failed to decode session record")

					return err
				}

				if record.Width != lastWidth || record.Height != lastHeight {
					if !slices.Contains(session.Events.Types, string(models.SessionEventTypeWindowChange)) {
						if _, err := db.Collection("sessions").UpdateOne(ctx,
							bson.M{"uid": uid},
							bson.M{
								"$addToSet": bson.M{
									"events.types": models.SessionEventTypeWindowChange,
									"events.seats": 0,
								},
							},
						); err != nil {
							logger.WithError(err).Error("Failed to update session events types to window-change")

							return err
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
							logger.WithError(err).Error("Failed to insert session event window-change")

							return err
						}
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
					logger.WithError(err).Error("Failed to update session events types to pty-output")

					return err
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
					logger.WithError(err).Error("Failed to insert session event pty-output")

					return err
				}

			}

			if _, err := db.Collection("recorded_sessions").DeleteMany(ctx, bson.M{
				"uid": uid,
			}); err != nil {
				logger.WithError(err).Error("failed to delete the recorded session")

				return err
			}

			logger.Debug("Successfully processed session")
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   95,
			"action":    "Down",
		}).Info("Reverting migration")

		if !envs.IsEnterprise() {
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

		for cursor.Next(ctx) {
			var session struct {
				UID string `bson:"uid"`
			}

			if err := cursor.Decode(&session); err != nil {
				log.WithError(err).Error("Failed to decode session")

				return err
			}

			uid := session.UID
			log.WithField("uid", uid).Debug("Reverting session")

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

				return err
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
					d := &models.SSHPty{}

					data, _ := bson.Marshal(event.Data.(primitive.D))
					if err := bson.Unmarshal(data, &d); err != nil {
						return err
					}

					ptyReq := d

					lastWidth, lastHeight = ptyReq.Columns, ptyReq.Rows
				case models.SessionEventTypeWindowChange:
					d := &models.SSHWindowChange{}

					data, _ := bson.Marshal(event.Data.(primitive.D))
					if err := bson.Unmarshal(data, &d); err != nil {
						return err
					}

					winChange := d

					lastWidth, lastHeight = winChange.Columns, winChange.Rows
				case models.SessionEventTypePtyOutput:
					d := &models.SSHPtyOutput{}

					data, _ := bson.Marshal(event.Data.(primitive.D))
					if err := bson.Unmarshal(data, &d); err != nil {
						return err
					}

					ptyOutput := d

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

			log.WithField("uid", uid).Debug("Successfully reverted session")
		}

		return nil
	}),
}
