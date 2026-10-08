package main

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

const sessionReaperCron = "* * * * *"

// TestSSHConnectionBehaviour covers what a legacy-mode client sees of the connection itself rather
// than of the session it runs: the cipher it negotiates, the banner that explains a refusal, and
// how the session ends when either side ends it.
func TestSSHConnectionBehaviour(t *testing.T) {
	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)
	signer := registerDeviceKey(t, ctx, compose)
	_, device := startAcceptedAgent(t, ctx, compose)

	t.Run("each cipher the server offers is the one negotiated", func(t *testing.T) {
		for _, cipher := range []string{
			"aes128-gcm@openssh.com",
			"aes256-gcm@openssh.com",
			"chacha20-poly1305@openssh.com",
			"aes128-ctr",
			"aes192-ctr",
			"aes256-ctr",
		} {
			t.Run(cipher, func(t *testing.T) {
				conn := dialClient(t, t.Context(), compose.SSHAddress(), &ssh.ClientConfig{
					User:            deviceSSHID(device),
					Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
					HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // the test stack's host key is ephemeral
					Config:          ssh.Config{Ciphers: []string{cipher}},
				})
				defer conn.Close() //nolint:errcheck // the test is over once the command answered

				negotiated := conn.Conn.(ssh.AlgorithmsConnMetadata).Algorithms() //nolint:forcetypeassert // every x/crypto client connection carries its algorithms
				assert.Equal(t, cipher, negotiated.Read.Cipher)
				assert.Equal(t, cipher, negotiated.Write.Cipher)

				assert.Equal(t, cipher, runOnDevice(t, conn, "echo -n "+cipher))
			})
		}
	})

	t.Run("a cipher the server does not offer fails the handshake", func(t *testing.T) {
		err := handshakeWith(t.Context(), compose.SSHAddress(), &ssh.ClientConfig{
			User:            deviceSSHID(device),
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // the test stack's host key is ephemeral
			Config:          ssh.Config{Ciphers: []string{"aes128-cbc"}},
		})
		require.ErrorContains(t, err, "ssh: no common algorithm for client to server cipher")
	})

	t.Run("a banner tells the client its SSHID is malformed", func(t *testing.T) {
		banner := dialForBanner(t, compose, "root-without-a-namespace", signer)

		assert.Contains(t, banner, "SSHID Format Error")
		assert.Contains(t, banner, "Correct format: username@namespace.device@host")
	})

	t.Run("a banner tells the client the device is offline", func(t *testing.T) {
		agent, offline := startAcceptedAgent(t, ctx, compose)
		require.NoError(t, agent.Stop(ctx, nil))
		compose.AwaitDeviceOffline(t, offline.UID)

		banner := dialForBanner(t, compose, deviceSSHID(offline), signer)

		assert.Contains(t, banner, "Connection Failed")
		assert.Contains(t, banner, "The target device is offline or cannot be reached.")
	})

	t.Run("a client that disconnects leaves no command running on the device", func(t *testing.T) {
		const command = "sleep 637"

		before := currentSessions(t, ctx, compose)

		var conn *ssh.Client

		session := sessionAfter(t, ctx, compose, before, func() {
			conn = dialDevice(t, ctx, compose, device, signer)

			sess, err := conn.NewSession()
			require.NoError(t, err)
			require.NoError(t, sess.Start(command))
		})

		observer := dialDevice(t, ctx, compose, device, signer)
		defer observer.Close() //nolint:errcheck // the test is over once the device was read

		const isRunning = "pgrep -f '[s]leep 637' > /dev/null && echo running || echo gone"

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			output, err := deviceOutput(observer, isRunning)
			assert.NoError(tt, err, output)
			assert.Equal(tt, "running\n", output)
		}, 30*time.Second, time.Second)

		require.NoError(t, conn.Close())

		requireSessionActive(t, ctx, compose, session.UID, false)
		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			output, err := deviceOutput(observer, isRunning)
			assert.NoError(tt, err, output)
			assert.Equal(tt, "gone\n", output)
		}, 30*time.Second, time.Second)
	})

	t.Run("a session the server is asked to close ends on the client", func(t *testing.T) {
		before := currentSessions(t, ctx, compose)

		var (
			conn *ssh.Client
			sess *ssh.Session
		)

		session := sessionAfter(t, ctx, compose, before, func() {
			conn = dialDevice(t, ctx, compose, device, signer)

			var err error

			sess, err = conn.NewSession()
			require.NoError(t, err)
			require.NoError(t, sess.RequestPty("xterm", 24, 80, ssh.TerminalModes{}))
			require.NoError(t, sess.Shell())
		})
		defer conn.Close() //nolint:errcheck // the server ends the session, not the connection

		ended := make(chan error, 1)

		go func() { ended <- sess.Wait() }()

		resp, err := compose.R(ctx).
			SetBody(map[string]string{"device": device.UID}).
			Post("/api/sessions/" + session.UID + "/close")
		require.NoError(t, err)
		require.Equal(t, 202, resp.StatusCode(), resp.String())

		select {
		case err := <-ended:
			var missing *ssh.ExitMissingError

			require.ErrorAs(t, err, &missing, "the shell was ended under the client, not exited")
		case <-time.After(30 * time.Second):
			t.Fatal("the session the server was asked to close is still open on the client")
		}
	})
}

// TestSSHServerRestart covers what survives the server process going away: the agent's tunnel,
// rebuilt against the new process, and the sessions it was serving, which only the reaper can
// retire once their keep-alive stops. A session still served is kept alive past the same window.
func TestSSHServerRestart(t *testing.T) {
	ctx := context.Background()

	compose := newConfiguredSSHEnvironment(t, ctx,
		environment.New(t, run).WithEnv("SHELLHUB_SESSION_KEEPALIVE_TIMEOUT", "2m").WithCronTrigger(),
		models.SSHAccessModeLegacy)
	signer := registerDeviceKey(t, ctx, compose)
	_, device := startAcceptedAgent(t, ctx, compose)

	_, orphaned := openSession(t, ctx, compose, device, signer, "")

	server := compose.Service(environment.ServiceServer)
	require.NoError(t, server.Stop(ctx, nil))
	require.NoError(t, server.Start(ctx))

	requireSessionActive(t, ctx, compose, orphaned.UID, true)
	awaitAgentTunnels(t, compose, 2)

	t.Run("the agent reconnects and serves new sessions", func(t *testing.T) {
		conn := dialDevice(t, ctx, compose, device, signer)
		defer conn.Close() //nolint:errcheck // the test is over once the command answered

		assert.Equal(t, "after the restart", runOnDevice(t, conn, "echo -n after the restart"))
	})

	t.Run("a keep-alive keeps an idle session past the reaper's window", func(t *testing.T) {
		_, idle := openSession(t, ctx, compose, device, signer, "")

		compose.AgeSessionKeepAlive(t, orphaned.UID, 10*time.Minute)
		compose.AgeSessionKeepAlive(t, idle.UID, 10*time.Minute)

		awaitSession(t, ctx, compose, idle.UID, time.Minute, func(tt *assert.CollectT, session models.Session) {
			assert.WithinDuration(tt, time.Now(), session.LastSeen, time.Minute, "no keep-alive refreshed the idle session") //nolint:forbidigo // compares a stamp the server wrote with its own wall clock
		})

		compose.RunCron(t, sessionReaperCron)

		requireSessionActive(t, ctx, compose, orphaned.UID, false)
		requireSessionActive(t, ctx, compose, idle.UID, true)
	})
}

func awaitAgentTunnels(t *testing.T, compose *environment.DockerCompose, count int) {
	t.Helper()

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		reader, err := compose.Service(environment.ServiceServer).Logs(t.Context())
		if !assert.NoError(tt, err) {
			return
		}

		defer func() { _ = reader.Close() }()

		logs, err := io.ReadAll(reader)
		assert.NoError(tt, err)
		assert.GreaterOrEqual(tt, strings.Count(string(logs), "v2 connection established"), count)
	}, 2*time.Minute, 2*time.Second)
}

func dialForBanner(t *testing.T, compose *environment.DockerCompose, sshid string, signer ssh.Signer) string {
	t.Helper()

	var banner string

	err := handshakeWith(t.Context(), compose.SSHAddress(), &ssh.ClientConfig{
		User:            sshid,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // the test stack's host key is ephemeral
		BannerCallback: func(message string) error {
			banner += message

			return nil
		},
	})
	require.Error(t, err, "a connection the server explains with a banner is refused")

	return banner
}
