package main

import (
	"regexp"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

// TestAgentCLI covers the agent's commands a person runs by hand next to the daemon. The login
// command prints a code a namespace member accepts the device with from the console, and returns
// once the server reports the device accepted. The cases share one stack.
func TestAgentCLI(t *testing.T) {
	t.Parallel()

	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	t.Run("login prints a code that accepts a pending device and returns once it is accepted", func(t *testing.T) {
		login := startAgentCommand(t, compose, []string{"./agent", "login"}, NewAgentContainerWithHostname("login-pending"))
		code := awaitAgentPairingCode(t, login)

		preview := requireLoginCodePreview(t, compose, code)
		require.Equal(t, "login-pending", preview.Name)
		require.Equal(t, models.DeviceStatusPending, preview.Status)

		compose.UpdateDeviceStatus(t, preview.UID, environment.DeviceActionAccept)

		assert.Zero(t, awaitAgentExit(t, login))
		awaitAgentLog(t, login, regexp.MustCompile(`(✓ Device accepted into namespace "`+ShellHubNamespaceName+`"\.)`), "login did not report the acceptance")
	})
}

func startAgentCommand(t *testing.T, compose *environment.DockerCompose, command []string, opts ...NewAgentContainerOption) testcontainers.Container {
	t.Helper()

	return startCustomAgent(t, t.Context(), compose, func(req *testcontainers.ContainerRequest) {
		req.Entrypoint = command
	}, opts...)
}

func awaitAgentExit(t *testing.T, agent testcontainers.Container) int {
	t.Helper()

	var exitCode int

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		state, err := agent.State(t.Context())
		if assert.NoError(tt, err) && assert.False(tt, state.Running, "the command is still running") {
			exitCode = state.ExitCode
		}
	}, 30*time.Second, time.Second)

	return exitCode
}
