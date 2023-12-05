package workers

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	log "github.com/sirupsen/logrus"
)

// registerSessionCleanup worker is designed to delete recorded sessions older than a specified number
// of days. The retention period is determined by the value of the `SHELLHUB_RECORD_RETENTION` environment
// variable. To disable this worker, set `SHELLHUB_RECORD_RETENTION` to 0 (default behavior). It uses
// a cron expression from `SHELLHUB_RECORD_RETENTION` to schedule its periodic execution.
func (w *Workers) registerSessionCleanup() {
	if w.env.SessionRecordCleanupRetention < 1 {
		log.WithFields(
			log.Fields{
				"component": "worker",
				"task":      TaskSessionCleanup,
			}).
			Warnf("Aborting cleanup worker due to SHELLHUB_RECORD_RETENTION equal to %d.", w.env.SessionRecordCleanupRetention)

		return
	}

	w.mux.HandleFunc(TaskSessionCleanup, func(ctx context.Context, _ *asynq.Task) error {
		log.WithFields(
			log.Fields{
				"component":       "worker",
				"cron_expression": w.env.SessionRecordCleanupSchedule,
				"task":            TaskSessionCleanup,
			}).
			Trace("Executing cleanup worker.")

		lte := time.Now().UTC().AddDate(0, 0, w.env.SessionRecordCleanupRetention*(-1))
		deletedCount, updatedCount, err := w.store.SessionDeleteRecordFrameByDate(ctx, lte)
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

		log.WithFields(
			log.Fields{
				"component":       "worker",
				"cron_expression": w.env.SessionRecordCleanupSchedule,
				"task":            TaskSessionCleanup,
				"lte":             lte.String(),
				"deleted_count":   deletedCount,
				"updated_count":   updatedCount,
			}).
			Trace("Finishing cleanup worker.")

		return nil
	})

	task := asynq.NewTask(TaskSessionCleanup, nil, asynq.TaskID(TaskSessionCleanup), asynq.Queue("session_record"))
	if _, err := w.scheduler.Register(w.env.SessionRecordCleanupSchedule, task); err != nil {
		log.WithFields(
			log.Fields{
				"component": "worker",
				"task":      TaskSessionCleanup,
			}).
			WithError(err).
			Error("Failed to register the scheduler.")
	}
}
