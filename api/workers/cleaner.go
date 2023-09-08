package workers

import (
	"context"
	"runtime"
	"time"

	"github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type Cleaner struct {
	store store.Store
}

var _ Worker = (*Cleaner)(nil)

func NewCleaner(store store.Store) *Cleaner {
	return &Cleaner{
		store: store,
	}
}

// Start starts a worker to delete session's records registers older than days defined by
// SHELLHUB_RECORD_RETENTION.
//
// If something inside the function does not work properly, it will panic.
// When SHELLHUB_RECORD_RETENTION is equals to zero, records will never be deleted.
// When SHELLHUB_RECORD_RETENTION is less than zero, nothing happen.
func (c *Cleaner) Start(ctx context.Context, msgs chan WorkerMessage) {
	envs, err := getEnvs()
	if err != nil {
		// return fmt.Errorf("failed to get the envs: %w", err)
		msgs <- NewWorkerMessage("failed to get the envs", err)

		return
	}

	if envs.SessionRecordCleanupRetention == 0 {
		msgs <- NewWorkerMessage("session retention time is zero", nil)

		return
	}

	if envs.SessionRecordCleanupRetention < 0 {
		msgs <- NewWorkerMessage("invalid time interval", err)

		return
	}

	limit := time.Now().UTC().AddDate(0, 0, envs.SessionRecordCleanupRetention*-1)

	store := c.store.(*mongo.Store)

	addr, err := asynq.ParseRedisURI(envs.RedisURI)
	if err != nil {
		msgs <- NewWorkerMessage("failed to parse redis uri", err)

		return
	}

	srv := asynq.NewServer(
		addr,
		asynq.Config{ //nolint:exhaustruct
			Concurrency: runtime.NumCPU(),
			BaseContext: func() context.Context { return ctx },
		},
	)

	mux := asynq.NewServeMux()

	// Handle session_record:cleanup task
	mux.HandleFunc("session_record:cleanup", func(ctx context.Context, task *asynq.Task) error {
		if _, err := store.Database().Collection("recorded_sessions").DeleteMany(ctx,
			bson.M{"time": bson.D{{"$lte", limit}}},
		); err != nil {
			return err
		}

		if _, err := store.Database().Collection("sessions").UpdateMany(ctx,
			bson.M{"started_at": bson.D{{"$lte", limit}}, "recorded": bson.M{"$eq": true}},
			bson.M{"$set": bson.M{"recorded": false}}); err != nil {
			return err
		}

		return nil
	})

	go func() {
		if err := srv.Run(mux); err != nil {
			msgs <- NewWorkerMessage("failed to run server", err)
		}
	}()

	scheduler := asynq.NewScheduler(addr, nil)

	// Schedule session_record:cleanup to run once a day
	if _, err := scheduler.Register(envs.SessionRecordCleanupSchedule,
		asynq.NewTask("session_record:cleanup", nil, asynq.TaskID("session_record:cleanup"))); err != nil {
		msgs <- NewWorkerMessage("failed to register task", err)

		return
	}

	msgs <- WorkerMessageStarted

	if err := scheduler.Run(); err != nil { //nolint:contextcheck
		msgs <- WorkerMessageStopped
	}
}
