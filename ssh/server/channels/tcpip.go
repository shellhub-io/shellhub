package channels

import (
	"io"
	"net"
	"strconv"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
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
)

// DefaultTCPIPHandler is the default handler for DirectTCPIPChannel and DynamicTCPIPChannel channels.
//
// It will reject the channel if the LocalPortForwardingCallback is not set or returns false.
// Otherwise, it will dial the agent and proxy the channel.
func DefaultTCPIPHandler(server *gliderssh.Server, _ *gossh.ServerConn, newChan gossh.NewChannel, ctx gliderssh.Context) {
	type channelData struct {
		DestAddr   string
		DestPort   uint32
		OriginAddr string
		OriginPort uint32
	}

	data := channelData{}
	if err := gossh.Unmarshal(newChan.ExtraData(), &data); err != nil {
		newChan.Reject(gossh.ConnectionFailed, "error parsing forward data: "+err.Error()) //nolint:errcheck

		return
	}

	if server.LocalPortForwardingCallback == nil || !server.LocalPortForwardingCallback(ctx, data.DestAddr, data.DestPort) {
		newChan.Reject(gossh.Prohibited, "port forwarding is disabled") //nolint:errcheck

		return
	}

	if !metadata.RestoreEstablished(ctx) {
		newChan.Reject(gossh.Prohibited, "connection between server and agent is not established yet") //nolint:errcheck

		return
	}

	dest := net.JoinHostPort(data.DestAddr, strconv.FormatInt(int64(data.DestPort), 10))

	agent := metadata.RestoreAgent(ctx)
	if agent == nil {
		newChan.Reject(gossh.ConnectionFailed, "error restoring the agent") //nolint:errcheck

		return
	}

	dialed, err := agent.Dial("tcp", dest)
	if err != nil {
		newChan.Reject(gossh.ConnectionFailed, "error dialing the agent to host and port: "+err.Error()) //nolint:errcheck

		return
	}

	channel, reqs, err := newChan.Accept()
	if err != nil {
		newChan.Reject(gossh.ConnectionFailed, "error accepting the channel: "+err.Error()) //nolint:errcheck

		return
	}

	go gossh.DiscardRequests(reqs)

	go func() {
		defer channel.Close()
		io.Copy(channel, dialed) //nolint:errcheck
	}()
	go func() {
		defer channel.Close()
		io.Copy(dialed, channel) //nolint:errcheck
	}()
}
