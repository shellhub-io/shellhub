package workers

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

// StartCleaner starts a worker to delete session's records registers older than days defined by
// SHELLHUB_RECORD_RETENTION.
//
// If something inside the function does not work properly, it will panic.
// When SHELLHUB_RECORD_RETENTION is equals to zero, records will never be deleted.
// When SHELLHUB_RECORD_RETENTION is less than zero, nothing happen.
func StartCleaner(ctx context.Context, store store.Store) (err error) {
	log.WithFields(log.Fields{
		"workder": "cleaner",
	}).Info("Starting cleaner worker")
	defer log.WithFields(log.Fields{
		"workder": "cleaner",
	}).Info("Cleaner worker done")
	defer func() {
		// NOTE: Due to named return, err, we can log what happened using this defer function, avoiding a `log.Error` on
		// each error's return.
		log.WithFields(log.Fields{
			"workder": "cleaner",
		}).Error(err)
	}()

	envs, err := getEnvs()
	if err != nil {
		return err
	}

	if envs.SessionRecordCleanupRetention == 0 {
		return fmt.Errorf("stopping cleaner worker due cleaup rentention set to zero")
	}

	if envs.SessionRecordCleanupRetention < 0 {
		return fmt.Errorf("invalid time interval: %w", fmt.Errorf("%d is not a valid time interval", envs.SessionRecordCleanupRetention))
	}

	limit := time.Now().UTC().AddDate(0, 0, envs.SessionRecordCleanupRetention*-1)

	mongoStore := store.(*mongo.Store)

	addr, err := asynq.ParseRedisURI(envs.RedisURI)
	if err != nil {
		return fmt.Errorf("failed to parse redis uri: %w", err)
	}

	srv := asynq.NewServer(
		addr,
		asynq.Config{ //nolint:exhaustruct
			BaseContext: func() context.Context {
				return ctx
			},
			Concurrency: runtime.NumCPU(),
		},
	)

	mux := asynq.NewServeMux()

	// Handle session_record:cleanup task
	mux.HandleFunc("session_record:cleanup", func(ctx context.Context, task *asynq.Task) error {
		if _, err := mongoStore.Database().Collection("recorded_sessions").DeleteMany(ctx,
			bson.M{"time": bson.D{{"$lte", limit}}},
		); err != nil {
			return err
		}

		if _, err := mongoStore.Database().Collection("sessions").UpdateMany(ctx,
			bson.M{"started_at": bson.D{{"$lte", limit}}, "recorded": bson.M{"$eq": true}},
			bson.M{"$set": bson.M{"recorded": false}}); err != nil {
			return err
		}

		return nil
	})

	go func() {
		if err := srv.Run(mux); err != nil {
			log.Fatal(err)
		}
	}()

	scheduler := asynq.NewScheduler(addr, nil)

	// Schedule session_record:cleanup to run once a day
	if _, err := scheduler.Register(envs.SessionRecordCleanupSchedule,
		asynq.NewTask("session_record:cleanup", nil, asynq.TaskID("session_record:cleanup"))); err != nil {
		log.Error(err)
	}

	return scheduler.Run() //nolint:contextcheck
}
