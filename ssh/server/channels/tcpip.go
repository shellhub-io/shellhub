package channels

import (
	"io"
	"net"
	"strconv"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	"github.com/shellhub-io/shellhub/ssh/session"
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
)

func TunnelDefaultDirectTCPIPHandler(server *gliderssh.Server, _ *gossh.ServerConn, newChan gossh.NewChannel, ctx gliderssh.Context) {
	target := metadata.RestoreTarget(ctx)

	log.WithFields(log.Fields{
		"username": target.Username,
		"sshid":    target.Data,
	}).Info("handling direct-tcpip channel")

	type channelData struct {
		DestAddr   string
		DestPort   uint32
		OriginAddr string
		OriginPort uint32
	}

	data := new(channelData)
	if err := gossh.Unmarshal(newChan.ExtraData(), data); err != nil {
		newChan.Reject(gossh.ConnectionFailed, "faild to parse forward data: "+err.Error()) //nolint:errcheck
		log.WithError(err).WithFields(log.Fields{
			"username":    target.Username,
			"sshid":       target.Data,
			"origin_port": data.OriginAddr,
			"origin_addr": data.OriginPort,
			"dest_port":   data.DestPort,
			"dest_addr":   data.DestAddr,
		}).Error("failed to parse forward data")

		return
	}

	if server.LocalPortForwardingCallback == nil || !server.LocalPortForwardingCallback(ctx, data.DestAddr, data.DestPort) {
		newChan.Reject(gossh.Prohibited, "port forwarding is disabled") //nolint:errcheck
		log.WithFields(log.Fields{
			"username":    target.Username,
			"sshid":       target.Data,
			"origin_port": data.OriginAddr,
			"origin_addr": data.OriginPort,
			"dest_port":   data.DestPort,
			"dest_addr":   data.DestAddr,
		}).Info("port forwarding is disabled")

		return
	}

	dest := net.JoinHostPort(data.DestAddr, strconv.FormatInt(int64(data.DestPort), 10))

	// NOTE: Certain SSH connections may not necessitate a dedicated handler, such as an SSH handler.
	// In such instances, a new connection to the agent is generated and saved in the metadata for
	// subsequent use.
	// An illustrative scenario is when the SSH connection is initiated with the "-N" flag.
	connection := ctx.Value("session").(*session.Session).AgentClient

	agent, err := connection.Dial("tcp", dest)
	if err != nil {
		newChan.Reject(gossh.ConnectionFailed, "failed dialing the agent to host and port: "+err.Error()) //nolint:errcheck
		log.WithError(err).WithFields(log.Fields{
			"username":    target.Username,
			"sshid":       target.Data,
			"origin_port": data.OriginAddr,
			"origin_addr": data.OriginPort,
			"dest_port":   data.DestPort,
			"dest_addr":   data.DestAddr,
		}).Error("failed dialing the agent to host and port")

		return
	}

	channel, reqs, err := newChan.Accept()
	if err != nil {
		newChan.Reject(gossh.ConnectionFailed, "failed accepting the channel: "+err.Error()) //nolint:errcheck
		log.WithError(err).WithFields(log.Fields{
			"username":    target.Username,
			"sshid":       target.Data,
			"origin_port": data.OriginAddr,
			"origin_addr": data.OriginPort,
			"dest_port":   data.DestPort,
			"dest_addr":   data.DestAddr,
		}).Error("failed accepting the channel")

		return
	}

	go gossh.DiscardRequests(reqs)

	log.WithFields(log.Fields{
		"username":    target.Username,
		"sshid":       target.Data,
		"origin_port": data.OriginAddr,
		"origin_addr": data.OriginPort,
		"dest_port":   data.DestPort,
		"dest_addr":   data.DestAddr,
	}).Info("piping data between client and agent")

	// TODO: control the running state of these goroutines.
	go func() {
		log.WithFields(log.Fields{
			"username":    target.Username,
			"sshid":       target.Data,
			"origin_port": data.OriginAddr,
			"origin_addr": data.OriginPort,
			"dest_port":   data.DestPort,
			"dest_addr":   data.DestAddr,
		}).Debug("copying data from client to agent")

		defer channel.Close()
		io.Copy(channel, agent) //nolint:errcheck
	}()
	go func() {
		log.WithFields(log.Fields{
			"username":    target.Username,
			"sshid":       target.Data,
			"origin_port": data.OriginAddr,
			"origin_addr": data.OriginPort,
			"dest_port":   data.DestPort,
			"dest_addr":   data.DestAddr,
		}).Debug("copying data from agent to client")

		defer channel.Close()
		io.Copy(agent, channel) //nolint:errcheck
	}()
}
