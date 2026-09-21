package main

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func dialClient(t *testing.T, ctx context.Context, addr string, config *ssh.ClientConfig) *ssh.Client {
	t.Helper()

	var client *ssh.Client

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		dialer := net.Dialer{} //nolint:exhaustruct // the zero dialer is what ssh.Dial uses

		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if !assert.NoError(tt, err) {
			return
		}

		sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
		if !assert.NoError(tt, err) {
			_ = conn.Close()

			return
		}

		client = ssh.NewClient(sshConn, chans, reqs)
	}, 30*time.Second, 1*time.Second)

	return client
}

func dialDevice(t *testing.T, ctx context.Context, compose *environment.DockerCompose, device *models.Device) *ssh.Client {
	t.Helper()

	return dialClient(t, ctx, compose.SSHAddress(), &ssh.ClientConfig{ //nolint:exhaustruct // the remaining fields keep their defaults
		User: deviceSSHID(device),
		Auth: []ssh.AuthMethod{
			ssh.Password(ShellHubAgentPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // the test stack's host key is ephemeral
	})
}

func openShellSession(t *testing.T, ctx context.Context, compose *environment.DockerCompose, device *models.Device) (*ssh.Client, *models.Session) {
	t.Helper()

	conn := dialDevice(t, ctx, compose, device)

	sess, err := conn.NewSession()
	require.NoError(t, err)

	require.NoError(t, sess.RequestPty("xterm", 100, 100, ssh.TerminalModes{ssh.ECHO: 1}))
	require.NoError(t, sess.Shell())

	session := &models.Session{} //nolint:exhaustruct // filled from the API response

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		sessions := []models.Session{}

		resp, err := compose.R(ctx).SetResult(&sessions).Get("/api/sessions")
		assert.NoError(tt, err)
		assert.Equal(tt, 200, resp.StatusCode())

		for _, s := range sessions {
			if s.Active {
				*session = s

				return
			}
		}

		assert.Fail(tt, "no active session yet")
	}, 30*time.Second, 1*time.Second)

	return conn, session
}

func requireSessionActive(t *testing.T, ctx context.Context, compose *environment.DockerCompose, uid string, active bool) {
	t.Helper()

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		session := models.Session{} //nolint:exhaustruct // filled from the API response

		resp, err := compose.R(ctx).SetResult(&session).Get("/api/sessions/" + uid)
		assert.NoError(tt, err)
		assert.Equal(tt, 200, resp.StatusCode())
		assert.Equal(tt, active, session.Active)
	}, 30*time.Second, 1*time.Second)
}

func TestSessionCloseDelegatesToTheGatewayThatOwnsIt(t *testing.T) {
	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, "legacy")
	_, device := startAcceptedAgent(t, ctx, compose)

	conn, session := openShellSession(t, ctx, compose, device)

	resp, err := compose.R(ctx).
		SetBody(map[string]string{"device": device.UID}).
		Post("/api/sessions/" + session.UID + "/close")
	require.NoError(t, err)
	assert.Equal(t, 202, resp.StatusCode(),
		"a session the gateway still owns is delivered to the device, not closed by the endpoint")

	_ = conn.Close()

	requireSessionActive(t, ctx, compose, session.UID, false)
}

func TestSessionCloseRetiresASessionNoGatewayOwns(t *testing.T) {
	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, "legacy")
	_, device := startAcceptedAgent(t, ctx, compose)

	conn, session := openShellSession(t, ctx, compose, device)
	defer conn.Close() //nolint:errcheck // the connection is already dead once the server restarts

	server := compose.Service(environment.ServiceServer)
	require.NoError(t, server.Stop(ctx, nil))
	require.NoError(t, server.Start(ctx))

	requireSessionActive(t, ctx, compose, session.UID, true)

	var status int

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		resp, err := compose.R(ctx).
			SetBody(map[string]string{"device": device.UID}).
			Post("/api/sessions/" + session.UID + "/close")
		assert.NoError(tt, err)
		assert.NotEqual(tt, 0, resp.StatusCode())

		status = resp.StatusCode()
	}, 60*time.Second, 2*time.Second)

	assert.Equal(t, 200, status,
		"no gateway owns this session, so the endpoint closes it instead of reporting delivery")

	requireSessionActive(t, ctx, compose, session.UID, false)
}
