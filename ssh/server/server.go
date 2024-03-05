package server

import (
	"net"
	"os"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/pires/go-proxyproto"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	"github.com/shellhub-io/shellhub/ssh/server/auth"
	"github.com/shellhub-io/shellhub/ssh/server/channels"
	"github.com/shellhub-io/shellhub/ssh/server/subsystems"
	log "github.com/sirupsen/logrus"
)

type Options struct {
	ConnectTimeout time.Duration `env:"CONNECT_TIMEOUT,default=30s"`
	RedisURI       string        `env:"REDIS_URI,default=redis://redis:6379"`
	// TODO: add default value for RECORD_URL.
	RecordURL string `env:"RECORD_URL"`
	// Allows SSH to connect with an agent via a public key when the agent version is less than 0.6.0.
	// Agents 0.5.x or earlier do not validate the public key request and may panic.
	// Please refer to: https://github.com/shellhub-io/shellhub/issues/3453
	AllowPublickeyAccessBelow060 bool `env:"ALLOW_PUBLIC_KEY_ACCESS_BELLOW_0_6_0,default=false"`
}

type Server struct {
	sshd   *gliderssh.Server
	opts   *Options
	tunnel *httptunnel.Tunnel
}

func NewServer(opts *Options, tunnel *httptunnel.Tunnel) *Server {
	server := &Server{ // nolint: exhaustruct
		opts:   opts,
		tunnel: tunnel,
	}

	server.sshd = &gliderssh.Server{ // nolint: exhaustruct
		Addr:             ":2222",
		PasswordHandler:  auth.PasswordHandlerWithTunnel(tunnel),
		PublicKeyHandler: auth.PublicKeyHandlerWithTunnel(tunnel),
		// Handler is the default one for "normal" SSH connection. It is called by [gliderssh.DefaultSessionHandler],
		// allowing its client to manager the SSH session with the client. We peform our Shell, Exec, and Heredoc
		// management in this function, what acts like a the "main" handler.
		//
		// It worth to notice that, when a SFTP, for example, session is requested, this handler isn't called, calling
		// the channel handler for this request instead.
		Handler: Handler(tunnel, opts),
		// Channels form the foundation of secure communication between clients and servers in SSH connections. A
		// channel, in the context of SSH, is a logical conduit through which data travels securely between the client
		// and the server. SSH channels serve as the infrastructure for executing commands, establishing shell sessions,
		// and securely forwarding network services.
		ChannelHandlers: map[string]gliderssh.ChannelHandler{
			channels.SessionChannel:     gliderssh.DefaultSessionHandler,
			channels.DirectTCPIPChannel: channels.DefaultDirectTCPIPHandler,
		},
		// SSH subsystems extend the functionality of SSH connections by offering specialized services beyond standard
		// shell access. A subsystem, in the context of SSH, refers to an additional feature or service that can be
		// executed securely over the SSH connection. A good example of SSH subsystems is the SFTP one.
		SubsystemHandlers: map[string]gliderssh.SubsystemHandler{
			subsystems.SFTPSubsystem: subsystems.SFTPSubsystemHandler,
		},
		LocalPortForwardingCallback: func(ctx gliderssh.Context, dhost string, dport uint32) bool {
			return true
		},
		ReversePortForwardingCallback: func(ctx gliderssh.Context, bindHost string, bindPort uint32) bool {
			return false
		},
		SessionRequestCallback: func(client gliderssh.Session, request string) bool {
			metadata.StoreRequest(client.Context(), request)

			target := metadata.RestoreTarget(client.Context())
			log.WithFields(log.Fields{
				"username": target.Username,
				"sshid":    target.Data,
				"request":  request,
			}).Info("Session request")

			return true
		},
	}

	if _, err := os.Stat(os.Getenv("PRIVATE_KEY")); os.IsNotExist(err) {
		log.WithError(err).Fatal("private key not found!")
	}

	if err := server.sshd.SetOption(gliderssh.HostKeyFile(os.Getenv("PRIVATE_KEY"))); err != nil {
		log.WithError(err).Fatal("host key not found!")
	}

	return server
}

func (s *Server) ListenAndServe() error {
	log.WithFields(log.Fields{
		"addr": s.sshd.Addr,
	}).Info("ssh server listening")

	list, err := net.Listen("tcp", s.sshd.Addr)
	if err != nil {
		log.WithError(err).Error("failed to listen an serve the TCP server")

		return err
	}

	proxy := &proxyproto.Listener{Listener: list} // nolint: exhaustruct
	defer proxy.Close()

	return s.sshd.Serve(proxy)
}
