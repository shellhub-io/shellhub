package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pkg/sftp"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// TestConnectorSessions covers the SSH sessions a connector serves inside the container it
// connects, and the transport it reaches the server over.
func TestConnectorSessions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

	name := "sessions-" + randomHex(t)
	target := startLabelledContainer(t, ctx, name)
	execInContainer(t, ctx, target, "echo "+name+" > /tmp/connector-target")
	connector := startConnectorWithEnv(t, ctx, compose, connectorLabel+"="+name, map[string]string{"SHELLHUB_TRANSPORT_VERSION": "1"})

	device := awaitContainerWithStatus(t, compose, models.DeviceStatusPending)
	compose.UpdateDeviceStatus(t, device.UID, environment.DeviceActionAccept)
	compose.AwaitDeviceOnline(t, device.UID)

	signer := registerDeviceKey(t, ctx, compose)

	t.Run("an exec session runs the command in the container and returns its output and exit status", func(t *testing.T) {
		conn := dialDevice(t, ctx, compose, &device, signer)
		defer conn.Close() //nolint:errcheck // the test is over once the command exited

		sess, err := conn.NewSession()
		require.NoError(t, err)

		defer sess.Close() //nolint:errcheck // Run already reaped the command

		var stdout, stderr bytes.Buffer

		sess.Stdout = &stdout
		sess.Stderr = &stderr

		err = sess.Run("sh -c 'cat /tmp/connector-target; echo on-stderr >&2; exit 3'")

		var exit *ssh.ExitError
		require.ErrorAs(t, err, &exit)
		assert.Equal(t, 3, exit.ExitStatus())
		assert.Equal(t, name, strings.TrimSpace(stdout.String()))
		assert.Equal(t, "on-stderr", strings.TrimSpace(stderr.String()))
	})

	t.Run("a heredoc session runs the script it reads from stdin in the container", func(t *testing.T) {
		conn := dialDevice(t, ctx, compose, &device, signer)
		defer conn.Close() //nolint:errcheck // the test is over once the shell exited

		sess, err := conn.NewSession()
		require.NoError(t, err)

		defer sess.Close() //nolint:errcheck // Wait already reaped the shell

		var stdout bytes.Buffer

		sess.Stdin = strings.NewReader("target=$(cat /tmp/connector-target)\necho \"heredoc in $target\"\nexit 4\n")
		sess.Stdout = &stdout

		require.NoError(t, sess.Shell())

		err = sess.Wait()

		var exit *ssh.ExitError
		require.ErrorAs(t, err, &exit)
		assert.Equal(t, 4, exit.ExitStatus())
		assert.Equal(t, "heredoc in "+name, strings.TrimSpace(stdout.String()))
	})

	t.Run("an SFTP session is refused", func(t *testing.T) {
		conn := dialDevice(t, ctx, compose, &device, signer)
		defer conn.Close() //nolint:errcheck // the test is over once the client gave up

		done := make(chan error, 1)

		go func() {
			client, err := sftp.NewClient(conn)
			if err == nil {
				_ = client.Close()
			}

			done <- err
		}()

		select {
		case err := <-done:
			require.ErrorContains(t, err, "server unexpectedly closed connection")
		case <-time.After(30 * time.Second):
			require.Fail(t, "the SFTP client neither started nor failed")
		}
	})

	t.Run("the connector tunnels over the V2 transport even when told to use V1", func(t *testing.T) {
		environment.AwaitLogLine(t, connector, "Server connection established", "transport_version=2")
		environment.AwaitLogLine(t, compose.Service(environment.ServiceServer), "v2 connection established", device.UID)
	})
}
