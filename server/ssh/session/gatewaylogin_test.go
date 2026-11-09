package session

import (
	"net"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionSniffer(t *testing.T) {
	cases := []struct {
		description string
		stream      string
		admits      bool
	}{
		{description: "an agent announcing the gateway login", stream: "SSH-2.0-ShellHub-Agent gateway-login\r\n", admits: true},
		{description: "an agent that predates it", stream: "SSH-2.0-Go\r\n", admits: false},
		{description: "lines ahead of the version", stream: "hello gateway-login\r\nSSH-2.0-ShellHub-Agent gateway-login\r\n", admits: true},
		{description: "the comment past the version line", stream: "SSH-2.0-Go\r\n gateway-login\r\n", admits: false},
		{description: "a version never finished", stream: "SSH-2.0-ShellHub-Agent gateway-login", admits: false},
		{description: "an agent announcing both", stream: "SSH-2.0-ShellHub-Agent gateway-login shell-on-open\r\n", admits: true},
		{description: "a capability inside another word", stream: "SSH-2.0-ShellHub-Agent gateway-login-shell-on-open\r\n", admits: false},
		{description: "a capability in the software version", stream: "SSH-2.0-shell-on-open\r\n", admits: false},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			server, device := net.Pipe()
			t.Cleanup(func() {
				_ = server.Close()
				_ = device.Close()
			})

			go func() {
				_, _ = device.Write([]byte(tc.stream))
				_ = device.Close()
			}()

			sniffer := sniffVersion(server)

			buf := make([]byte, 1)
			for {
				if _, err := sniffer.Read(buf); err != nil {
					break
				}
			}

			assert.Equal(t, tc.admits, sniffer.admitsGatewayLogin())
		})
	}
}

func TestGatewayLogin(t *testing.T) {
	t.Run("logs in with the stream's secret where the agent admits it", func(t *testing.T) {
		device := newPasswordDevice(t)
		device.version = "SSH-2.0-ShellHub-Agent gateway-login"

		sess, ctx, _ := warmSession(t, device)
		sess.agentLogin = "secret"
		sess.Namespace.Settings = &models.NamespaceSettings{Legacy: true} //nolint:exhaustruct // only the model matters

		sess.Warm(ctx)

		require.NoError(t, sess.connect(ctx, gatewayCredential), "the session has no service to mint a key with")

		_, attempts := device.seen()
		assert.Equal(t, []string{"secret"}, attempts)
	})

	t.Run("never sends the secret to an agent that does not announce it", func(t *testing.T) {
		device := newPasswordDevice(t)

		sess, ctx, _ := warmSession(t, device)
		sess.agentLogin = "secret"

		require.Error(t, sess.connect(ctx, gatewayCredential))

		_, attempts := device.seen()
		assert.Empty(t, attempts)
	})

	t.Run("sends a client's password as typed to an agent that admits the secret", func(t *testing.T) {
		device := newPasswordDevice(t)
		device.version = "SSH-2.0-ShellHub-Agent gateway-login"

		sess, ctx, _ := warmSession(t, device)
		sess.agentLogin = "other"

		require.NoError(t, sess.connect(ctx, credential{password: "secret"}))

		_, attempts := device.seen()
		assert.Equal(t, []string{"secret"}, attempts)
	})
}
