package server

import (
	"crypto/subtle"
	"errors"
	"net"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

const agentVersion = "SSH-2.0-ShellHub-Agent"

func serverVersion(ctx gliderssh.Context) string {
	if gatewayStream(ctx).Secret != "" {
		return agentVersion + " " + models.AgentCapabilityGatewayLogin + " " + models.AgentCapabilityShellOnOpen
	}

	return agentVersion + " " + models.AgentCapabilityShellOnOpen
}

type gatewayStreamKey struct{}

var errNotAdmitted = errors.New("the gateway's login names no account the device admits")

// GatewayStream is what the gateway said in the header of a stream it opened for one SSH
// connection. Secret is what it logs in with once it has authorized the client.
//
// Vouched is the gateway's word that every login on the stream is its own: the stream's SSH user
// is then admitted without a credential, as soon as it names an account the device admits.
type GatewayStream struct {
	Secret  string
	Vouched bool
}

// GatewayLoginConn is a stream the gateway opened for one SSH connection, with what it said in
// the stream's header.
type GatewayLoginConn struct {
	net.Conn
	GatewayStream
}

func (s *Server) connCallback(ctx gliderssh.Context, conn net.Conn) net.Conn {
	if login, ok := conn.(*GatewayLoginConn); ok {
		ctx.SetValue(gatewayStreamKey{}, login.GatewayStream)
	}

	return conn
}

func gatewayStream(ctx gliderssh.Context) GatewayStream {
	stream, _ := ctx.Value(gatewayStreamKey{}).(GatewayStream)

	return stream
}

func (s *Server) serverConfig(ctx gliderssh.Context) *gossh.ServerConfig {
	config := &gossh.ServerConfig{ServerVersion: serverVersion(ctx)}

	stream := gatewayStream(ctx)

	if stream.Vouched {
		config.NoClientAuth = true
		config.NoClientAuthCallback = func(conn gossh.ConnMetadata) (*gossh.Permissions, error) {
			admitted := s.mode.GatewayLogin(ctx, conn.User())

			log.WithFields(log.Fields{"user": conn.User(), "admitted": admitted}).Info("Admitting the login the gateway vouched for")

			if !admitted {
				return nil, errNotAdmitted
			}

			return &gossh.Permissions{}, nil
		}
	}

	return config
}

func (s *Server) passwordHandler(ctx gliderssh.Context, pass string) bool {
	if secret := gatewayStream(ctx).Secret; secret != "" && subtle.ConstantTimeCompare([]byte(secret), []byte(pass)) == 1 {
		admitted := s.mode.GatewayLogin(ctx, ctx.User())

		log.WithFields(log.Fields{"user": ctx.User(), "admitted": admitted}).Info("Using the gateway's login")

		return admitted
	}

	return s.mode.Password(ctx, ctx.User(), pass)
}

func (s *Server) publicKeyHandler(ctx gliderssh.Context, key gliderssh.PublicKey) bool {
	return s.mode.PublicKey(ctx, ctx.User(), key)
}
