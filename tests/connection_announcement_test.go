package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// TestConnectionAnnouncement covers the namespace's connection announcement, the text the gateway
// writes to an SSH client before the device's shell. A namespace starts with its edition's, the
// owner replaces it through the namespace settings, and a command run without a terminal gets
// only its own output.
func TestConnectionAnnouncement(t *testing.T) {
	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)
	signer := registerDeviceKey(t, ctx, compose)
	_, device := startAcceptedAgent(t, ctx, compose)

	connected := "Connected to " + deviceSSHID(device) + " via ShellHub.\n"

	t.Run("a namespace starts with its edition's announcement", func(t *testing.T) {
		greeting := shellGreeting(t, ctx, compose, device, signer)

		assert.Contains(t, greeting, connected)
		assert.Contains(t, greeting, "Welcome to ShellHub Community!")
	})

	t.Run("the client sees the announcement the namespace sets", func(t *testing.T) {
		announcement := "Maintenance tonight at 22:00.\nSave your work."

		resp, err := compose.R(t.Context()).
			SetBody(map[string]any{"settings": map[string]string{"connection_announcement": announcement}}).
			Put("/api/namespaces/" + ShellHubNamespace)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		greeting := shellGreeting(t, ctx, compose, device, signer)

		assert.Contains(t, greeting, connected+announcement+"\n")
		assert.NotContains(t, greeting, "Welcome to ShellHub")
	})

	t.Run("a command without a terminal gets no announcement", func(t *testing.T) {
		conn := dialDevice(t, ctx, compose, device, signer)
		defer conn.Close() //nolint:errcheck // the test is over once the command answered

		assert.Equal(t, "in", runOnDevice(t, conn, "echo -n in"))
	})
}

func shellGreeting(t *testing.T, ctx context.Context, compose *environment.DockerCompose, device *models.Device, signer ssh.Signer) string {
	t.Helper()

	conn := dialDevice(t, ctx, compose, device, signer)
	defer conn.Close() //nolint:errcheck // the shell has exited, or the deadline closed the connection

	deadline := time.AfterFunc(30*time.Second, func() { _ = conn.Close() })
	defer deadline.Stop()

	sess, err := conn.NewSession()
	require.NoError(t, err)

	stdout := new(bytes.Buffer)
	sess.Stdout = stdout

	stdin, err := sess.StdinPipe()
	require.NoError(t, err)

	require.NoError(t, sess.RequestPty("xterm", 24, 80, ssh.TerminalModes{}))
	require.NoError(t, sess.Shell())

	_, err = io.WriteString(stdin, "exit\n")
	require.NoError(t, err)

	require.NoError(t, sess.Wait(), "the shell did not exit within 30 seconds")

	return strings.ReplaceAll(stdout.String(), "\r", "")
}
