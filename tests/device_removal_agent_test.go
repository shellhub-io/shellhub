package main

import (
	"bytes"
	"regexp"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

var agentRemovedRepairingLine = regexp.MustCompile(`msg="(The device was removed from its namespace; pairing it again)"`)

const agentPairedTenantFile = "/tmp/shellhub.key.tenant"

// TestRemovedPairedDeviceAgent runs a real agent whose device was enrolled by pairing and is then
// removed from its namespace. The server closes the device's tunnel and refuses its next
// authentication. An agent that learned its tenant from the pairing forgets it and pairs again; one
// given its tenant in the environment cannot forget it, so it exits. The cases share one stack, so
// every agent carries a hostname no other case uses.
func TestRemovedPairedDeviceAgent(t *testing.T) {
	t.Parallel()

	compose := newSSHEnvironment(t, t.Context(), newNamespace)

	t.Run("an agent that learned its tenant from pairing pairs again", func(t *testing.T) {
		agent := startAgent(t, t.Context(), compose, NewAgentContainerUnpaired(), NewAgentContainerWithHostname("removed-repairs"))

		first := acceptPairing(t, compose.R(t.Context()), awaitAgentPairingCode(t, agent))
		awaitAgentConnected(t, agent)
		compose.AwaitDeviceOnline(t, first.UID)

		compose.DeleteDevice(t, first.UID)

		compose.AwaitDeviceOffline(t, first.UID)
		awaitAgentLog(t, agent, agentRemovedRepairingLine, "the agent did not go back to pairing")
		execInAgent(t, agent, "test", "!", "-e", agentPairedTenantFile)

		second := acceptPairing(t, compose.R(t.Context()), awaitNthAgentLog(t, agent, agentPairingLink, 2, "the agent has not printed a second pairing code"))
		assert.Equal(t, ShellHubNamespace, awaitNthAgentLog(t, agent, agentConnectedLine, 2, "the agent has not connected again"))

		device := compose.AwaitDeviceOnline(t, second.UID)
		assert.Equal(t, models.DeviceStatusAccepted, device.Status)
		assert.Equal(t, ShellHubNamespace, readAgentFile(t, agent, agentPairedTenantFile), "the agent did not keep the tenant it paired with again")

		logs, err := agentLogs(t.Context(), agent)
		require.NoError(t, err)
		assert.Equal(t, 1, bytes.Count(logs, []byte(`msg="Starting ShellHub"`)), "the agent restarted to pair again")
	})

	t.Run("an agent given its tenant in the environment exits", func(t *testing.T) {
		key := newAgentPrivateKey(t)
		identity := randomHex(t)

		withKey := withAgentPrivateKey(key)
		sameDevice := []NewAgentContainerOption{NewAgentContainerWithIdentity(identity), NewAgentContainerWithHostname("removed-exits")}

		pairing := startCustomAgent(t, t.Context(), compose, withKey, append(sameDevice, NewAgentContainerUnpaired())...)
		paired := acceptPairing(t, compose.R(t.Context()), awaitAgentPairingCode(t, pairing))
		awaitAgentConnected(t, pairing)
		compose.AwaitDeviceOnline(t, paired.UID)

		require.NoError(t, pairing.Stop(t.Context(), nil))
		compose.AwaitDeviceOffline(t, paired.UID)

		configured := startCustomAgent(t, t.Context(), compose, withKey, append(sameDevice, NewAgentContainerWithTenantOnly(ShellHubNamespace))...)
		require.Equal(t, ShellHubNamespace, awaitAgentConnected(t, configured))
		compose.AwaitDeviceOnline(t, paired.UID)

		compose.DeleteDevice(t, paired.UID)

		assert.NotZero(t, awaitAgentExit(t, configured), "the agent kept running for a device it can no longer serve")
		compose.AwaitDeviceOffline(t, paired.UID)

		logs, err := agentLogs(t.Context(), configured)
		require.NoError(t, err)
		assert.Contains(t, string(logs), "Pair it again without SHELLHUB_TENANT_ID")
		assert.NotRegexp(t, agentPairingLink, string(logs), "an agent given its tenant asked for a pairing code")
	})
}

func awaitNthAgentLog(t *testing.T, agent testcontainers.Container, line *regexp.Regexp, n int, missing string) string {
	t.Helper()

	return awaitNthAgentLogWithin(t, agent, line, n, missing, 30*time.Second)
}
