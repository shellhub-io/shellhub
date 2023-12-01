package workers

import (
	"context"
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
func StartCleaner(_ context.Context, store store.Store) {
	envs, err := getEnvs()
	if err != nil {
		log.WithFields(log.Fields{"component": "worker", "task": TaskSessionCleanup}).
			WithError(err).
			Error("Failed to parse the envs.")

		return
	}

	if envs.SessionRecordCleanupRetention < 1 {
		log.WithFields(log.Fields{"component": "worker", "task": TaskSessionCleanup}).
			Warnf("Aborting cleanup worker due to SHELLHUB_RECORD_RETENTION equal to %d.", envs.SessionRecordCleanupRetention)

		return
	}

	lte := time.Now().UTC().AddDate(0, 0, envs.SessionRecordCleanupRetention*-1)

	mongoStore := store.(*mongo.Store)

	addr, err := asynq.ParseRedisURI(envs.RedisURI)
	if err != nil {
		log.WithFields(log.Fields{"component": "worker", "task": TaskSessionCleanup}).
			WithError(err).
			Errorf("Failed to parse redis URI: %s.", envs.RedisURI)

		return
	}

	srv := asynq.NewServer(
		addr,
		asynq.Config{ //nolint:exhaustruct
			Concurrency: runtime.NumCPU(),
		},
	)

	mux := asynq.NewServeMux()

	mux.HandleFunc("session_record:cleanup", func(ctx context.Context, task *asynq.Task) error {
		log.WithFields(
			log.Fields{
				"component":       "worker",
				"cron_expression": envs.SessionRecordCleanupSchedule,
				"task":            TaskSessionCleanup,
				"lte":             lte.String(),
			}).
			Info("Executing cleanup worker.")

		_, err := mongoStore.Database().Collection("recorded_sessions").DeleteMany(
			ctx,
			bson.M{
				"time": bson.D{
					{"$lte", lte},
				},
			},
		)
		if err != nil {
			log.WithFields(
				log.Fields{
					"component": "worker",
					"task":      TaskSessionCleanup,
				}).
				WithError(err).
				Error("Failed to delete recorded sessions")

			return err
		}

		_, err = mongoStore.Database().Collection("sessions").UpdateMany(
			ctx,
			bson.M{
				"started_at": bson.D{
					{"$lte", lte},
				},
				"recorded": bson.M{
					"$eq": true,
				},
			},
			bson.M{
				"$set": bson.M{
					"recorded": false,
				},
			},
		)
		if err != nil {
			log.WithFields(
				log.Fields{
					"component": "worker",
					"task":      TaskSessionCleanup,
				}).
				WithError(err).
				Error("Failed to update sessions")

			return err
		}

		log.WithFields(
			log.Fields{
				"component":       "worker",
				"cron_expression": envs.SessionRecordCleanupSchedule,
				"task":            TaskSessionCleanup,
				"lte":             lte.String(),
			}).
			Info("Finishing cleanup worker.")

		return nil
	})

	go func() {
		if err := srv.Run(mux); err != nil {
			log.WithFields(log.Fields{"component": "worker", "task": TaskSessionCleanup}).
				WithError(err).
				Fatal("Unable to run the server.")
		}
	}()

	scheduler := asynq.NewScheduler(addr, nil)

	task := asynq.NewTask("session_record:cleanup", nil, asynq.TaskID("session_record:cleanup"))
	if _, err := scheduler.Register(envs.SessionRecordCleanupSchedule, task); err != nil {
		log.WithFields(log.Fields{"component": "worker", "task": TaskSessionCleanup}).
			WithError(err).
			Error("Failed to register the scheduler.")
	}

	if err := scheduler.Run(); err != nil {
		log.WithFields(log.Fields{"component": "worker", "task": TaskSessionCleanup}).
			WithError(err).
			Fatal("Unable to run the scheduler.")
	}
}
