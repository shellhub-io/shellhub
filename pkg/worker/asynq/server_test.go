package asynq_test

import (
	"context"
	"testing"
	"time"

	asynqlib "github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/pkg/worker/asynq"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	t.Parallel()

	redisConnStr := startValkey(t)

	srv := asynq.NewServer(redisConnStr)

	t.Cleanup(srv.Shutdown)

	taskCalled := make(chan string, 1)
	srv.HandleTask("queue:task", func(_ context.Context, payload []byte) error {
		taskCalled <- string(payload)

		return nil
	})

	cronCalled := make(chan struct{})
	srv.HandleCron("* * * * *", func(_ context.Context) error {
		select {
		case <-cronCalled:
		default:
			close(cronCalled)
		}

		return nil
	})

	require.NoError(t, srv.Start())

	opt, err := asynqlib.ParseRedisURI(redisConnStr)
	require.NoError(t, err)
	asynqClient := asynqlib.NewClient(opt)
	defer asynqClient.Close() //nolint:errcheck
	_, err = asynqClient.Enqueue(asynqlib.NewTask("queue:task", []byte("task was called")), asynqlib.Queue("queue"))
	require.NoError(t, err)

	select {
	case payload := <-taskCalled:
		require.Equal(t, "task was called", payload)
	case <-time.After(10 * time.Second):
		t.Fatal("task was not processed within 10s")
	}

	select {
	case <-cronCalled:
	case <-time.After(2 * time.Minute):
		t.Fatal("cron did not fire within 2 minutes")
	}
}
