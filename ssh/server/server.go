package server

import (
	"net"
	"os"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/pires/go-proxyproto"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/ssh/pkg/banner"
	"github.com/shellhub-io/shellhub/ssh/pkg/dialer"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	"github.com/shellhub-io/shellhub/ssh/server/auth"
	"github.com/shellhub-io/shellhub/ssh/server/channels"
	"github.com/shellhub-io/shellhub/ssh/session"
	log "github.com/sirupsen/logrus"
)

type Options struct {
	ConnectTimeout time.Duration
	// Allows SSH to connect with an agent via a public key when the agent version is less than 0.6.0.
	// Agents 0.5.x or earlier do not validate the public key request and may panic.
	// Please refer to: https://github.com/shellhub-io/shellhub/issues/3453
	AllowPublickeyAccessBelow060 bool
}

type Server struct {
	sshd   *gliderssh.Server
	opts   *Options
	dialer *dialer.Dialer
}

// bannerDeps holds the injectable operations used by newBannerHandler. The
// defaults (set by defaultBannerDeps) delegate to the real session package;
// tests supply stubs to exercise individual branches without network I/O.
type bannerDeps struct {
	newSession func(ctx gliderssh.Context, d *dialer.Dialer, c cache.Cache) (*session.Session, error)
	dial       func(sess *session.Session, ctx gliderssh.Context) error
	evaluate   func(sess *session.Session, ctx gliderssh.Context) error
}

func defaultBannerDeps() bannerDeps {
	return bannerDeps{
		newSession: session.NewSession,
		dial:       (*session.Session).Dial,
		evaluate:   (*session.Session).Evaluate,
	}
}

// newBannerHandler returns a gliderssh.BannerHandler that validates the SSHID,
// establishes the session, and dials the target device. It returns a banner
// message (using ssh/pkg/banner) when any step fails, or an empty string on
// success so the SSH handshake continues normally.
func newBannerHandler(d *dialer.Dialer, c cache.Cache) gliderssh.BannerHandler {
	return newBannerHandlerWithDeps(d, c, defaultBannerDeps())
}

// newBannerHandlerWithDeps is the testable core of newBannerHandler. Callers
// supply a bannerDeps to stub out network-dependent operations.
func newBannerHandlerWithDeps(d *dialer.Dialer, c cache.Cache, deps bannerDeps) gliderssh.BannerHandler {
	return func(ctx gliderssh.Context) string {
		logger := log.WithFields(
			log.Fields{
				"uid":   ctx.SessionID(),
				"sshid": ctx.User(),
			})

		logger.Info("new connection established")

		if _, err := target.NewTarget(ctx.User()); err != nil {
			logger.WithError(err).Error("invalid SSHID")

			return banner.Message(banner.KindInvalidSSHID)
		}

		sess, err := deps.newSession(ctx, d, c)
		if err != nil {
			logger.WithError(err).Error("failed to create the session")

			return banner.Message(banner.KindConnectionFailed)
		}

		if err := deps.dial(sess, ctx); err != nil {
			logger.WithError(err).Error("destination device is offline or cannot be reached")

			return banner.Message(banner.KindConnectionFailed)
		}

		if err := deps.evaluate(sess, ctx); err != nil {
			logger.WithError(err).Error("destination device has a firewall to blocked it or a billing issue")

			return banner.Message(banner.KindAccessDenied)
		}

		return ""
	}
}

func NewServer(dialer *dialer.Dialer, cache cache.Cache, opts *Options) *Server {
	server := &Server{ // nolint: exhaustruct
		opts:   opts,
		dialer: dialer,
	}

	server.sshd = &gliderssh.Server{ // nolint: exhaustruct
		Addr: ":2222",
		ConnCallback: func(ctx gliderssh.Context, conn net.Conn) net.Conn {
			ctx.SetValue("conn", conn)

			return conn
		},
		BannerHandler:    newBannerHandler(dialer, cache),
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

	if _, err := os.Stat(os.Getenv("PRIVATE_KEY")); os.IsNotExist(err) { //nolint:gosec // G703: path comes from trusted env var
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
	defer proxy.Close()                           //nolint:errcheck

	return s.sshd.Serve(proxy)
}
