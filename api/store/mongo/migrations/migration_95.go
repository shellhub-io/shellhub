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
	Version:     MigrationVersion95,
	Description: "Convert recorded sessions into session's events",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.WithFields(log.Fields{
			"component": "migration",
			"version":   MigrationVersion95,
			"action":    "Up",
		}).Info("Applying migration")

		if !envs.IsEnterprise() {
			log.Info("skipping migration as the ShellHub instance isn't enterprise")

			return nil
		}

		/*sessionUIDsCursor, err := db.Collection("recorded_sessions").Aggregate(ctx, []bson.M{
			{"$group": bson.M{"_id": "$uid"}},
		})
		if err != nil {
			return fmt.Errorf("failed to query session UIDs: %w", err)
		}
		defer sessionUIDsCursor.Close(ctx)

		var sessionUIDs []string
		for sessionUIDsCursor.Next(ctx) {
			var result struct {
				UID string `bson:"_id"`
			}
			if err := sessionUIDsCursor.Decode(&result); err != nil {
				log.WithError(err).Error("Failed to decode UID result")

				return err
			}
			sessionUIDs = append(sessionUIDs, result.UID)
		}*/

		sessionUIDs, err := db.Collection("recorded_sessions").Distinct(ctx, "uid", bson.M{})
		if err != nil {
			log.WithError(err).Error("failed to get all recorded_sessions uids")

			return fmt.Errorf("failed to query session UIDs: %w", err)
		}

		if len(sessionUIDs) == 0 {
			log.Info("No recorded sessions found")

			return nil
		}

		sessionsCursor, err := db.Collection("sessions").Find(ctx, bson.M{
			"uid": bson.M{"$in": sessionUIDs},
		})
		if err != nil {
			return fmt.Errorf("failed to query sessions: %w", err)
		}

		defer sessionsCursor.Close(ctx)

		existingSessions := make(map[string]*models.Session)
		for sessionsCursor.Next(ctx) {
			var session models.Session
			if err := sessionsCursor.Decode(&session); err != nil {
				log.WithError(err).Error("Failed to decode session")

				return err
			}
			existingSessions[session.UID] = &session
		}

		for _, uid := range sessionUIDs {
			session := existingSessions[uid.(string)]

			logger := log.WithField("uid", uid)

			logger.Debug("Processing session")
			if session == nil {
				if _, err := db.Collection("recorded_sessions").DeleteMany(ctx, bson.M{"uid": uid}); err != nil {
					logger.WithError(err).Error("failed to delete the recorded session when session isn't found")

					return err
				}

				logger.Debug("Deleted recorded session for a not found session")

				continue
			}

			recordsCursor, err := db.Collection("recorded_sessions").Find(ctx, bson.M{"uid": uid}, options.Find().SetSort(bson.D{{Key: "time", Value: 1}}))
			if err != nil {
				logger.WithError(err).Error("Failed to query session records")

				return err
			}

			defer recordsCursor.Close(ctx)

			var records []models.RecordedSession
			if err := recordsCursor.All(ctx, &records); err != nil {
				logger.WithError(err).Error("Failed to decode all records")

				return err
			}

			if len(records) == 0 {
				logger.Debug("No records found for session")

				return nil
			}

			var sessionEvents []interface{}
			var sessionUpdates []mongo.WriteModel
			eventTypesToAdd := make(map[string]bool)
			firstRecord := records[0]

			lastWidth, lastHeight := firstRecord.Width, firstRecord.Height

			if !slices.Contains(session.Events.Types, string(models.SessionEventTypePtyRequest)) {
				eventTypesToAdd[string(models.SessionEventTypePtyRequest)] = true
				sessionEvents = append(
					sessionEvents,
					&models.SessionEvent{
						Session:   uid.(string),
						Type:      models.SessionEventTypePtyRequest,
						Timestamp: firstRecord.Time,
						Data: &models.SSHPty{
							Term:     "",
							Columns:  uint32(firstRecord.Width),
							Rows:     uint32(firstRecord.Height),
							Width:    0,
							Height:   0,
							Modelist: []byte{},
						},
						Seat: 0,
					},
				)
			}

			if !slices.Contains(session.Events.Types, string(models.SessionEventTypePtyOutput)) {
				eventTypesToAdd[string(models.SessionEventTypePtyOutput)] = true
			}

			for _, record := range records {
				if record.Width != lastWidth || record.Height != lastHeight {
					if !slices.Contains(session.Events.Types, string(models.SessionEventTypeWindowChange)) {
						eventTypesToAdd[string(models.SessionEventTypeWindowChange)] = true
					}
					sessionEvents = append(
						sessionEvents,
						&models.SessionEvent{
							Session:   uid.(string),
							Type:      models.SessionEventTypeWindowChange,
							Timestamp: record.Time,
							Data: &models.SSHWindowChange{
								Columns: uint32(record.Width),
								Rows:    uint32(record.Height),
								Width:   0,
								Height:  0,
							},
							Seat: 0,
						},
					)
					lastWidth, lastHeight = record.Width, record.Height
				}
				sessionEvents = append(
					sessionEvents,
					&models.SessionEvent{
						Session:   uid.(string),
						Type:      models.SessionEventTypePtyOutput,
						Timestamp: record.Time,
						Data:      &models.SSHPtyOutput{Output: record.Message},
						Seat:      0,
					},
				)
			}

			if len(sessionEvents) > 0 {
				if _, err := db.Collection("sessions_events").InsertMany(ctx, sessionEvents); err != nil {
					logger.WithError(err).Error("Failed to bulk insert session events")

					return err
				}
			}

			if len(eventTypesToAdd) > 0 {
				var typesToAdd []string
				for eventType := range eventTypesToAdd {
					typesToAdd = append(typesToAdd, eventType)
				}
				updateDoc := bson.M{"$addToSet": bson.M{
					"events.types": bson.M{"$each": typesToAdd},
					"events.seats": 0,
				}}
				sessionUpdates = append(sessionUpdates, mongo.NewUpdateOneModel().SetFilter(bson.M{"uid": uid}).SetUpdate(updateDoc))
			}

			if len(sessionUpdates) > 0 {
				if _, err := db.Collection("sessions").BulkWrite(ctx, sessionUpdates); err != nil {
					logger.WithError(err).Error("Failed to bulk update session")

					return err
				}
			}

			if _, err := db.Collection("recorded_sessions").DeleteMany(ctx, bson.M{"uid": uid}); err != nil {
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
			"version":   MigrationVersion95,
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

			var recordedSessions []interface{}
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
					lastWidth, lastHeight = d.Columns, d.Rows

				case models.SessionEventTypeWindowChange:
					d := &models.SSHWindowChange{}
					data, _ := bson.Marshal(event.Data.(primitive.D))
					if err := bson.Unmarshal(data, &d); err != nil {
						return err
					}
					lastWidth, lastHeight = d.Columns, d.Rows

				case models.SessionEventTypePtyOutput:
					d := &models.SSHPtyOutput{}
					data, _ := bson.Marshal(event.Data.(primitive.D))
					if err := bson.Unmarshal(data, &d); err != nil {
						return err
					}

					recordedSessions = append(recordedSessions, bson.M{
						"uid":     uid,
						"message": d.Output,
						"time":    event.Timestamp,
						"width":   lastWidth,
						"height":  lastHeight,
					})
				}
			}

			if len(recordedSessions) > 0 {
				if _, err := db.Collection("recorded_sessions").InsertMany(ctx, recordedSessions); err != nil {
					log.WithError(err).WithField("uid", uid).Error("Failed to bulk insert recorded sessions")
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
