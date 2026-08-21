package channels

import (
	"context"
	"net"

	"github.com/shellhub-io/shellhub/server/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

const (
	// DirectTCPIPChannel is the channel type for direct-tcpip channels like "local port forwarding" and "dynamic
	// application-level port forwarding".
	//
	// Local port forwarding is used to forward a port from the client to the server, and dynamic application-level
	// is a method for securely tunneling and routing network traffic through an SSH connection to access remote
	// resources as if they were local.
	//
	// Example of local port forwarding: `ssh -L 8080:localhost:80 user@sshid`.
	//
	// Example of dynamic application-level port forwarding: `ssh -D 1080 user@sshid`.
	DirectTCPIPChannel = "direct-tcpip"
	SessionChannel     = "session"
)

// Session is what the channel handlers need from an SSH session: the operations
// for running a channel over it, plus the fields identifying it in a log.
//
// It is declared here, at the consumer, rather than in the session package. The
// handlers use a dozen operations out of a much larger surface, and naming just
// those is what lets them be exercised against a fake instead of a real SSH
// handshake and a live agent.
type Session interface {
	LogFields() log.Fields
	Event(t string, data any, seat int)

	NewSeat() (int, error)
	Seat(seat int) (session.Seat, bool)
	SetSeatPty(seat int, status bool)
	SetSeatType(seat int, kind string)

	NewAgentChannel(name string, seat int) (*session.AgentChannel, error)
	NewClientChannel(newChannel gossh.NewChannel, seat int) (*session.ClientChannel, error)
	DropAgentChannel(seat int)
	OpenAgentForwards(name string) <-chan gossh.NewChannel
	DialAgent(ctx context.Context, network, addr string) (net.Conn, error)
	CloseAgentWrite(seat int) error

	Recorded(seat int) error
	Announce(client gossh.Channel) error
}
