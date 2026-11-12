package environment

import (
	"context"
	"testing"
	"time"

	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CrashServer kills the server with SIGKILL and starts it again, failing t when the container
// cannot be killed or started, when it is still running a minute after the kill, and unless the
// API answers again within three minutes. Unlike [DockerCompose.RestartServer], the server gets
// no chance to shut down, so the sessions it was serving stay active until something retires them.
// It outlives t's context, as RestartServer does.
func (dc *DockerCompose) CrashServer(t *testing.T) {
	t.Helper()

	ctx := context.WithoutCancel(t.Context())
	server := dc.Service(ServiceServer)

	_, err := dc.stack.run.client.ContainerKill(ctx, server.GetContainerID(), client.ContainerKillOptions{Signal: "KILL"})
	require.NoError(t, err)

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		state, err := server.State(ctx)
		if assert.NoError(tt, err) {
			assert.False(tt, state.Running, "the server is still running")
		}
	}, time.Minute, 100*time.Millisecond)

	require.NoError(t, server.Start(ctx))
	require.NoError(t, dc.stack.AwaitAPI(ctx))
}
