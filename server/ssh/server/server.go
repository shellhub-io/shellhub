package server

import (
	"errors"
	"net"
	"os"
	"runtime/debug"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/pires/go-proxyproto"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
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

type Options struct {
	ConnectTimeout time.Duration
	// HostKeyFile is the path to the SSH host key. It is deliberately not read
	// from PRIVATE_KEY: the API signs its JWTs with a key under that name, and
	// silently adopting it would change the host key fingerprint of every
	// existing installation.
	HostKeyFile string
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
	newSession func(ctx gliderssh.Context, d *dialer.Dialer, service services.Service, api internalclient.Client, handoff *webhandoff.Store) (*session.Session, error)
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
func newBannerHandler(d *dialer.Dialer, service services.Service, api internalclient.Client, handoff *webhandoff.Store) gliderssh.BannerHandler {
	return newBannerHandlerWithDeps(d, service, api, handoff, defaultBannerDeps())
}

// newBannerHandlerWithDeps is the testable core of newBannerHandler. Callers
// supply a bannerDeps to stub out network-dependent operations.
func newBannerHandlerWithDeps(d *dialer.Dialer, service services.Service, api internalclient.Client, handoff *webhandoff.Store, deps bannerDeps) gliderssh.BannerHandler {
	return func(ctx gliderssh.Context) (message string) {
		logger := log.WithFields(
			log.Fields{
				"uid":   ctx.SessionID(),
				"sshid": ctx.User(),
			})

		// gliderlabs/ssh runs this handler in the per-connection goroutine with
		// no recover of its own, so an unhandled panic here crashes the whole
		// service. Contain it and fail just this connection.
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

		sess, err := deps.newSession(ctx, d, service, api, handoff)
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

		// No pre-auth banner on success. In identity mode the enrollment URL is
		// sent later, mid-handshake, only if the presented key is unenrolled, so
		// an enrolled key connects cleanly with no banner.
		return ""
	}
}

var (
	// errNoneAuthUnsupported fails the `none` method so the server falls back to
	// advertising the standard publickey+password methods (legacy behavior).
	errNoneAuthUnsupported = errors.New("ssh: none authentication is not supported")
	// errPermissionDenied denies a publickey attempt in identity mode.
	errPermissionDenied = errors.New("ssh: permission denied")
)

// newServerConfigCallback builds the per-connection SSH server config. It enables
// the `none` auth method so the offered authentication methods can be decided
// AFTER the SSHID — and thus the namespace's access mode — is known, which only
// happens once the client sends its username. The banner handler runs before
// `none` is processed (x/crypto sends the banner first), so the session and its
// access mode are already resolved here.
//
// In identity mode the identity is an SSH key, so the connection is restricted to
// publickey only via a PartialSuccessError: password is never advertised, so a
// stock OpenSSH client never prompts for one, and a keyless client gets a clean
// "publickey" denial instead of a password prompt. Legacy namespaces are
// untouched — `none` fails and the standard publickey+password set is advertised
// exactly as before, so password login keeps working there.
//
// The gliderssh fork overlays the host key and the publickey/password/banner
// callbacks (from the server's handlers) onto the returned config; it never
// touches NoClientAuth/NoClientAuthCallback, so what is set here survives.
func newServerConfigCallback(ctx gliderssh.Context) *gossh.ServerConfig {
	return &gossh.ServerConfig{ //nolint:exhaustruct
		NoClientAuth: true,
		// Capture the pre-auth connection so the enrollment/step-up banner can be
		// sent mid-handshake, after the presented key is resolved, instead of
		// unconditionally up front.
		PreAuthConnCallback: func(conn gossh.ServerPreAuthConn) {
			session.StorePreAuthConn(ctx, conn)
		},
		NoClientAuthCallback: func(gossh.ConnMetadata) (*gossh.Permissions, error) {
			sess, state := session.ObtainSession(ctx)
			if state < session.StateEvaluated || sess.Web || !sess.IsIdentityMode() {
				return nil, errNoneAuthUnsupported
			}

			return nil, &gossh.PartialSuccessError{
				Next: gossh.ServerAuthCallbacks{ //nolint:exhaustruct
					PublicKeyCallback: func(_ gossh.ConnMetadata, key gossh.PublicKey) (*gossh.Permissions, error) {
						if ok := auth.PublicKeyHandler(ctx, key); !ok {
							return nil, errPermissionDenied
						}

						return ctx.Permissions().Permissions, nil
					},
				},
			}
		},
	}
}

func NewServer(dialer *dialer.Dialer, service services.Service, api internalclient.Client, handoff *webhandoff.Store, opts *Options) (*Server, error) {
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
		ServerConfigCallback: newServerConfigCallback,
		BannerHandler:        newBannerHandler(dialer, service, api, handoff),
		PasswordHandler:      auth.PasswordHandler,
		PublicKeyHandler:     auth.PublicKeyHandler,
		// Channels form the foundation of secure communication between clients and servers in SSH connections. A
		// channel, in the context of SSH, is a logical conduit through which data travels securely between the client
		// and the server. SSH channels serve as the infrastructure for executing commands, establishing shell sessions,
		// and securely forwarding network services.
		ChannelHandlers: map[string]gliderssh.ChannelHandler{
			channels.SessionChannel:     recoverChannel(channels.SessionChannel, channels.DefaultSessionHandler()),
			channels.DirectTCPIPChannel: recoverChannel(channels.DirectTCPIPChannel, channels.DefaultDirectTCPIPHandler),
		},
		// Answers the web terminal bridge with this connection's session UID, so a
		// client-side recording can be tied to its server session.
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

// recoverChannel keeps a panic in one channel from reaching the runtime. The
// upstream server runs these in a per-connection goroutine with no recover of
// its own, and since the HTTP server shares this process a single bad session
// would take the API down with it.
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

// loopbackProxyPolicy honours a PROXY protocol header only when the real TCP
// peer is loopback, and rejects the connection outright otherwise.
//
// Nothing produces such a header today: the web terminal bridge dials localhost
// as a plain SSH client and passes the browser's address through the handoff
// store instead. This is hardening, not a live path — the source address feeds
// the firewall ip_address rule, the audit trail and the geoip record, so a peer
// that is not us must never be able to declare one.
var loopbackProxyPolicy = proxyproto.MustPolicyFromRanges([]string{"127.0.0.0/8", "::1/128"}, proxyproto.USE, proxyproto.REJECT)

func newProxyListener(lis net.Listener) *proxyproto.Listener {
	return &proxyproto.Listener{ // nolint: exhaustruct
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

	list, err := net.Listen("tcp", s.sshd.Addr)
	if err != nil {
		log.WithError(err).Error("failed to listen an serve the TCP server")

		return err
	}

	proxy := newProxyListener(list)
	defer proxy.Close() //nolint:errcheck

	return s.sshd.Serve(proxy)
}
