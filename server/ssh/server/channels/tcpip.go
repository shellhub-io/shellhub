package channels

import (
	"errors"
	"io"
	"net"
	"strconv"
	"sync"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// forwardData is the direct-tcpip channel payload: the address the client wants
// the agent to reach, and the address it is forwarding from.
type forwardData struct {
	DestAddr   string `json:"dest_addr"`
	DestPort   uint32 `json:"dest_port"`
	OriginAddr string `json:"origin_addr"`
	OriginPort uint32 `json:"origin_port"`
}

func (d *forwardData) logFields() log.Fields {
	return log.Fields{
		"origin_port": d.OriginPort,
		"origin_addr": d.OriginAddr,
		"dest_port":   d.DestPort,
		"dest_addr":   d.DestAddr,
	}
}

// DefaultDirectTCPIPHandler is the channel's handler for direct-tcpip channels like "local port forwarding" and "dynamic
// application-level port forwarding".
func DefaultDirectTCPIPHandler(server *gliderssh.Server, _ *gossh.ServerConn, newChan gossh.NewChannel, ctx gliderssh.Context) {
	sess, state := session.ObtainSession(ctx)
	if sess == nil || !state.Established() {
		// See the same guard in the session channel handler: unreachable today,
		// and here the goroutine below sits outside the panic recovery.
		log.WithFields(log.Fields{"session": ctx.SessionID(), "state": state}).
			Error("direct-tcpip channel opened without an established session")

		newChan.Reject(gossh.ConnectionFailed, "session is not established") //nolint:errcheck

		return
	}

	directTCPIPChannel(ctx, sess, newChan, server.LocalPortForwardingCallback)
}

// directTCPIPChannel forwards one client-requested connection through the agent.
func directTCPIPChannel(ctx gliderssh.Context, sess Session, newChan gossh.NewChannel, allow gliderssh.LocalPortForwardingCallback) {
	logger := log.WithFields(sess.LogFields())

	logger.Trace("handling direct-tcpip channel")

	data := new(forwardData)
	if err := gossh.Unmarshal(newChan.ExtraData(), data); err != nil {
		newChan.Reject(gossh.ConnectionFailed, "failed to parse forward data: "+err.Error()) //nolint:errcheck
		logger.WithError(err).WithFields(data.logFields()).Error("failed to parse forward data")

		return
	}

	logger = logger.WithFields(data.logFields())

	if allow == nil || !allow(ctx, data.DestAddr, data.DestPort) {
		newChan.Reject(gossh.Prohibited, "port forwarding is disabled") //nolint:errcheck
		logger.Info("port forwarding is disabled")

		return
	}

	dest := net.JoinHostPort(data.DestAddr, strconv.FormatInt(int64(data.DestPort), 10))

	// NOTE: Certain SSH connections may not necessitate a dedicated handler, such as an SSH handler.
	// In such instances, a new connection to the agent is generated and saved in the metadata for
	// subsequent use.
	// An illustrative scenario is when the SSH connection is initiated with the "-N" flag.
	agent, err := sess.DialAgent(ctx, "tcp", dest)
	if err != nil {
		newChan.Reject(gossh.ConnectionFailed, "failed dialing the agent to host and port: "+err.Error()) //nolint:errcheck
		logger.WithError(err).Error("failed dialing the agent to host and port")

		return
	}

	defer agent.Close() //nolint:errcheck

	client, reqs, err := newChan.Accept()
	if err != nil {
		newChan.Reject(gossh.ConnectionFailed, "failed accepting the channel: "+err.Error()) //nolint:errcheck
		logger.WithError(err).Error("failed accepting the channel")

		return
	}

	defer client.Close() //nolint:errcheck

	go gossh.DiscardRequests(reqs)

	logger.Info("piping data between client and agent")

	wg := new(sync.WaitGroup)

	// TODO: control the running state of these goroutines.

	wg.Go(func() {
		logger.Trace("copying data from client to agent")

		if _, err := io.Copy(client, &deadReadGuard{r: agent}); err != nil && !errors.Is(err, io.EOF) {
			logger.WithError(err).Error("failed to copy data from agent to client")

			// Close both ends so the peer goroutine unblocks and wg.Wait can return.
			_ = agent.Close()
			_ = client.Close()

			return
		}
	})

	wg.Go(func() {
		logger.Trace("copying data from agent to client")

		if _, err := io.Copy(agent, &deadReadGuard{r: client}); err != nil && !errors.Is(err, io.EOF) {
			logger.WithError(err).Error("failed to copy data from client to agent")

			_ = agent.Close()
			_ = client.Close()

			return
		}
	})

	wg.Wait()

	logger.Trace("handling direct-tcpip finished")
}
