package session

import (
	"context"
	"net"

	"github.com/Masterminds/semver/v3"
	gossh "golang.org/x/crypto/ssh"
)

// SeatTypeExec is the seat type recorded for an exec or subsystem request. The
// channel handlers set it; the session reads it back in [Session.CloseAgentWrite].
const SeatTypeExec = "exec"

// agentExecCloseFloor is the first agent release that survives a write-side
// close after an exec. Below it the agent waits on a read that never ends, so
// the whole channel has to go.
var agentExecCloseFloor = semver.MustParse("v0.9.3")

// OpenAgentForwards returns the channels the agent opens back towards the
// server for the given channel type, used to relay agent forwarding to the
// client.
func (s *Session) OpenAgentForwards(name string) <-chan gossh.NewChannel {
	return s.agent.client.HandleChannelOpen(name)
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
// lesser failure, since a full close would truncate a live shell.
func (s *Session) CloseAgentWrite(seat int) error {
	s.agent.mu.Lock()
	channel, ok := s.agent.channels[seat]
	s.agent.mu.Unlock()

	if !ok {
		return ErrSeatNotFound
	}

	if s.closesAgentOnExec(seat) {
		return channel.Channel.Close()
	}

	return channel.Channel.CloseWrite()
}

func (s *Session) closesAgentOnExec(seat int) bool {
	if s.Device == nil || s.Device.Info == nil {
		return false
	}

	version, err := semver.NewVersion(s.Device.Info.Version)
	if err != nil || version == nil {
		return false
	}

	if !version.LessThan(agentExecCloseFloor) {
		return false
	}

	item, ok := s.seats.Get(seat)

	return ok && item.Type == SeatTypeExec
}
