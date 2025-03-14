package server

import (
	"fmt"
	"net"
	"os"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/pires/go-proxyproto"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	"github.com/shellhub-io/shellhub/ssh/server/auth"
	"github.com/shellhub-io/shellhub/ssh/server/channels"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
)

type Options struct {
	ConnectTimeout time.Duration
	// TODO: add default value for RECORD_URL.
	RecordURL string
	// Allows SSH to connect with an agent via a public key when the agent version is less than 0.6.0.
	// Agents 0.5.x or earlier do not validate the public key request and may panic.
	// Please refer to: https://github.com/shellhub-io/shellhub/issues/3453
	AllowPublickeyAccessBelow060 bool
}

type Server struct {
	sshd   *gliderssh.Server
	opts   *Options
	tunnel *httptunnel.Tunnel
}

func NewServer(opts *Options, tunnel *httptunnel.Tunnel, cache cache.Cache) *Server {
	server := &Server{ // nolint: exhaustruct
		opts:   opts,
		tunnel: tunnel,
	}

	server.sshd = &gliderssh.Server{ // nolint: exhaustruct
		Addr: ":2222",
		ConnCallback: func(ctx gliderssh.Context, conn net.Conn) net.Conn {
			ctx.SetValue("conn", conn)
			ctx.SetValue("RECORD_URL", opts.RecordURL)

			return conn
		},
		BannerHandler: func(ctx gliderssh.Context) string {
			logger := log.WithFields(
				log.Fields{
					"uid":   ctx.SessionID(),
					"sshid": ctx.User(),
				})

			logger.Info("new connection established")

			message := func(msg string) string {
				return fmt.Sprintf("%s\r\n", msg)
			}

			if _, err := target.NewTarget(ctx.User()); err != nil {
				logger.WithError(err).Error("invalid SSHID")

				return message("it is not a valid SSHID")
			}

			sess, err := session.NewSession(ctx, tunnel, cache)
			if err != nil {
				logger.WithError(err).Error("failed to create the session")

				return message("device is offline or cannot be reached")
			}

			if err := sess.Dial(ctx); err != nil {
				logger.WithError(err).Error("destination device is offline or cannot be reached")

				return message("device is offline or cannot be reached")
			}

			if err := sess.Evaluate(ctx); err != nil {
				logger.WithError(err).Error("destination device has a firewall to blocked it or a billing issue")

				return message("you cannot access the device due a policy rule")
			}

			return ""
		},
		PasswordHandler:  auth.PasswordHandler,
		PublicKeyHandler: auth.PublicKeyHandler,
		// Channels form the foundation of secure communication between clients and servers in SSH connections. A
		// channel, in the context of SSH, is a logical conduit through which data travels securely between the client
		// and the server. SSH channels serve as the infrastructure for executing commands, establishing shell sessions,
		// and securely forwarding network services.
		ChannelHandlers: map[string]gliderssh.ChannelHandler{
			channels.SessionChannel:     channels.DefaultSessionHandler(),
			channels.DirectTCPIPChannel: channels.DefaultDirectTCPIPHandler,
		},
		LocalPortForwardingCallback: func(_ gliderssh.Context, _ string, _ uint32) bool {
			return true
		},
		ReversePortForwardingCallback: func(_ gliderssh.Context, _ string, _ uint32) bool {
			return false
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
