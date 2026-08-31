package server

import (
	"context"
	"errors"
	"net"
	"os"
	"runtime/debug"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/pires/go-proxyproto"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/banner"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/target"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/webhandoff"
	"github.com/shellhub-io/shellhub/server/ssh/server/auth"
	"github.com/shellhub-io/shellhub/server/ssh/server/channels"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

type finishingConn struct {
	net.Conn

	ctx gliderssh.Context
}

func (c *finishingConn) Close() error {
	if sess, _ := session.ObtainSession(c.ctx); sess != nil {
		sess.Finish() //nolint:errcheck
	}

	return c.Conn.Close()
}

type Options struct {
	ConnectTimeout time.Duration
	// HostKeyFile is the path to the SSH host key. It is deliberately not read
	// from PRIVATE_KEY: the API signs its JWTs with a key under that name, and
	// silently adopting it would change the host key fingerprint of every
	// existing installation.
	HostKeyFile string
	// Allows SSH to connect with an agent via a public key when the agent version is less than 0.6.0.
	// Agents 0.5.x or earlier do not validate the public key request and may panic.
	AllowPublickeyAccessBelow060 bool
	// Domain is the instance's base domain, used to build the browser approval
	// URL shown in the terminal when a namespace requires SSH login approval.
	Domain string
	// AutoSSL reports whether the console is served over HTTPS; it selects the
	// scheme of that approval URL.
	AutoSSL bool
}

var keepAlive = net.KeepAliveConfig{
	Enable:   true,
	Idle:     15 * time.Second,
	Interval: 15 * time.Second,
	Count:    9,
}

const handshakeBudget = 2*session.ApprovalWaitTimeout + 30*time.Second

type Server struct {
	sshd   *gliderssh.Server
	opts   *Options
	dialer *dialer.Dialer
}

type bannerDeps struct {
	newSession func(ctx gliderssh.Context, d *dialer.Dialer, service services.Service, handoff *webhandoff.Store) (*session.Session, error)
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

func newBannerHandler(d *dialer.Dialer, service services.Service, handoff *webhandoff.Store) gliderssh.BannerHandler {
	return newBannerHandlerWithDeps(d, service, handoff, defaultBannerDeps())
}

func newBannerHandlerWithDeps(d *dialer.Dialer, service services.Service, handoff *webhandoff.Store, deps bannerDeps) gliderssh.BannerHandler {
	return func(ctx gliderssh.Context) (message string) {
		logger := log.WithFields(
			log.Fields{
				"uid":   ctx.SessionID(),
				"sshid": ctx.User(),
			})

		defer func() {
			if r := recover(); r != nil {
				logger.WithField("panic", r).Error("recovered from panic while establishing the session")

				message = banner.Message(banner.KindConnectionFailed)
			}
		}()

		logger.Info("new connection established")

		if _, err := target.NewTarget(ctx.User()); err != nil {
			logger.WithError(err).Error("invalid SSHID")

			return banner.Message(banner.KindInvalidSSHID)
		}

		sess, err := deps.newSession(ctx, d, service, handoff)
		if err != nil {
			logger.WithError(err).Error("failed to create the session")

			return banner.Message(banner.KindConnectionFailed)
		}

		if err := deps.dial(sess, ctx); err != nil {
			logger.WithError(err).Error("destination device is offline or cannot be reached")

			return banner.Message(banner.KindConnectionFailed)
		}

		if err := deps.evaluate(sess, ctx); err != nil {
			evaluated := logger.WithError(err)
			if sess.Device != nil {
				evaluated = evaluated.WithField("tenant", sess.Device.TenantID)
			}

			switch {
			case errors.Is(err, session.ErrBillingBlock):
				evaluated.Error("destination device is blocked by the namespace's billing state")
			case errors.Is(err, session.ErrFirewallBlock), errors.Is(err, session.ErrFirewallUnknown):
				evaluated.Error("destination device is blocked by a firewall rule")
			default:
				evaluated.Error("destination device did not pass the connection evaluation")
			}

			return banner.Message(banner.KindAccessDenied)
		}

		return ""
	}
}

var (
	errNoneAuthUnsupported = errors.New("ssh: none authentication is not supported")
	errPermissionDenied    = errors.New("ssh: permission denied")
)

func newServerConfigCallback(ctx gliderssh.Context) *gossh.ServerConfig {
	return &gossh.ServerConfig{ //nolint:exhaustruct
		NoClientAuth: true,
		VerifiedPublicKeyCallback: func(_ gossh.ConnMetadata, key gossh.PublicKey, _ *gossh.Permissions, _ string) (*gossh.Permissions, error) {
			if ok := auth.PublicKeyVerified(ctx, key); !ok {
				return nil, errPermissionDenied
			}

			return ctx.Permissions().Permissions, nil
		},
		PreAuthConnCallback: func(conn gossh.ServerPreAuthConn) {
			session.StorePreAuthConn(ctx, conn)
		},
		NoClientAuthCallback: func(gossh.ConnMetadata) (*gossh.Permissions, error) {
			sess, state := session.ObtainSession(ctx)
			if !state.Evaluated() || sess.Web || !sess.IsIdentityMode() {
				return nil, errNoneAuthUnsupported
			}

			return nil, &gossh.PartialSuccessError{
				Next: gossh.ServerAuthCallbacks{ //nolint:exhaustruct
					PublicKeyCallback: func(_ gossh.ConnMetadata, key gossh.PublicKey) (*gossh.Permissions, error) {
						if ok := auth.PublicKeyOffer(ctx, key); !ok {
							return nil, errPermissionDenied
						}

						return ctx.Permissions().Permissions, nil
					},
				},
			}
		},
	}
}

func NewServer(dialer *dialer.Dialer, service services.Service, handoff *webhandoff.Store, opts *Options) (*Server, error) {
	session.Configure(session.Config{
		AllowPublickeyAccessBelow060: opts.AllowPublickeyAccessBelow060,
		Domain:                       opts.Domain,
		AutoSSL:                      opts.AutoSSL,
		ConnectTimeout:               opts.ConnectTimeout,
	})

	server := &Server{ //nolint:exhaustruct
		opts:   opts,
		dialer: dialer,
	}

	server.sshd = &gliderssh.Server{ //nolint:exhaustruct
		Addr:             ":2222",
		HandshakeTimeout: handshakeBudget,
		ConnCallback: func(ctx gliderssh.Context, conn net.Conn) net.Conn {
			wrapped := &finishingConn{Conn: conn, ctx: ctx}

			session.StoreConn(ctx, wrapped)

			return wrapped
		},
		ServerConfigCallback: newServerConfigCallback,
		BannerHandler:        newBannerHandler(dialer, service, handoff),
		PasswordHandler:      auth.PasswordHandler,
		PublicKeyHandler:     auth.PublicKeyOffer,
		ChannelHandlers: map[string]gliderssh.ChannelHandler{
			channels.SessionChannel:     recoverChannel(channels.SessionChannel, channels.DefaultSessionHandler()),
			channels.DirectTCPIPChannel: recoverChannel(channels.DirectTCPIPChannel, channels.DefaultDirectTCPIPHandler),
		},
		RequestHandlers: map[string]gliderssh.RequestHandler{
			"session-uid@shellhub.io": func(ctx gliderssh.Context, _ *gliderssh.Server, _ *gossh.Request) (bool, []byte) {
				return true, []byte(ctx.SessionID())
			},
		},
		LocalPortForwardingCallback: func(_ gliderssh.Context, _ string, _ uint32) bool {
			return true
		},
		ReversePortForwardingCallback: func(_ gliderssh.Context, _ string, _ uint32) bool {
			return false
		},
	}

	if opts.HostKeyFile == "" {
		return nil, errors.New("no ssh host key configured")
	}

	if _, err := os.Stat(opts.HostKeyFile); err != nil {
		return nil, errors.Join(errors.New("failed to read the ssh host key"), err)
	}

	if err := server.sshd.SetOption(gliderssh.HostKeyFile(opts.HostKeyFile)); err != nil {
		return nil, errors.Join(errors.New("failed to load the ssh host key"), err)
	}

	return server, nil
}

func recoverChannel(name string, next gliderssh.ChannelHandler) gliderssh.ChannelHandler {
	return func(srv *gliderssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx gliderssh.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.WithFields(log.Fields{
					"channel": name,
					"session": ctx.SessionID(),
					"panic":   r,
					"stack":   string(debug.Stack()),
				}).Error("recovered from a panic on an ssh channel")
			}
		}()

		next(srv, conn, newChan, ctx)
	}
}

var loopbackProxyPolicy = proxyproto.MustPolicyFromRanges([]string{"127.0.0.0/8", "::1/128"}, proxyproto.USE, proxyproto.REJECT)

func newProxyListener(lis net.Listener) *proxyproto.Listener {
	return &proxyproto.Listener{ //nolint:exhaustruct
		Listener:   lis,
		ConnPolicy: loopbackProxyPolicy,
	}
}

// Close stops the SSH server and drops every connection it is holding.
func (s *Server) Close() error {
	return s.sshd.Close()
}

func (s *Server) ListenAndServe() error {
	log.WithFields(log.Fields{
		"addr": s.sshd.Addr,
	}).Info("ssh server listening")

	lc := net.ListenConfig{KeepAliveConfig: keepAlive} //nolint:exhaustruct

	list, err := lc.Listen(context.Background(), "tcp", s.sshd.Addr)
	if err != nil {
		log.WithError(err).Error("failed to listen an serve the TCP server")

		return err
	}

	proxy := newProxyListener(list)
	defer proxy.Close() //nolint:errcheck

	return s.sshd.Serve(proxy)
}
