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

	return dialClientWithin(t, ctx, addr, config, 30*time.Second)
}

func dialClientWithin(t *testing.T, ctx context.Context, addr string, config *ssh.ClientConfig, timeout time.Duration) *ssh.Client {
	t.Helper()

	var client *ssh.Client

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		var err error

		client, err = dialClientOnce(ctx, addr, config)
		assert.NoError(tt, err)
	}, timeout, 1*time.Second)

	return client
}

func dialClientOnce(ctx context.Context, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
	dialer := net.Dialer{} //nolint:exhaustruct // the zero dialer is what ssh.Dial uses

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		_ = conn.Close()

		return nil, err
	}

	return ssh.NewClient(sshConn, chans, reqs), nil
}

func dialDevice(t *testing.T, ctx context.Context, compose *environment.DockerCompose, device *models.Device, signer ssh.Signer) *ssh.Client {
	t.Helper()

	return dialDeviceWithin(t, ctx, compose, device, signer, 30*time.Second)
}

func dialDeviceWithin(t *testing.T, ctx context.Context, compose *environment.DockerCompose, device *models.Device, signer ssh.Signer, timeout time.Duration) *ssh.Client {
	t.Helper()

	return dialClientWithin(t, ctx, compose.SSHAddress(), &ssh.ClientConfig{ //nolint:exhaustruct // the remaining fields keep their defaults
		User: deviceSSHID(device),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // the test stack's host key is ephemeral
	}, timeout)
}

func openShellSession(t *testing.T, ctx context.Context, compose *environment.DockerCompose, device *models.Device) (*ssh.Client, *models.Session) {
	t.Helper()

	conn := dialDevice(t, ctx, compose, device, registerDeviceKey(t, ctx, compose))

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

	awaitSession(t, ctx, compose, uid, 30*time.Second, func(tt *assert.CollectT, session models.Session) {
		assert.Equal(tt, active, session.Active)
	})
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
