package asynq

import (
	"bytes"
	"context"
	"time"

	"github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/pkg/worker"
)

// batchConfig configures the asynq batch settings.
type batchConfig struct {
	// maxSize is the maximum number of tasks that a batch task can handle before
	// processing.
	maxSize int
	// maxDelay is the maximum amount of time that a batch task can wait before
	// processing.
	maxDelay time.Duration
	// gracePeriod is the amount of time that the server will wait before aggregating
	// batch tasks.
	gracePeriod time.Duration
}

// queues is a map of queues where the key is the name and the value is the priority.
type queues map[string]int

// cronQueue is the queue where's all the cronjobs will send tasks.
const cronQueue = "cron"

// aggregate is the handler that Asynq will execute to aggregate the tasks.
// It will combine all task payloads into one, separated by '\n', and then
// execute a new task with the name "{group}+:aggregated".
func aggregate(group string, tasks []*asynq.Task) *asynq.Task {
	buf := new(bytes.Buffer)
	for _, t := range tasks {
		buf.Write(t.Payload())
		buf.WriteByte('\n')
	}

	return asynq.NewTask(group+":batch", buf.Bytes())
}

// cronToAsynq converts a [github.com/shellhub-io/shellhub/pkg/api/worker.CronHandler] to an asynq handler.
func cronToAsynq(h worker.CronHandler) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, _ *asynq.Task) error {
		return h(ctx)
	}
}

// taskToAsynq converts a [github.com/shellhub-io/shellhub/pkg/api/worker.TaskHandler] to an asynq handler.
func taskToAsynq(h worker.TaskHandler) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		return h(ctx, task.Payload())
	}
}
