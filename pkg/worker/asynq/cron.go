package asynq

import (
	"github.com/shellhub-io/shellhub/pkg/worker"
)

// Unique configures a cron job to prevent concurrent processing.
// When enabled, the job will not be enqueued or executed again until it completes
// or the timeout specified in `SHELLHUB_ASYNQ_UNIQUENESS_TIMEOUT` is reached.
func Unique() worker.CronjobOption {
	return func(c *worker.Cronjob) {
		c.Unique = true
	}
}
