package session

import (
	"context"
	"errors"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

var (
	errCredentialKind = errors.New("the device asked for a credential of another kind")
	errWarmUnclaimed  = errors.New("no client authenticated in time to claim the warm connection")
)

var warmClaimWindow = 30 * time.Second

type warmup struct {
	vouched     bool
	credentials chan credential
	result      chan warmed
	taken       chan struct{}
}

type warmed struct {
	conn    net.Conn
	client  gossh.Conn
	chans   <-chan gossh.NewChannel
	reqs    <-chan *gossh.Request
	err     error
	final   bool
	vouched bool
	dialled bool
}

// Warm opens the tunnel to the device and runs the SSH handshake with it up to authentication,
// while the client is still authenticating to the server. The first [Session.Auth] that reaches
// the device hands its credential to that connection instead of dialling one, so the round trips
// to the device the handshake costs overlap the client's own.
//
// Nothing runs on the connection before Auth has authorized the client. It waits for the
// credential before it presents one, and it is closed when ctx ends without one; the exception
// is a stream the gateway vouches for, which the device admits during the handshake for the
// account the client named, before the client has proven anything, and which opens no channel
// until Auth has authorized it. A connection no client has claimed
// within warmClaimWindow is closed too, and the Auth that comes later dials again. When the client
// leaves without claiming it, the device is asked to close the session the stream opened. Warm
// must be called at most once, before the first Auth, from the goroutine that runs the handshake.
func (s *Session) Warm(ctx gliderssh.Context) {
	w := &warmup{
		vouched:     !s.IsLegacy(),
		credentials: make(chan credential, 1),
		result:      make(chan warmed, 1),
		taken:       make(chan struct{}),
	}

	s.warm = w

	go w.run(ctx, s)
}

func (w *warmup) deliver(cred credential) warmed {
	close(w.taken)
	w.credentials <- cred

	return <-w.result
}

func (w *warmup) run(ctx gliderssh.Context, s *Session) {
	logger := log.WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID})

	unclaimed := make(chan struct{})
	expiry := time.AfterFunc(warmClaimWindow, func() { close(unclaimed) })
	defer expiry.Stop()

	result := w.handshake(ctx, s, unclaimed)

	if result.err != nil {
		logger.WithError(result.err).Debug("the warm connection to the device did not complete")
	}

	select {
	case <-w.taken:
	case <-ctx.Done():
	case <-unclaimed:
		select {
		case <-w.taken:
		default:
			result.close()
			result = warmed{err: errWarmUnclaimed, dialled: result.dialled}

			select {
			case <-w.taken:
			case <-ctx.Done():
			}
		}
	}

	select {
	case <-w.taken:
		w.result <- result
	default:
		result.close()

		if result.dialled {
			s.closingOnAgent.Go(s.closeOnAgent)
		}

		w.result <- warmed{err: context.Cause(ctx)}
	}
}

func (s *Session) abandonAgentConn(w warmed) {
	w.close()

	s.closingOnAgent.Go(s.closeOnAgent)
}

func (s *Session) adoptWarmed(ctx gliderssh.Context, w warmed) error {
	if !s.adoptAgentConn(w.conn) {
		s.abandonAgentConn(w)

		return ErrClientClosing
	}

	s.attachAgent(ctx, w.client, w.chans, w.reqs)

	return nil
}

func (w warmed) close() {
	if w.client != nil {
		_ = w.client.Close()
	}

	if w.conn != nil {
		_ = w.conn.Close()
	}
}

func (w *warmup) handshake(ctx gliderssh.Context, s *Session, unclaimed <-chan struct{}) warmed {
	config, err := s.clientConfig()
	if err != nil {
		return warmed{err: err}
	}

	conn, err := s.dialer.DialTo(ctx, s.Device.TenantID, s.Device.UID, dialer.SSHOpenTarget{SessionID: s.UID, Login: s.agentLogin, Vouched: w.vouched})
	if err != nil {
		if waitedOut(err) {
			logger := log.WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID})
			logger.WithError(err).Error("failed to open the ssh session: " + dialer.Describe(err))
		}

		return warmed{err: errors.Join(ErrDial, err), final: waitedOut(err)}
	}

	timeout := config.Timeout
	deadline := func() {
		if timeout > 0 {
			_ = conn.SetDeadline(time.Now().Add(timeout)) //nolint:forbidigo // a network deadline needs the wall clock
		}
	}

	deadline()

	greeted, err := s.greet(conn)
	if err != nil {
		_ = conn.Close()

		return warmed{err: err, final: waitedOut(err), dialled: true}
	}

	device := sniffVersion(greeted)

	var delivered bool

	await := sync.OnceValues(func() (credential, error) {
		_ = conn.SetDeadline(time.Time{})

		select {
		case cred := <-w.credentials:
			delivered = true

			deadline()

			return cred, nil
		case <-ctx.Done():
			return credential{}, context.Cause(ctx)
		case <-unclaimed:
			return credential{}, errWarmUnclaimed
		}
	})

	config.Auth = s.authMethods(ctx, await, device)
	config.Timeout = 0

	client, chans, reqs, err := gossh.NewClientConn(device, "tcp", config)
	if err != nil {
		_ = conn.Close()

		return warmed{err: err, final: waitedOut(err) || (delivered && refusedByDevice(err)), dialled: true}
	}

	if timeout > 0 {
		_ = conn.SetDeadline(time.Time{})
	}

	return warmed{conn: conn, client: client, chans: chans, reqs: reqs, vouched: w.vouched && !delivered, dialled: true}
}

func waitedOut(err error) bool {
	return errors.Is(err, dialer.ErrUnreachable) || errors.Is(err, os.ErrDeadlineExceeded)
}

func refusedByDevice(err error) bool {
	return errors.Is(err, errCredentialKind) ||
		errors.Is(err, ErrHostKey) ||
		strings.Contains(err.Error(), "unable to authenticate")
}
