package session

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Masterminds/semver/v3"
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/devicekey"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// AgentChannel represents a channel open between agent and server.
type AgentChannel struct {
	// Channel is an open channel for communication between the agent and the server.
	Channel gossh.Channel
	// Requests is the channel to handle SSH requests.
	Requests <-chan *gossh.Request

	closedByGateway atomic.Bool
}

// Close closes the underlying agent channel connection. It records that the gateway closed it
// before closing, so [AgentChannel.ClosedByGateway] already reports true by the time Requests
// closes. It returns the error from closing the underlying channel, which a channel already
// closed also returns, and records the close either way.
func (a *AgentChannel) Close() error {
	a.closedByGateway.Store(true)

	return a.Channel.Close()
}

// ClosedByGateway reports whether the gateway has closed this channel through [AgentChannel.Close].
func (a *AgentChannel) ClosedByGateway() bool {
	return a.closedByGateway.Load()
}

// Agent represents a connection to an agent.
type Agent struct {
	conn     net.Conn
	connMu   sync.Mutex
	client   *gossh.Client
	mu       sync.Mutex
	channels map[int]*AgentChannel
	finished bool
}

// Close closes the underlying ssh client connection.
func (a *Agent) Close() error {
	a.mu.Lock()
	for _, channel := range a.channels {
		channel.Close() //nolint:errcheck
	}
	a.mu.Unlock()

	return a.client.Close()
}

// NewAgentChannel opens a session channel on the device and sets it on seat. It returns
// ErrSeatAlreadySet when seat already holds an agent channel, and the open's error unchanged
// otherwise, a *gossh.OpenChannelError when the device rejects the channel.
func (s *Session) NewAgentChannel(seat int) (*AgentChannel, error) {
	s.agent.mu.Lock()
	defer s.agent.mu.Unlock()

	if _, ok := s.agent.channels[seat]; ok {
		return nil, ErrSeatAlreadySet
	}

	channel, requests, err := s.agent.client.OpenChannel(sessionChannelType, nil)
	if err != nil {
		return nil, err
	}

	a := &AgentChannel{Channel: channel, Requests: requests}
	s.agent.channels[seat] = a

	return a, nil
}

// DropAgentChannel closes the agent channel on a seat and forgets it, so a seat
// whose client side failed to open does not keep a live channel behind it.
func (s *Session) DropAgentChannel(seat int) {
	s.agent.mu.Lock()
	defer s.agent.mu.Unlock()

	channel, ok := s.agent.channels[seat]
	if !ok {
		return
	}

	channel.Close() //nolint:errcheck
	delete(s.agent.channels, seat)
}

func (s *Session) clientConfig() (*gossh.ClientConfig, error) {
	hostKey, err := devicekey.ParsePublicKey(s.Device.PublicKey)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID, "device": s.Device.UID}).
			Error("refusing to connect to a device whose host key cannot be verified")

		return nil, ErrHostKey
	}

	return &gossh.ClientConfig{
		User:            s.Target.Username,
		HostKeyCallback: deviceHostKey(hostKey),
		Timeout:         sshconf.ConnectTimeout,
	}, nil
}

func (s *Session) attachAgent(ctx gliderssh.Context, conn gossh.Conn, chans <-chan gossh.NewChannel, reqs <-chan *gossh.Request) {
	ch := make(chan *gossh.Request)
	close(ch)

	s.agent.client = gossh.NewClient(conn, chans, ch)

	clientConn, _ := ctx.Value(connCtxKey{}).(net.Conn)

	go s.drainAgentRequests(ctx, reqs, clientConn)
}

func (s *Session) connect(ctx gliderssh.Context, cred credential) error {
	config, err := s.clientConfig()
	if err != nil {
		return err
	}

	return s.login(ctx, config, cred)
}

func (s *Session) login(ctx gliderssh.Context, config *gossh.ClientConfig, cred credential) error {
	const Addr = "tcp"

	if s.agent.conn == nil {
		if err := s.Dial(ctx); err != nil {
			return err
		}
	}

	if config.Timeout > 0 {
		if err := s.agent.conn.SetDeadline(time.Now().Add(config.Timeout)); err != nil { //nolint:forbidigo // a network deadline needs the wall clock
			log.WithError(err).
				WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
				Error("Error when trying to set dial deadline")

			return err
		}
	}

	greeted, err := s.greet(s.agent.conn)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
			Error("Error when trying to hear the agent out before the handshake")

		s.dropAgentConn()

		return err
	}

	config.Auth = s.authMethods(ctx, cred)

	conn, chans, reqs, err := gossh.NewClientConn(greeted, Addr, config)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": s.UID}).
			Error("Error when trying to create the client's connection")

		s.dropAgentConn()

		return err
	}

	if config.Timeout > 0 {
		if err := s.agent.conn.SetDeadline(time.Time{}); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
				Error("Error when trying to set dial deadline with Time{}")

			return err
		}
	}

	s.attachAgent(ctx, conn, chans, reqs)

	return nil
}

const keepAliveRequestPrefix = "keepalive"

const clientChannelsCloseTimeout = 5 * time.Second

// KeepAliveRequestType is the keepalive the gateway forwards to the client.
const KeepAliveRequestType = keepAliveRequestPrefix + "@shellhub.io"

func (s *Session) drainAgentRequests(ctx gliderssh.Context, reqs <-chan *gossh.Request, clientConn net.Conn) {
	logger := log.WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID})

	defer logger.Trace("agent global requests drained")

	for req := range reqs {
		if !strings.HasPrefix(req.Type, keepAliveRequestPrefix) {
			if req.WantReply {
				if err := req.Reply(false, nil); err != nil {
					logger.WithError(err).Warn("failed to refuse a global request from the agent")
				}
			}

			continue
		}

		if req.WantReply {
			if err := req.Reply(true, nil); err != nil {
				logger.WithError(err).Warn("failed to reply to the keepalive from the agent")
			}
		}

		if err := s.KeepAlive(ctx); err != nil {
			logger.WithError(err).Warn("failed to record the session as alive")
		}

		if conn, ok := ctx.Value(gliderssh.ContextKeyConn).(gossh.Conn); ok && conn != nil {
			if _, _, err := conn.SendRequest(KeepAliveRequestType, false, req.Payload); err != nil {
				logger.WithError(err).Warn("failed to forward the keepalive to the client")
			}
		}
	}

	if clientConn != nil {
		if !s.client.awaitChannels(clientChannelsCloseTimeout) {
			logger.Warn("closing the client connection before its channels were told the session ended")
		}

		if err := clientConn.Close(); err != nil {
			logger.WithError(err).Debug("failed to close the client connection after the agent connection ended")
		}
	}
}

func (s *Session) dropAgentConn() {
	if s.agent.conn == nil {
		return
	}

	if err := s.agent.conn.Close(); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
			Debug("failed to close the tunnel stream of a handshake that did not finish")
	}

	s.agent.connMu.Lock()
	s.agent.conn = nil
	s.agent.connMu.Unlock()
}

func (s *Session) adoptAgentConn(conn net.Conn) bool {
	s.agent.connMu.Lock()
	defer s.agent.connMu.Unlock()

	if s.agent.finished {
		return false
	}

	s.agent.conn = conn

	return true
}

var agentEarlyPayloadFloor = semver.MustParse("v0.27.0")

func (s *Session) greet(conn net.Conn) (net.Conn, error) { //nolint:ireturn // net.Conn is what the handshake takes
	if dialer.Pipelined(conn) && keepsEarlyPayload(s.Device) {
		return conn, nil
	}

	return awaitAgentGreeting(conn)
}

func keepsEarlyPayload(device *models.Device) bool {
	if device == nil || device.Info == nil {
		return false
	}

	if device.Info.Version == "latest" {
		return true
	}

	version, err := semver.NewVersion(device.Info.Version)
	if err != nil {
		return false
	}

	return !version.LessThan(agentEarlyPayloadFloor)
}

func awaitAgentGreeting(conn net.Conn) (net.Conn, error) { //nolint:ireturn // net.Conn is what the handshake takes
	var greeting [1]byte

	if _, err := io.ReadFull(conn, greeting[:]); err != nil {
		return nil, err
	}

	return &greetedConn{Conn: conn, rest: io.MultiReader(bytes.NewReader(greeting[:]), conn)}, nil
}

type greetedConn struct {
	net.Conn
	rest io.Reader
}

func (c *greetedConn) Read(p []byte) (int, error) {
	return c.rest.Read(p)
}

// Online reports whether the device has sent a heartbeat recently enough to be
// worth connecting to. It reads the flag the device lookup already computed, so
// it costs nothing and opens no tunnel, and it is a filter rather than a
// guarantee: a device that stopped answering within the heartbeat window still
// reads online here and fails later, when the session actually dials it.
func (s *Session) Online() bool {
	return s.Device != nil && s.Device.Online
}

// Dial establishes the underlying transport to the target device. For V1
// transports an HTTP GET request is issued (legacy reverse tunnel). For
// V2 transports the ProtoSSHOpen multistream selection and a JSON envelope
// with the session id are written together, without waiting for the agent
// to confirm the selection. After a nil return s.agent.conn is the raw stream
// on V1, and on V2 a stream whose first read checks that confirmation. Unless
// the stream was pipelined to an agent that keeps what arrives behind the
// header, it is not yet ready to be written to: the device has still to be
// heard from first, which connect does, because a version line written before
// the device has read the stream's header is taken into its header reader and
// never reaches its SSH server. It returns ErrDial joined with the dialer's
// error when the device cannot be reached, and ErrClientClosing, having closed
// the new stream and asked the device to close the session, when the session
// has finished by the time the stream arrives. On V2 an agent that refuses the
// selection or never answers it does not fail Dial: it fails the first read in
// connect, with multistream's ErrNotSupported or the connect deadline.
func (s *Session) Dial(ctx gliderssh.Context) error {
	var err error

	ctx.Lock()
	defer ctx.Unlock()

	conn, err := s.dialer.DialTo(ctx, s.Device.TenantID, s.Device.UID, dialer.SSHOpenTarget{SessionID: s.UID})
	if err != nil {
		log.WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
			WithError(err).
			Error("failed to open the ssh session: " + dialer.Describe(err))

		return errors.Join(ErrDial, err)
	}

	if !s.adoptAgentConn(conn) {
		_ = conn.Close()

		s.closingOnAgent.Go(s.closeOnAgent)

		return ErrClientClosing
	}

	return nil
}

const closeRequestTimeout = 10 * time.Second

func (s *Session) closeOnAgent() {
	logger := log.WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID})

	ctx, cancel := context.WithTimeout(context.Background(), closeRequestTimeout)
	defer cancel()

	conn, err := s.dialer.DialTo(ctx, s.Device.TenantID, s.Device.UID, dialer.SSHCloseTarget{SessionID: s.UID})
	if err != nil {
		logger.WithError(err).
			Warning("failed to ask the device to close the session: " + dialer.Describe(err))

		return
	}

	if err := conn.Close(); err != nil {
		logger.WithError(err).Warning("failed to close the connection the session close was sent over")
	}
}
