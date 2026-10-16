package main

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"golang.org/x/crypto/ssh"
)

const (
	connectorShadowUser     = "shadowed"
	connectorShadowPassword = "shadowed-password"
)

// TestConnectorSSH logs in through a connector to the container it serves. The connector checks a
// password against the container's own /etc/shadow, so the password cases log in as a user that
// exists only in that container, with the password set there.
func TestConnectorSSH(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, legacyNamespace)

	name := "ssh-" + randomHex(t)
	target := startLabelledContainer(t, ctx, name)
	execInConnectorTarget(t, ctx, target, "adduser -D "+connectorShadowUser+" && echo "+connectorShadowUser+":"+connectorShadowPassword+" | chpasswd")
	startConnector(t, ctx, compose, connectorLabel+"="+name)

	device := awaitContainerWithStatus(t, compose, models.DeviceStatusPending)
	compose.UpdateDeviceStatus(t, device.UID, environment.DeviceActionAccept)
	compose.AwaitDeviceOnline(t, device.UID)

	t.Run("a password set in the container's /etc/shadow logs in", func(t *testing.T) {
		skipUnlessAgentAcceptsPasswords(t)

		conn := dialClient(t, ctx, compose.SSHAddress(), connectorShadowLogin(&device, connectorShadowPassword))
		defer conn.Close() //nolint:errcheck // the test is over once the command answered

		assert.Equal(t, connectorShadowUser, strings.TrimSpace(runOnDevice(t, conn, "id -un")))
	})

	t.Run("a password the container's /etc/shadow does not hold is refused", func(t *testing.T) {
		skipUnlessAgentAcceptsPasswords(t)

		err := handshakeWith(ctx, compose.SSHAddress(), connectorShadowLogin(&device, "wrong-password"))
		require.ErrorContains(t, err, "ssh: unable to authenticate, attempted methods [none password]")
	})

	t.Run("a registered public key logs in", func(t *testing.T) {
		conn := dialDevice(t, ctx, compose, &device, registerDeviceKey(t, ctx, compose))
		defer conn.Close() //nolint:errcheck // the test is over once the command answered

		assert.Equal(t, ShellHubAgentUsername, strings.TrimSpace(runOnDevice(t, conn, "id -un")))
	})
}

func connectorShadowLogin(device *models.Device, password string) *ssh.ClientConfig {
	return &ssh.ClientConfig{ //nolint:exhaustruct // the remaining fields keep their defaults
		User:            connectorShadowUser + "@" + ShellHubNamespaceName + "." + device.Name,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // the test stack's host key is ephemeral
	}
}

func execInConnectorTarget(t *testing.T, ctx context.Context, c testcontainers.Container, script string) {
	t.Helper()

	code, reader, err := c.Exec(ctx, []string{"sh", "-c", script}, tcexec.Multiplexed())
	require.NoError(t, err)

	output, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Zero(t, code, string(output))
}
