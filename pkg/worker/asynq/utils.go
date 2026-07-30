package asynq

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/pkg/worker"
)

// queues is a map of queues where the key is the name and the value is the priority.
type queues map[string]int

// cronQueue is the queue where's all the cronjobs will send tasks.
const cronQueue = "cron"

// cronToAsynq converts a [github.com/shellhub-io/shellhub/pkg/api/worker.CronHandler] to an asynq handler.
func cronToAsynq(h worker.CronHandler) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, _ *asynq.Task) error {
		return h(ctx)
	}
}

// buildCronOptions generates a slice of asynq.Options for configuring a cron job.
func buildCronOptions(s *server, c *worker.Cronjob) []asynq.Option {
	opts := make([]asynq.Option, 0)

	if c.Unique && s.uniquenessTimeout > 0 {
		opts = append(opts, asynq.Unique(time.Duration(s.uniquenessTimeout)*time.Hour))
	}

	return opts
}

// taskToAsynq converts a [github.com/shellhub-io/shellhub/pkg/api/worker.TaskHandler] to an asynq handler.
func taskToAsynq(h worker.TaskHandler) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		return h(ctx, task.Payload())
	}
}
