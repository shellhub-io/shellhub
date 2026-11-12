package main

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServerShutdown stops the server the way Docker does, with SIGTERM, and reads how it went
// from the container: the exit status, and the line the server logs once every shutdown step,
// the heartbeat drain and the telemetry flush included, has run.
func TestServerShutdown(t *testing.T) {
	t.Parallel()

	compose := environment.New(t, run).Up(t.Context())
	t.Cleanup(compose.Down)

	t.Run("SIGTERM finishes the shutdown and exits with status zero", func(t *testing.T) {
		server := compose.Service(environment.ServiceServer)

		grace := 30 * time.Second
		require.NoError(t, server.Stop(context.WithoutCancel(t.Context()), &grace))

		state, err := server.State(t.Context())
		require.NoError(t, err)

		assert.Equal(t, 0, state.ExitCode)
		environment.AwaitLogContains(t, server, "Server shutdown complete")
	})
}
