package asynq

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/pkg/worker"
)

type queues map[string]int

const cronQueue = "cron"

func cronToAsynq(h worker.CronHandler) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, _ *asynq.Task) error {
		return h(ctx)
	}
}

func buildCronOptions(s *server, c *worker.Cronjob) []asynq.Option {
	opts := make([]asynq.Option, 0)

	if c.Unique && s.uniquenessTimeout > 0 {
		opts = append(opts, asynq.Unique(time.Duration(s.uniquenessTimeout)*time.Hour))
	}

	return opts
}

func taskToAsynq(h worker.TaskHandler) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		return h(ctx, task.Payload())
	}
}
