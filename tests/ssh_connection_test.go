package main

import (
	"context"
	"testing"
	"time"

	"github.com/moby/moby/client"
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
	t.Parallel()

	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, legacyNamespace)
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
		defer conn.Close() //nolint:errcheck // only a failed run reaches this with the connection still open

		ended := make(chan error, 1)

		go func() { ended <- sess.Wait() }()

		disconnected := make(chan struct{})

		go func() {
			defer close(disconnected)

			_ = conn.Wait()
		}()

		resp, err := compose.R(ctx).
			SetBody(map[string]string{"device": device.UID}).
			Post("/api/sessions/" + session.UID + "/close")
		require.NoError(t, err)
		require.Equal(t, 202, resp.StatusCode(), resp.String())

		select {
		case err := <-ended:
			var exit *ssh.ExitError

			require.ErrorAs(t, err, &exit, "the shell ended without the client being told how")
			assert.Equal(t, "HUP", exit.Signal())
			assert.Equal(t, "shellhub: the session was ended on the device", exit.Msg())
		case <-time.After(30 * time.Second):
			t.Fatal("the session the server was asked to close is still open on the client")
		}

		select {
		case <-disconnected:
		case <-time.After(30 * time.Second):
			t.Fatal("the server ended the session but kept the client's connection open")
		}

		requireSessionActive(t, ctx, compose, session.UID, false)
	})
}

// TestSSHServerRestart covers what survives the server process going away: the agent's tunnel,
// rebuilt against the new process, and the sessions it was serving, which only the reaper can
// retire once their keep-alive stops. A session still served is kept alive past the same window.
func TestSSHServerRestart(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := newConfiguredSSHEnvironment(t, ctx,
		environment.New(t, run).WithEnv("SHELLHUB_SESSION_KEEPALIVE_TIMEOUT", "2m").WithCronTrigger(),
		legacyNamespace)
	signer := registerDeviceKey(t, ctx, compose)
	_, device := startAcceptedAgent(t, ctx, compose)

	_, orphaned := openSession(t, ctx, compose, device, signer, "")

	server := compose.Service(environment.ServiceServer)
	require.NoError(t, server.Stop(ctx, nil))
	require.NoError(t, server.Start(ctx))

	requireSessionActive(t, ctx, compose, orphaned.UID, true)

	t.Run("the agent reconnects and serves new sessions", func(t *testing.T) {
		conn := dialDeviceWithin(t, ctx, compose, device, signer, 90*time.Second)
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

// TestSSHUnresponsiveDevice covers a device that holds its tunnel but never dials back: a V1 agent
// paused as soon as it comes online, whose kernel keeps the tunnel's socket open while the agent
// reads nothing from it. The tunnel is pinged once a minute, so it outlives the server's 30 second
// wait for the dial-back by about half a minute; the client fails at that wait, while the device
// still holds its tunnel, and does not wait out the 180 seconds the server allows a handshake.
func TestSSHUnresponsiveDevice(t *testing.T) {
	t.Parallel()

	const (
		tunnelPingInterval = 59 * time.Second
		failWithin         = 45 * time.Second
	)

	ctx := context.Background()

	compose := newConfiguredSSHEnvironment(t, ctx,
		environment.New(t, run).WithEnv("SHELLHUB_TUNNEL_PING_INTERVAL", tunnelPingInterval.String()),
		legacyNamespace)
	signer := registerDeviceKey(t, ctx, compose)
	agent, device := startAcceptedAgent(t, ctx, compose, NewAgentContainerWithConnectionVersion(1))

	docker := newDockerClient(t)

	_, err := docker.ContainerPause(ctx, agent.GetContainerID(), client.ContainerPauseOptions{})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = docker.ContainerUnpause(context.Background(), agent.GetContainerID(), client.ContainerUnpauseOptions{})
	})

	mark := compose.ServerLogMark(t)

	dialCtx, cancel := context.WithTimeout(ctx, 4*time.Minute)
	defer cancel()

	started := time.Now() //nolint:forbidigo // measures how long a real client waits
	err = handshakeWith(dialCtx, compose.SSHAddress(), &ssh.ClientConfig{
		User:            deviceSSHID(device),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // the test stack's host key is ephemeral
	})
	waited := time.Since(started)

	require.Error(t, err, "the client reached a device that cannot answer")
	t.Logf("the client failed after %s: %v", waited, err)
	assert.Less(t, waited, failWithin, "the client waited past the server's dial-back bound")
	assert.True(t, requireDevice(t, compose, device.UID).Online, "the tunnel closed before the client failed, so the dial-back bound was not what ended the wait")

	compose.AwaitServerLogLine(t, mark, "sshid="+deviceSSHID(device), "the device holds a tunnel but did not answer")
}

func requireAccessDenied(t *testing.T, compose *environment.DockerCompose, sshid string, signer ssh.Signer, reasons ...string) {
	t.Helper()

	mark := compose.ServerLogMark(t)

	assert.Contains(t, dialForBanner(t, compose, sshid, signer), "Access Denied")
	compose.AwaitServerLogLine(t, mark, append([]string{"sshid=" + sshid}, reasons...)...)
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
