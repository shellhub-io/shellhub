package environment

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

const (
	redisPortEnv = "SHELLHUB_TEST_REDIS_PORT"
	cronQueue    = "cron"
)

var (
	// ErrCronUnreachable is returned by [Stack.RunCron] for a stack brought up without
	// [DockerComposeConfigurator.WithCronTrigger], whose redis sits on a port the test does not know.
	ErrCronUnreachable = errors.New("the stack's redis port is unknown; bring it up WithCronTrigger")
	// ErrCronNotScheduled is returned by [Stack.RunCron] when the server has registered no cron job
	// on the spec, which it also is in the seconds before its scheduler first reports in.
	ErrCronNotScheduled = errors.New("no cron job is scheduled on the spec")
)

// RunCron enqueues, once and now, every cron job the server scheduled on spec, the five-field
// expression they were registered with, so the jobs run in the server as they would at their next
// tick without waiting for it. A job is told apart only by its spec, so jobs sharing one, such as
// the enterprise recording archive and the active-session reaper on "* * * * *", all run. It
// returns once the jobs are queued, not once they have run, so read their outcome for that. It
// returns ErrCronUnreachable for a stack whose redis port is unknown, ErrCronNotScheduled when no
// job is scheduled on spec, and redis's error when it cannot read the schedule or queue a job,
// having queued the jobs before it. ctx bounds the enqueue.
func (s *Stack) RunCron(ctx context.Context, spec string) error {
	port := s.envs[redisPortEnv]
	if port == "" {
		return ErrCronUnreachable
	}

	redis := asynq.RedisClientOpt{Addr: "localhost:" + port} //nolint:exhaustruct // the remaining fields keep asynq's defaults

	inspector := asynq.NewInspector(redis)
	defer inspector.Close() //nolint:errcheck // the inspector only read the schedule

	entries, err := inspector.SchedulerEntries()
	if err != nil {
		return err
	}

	var scheduled []*asynq.Task

	for _, entry := range entries {
		if entry.Spec == spec {
			scheduled = append(scheduled, entry.Task)
		}
	}

	if len(scheduled) == 0 {
		return fmt.Errorf("%w: %q", ErrCronNotScheduled, spec)
	}

	client := asynq.NewClient(redis)
	defer client.Close() //nolint:errcheck // every task is queued once EnqueueContext returns

	for _, task := range scheduled {
		if _, err := client.EnqueueContext(ctx, asynq.NewTask(task.Type(), task.Payload()), asynq.Queue(cronQueue)); err != nil {
			return err
		}
	}

	return nil
}

// RunCron enqueues every cron job the server scheduled on spec, as [Stack.RunCron] does, waiting
// up to 30 seconds for the server's scheduler to report a job in. It fails t on any other error,
// and if no job is scheduled when the wait ends.
func (dc *DockerCompose) RunCron(t *testing.T, spec string) {
	t.Helper()

	var err error

	require.Eventually(t, func() bool {
		err = dc.stack.RunCron(t.Context(), spec)

		return !errors.Is(err, ErrCronNotScheduled)
	}, 30*time.Second, 1*time.Second, "no cron job was scheduled on %q", spec)
	require.NoError(t, err)
}
