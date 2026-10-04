package asynq_test

import (
	"context"
	"runtime"
	"testing"
	"time"

	asynqlib "github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/pkg/testport"
	"github.com/shellhub-io/shellhub/pkg/worker/asynq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

func TestClient(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	redisConnStr := startValkey(t)

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

	handled := make(chan string, 1)
	asynqMux.HandleFunc("queue:kind", func(_ context.Context, t *asynqlib.Task) error {
		handled <- string(t.Payload())

		return nil
	})

	require.NoError(t, asynqSrv.Start(asynqMux))

	client, err := asynq.NewClient(redisConnStr)
	require.NoError(t, err)
	defer client.Close() //nolint:errcheck

	require.NoError(t, client.Submit(ctx, "queue:kind", []byte("task was called")))

	select {
	case payload := <-handled:
		require.Equal(t, "task was called", payload)
	case <-time.After(30 * time.Second):
		t.Fatal("the task was never handled")
	}
}

func startValkey(t *testing.T) string {
	t.Helper()

	valkey, err := redis.Run(t.Context(), "docker.io/valkey/valkey:9.1-alpine", testport.Bind("6379/tcp"))
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, valkey.Terminate(context.WithoutCancel(t.Context())))
	})

	uri, err := valkey.ConnectionString(t.Context())
	require.NoError(t, err)

	return uri
}
