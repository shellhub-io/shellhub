package server

import (
	"net"
	"sync"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/agent/server/modes"
	"github.com/shellhub-io/shellhub/agent/server/modes/host"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// List of SSH subsystems names supported by the agent.
const (
	// SFTPSubsystemName is the name of the SFTP subsystem.
	SFTPSubsystemName = "sftp"
)

// Server is the agent's SSH server. It does not listen on a public port: the connections it
// serves arrive over the tunnel the agent holds open to the ShellHub server.
type Server struct {
	sshd              *gliderssh.Server
	ContainerID       string
	keepAliveInterval uint32

	mode     modes.Mode
	Sessions sync.Map
}

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
)

// Feature is a bit set of the optional capabilities a server is willing to serve.
type Feature uint

const (
	// NoFeature no features enable.
	NoFeature Feature = 0
	// LocalPortForwardFeature enable local port forward feature.
	LocalPortForwardFeature Feature = iota << 1
	// ReversePortForwardFeature enable reverse port forward feature.
	ReversePortForwardFeature
)

// Config stores configuration needs for the SSH server.
type Config struct {
	// PrivateKey is the path for the SSH server private key.
	PrivateKey string
	// KeepAliveInterval stores the time between each SSH keep alive request.
	KeepAliveInterval uint32
	// Features list of featues on SSH server.
	Features Feature
}

// NewServer creates a new server SSH agent server.
func NewServer(mode modes.Mode, cfg *Config) *Server {
	server := &Server{
		mode:              mode,
		keepAliveInterval: cfg.KeepAliveInterval,
		Sessions:          sync.Map{},
	}

	server.sshd = &gliderssh.Server{
		PasswordHandler:        server.passwordHandler,
		PublicKeyHandler:       server.publicKeyHandler,
		Handler:                server.sessionHandler,
		SessionRequestCallback: server.sessionRequestCallback,
		SubsystemHandlers: map[string]gliderssh.SubsystemHandler{
			SFTPSubsystemName: server.sftpSubsystemHandler,
		},
		LocalPortForwardingCallback: func(_ gliderssh.Context, _ string, _ uint32) bool {
			return cfg.Features&LocalPortForwardFeature > 0
		},
		ReversePortForwardingCallback: func(_ gliderssh.Context, _ string, _ uint32) bool {
			return cfg.Features&ReversePortForwardFeature > 0
		},
		ChannelHandlers: map[string]gliderssh.ChannelHandler{
			ChannelSession:     gliderssh.DefaultSessionHandler,
			ChannelDirectTcpip: gliderssh.DirectTCPIPHandler,
		},
	}

	if _, ok := mode.(*host.Mode); ok {
		server.sshd.PtyHandler = func(ctx gliderssh.Context, sess gliderssh.Session, pty gliderssh.Pty) (func() error, error) {
			closer, err := gliderssh.AllocatePtyHandler(ctx, sess, pty)
			if err != nil {
				entry := log.WithError(err)
				if hint := host.PtyFailureHint(err); hint != "" {
					entry = entry.WithField("hint", hint)
				}

				entry.Error("failed to allocate a pty")
			}

			return closer, err
		}
	}

	err := server.sshd.SetOption(gliderssh.HostKeyFile(cfg.PrivateKey))
	if err != nil {
		log.Warn(err)
	}

	return server
}

func (s *Server) startKeepAliveLoop(session gliderssh.Session) {
	interval := time.Duration(s.keepAliveInterval) * time.Second

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.WithFields(log.Fields{
		"interval": interval,
	}).Debug("Starting keep alive loop")

loop:
	for {
		select {
		case <-ticker.C:
			if conn, ok := session.Context().Value(gliderssh.ContextKeyConn).(gossh.Conn); ok {
				ok, _, err := conn.SendRequest("keepalive", true, nil)
				if err != nil {
					log.Error(err)
				}

				log.WithField("reply", ok).Info("keepalive sent with WantReply=true")
			}
		case <-session.Context().Done():
			log.Debug("Stopping keep alive loop after session closed")
			ticker.Stop()

			break loop
		}
	}
}

// List of request types that are supported by SSH.
//
// Once the session has been set up, a program is started at the remote end.  The program can be a shell, an application
// program, or a subsystem with a host-independent name.  Only one of these requests can succeed per channel.
//
// Check www.ietf.org/rfc/rfc4254.txt at section 6.5 for more information.
const (
	// RequestTypeShell is the request type for shell.
	RequestTypeShell = "shell"
	// RequestTypeExec is the request type for exec.
	RequestTypeExec = "exec"
	// RequestTypeSubsystem is the request type for any subsystem.
	RequestTypeSubsystem = "subsystem"
)

func (s *Server) sessionRequestCallback(session gliderssh.Session, requestType string) bool {
	session.Context().SetValue("request_type", requestType)

	go s.startKeepAliveLoop(session)

	return true
}

// HandleConn serves conn as an SSH connection. It is how the tunnel hands a session over.
func (s *Server) HandleConn(conn net.Conn) {
	s.sshd.HandleConn(conn)
}

// SetContainerID records the container sessions should target in connector mode.
func (s *Server) SetContainerID(id string) {
	s.ContainerID = id
}

// CloseSession tears down the session with the given id, if it is still open.
func (s *Server) CloseSession(id string) {
	if session, ok := s.Sessions.Load(id); ok {
		if conn, ok := session.(net.Conn); ok {
			_ = conn.Close()
		}

		s.Sessions.Delete(id)
	}
}
