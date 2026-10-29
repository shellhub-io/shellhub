package environment

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	redisPortEnv = "SHELLHUB_TEST_REDIS_PORT"
	cronQueue    = "cron"
)

var (
	// ErrCronUnreachable is returned by [Stack.RunCron], and fails the test in
	// [DockerCompose.RunCron], [DockerCompose.PauseCron] and [DockerCompose.AwaitCronIdle], for a
	// stack brought up without [DockerComposeConfigurator.WithCronTrigger], whose redis sits on a
	// port the test does not know.
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
	redis, err := s.redisOpt()
	if err != nil {
		return err
	}

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

// PauseCron stops the server from starting any cron job, the active-session reaper included, not
// only the jobs of one spec. It returns once no cron job is still running, waiting up to 30
// seconds for one that started before the pause. A stack whose cron queue does not exist yet has
// no job running. While paused, the scheduler and [DockerCompose.RunCron] keep queuing jobs, which
// run once the queue is unpaused. The returned function unpauses the queue; t's cleanup also
// unpauses it, and unpausing again does nothing. It fails t when the stack's redis port is
// unknown, as [Stack.RunCron] does, when the queue is already paused, when a job is still running
// once the wait ends, when redis cannot pause or unpause the queue, and when it cannot read the
// queue for the whole wait. Its redis calls are not bound to t's context.
func (dc *DockerCompose) PauseCron(t *testing.T) func() {
	t.Helper()

	inspector := dc.cronInspector(t)

	require.NoError(t, inspector.PauseQueue(cronQueue))

	var once sync.Once

	unpause := func() {
		once.Do(func() {
			require.NoError(t, inspector.UnpauseQueue(cronQueue))
		})
	}
	t.Cleanup(unpause)

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		active, err := cronJobs(inspector.ListActiveTasks)
		require.NoError(tt, err)
		assert.Empty(tt, active, "a cron job that started before the pause is still running")
	}, 30*time.Second, 500*time.Millisecond)

	return unpause
}

// AwaitCronIdle waits up to 30 seconds until the cron queue has no job waiting to run and none
// running, so every cron job queued before the call has run once. A job that failed and waits for
// a retry does not count. A job the scheduler queues during the wait delays it, and on a queue
// [DockerCompose.PauseCron] left paused it always fails t. A stack whose cron queue does not
// exist yet is idle. It fails t when the stack's redis port is unknown, as [Stack.RunCron] does,
// when a job is still waiting or running once the wait ends, and when redis cannot read the queue
// for the whole wait. Its redis calls are not bound to t's context.
func (dc *DockerCompose) AwaitCronIdle(t *testing.T) {
	t.Helper()

	inspector := dc.cronInspector(t)

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		pending, err := cronJobs(inspector.ListPendingTasks)
		require.NoError(tt, err)

		active, err := cronJobs(inspector.ListActiveTasks)
		require.NoError(tt, err)

		assert.Empty(tt, pending, "a cron job is still waiting to run")
		assert.Empty(tt, active, "a cron job is still running")
	}, 30*time.Second, 500*time.Millisecond)
}

func (dc *DockerCompose) cronInspector(t *testing.T) *asynq.Inspector {
	t.Helper()

	redis, err := dc.stack.redisOpt()
	require.NoError(t, err)

	inspector := asynq.NewInspector(redis)
	t.Cleanup(func() {
		inspector.Close() //nolint:errcheck // the test is done with redis and a failed close leaks only a connection
	})

	return inspector
}

func cronJobs(list func(queue string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error)) ([]*asynq.TaskInfo, error) {
	jobs, err := list(cronQueue)
	if errors.Is(err, asynq.ErrQueueNotFound) {
		return nil, nil
	}

	return jobs, err
}

func (s *Stack) redisOpt() (asynq.RedisClientOpt, error) {
	port := s.envs[redisPortEnv]
	if port == "" {
		return asynq.RedisClientOpt{}, ErrCronUnreachable //nolint:exhaustruct // callers discard the options on error
	}

	return asynq.RedisClientOpt{Addr: "localhost:" + port}, nil //nolint:exhaustruct // the remaining fields keep asynq's defaults
}
