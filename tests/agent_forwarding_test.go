package main

import (
	"crypto/rand"
	"crypto/rsa"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const forwardingUser = "forwarded"

// TestAgentSSHAgentForwarding opens a shell as an unprivileged account with agent forwarding
// requested, so the socket's owner tells the session user apart from the root the agent runs as.
// Only a shell session receives SSH_AUTH_SOCK.
func TestAgentSSHAgentForwarding(t *testing.T) {
	t.Parallel()

	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	shellhubAgent, device := startAcceptedAgent(t, t.Context(), compose)
	execInAgent(t, shellhubAgent, "adduser", "-D", forwardingUser)

	forwardedKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	forwardedPublicKey, err := ssh.NewPublicKey(&forwardedKey.PublicKey)
	require.NoError(t, err)

	keyring := agent.NewKeyring()
	require.NoError(t, keyring.Add(agent.AddedKey{PrivateKey: forwardedKey}))

	conn := dialClientWithin(t, t.Context(), compose.SSHAddress(), &ssh.ClientConfig{
		User:            forwardingUser + "@" + ShellHubNamespaceName + "." + device.Name,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(registerDeviceKey(t, t.Context(), compose))},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // the test stack's host key is ephemeral
	}, 30*time.Second)

	defer conn.Close() //nolint:errcheck // the test is over once the shell exited

	require.NoError(t, agent.ForwardToAgent(conn, keyring))

	sess, err := conn.NewSession()
	require.NoError(t, err)

	defer sess.Close() //nolint:errcheck // the shell already exited once stdout reached EOF

	require.NoError(t, agent.RequestAgentForwarding(sess))
	require.NoError(t, sess.RequestPty("xterm", 40, 400, ssh.TerminalModes{ssh.ECHO: 0}))

	stdout, err := sess.StdoutPipe()
	require.NoError(t, err)

	sess.Stdin = strings.NewReader(`echo "OWNER=$(stat -c %U:%G "$SSH_AUTH_SOCK" "${SSH_AUTH_SOCK%/*}" | tr '\n' ' ')"; echo "KEYS=$(ssh-add -L)"; exit` + "\n")

	require.NoError(t, sess.Shell())

	output := make(chan string, 1)

	go func() {
		read, _ := io.ReadAll(stdout)
		output <- string(read)
	}()

	var transcript string

	select {
	case transcript = <-output:
	case <-time.After(60 * time.Second):
		require.FailNow(t, "the shell never exited")
	}

	lines := map[string]string{}

	for line := range strings.Lines(transcript) {
		for _, marker := range []string{"OWNER", "KEYS"} {
			if _, value, ok := strings.Cut(line, marker+"="); ok && !strings.Contains(value, "$(") {
				lines[marker] = strings.TrimSpace(value)
			}
		}
	}

	t.Run("the forwarded agent answers inside the session", func(t *testing.T) {
		assert.Equal(t, strings.TrimSpace(string(ssh.MarshalAuthorizedKey(forwardedPublicKey))), lines["KEYS"], transcript)
	})

	t.Run("the agent socket and its directory belong to the session user", func(t *testing.T) {
		owner := forwardingUser + ":" + forwardingUser

		assert.Equal(t, owner+" "+owner, lines["OWNER"], transcript)
	})
}
