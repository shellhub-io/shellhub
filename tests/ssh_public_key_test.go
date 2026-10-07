package main

import (
	"context"
	"regexp"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

const (
	errPublicKeyNotEvaluated = "failed to evaluate the provided public key"
	errPublicKeyAgentTooOld  = "connections using public keys are not permitted when the agent version is 0.5.x or earlier"
)

// TestSSHPublicKeyRules covers the rules a public key carries in the legacy model: the devices it
// reaches, by name or by tag, and the device users it logs in as. The same key reaches the device
// its rule names and is refused, with the reason in the server's log, by the device or user it
// does not.
func TestSSHPublicKeyRules(t *testing.T) {
	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)
	_, named := startAcceptedAgent(t, ctx, compose)
	_, other := startAcceptedAgent(t, ctx, compose)

	resp, err := compose.R(ctx).SetBody(map[string]string{"name": "filtered"}).Post("/api/tags")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())

	t.Run("a hostname filter lets the key into the device it matches", func(t *testing.T) {
		signer := registerPublicKey(t, compose, ".*", requests.PublicKeyFilter{Hostname: regexp.QuoteMeta(named.Name)})

		requireKeyLogsIn(t, compose, deviceSSHID(named), signer)
	})

	t.Run("a hostname filter refuses the key on a device it does not match", func(t *testing.T) {
		signer := registerPublicKey(t, compose, ".*", requests.PublicKeyFilter{Hostname: regexp.QuoteMeta(named.Name)})

		requireKeyRefused(t, compose, deviceSSHID(other), signer, errPublicKeyNotEvaluated)
	})

	t.Run("a tag filter lets the key into a device carrying the tag", func(t *testing.T) {
		tagDevice(t, compose, named.UID, "filtered")
		signer := registerPublicKey(t, compose, ".*", requests.PublicKeyFilter{Tags: []string{"filtered"}})

		requireKeyLogsIn(t, compose, deviceSSHID(named), signer)
	})

	t.Run("a tag filter refuses the key on a device without the tag", func(t *testing.T) {
		tagDevice(t, compose, named.UID, "filtered")
		signer := registerPublicKey(t, compose, ".*", requests.PublicKeyFilter{Tags: []string{"filtered"}})

		requireKeyRefused(t, compose, deviceSSHID(other), signer, errPublicKeyNotEvaluated)
	})

	t.Run("a username restriction lets the key log in as the user it matches", func(t *testing.T) {
		signer := registerPublicKey(t, compose, ShellHubAgentUsername, requests.PublicKeyFilter{Hostname: ".*"})

		requireKeyLogsIn(t, compose, deviceSSHID(named), signer)
	})

	t.Run("a username restriction refuses the key as a user it does not match", func(t *testing.T) {
		signer := registerPublicKey(t, compose, ShellHubAgentUsername, requests.PublicKeyFilter{Hostname: ".*"})

		requireKeyRefused(t, compose, "daemon@"+ShellHubNamespaceName+"."+named.Name, signer, errPublicKeyNotEvaluated)
	})
}

// TestSSHPublicKeyRefusedBelowAgent060 covers the gate that keeps public keys away from agents
// older than 0.6.0, which do not check the key the server asks them to accept. A password still
// logs in to such an agent.
func TestSSHPublicKeyRefusedBelowAgent060(t *testing.T) {
	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)
	signer := registerDeviceKey(t, ctx, compose)

	agent, err := NewAgentContainerAtVersion(ctx, compose.Service(environment.ServiceGateway).GetContainerID(), "0.5.2")
	require.NoError(t, err)

	t.Cleanup(func() { _ = agent.Terminate(context.Background()) })

	require.NoError(t, agent.Start(ctx))

	pending := compose.AwaitDeviceWithStatus(t, models.DeviceStatusPending)
	compose.UpdateDeviceStatus(t, pending.UID, environment.DeviceActionAccept)
	device := compose.AwaitDeviceOnline(t, pending.UID)

	require.NotNil(t, device.Info)
	require.Equal(t, "0.5.2", device.Info.Version, "the agent under test must report the version it was built with")

	requireKeyRefused(t, compose, deviceSSHID(&device), signer, errPublicKeyAgentTooOld)

	t.Run("a password still logs in", func(t *testing.T) {
		skipUnlessAgentAcceptsPasswords(t)

		conn := dialClient(t, t.Context(), compose.SSHAddress(), &ssh.ClientConfig{
			User:            deviceSSHID(&device),
			Auth:            []ssh.AuthMethod{ssh.Password(ShellHubAgentPassword)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // the test stack's host key is ephemeral
		})
		require.NoError(t, conn.Close())
	})
}

func registerPublicKey(t *testing.T, compose *environment.DockerCompose, username string, filter requests.PublicKeyFilter) ssh.Signer {
	t.Helper()

	signer, data := newSigner(t)

	resp, err := compose.R(t.Context()).
		SetBody(&requests.PublicKeyCreate{
			Name:     t.Name(),
			Username: username,
			Data:     []byte(data),
			Filter:   filter,
		}).
		Post("/api/sshkeys/public-keys")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())

	fingerprint := ssh.FingerprintLegacyMD5(signer.PublicKey())
	t.Cleanup(func() {
		resp, err := compose.R(context.WithoutCancel(t.Context())).Delete("/api/sshkeys/public-keys/" + fingerprint)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode(), resp.String())
	})

	return signer
}

func requireKeyLogsIn(t *testing.T, compose *environment.DockerCompose, sshid string, signer ssh.Signer) {
	t.Helper()

	conn := dialClient(t, t.Context(), compose.SSHAddress(), &ssh.ClientConfig{
		User:            sshid,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // the test stack's host key is ephemeral
	})
	defer conn.Close() //nolint:errcheck // the test is over once the command answered

	assert.Equal(t, "in", runOnDevice(t, conn, "echo -n in"))
}

func requireKeyRefused(t *testing.T, compose *environment.DockerCompose, sshid string, signer ssh.Signer, reason string) {
	t.Helper()

	mark := compose.ServerLogMark(t)

	err := handshakeWith(t.Context(), compose.SSHAddress(), &ssh.ClientConfig{
		User:            sshid,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // the test stack's host key is ephemeral
	})
	require.ErrorContains(t, err, "ssh: unable to authenticate, attempted methods [none publickey]")

	compose.AwaitServerLogLine(t, mark, "refused the offered public key", sshid, reason)
}
