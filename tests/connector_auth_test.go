package main

import (
	"context"
	"io"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"golang.org/x/crypto/ssh"
)

// TestConnectorPasswordAuth covers the password logins a connector refuses for an account in the
// container it serves, each read from the container's own /etc/passwd and /etc/shadow.
func TestConnectorPasswordAuth(t *testing.T) {
	t.Parallel()

	skipUnlessAgentAcceptsPasswords(t)

	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, legacyNamespace)

	name := "auth-" + randomHex(t)
	target := startLabelledContainer(t, ctx, name)
	execInContainer(t, ctx, target, "adduser -D nopasswd && sed -i 's/^nopasswd:x:/nopasswd::/' /etc/passwd")
	execInContainer(t, ctx, target, "adduser -D noshadow && passwd -d noshadow")
	connector := startConnector(t, ctx, compose, connectorLabel+"="+name)

	device := awaitContainerWithStatus(t, compose, models.DeviceStatusPending)
	compose.UpdateDeviceStatus(t, device.UID, environment.DeviceActionAccept)
	compose.AwaitDeviceOnline(t, device.UID)

	t.Run("a user whose /etc/passwd entry holds no password cannot log in with an empty password", func(t *testing.T) {
		err := connectorPasswordHandshake(ctx, compose, &device, "nopasswd", "")
		require.ErrorContains(t, err, "ssh: unable to authenticate, attempted methods [none password]")

		environment.AwaitLogContains(t, connector, "user passwd is empty, so the authentication via password is blocked")
	})

	t.Run("a user whose /etc/shadow entry holds no password cannot log in with an empty password", func(t *testing.T) {
		err := connectorPasswordHandshake(ctx, compose, &device, "noshadow", "")
		require.ErrorContains(t, err, "ssh: unable to authenticate, attempted methods [none password]")

		environment.AwaitLogContains(t, connector, "User cannot login with empty password")
	})
}

func connectorPasswordHandshake(ctx context.Context, compose *environment.DockerCompose, device *models.Device, user, password string) error {
	return handshake(ctx, compose.SSHAddress(), user+"@"+ShellHubNamespaceName+"."+device.Name, []ssh.AuthMethod{ssh.Password(password)})
}

func execInContainer(t *testing.T, ctx context.Context, c testcontainers.Container, script string) {
	t.Helper()

	code, reader, err := c.Exec(ctx, []string{"sh", "-c", script}, tcexec.Multiplexed())
	require.NoError(t, err)

	output, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Zero(t, code, string(output))
}
