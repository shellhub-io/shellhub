package asynq_test

import (
	"context"
	"testing"
	"time"

	asynqlib "github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/pkg/worker/asynq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestServer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Start the redis container
	redisContainer, err := redis.Run(ctx, "docker.io/redis:7")
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, redisContainer.Terminate(ctx))
	})

	redisConnStr, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Setup server and handlers
	srv := asynq.NewServer(redisConnStr, asynq.BatchConfig(2, 1, 1))
	defer srv.Shutdown()

	assertTaskPayload := ""
	srv.HandleTask("queue:task", func(_ context.Context, payload []byte) error {
		assertTaskPayload = string(payload)

		return nil
	})

	assertCronPayload := ""
	srv.HandleCron("* * * * *", func(_ context.Context) error {
		assertCronPayload = "cron was called"

		return nil
	})

	require.NoError(t, srv.Start())

	// Setup asynq client and enqueue task
	opt, err := asynqlib.ParseRedisURI(redisConnStr)
	require.NoError(t, err)
	asynqClient := asynqlib.NewClient(opt)
	_, err = asynqClient.Enqueue(asynqlib.NewTask("queue:task", []byte("task was called")), asynqlib.Queue("queue"))
	require.NoError(t, err)

	// Assert that tasks was called. We sleep for 1 minute to wait the server process the cronjob
	time.Sleep(1 * time.Minute)
	require.Equal(t, assertTaskPayload, "task was called")
	require.Equal(t, assertCronPayload, "cron was called")
}
