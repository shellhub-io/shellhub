package main

import (
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const agentUpdateCheckWindow = 5 * time.Second

var agentUpdateApplied = regexp.MustCompile(`msg="Update successfully applied".* next_version=(\S+)`)

var agentUpdateCheckLines = regexp.MustCompile(`msg="(Update successfully applied|Failed to check update|Failed to apply update|Waiting for the next update check)"`)

var agentRetriedRequest = regexp.MustCompile(`msg="(Cannot reach the server, retrying until it answers|Cannot authorize the device, retrying until the server accepts it)"`)

// TestAgentUpdateCheck covers the update check an agent runs as it starts, comparing its own
// version with the one the server reports. The test agents run outside Docker, so applying an
// update replaces nothing and the agent's log is all the check leaves behind. An agent reporting
// latest is a development build and never checks. Connecting does not wait for the check, so both
// cases watch the log for the same window after the agent connects, and a retried request in that
// window fails the latest case, since it could be a check the retries pushed past the window.
func TestAgentUpdateCheck(t *testing.T) {
	t.Parallel()

	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	t.Run("a versioned agent checks the server's version as it starts", func(t *testing.T) {
		agent, err := NewAgentContainerAtVersion(t.Context(), compose.Service(environment.ServiceGateway).GetContainerID(), "0.20.0")
		require.NoError(t, err)
		runAgent(t, t.Context(), agent)

		awaitAgentConnected(t, agent)

		checked := awaitAgentLogWithin(t, agent, agentUpdateApplied, "the versioned agent has not checked the server's version", agentUpdateCheckWindow)

		assert.Equal(t, strings.TrimPrefix(serverVersion(t, compose), "v"), checked)
	})

	t.Run("an agent reporting latest never checks for updates", func(t *testing.T) {
		agent := startAgent(t, t.Context(), compose)
		awaitAgentConnected(t, agent)

		assert.Never(t, func() bool {
			logs, err := agentLogs(t.Context(), agent)

			return err != nil || agentUpdateCheckLines.Match(logs) || agentRetriedRequest.Match(logs)
		}, agentUpdateCheckWindow, time.Second, "the latest agent checked for updates, or retried a request that could have been a check")
	})
}

func serverVersion(t *testing.T, compose *environment.DockerCompose) string {
	t.Helper()

	info := new(models.Info)

	resp, err := compose.Anonymous(t.Context()).SetResult(info).Get("/info")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	require.NotEmpty(t, info.Version)

	return info.Version
}
