package main

import (
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"golang.org/x/crypto/ssh"
)

const (
	singleUserPassword = "single-user-password"
	singleUserHash     = "$6$e2esingleuser$.sfZR6UlxyHc3YTk7m.S8yMCSvao0kjGSX.mLOKMwJhoEzws3YUFzBN3PzHwxAQ2ZfvZNzXkEDMnuGvdt6EqL/"
	singleUserAccount  = "nobody"
)

func NewAgentContainerInSingleUserMode() NewAgentContainerOption {
	return func(envs map[string]string) {
		envs["SHELLHUB_SINGLE_USER_PASSWORD"] = singleUserHash
		envs["SHELL"] = "/bin/sh"
	}
}

func asUser(user string) func(*testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.ConfigModifier = func(config *container.Config) {
			config.User = user
		}
	}
}

// TestAgentSingleUserMode runs the agent unprivileged with SHELLHUB_SINGLE_USER_PASSWORD, where it
// checks every login against that hash instead of the system's accounts and runs every session as
// the account it runs as, whatever user the client names.
func TestAgentSingleUserMode(t *testing.T) {
	t.Parallel()

	compose := newSSHEnvironment(t, t.Context(), legacyNamespace)

	t.Run("an agent running as root refuses to start", func(t *testing.T) {
		agent := startAgent(t, t.Context(), compose, NewAgentContainerInSingleUserMode())

		var state *container.State

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			var err error

			state, err = agent.State(t.Context())
			if assert.NoError(tt, err) {
				assert.False(tt, state.Running, "the agent is still running")
			}
		}, 30*time.Second, time.Second)

		assert.Equal(t, 1, state.ExitCode)

		logs, err := agentLogs(t.Context(), agent)
		require.NoError(t, err)
		assert.Contains(t, string(logs), "cannot run as root when single-user mode is enabled")

		assert.Empty(t, compose.ListDevices(t, models.DeviceStatusEmpty), "the refused agent enrolled a device")
	})

	startCustomAgent(t, t.Context(), compose, asUser(singleUserAccount), NewAgentContainerInSingleUserMode())

	device := acceptPendingDevice(t, compose)

	passwordConfig := func(password string) *ssh.ClientConfig {
		return &ssh.ClientConfig{
			User:            deviceSSHID(device),
			Auth:            []ssh.AuthMethod{ssh.Password(password)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // the test stack's host key is ephemeral
		}
	}

	t.Run("the configured password logs root in as the agent's own account", func(t *testing.T) {
		conn := dialClientWithin(t, t.Context(), compose.SSHAddress(), passwordConfig(singleUserPassword), 30*time.Second)
		defer conn.Close() //nolint:errcheck // the test is over once the command answered

		assert.Equal(t, singleUserAccount, runOnDevice(t, conn, "echo -n $(id -un)"))
		assert.Equal(t, singleUserAccount, runOnDevice(t, conn, "echo -n $USER"))
	})

	t.Run("a password other than the configured one is refused", func(t *testing.T) {
		err := handshakeWith(t.Context(), compose.SSHAddress(), passwordConfig(ShellHubAgentPassword))
		require.ErrorContains(t, err, "ssh: unable to authenticate")
	})
}
