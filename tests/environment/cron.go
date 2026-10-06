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
	// ErrCronAmbiguous is returned by [Stack.RunCron] when more than one cron job is scheduled on
	// the spec, since a job is told apart only by its spec.
	ErrCronAmbiguous = errors.New("more than one cron job is scheduled on the spec")
)

// RunCron enqueues, once and now, the cron job the server scheduled on spec, the five-field
// expression it was registered with, so the job runs in the server as it would at its next tick
// without waiting for it. It returns once the job is queued, not once it has run, so read the
// job's outcome for that. It returns ErrCronUnreachable for a stack whose redis port is unknown,
// ErrCronNotScheduled when no job is scheduled on spec, ErrCronAmbiguous when several are, and
// redis's error when it cannot read the schedule or queue the job. ctx bounds the enqueue.
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

	switch len(scheduled) {
	case 0:
		return fmt.Errorf("%w: %q", ErrCronNotScheduled, spec)
	case 1:
	default:
		return fmt.Errorf("%w: %q", ErrCronAmbiguous, spec)
	}

	client := asynq.NewClient(redis)
	defer client.Close() //nolint:errcheck // the task is queued once EnqueueContext returns

	_, err = client.EnqueueContext(ctx, asynq.NewTask(scheduled[0].Type(), scheduled[0].Payload()), asynq.Queue(cronQueue))

	return err
}

// RunCron enqueues the cron job the server scheduled on spec, as [Stack.RunCron] does, waiting up
// to 30 seconds for the server's scheduler to report the job in. It fails t on any other error,
// and if the job is still not scheduled when the wait ends.
func (dc *DockerCompose) RunCron(t *testing.T, spec string) {
	t.Helper()

	var err error

	require.Eventually(t, func() bool {
		err = dc.stack.RunCron(t.Context(), spec)

		return !errors.Is(err, ErrCronNotScheduled)
	}, 30*time.Second, 1*time.Second, "no cron job was scheduled on %q", spec)
	require.NoError(t, err)
}
