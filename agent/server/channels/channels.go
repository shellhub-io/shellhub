package channels

// SSH channels supported by the SSH server.
//
// An SSH channel refers to a communication link established between a client and a server. SSH channels are multiplexed
// over a single encrypted connection, facilitating concurrent and secure communication for various purposes.
//
// SSH_MSG_CHANNEL_OPEN
//
// Check www.ietf.org/rfc/rfc4254.txt for more information.
const (
	// ChannelSession refers to a type of SSH channel that is established between a client and a server for interactive
	// shell sessions or command execution. SSH channels are used to multiplex multiple logical communication channels
	// over a single SSH connection.
	//
	// Check www.ietf.org/rfc/rfc4254.txt at section 6.1 for more information.
	ChannelSession string = "session"
	// ChannelDirectTcpip is the channel type in SSH is used to establish a direct TCP/IP connection between the SSH
	// client and a target host through the SSH server. This channel type allows the client to initiate a connection to
	// a specific destination host and port, and the SSH server acts as a bridge to facilitate this connection.
	//
	// Check www.ietf.org/rfc/rfc4254.txt at section 7.2 for more information.
	ChannelDirectTcpip string = "direct-tcpip"
	// ChannelSRDP is the channel type used for SRDP connections over SSH.
	ChannelSRDP string = "srdp"
)

type Data struct {
	Display string
}
