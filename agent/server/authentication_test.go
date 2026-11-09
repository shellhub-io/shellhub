package server

import (
	"context"
	"net"
	"sync"
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/agent/server/modes"
	"github.com/shellhub-io/shellhub/agent/server/modes/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

type sshContext struct {
	context.Context
	sync.Mutex

	user string
}

func newSSHContext(user string) *sshContext {
	return &sshContext{Context: context.Background(), user: user}
}

func (c *sshContext) SetValue(key, value any)             { c.Context = context.WithValue(c.Context, key, value) }
func (c *sshContext) User() string                        { return c.user }
func (c *sshContext) SessionID() string                   { return "session" }
func (c *sshContext) ClientVersion() string               { return "" }
func (c *sshContext) ServerVersion() string               { return "" }
func (c *sshContext) RemoteAddr() net.Addr                { return nil }
func (c *sshContext) LocalAddr() net.Addr                 { return nil }
func (c *sshContext) Permissions() *gliderssh.Permissions { return &gliderssh.Permissions{} }

type authOnlyMode struct {
	*mocks.MockAuthenticator
	modes.Sessioner
}

func TestPasswordHandler(t *testing.T) {
	t.Run("admits the gateway's login secret through GatewayLogin", func(t *testing.T) {
		auth := mocks.NewMockAuthenticator(t)
		auth.EXPECT().GatewayLogin(mock.Anything, "root").Return(true).Once()

		s := &Server{mode: authOnlyMode{MockAuthenticator: auth}}
		ctx := newSSHContext("root")
		s.connCallback(ctx, &GatewayLoginConn{GatewayStream: GatewayStream{Secret: "secret"}})

		assert.True(t, s.passwordHandler(ctx, "secret"))
	})

	t.Run("checks anything else as the user's password", func(t *testing.T) {
		auth := mocks.NewMockAuthenticator(t)
		auth.EXPECT().Password(mock.Anything, "root", "guess").Return(false).Once()

		s := &Server{mode: authOnlyMode{MockAuthenticator: auth}}
		ctx := newSSHContext("root")
		s.connCallback(ctx, &GatewayLoginConn{GatewayStream: GatewayStream{Secret: "secret"}})

		assert.False(t, s.passwordHandler(ctx, "guess"))
	})

	t.Run("checks the password on a stream the gateway sent no secret for", func(t *testing.T) {
		auth := mocks.NewMockAuthenticator(t)
		auth.EXPECT().Password(mock.Anything, "root", "").Return(false).Once()

		s := &Server{mode: authOnlyMode{MockAuthenticator: auth}}
		ctx := newSSHContext("root")
		s.connCallback(ctx, &GatewayLoginConn{GatewayStream: GatewayStream{Secret: ""}})

		assert.False(t, s.passwordHandler(ctx, ""))
	})
}

type connMetadata struct {
	gossh.ConnMetadata

	user string
}

func (c connMetadata) User() string { return c.user }

func TestServerConfig(t *testing.T) {
	t.Run("admits the user of a vouched stream without a credential", func(t *testing.T) {
		auth := mocks.NewMockAuthenticator(t)
		auth.EXPECT().GatewayLogin(mock.Anything, "root").Return(true).Once()

		s := &Server{mode: authOnlyMode{MockAuthenticator: auth}}
		ctx := newSSHContext("")
		s.connCallback(ctx, &GatewayLoginConn{GatewayStream: GatewayStream{Vouched: true}})

		config := s.serverConfig(ctx)
		require.True(t, config.NoClientAuth)

		_, err := config.NoClientAuthCallback(connMetadata{user: "root"})
		assert.NoError(t, err)
	})

	t.Run("refuses a vouched user the device has no account for", func(t *testing.T) {
		auth := mocks.NewMockAuthenticator(t)
		auth.EXPECT().GatewayLogin(mock.Anything, "ghost").Return(false).Once()

		s := &Server{mode: authOnlyMode{MockAuthenticator: auth}}
		ctx := newSSHContext("")
		s.connCallback(ctx, &GatewayLoginConn{GatewayStream: GatewayStream{Secret: "secret", Vouched: true}})

		_, err := s.serverConfig(ctx).NoClientAuthCallback(connMetadata{user: "ghost"})
		assert.ErrorIs(t, err, errNotAdmitted)
	})

	t.Run("asks for a credential on a stream the gateway did not vouch for", func(t *testing.T) {
		s := &Server{mode: authOnlyMode{MockAuthenticator: mocks.NewMockAuthenticator(t)}}
		ctx := newSSHContext("")
		s.connCallback(ctx, &GatewayLoginConn{GatewayStream: GatewayStream{Secret: "secret"}})

		config := s.serverConfig(ctx)
		assert.False(t, config.NoClientAuth)
		assert.Nil(t, config.NoClientAuthCallback)
	})

	t.Run("announces the gateway login on a stream that carried its secret", func(t *testing.T) {
		s := &Server{}
		ctx := newSSHContext("")
		s.connCallback(ctx, &GatewayLoginConn{GatewayStream: GatewayStream{Secret: "secret"}})

		assert.Equal(t, "SSH-2.0-ShellHub-Agent gateway-login", s.serverConfig(ctx).ServerVersion)
	})

	t.Run("does not announce the gateway login on a stream with no secret", func(t *testing.T) {
		s := &Server{}
		ctx := newSSHContext("")
		s.connCallback(ctx, &net.TCPConn{})

		assert.Equal(t, "SSH-2.0-ShellHub-Agent", s.serverConfig(ctx).ServerVersion,
			"a V1 stream never carries the secret, so a password the gateway sends would match nothing")
	})
}
