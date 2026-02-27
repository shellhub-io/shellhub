package asynq_test

import (
	"context"
	"testing"
	"time"

	asynqlib "github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/worker/asynq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestServer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	image := "docker.io/redis:7"
	if envs.DefaultBackend.Get("CI") == "true" {
		image = "registry.infra.ossystems.io/cache/redis:7"
	}

	redisContainer, err := redis.Run(ctx, image)
	require.NoError(t, err)

	redisConnStr, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Setup server and handlers
	srv := asynq.NewServer(redisConnStr, asynq.BatchConfig(2, 1, 1))

	// Shutdown the server before terminating Redis to avoid connection refused spam.
	t.Cleanup(func() {
		srv.Shutdown()
		require.NoError(t, redisContainer.Terminate(ctx))
	})

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

	// Setup asynq client and enqueue task
	opt, err := asynqlib.ParseRedisURI(redisConnStr)
	require.NoError(t, err)
	asynqClient := asynqlib.NewClient(opt)
	defer asynqClient.Close()
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
