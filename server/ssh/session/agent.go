package session

import (
	"context"
	"net"

	"github.com/Masterminds/semver/v3"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// SeatTypeExec is the seat type recorded for an exec or subsystem request. The
// channel handlers set it; the session reads it back in [Session.CloseAgentWrite].
const SeatTypeExec = "exec"

var agentExecCloseFloor = semver.MustParse("v0.9.3")

var agentForwardHalfCloseFloor = semver.MustParse("v0.29.1")

// AuthRequestOpenSSHChannel is the channel the server opens to reach a forwarded agent.
//
// After a client has requested that a session have agent forwarding enabled, the server later may request a connection
// to the forwarded agent. The server does this by requesting a dedicated channel to communicate with the client's
// agent.
//
// https://www.ietf.org/archive/id/draft-miller-ssh-agent-11.html#section-4.2
const AuthRequestOpenSSHChannel = "auth-agent@openssh.com"

// ForwardAgent relays every agent channel the device opens to client, for a client that asked
// for agent forwarding. The first call starts relaying and every later call, from any seat, does
// nothing and ignores its arguments. Relaying ends when the connection to the device closes.
//
// Each agent use runs relay in a goroutine of its own, so relay must be safe to run
// concurrently, and both channels are closed once it returns. A use the client refuses is
// rejected towards the device, and relaying carries on with the next one.
func (s *Session) ForwardAgent(client gossh.Conn, relay func(device, client gossh.Channel)) {
	s.agentForwarding.Do(func() {
		forwards := s.agent.client.HandleChannelOpen(AuthRequestOpenSSHChannel)

		go s.relayAgentForwards(forwards, client, relay)
	})
}

func (s *Session) relayAgentForwards(forwards <-chan gossh.NewChannel, client gossh.Conn, relay func(device, client gossh.Channel)) {
	logger := log.WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID})

	for forward := range forwards {
		go relayAgentForward(logger, forward, client, relay)
	}

	logger.Trace("agent forwarding done")
}

func relayAgentForward(logger *log.Entry, forward gossh.NewChannel, client gossh.Conn, relay func(device, client gossh.Channel)) {
	clientChannel, clientRequests, err := client.OpenChannel(AuthRequestOpenSSHChannel, nil)
	if err != nil {
		logger.WithError(err).Debug("the client did not open its agent for a forwarded use")

		_ = forward.Reject(gossh.ConnectionFailed, "the client did not open its agent")

		return
	}

	defer clientChannel.Close() //nolint:errcheck // the use is over, and a channel the peer already closed reports an error nobody acts on

	go gossh.DiscardRequests(clientRequests)

	deviceChannel, deviceRequests, err := forward.Accept()
	if err != nil {
		logger.WithError(err).Error("failed to accept the agent forwarding channel from the device")

		return
	}

	defer deviceChannel.Close() //nolint:errcheck // the use is over, and a channel the peer already closed reports an error nobody acts on

	go gossh.DiscardRequests(deviceRequests)

	relay(deviceChannel, clientChannel)
}

// DialAgent opens a connection from the agent to addr, for a client that asked
// the agent to reach something on its own network.
func (s *Session) DialAgent(ctx context.Context, network, addr string) (net.Conn, error) {
	return s.agent.client.DialContext(ctx, network, addr)
}

// CloseAgentWrite closes the server's sending side of the seat's agent channel,
// once nothing more will be written to it.
//
// Agents older than [agentExecCloseFloor] hang after an exec when only the
// write side is closed, so for those the whole channel is closed instead. A
// version that does not parse is treated as new: closing the write side is the
// lesser failure, since a full close would truncate a live shell. A full close
// goes through [AgentChannel.Close], so the channel reports
// [AgentChannel.ClosedByGateway] from then on.
//
// It returns [ErrSeatNotFound] for a seat with no agent channel, and otherwise
// the error from closing the channel, which a channel already closed also
// returns.
func (s *Session) CloseAgentWrite(seat int) error {
	s.agent.mu.Lock()
	channel, ok := s.agent.channels[seat]
	s.agent.mu.Unlock()

	if !ok {
		return ErrSeatNotFound
	}

	if s.closesAgentOnExec(seat) {
		return channel.Close()
	}

	return channel.Channel.CloseWrite()
}

// CloseAgentForwardWrite tells the device that the client has finished sending
// on a port forward, by closing the write side of conn, a connection from
// [Session.DialAgent].
//
// Agents older than [agentForwardHalfCloseFloor] close the device's socket as
// soon as a forward's write side closes, which loses the device's reply, so
// for those it does nothing, returns nil, and conn stays open until the
// forward ends. A version that does not parse, and a device with no reported
// version, are treated as new. A conn without CloseWrite is closed fully.
// Otherwise it returns the error from closing conn, which a conn already
// closed also returns.
func (s *Session) CloseAgentForwardWrite(conn net.Conn) error {
	if s.agentOlderThan(agentForwardHalfCloseFloor) {
		return nil
	}

	if writeCloser, ok := conn.(interface{ CloseWrite() error }); ok {
		return writeCloser.CloseWrite()
	}

	return conn.Close()
}

func (s *Session) closesAgentOnExec(seat int) bool {
	if !s.agentOlderThan(agentExecCloseFloor) {
		return false
	}

	item, ok := s.seats.Get(seat)

	return ok && item.Type == SeatTypeExec
}

func (s *Session) agentOlderThan(floor *semver.Version) bool {
	if s.Device == nil || s.Device.Info == nil {
		return false
	}

	version, err := semver.NewVersion(s.Device.Info.Version)
	if err != nil || version == nil {
		return false
	}

	return version.LessThan(floor)
}
