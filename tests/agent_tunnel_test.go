package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	agentTunnelPath        = "/agent/connection"
	agentRedialInterval    = 10 * time.Second
	redialGapFloor         = agentRedialInterval - time.Second
	redialGapCeiling       = agentRedialInterval + 5*time.Second
	acceptedTunnelDeadline = agentRedialInterval + 2*time.Second
)

// TestAgentTunnelRedial covers how often an agent dials its tunnel again once the server refuses it.
// The server refuses the tunnel of a device that is not accepted when it requires accepted tunnels,
// while still authorizing the device, so the agent sees its tunnel dial fail and nothing else. The
// cases read the dials from the server's request log. Each gap between refused dials must fall
// between redialGapFloor and redialGapCeiling, and the tunnel must be up within
// acceptedTunnelDeadline of the accept.
func TestAgentTunnelRedial(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := newConfiguredSSHEnvironment(t, ctx,
		environment.New(t, run).WithEnv("SHELLHUB_REQUIRE_ACCEPTED_TUNNEL", "true"),
		models.SSHAccessModeLegacy)

	mark := compose.ServerLogMark(t)

	startAgent(t, ctx, compose)
	device := compose.AwaitDeviceWithStatus(t, models.DeviceStatusPending)

	t.Run("a refused tunnel is dialled again every ten seconds", func(t *testing.T) {
		var refused []time.Time

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			refused = tunnelDials(tt, ctx, compose, mark, "403")
			assert.GreaterOrEqual(tt, len(refused), 4, "the agent stopped dialling its refused tunnel")
		}, 90*time.Second, time.Second)

		for i := 1; i < len(refused); i++ {
			gap := refused[i].Sub(refused[i-1])

			assert.GreaterOrEqual(t, gap, redialGapFloor, "dial %d came %s after the one before it", i, gap)
			assert.LessOrEqual(t, gap, redialGapCeiling, "dial %d came %s after the one before it", i, gap)
		}
	})

	t.Run("the tunnel is up within one interval of the device being accepted", func(t *testing.T) {
		accepting := compose.ServerLogMark(t)

		compose.UpdateDeviceStatus(t, device.UID, environment.DeviceActionAccept)

		accepted := logTime(t, compose.AwaitServerLogLine(t, accepting, "uri=/api/devices/"+device.UID+"/accept "))

		var established []time.Time

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			established = tunnelDials(tt, ctx, compose, accepting, "200")
			assert.NotEmpty(tt, established, "the agent never dialled its tunnel again once accepted")
		}, 30*time.Second, time.Second)

		assert.LessOrEqual(t, established[0].Sub(accepted), acceptedTunnelDeadline)
	})
}

func tunnelDials(tt *assert.CollectT, ctx context.Context, compose *environment.DockerCompose, mark int, status string) []time.Time {
	logs, err := compose.ReadServerLogSince(ctx, mark)
	require.NoError(tt, err)

	var dials []time.Time

	for line := range strings.Lines(logs) {
		fields := logFields(line)
		if fields["uri"] != agentTunnelPath || fields["status"] != status {
			continue
		}

		at, err := time.Parse(time.RFC3339, fields["time"])
		if assert.NoError(tt, err, "the server logged a dial without a readable time: %s", line) {
			dials = append(dials, at)
		}
	}

	return dials
}

func logTime(t *testing.T, line string) time.Time {
	t.Helper()

	at, err := time.Parse(time.RFC3339, logFields(line)["time"])
	require.NoError(t, err, "the server logged a line without a readable time: %s", line)

	return at
}
