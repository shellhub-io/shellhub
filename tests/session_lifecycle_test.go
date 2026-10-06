package main

import (
	"context"
	"encoding/json"
	"net"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/websocket"
)

const (
	outsiderUsername      = "outsider"
	outsiderEmail         = "outsider@ossystems.com.br"
	outsiderPassword      = "password"
	outsiderNamespaceName = "outsiderspace"
	outsiderNamespace     = "00000000-0000-4000-0000-000000000003"
)

const (
	webMessageKindError   = 4
	webMessageKindSession = 5
)

func awaitSession(t *testing.T, ctx context.Context, compose *environment.DockerCompose, uid string, timeout time.Duration, check func(tt *assert.CollectT, session models.Session)) models.Session {
	t.Helper()

	var last models.Session

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		session := models.Session{} //nolint:exhaustruct // filled from the API response

		resp, err := compose.R(ctx).SetResult(&session).Get("/api/sessions/" + uid)
		if !assert.NoError(tt, err) || !assert.Equal(tt, 200, resp.StatusCode(), "the strict validator names the mismatch: %s", resp.String()) {
			return
		}

		last = session

		check(tt, session)
	}, timeout, 1*time.Second)

	return last
}

func getSession(t *testing.T, ctx context.Context, compose *environment.DockerCompose, uid string) models.Session {
	t.Helper()

	return awaitSession(t, ctx, compose, uid, 30*time.Second, func(*assert.CollectT, models.Session) {})
}

func openSession(t *testing.T, ctx context.Context, compose *environment.DockerCompose, device *models.Device, signer ssh.Signer, command string) (*ssh.Client, *models.Session) {
	t.Helper()

	before := currentSessions(t, ctx, compose)

	conn := dialDevice(t, ctx, compose, device, signer)
	t.Cleanup(func() { _ = conn.Close() })

	opened := sessionAfter(t, ctx, compose, before, func() {
		sess, err := conn.NewSession()
		require.NoError(t, err)

		if command != "" {
			_, err = sess.CombinedOutput(command)
			require.NoError(t, err)

			return
		}

		require.NoError(t, sess.RequestPty("xterm", 24, 80, ssh.TerminalModes{ssh.ECHO: 1}))
		require.NoError(t, sess.Shell())
	})

	return conn, opened
}

func finishSession(t *testing.T, ctx context.Context, compose *environment.DockerCompose, device *models.Device, signer ssh.Signer, command string) string {
	t.Helper()

	conn, opened := openSession(t, ctx, compose, device, signer, command)

	require.NoError(t, conn.Close())
	requireSessionActive(t, ctx, compose, opened.UID, false)

	return opened.UID
}

func openWebTerminal(t *testing.T, ctx context.Context, compose *environment.DockerCompose, device *models.Device) string {
	t.Helper()

	issued := struct {
		Token string `json:"token"`
	}{}

	resp, err := compose.R(ctx).
		SetBody(map[string]string{
			"device":   device.UID,
			"username": ShellHubAgentUsername,
			"password": ShellHubAgentPassword,
		}).
		SetResult(&issued).
		Post("/ws/ssh/session")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())

	query := url.Values{"token": {issued.Token}, "cols": {"80"}, "rows": {"24"}}

	config, err := websocket.NewConfig(
		strings.Replace(compose.BaseURL(), "http", "ws", 1)+"/ws/ssh?"+query.Encode(),
		compose.BaseURL(),
	)
	require.NoError(t, err)

	conn, err := config.DialContext(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	require.NoError(t, conn.SetReadDeadline(clock.Now().Add(30*time.Second)))

	for {
		var frame []byte
		require.NoError(t, websocket.Message.Receive(conn, &frame), "the terminal closed before naming its session")

		message := struct {
			Kind int    `json:"kind"`
			Data string `json:"data"`
		}{}

		if json.Unmarshal(frame, &message) != nil {
			continue
		}

		require.NotEqual(t, webMessageKindError, message.Kind, "the web terminal refused: %s", message.Data)

		if message.Kind == webMessageKindSession {
			return message.Data
		}
	}
}

func TestSessionLifecycle(t *testing.T) {
	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, "legacy")
	_, device := startAcceptedAgent(t, ctx, compose)
	signer := registerDeviceKey(t, ctx, compose)

	t.Run("a session is active while its client is connected and inactive once it disconnects", func(t *testing.T) {
		conn, opened := openSession(t, ctx, compose, device, signer, "")

		assert.False(t, opened.Web, "a session opened over SSH did not come from the web terminal")

		requireSessionActive(t, ctx, compose, opened.UID, true)

		require.NoError(t, conn.Close())

		requireSessionActive(t, ctx, compose, opened.UID, false)
	})

	t.Run("keep-alives advance last_seen while the session is open", func(t *testing.T) {
		_, opened := openSession(t, ctx, compose, device, signer, "")

		requireSessionActive(t, ctx, compose, opened.UID, true)

		seen := getSession(t, ctx, compose, opened.UID).LastSeen

		awaitSession(t, ctx, compose, opened.UID, 45*time.Second, func(tt *assert.CollectT, session models.Session) {
			assert.True(tt, session.Active, "the session is still open")
			assert.True(tt, session.LastSeen.After(seen), "last_seen stayed at %s", seen)
		})
	})

	t.Run("a login the device accepts marks the session authenticated", func(t *testing.T) {
		_, opened := openSession(t, ctx, compose, device, signer, "")

		awaitSession(t, ctx, compose, opened.UID, 30*time.Second, func(tt *assert.CollectT, session models.Session) {
			assert.True(tt, session.Authenticated)
		})
	})

	t.Run("a password the device refuses leaves the session unauthenticated", func(t *testing.T) {
		before := currentSessions(t, ctx, compose)

		refused := sessionAfter(t, ctx, compose, before, func() {
			dialer := net.Dialer{} //nolint:exhaustruct // the zero dialer is what ssh.Dial uses

			conn, err := dialer.DialContext(ctx, "tcp", compose.SSHAddress())
			require.NoError(t, err)

			defer conn.Close() //nolint:errcheck // the handshake failed, so nothing is left to flush

			_, _, _, err = ssh.NewClientConn(conn, compose.SSHAddress(), &ssh.ClientConfig{ //nolint:exhaustruct // the remaining fields keep their defaults
				User:            deviceSSHID(device),
				Auth:            []ssh.AuthMethod{ssh.Password("not-the-password")},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // the test stack's host key is ephemeral
			})
			require.Error(t, err)
		})

		session := getSession(t, ctx, compose, refused.UID)
		assert.False(t, session.Authenticated, "the device never accepted the credential")
		assert.False(t, session.Active, "a session the device refused never went live")
	})

	t.Run("a session opened from the web terminal is flagged web", func(t *testing.T) {
		skipUnlessAgentAcceptsPasswords(t)

		uid := openWebTerminal(t, ctx, compose, device)

		session := getSession(t, ctx, compose, uid)
		assert.True(t, session.Web)
		assert.Equal(t, device.UID, string(session.DeviceUID))
	})

	t.Run("a session is visible only inside its own namespace", func(t *testing.T) {
		compose.NewUser(t, outsiderUsername, outsiderEmail, outsiderPassword)
		compose.NewNamespace(t, outsiderUsername, outsiderNamespaceName, outsiderNamespace, "")

		outsider := compose.AuthUser(t, outsiderUsername, outsiderPassword)
		require.NotNil(t, outsider.Tenant)
		require.Equal(t, outsiderNamespace, *outsider.Tenant)

		_, opened := openSession(t, ctx, compose, device, signer, "")

		assert.Equal(t, ShellHubNamespace, getSession(t, ctx, compose, opened.UID).TenantID)

		resp, err := compose.Anonymous(ctx).SetAuthToken(outsider.Token).Get("/api/sessions/" + opened.UID)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode(), resp.String())

		listed := []models.Session{}

		resp, err = compose.Anonymous(ctx).SetAuthToken(outsider.Token).SetResult(&listed).Get("/api/sessions")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode(), resp.String())
		assert.Empty(t, listed, "another namespace lists none of this one's sessions")
	})
}
