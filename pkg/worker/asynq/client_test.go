package asynq_test

import (
	"context"
	"runtime"
	"testing"
	"time"

	asynqlib "github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/pkg/worker/asynq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestClient(t *testing.T) {
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
	addr, err := asynqlib.ParseRedisURI(redisConnStr)
	require.NoError(t, err)

	asynqMux := asynqlib.NewServeMux()
	asynqSrv := asynqlib.NewServer(
		addr,
		asynqlib.Config{ //nolint:exhaustruct
			Concurrency: runtime.NumCPU(),
			Queues:      map[string]int{"queue": 1},
		},
	)

	assertTaskPayload := ""
	asynqMux.HandleFunc("queue:kind", func(_ context.Context, t *asynqlib.Task) error {
		assertTaskPayload = string(t.Payload())

		return nil
	})

	require.NoError(t, asynqSrv.Start(asynqMux))

	// Setup client
	client, err := asynq.NewClient(redisConnStr)
	require.NoError(t, err)
	defer client.Close()

	require.NoError(t, client.Submit(ctx, "queue:kind", []byte("task was called")))
	time.Sleep(10 * time.Second)
	require.Equal(t, "task was called", assertTaskPayload)
}
