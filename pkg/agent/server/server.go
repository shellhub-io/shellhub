package server

import (
	"net"
	"os/exec"
	"sync"
	"time"

	dockerclient "github.com/docker/docker/client"
	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/agent/server/modes"
	"github.com/shellhub-io/shellhub/pkg/agent/server/modes/connector"
	"github.com/shellhub-io/shellhub/pkg/agent/server/modes/host"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// List of SSH subsystems names supported by the agent.
const (
	// SFTPSubsystemName is the name of the SFTP subsystem.
	SFTPSubsystemName = "sftp"
)

type sshConn struct {
	net.Conn
	closeCallback func(string)
	ctx           gliderssh.Context
}

func (c *sshConn) Close() error {
	if id, ok := c.ctx.Value(gliderssh.ContextKeySessionID).(string); ok {
		c.closeCallback(id)
	}

	return c.Conn.Close()
}

type Server struct {
	sshd               *gliderssh.Server
	api                client.Client
	authData           *models.DeviceAuthResponse
	cmds               map[string]*exec.Cmd
	Sessions           map[string]net.Conn
	deviceName         string
	containerID        string
	mu                 sync.Mutex
	keepAliveInterval  int
	singleUserPassword string
	// mode is the mode of the server, identifing where and how the SSH's server is running.
	//
	// For example, the [modes.HostMode] means that the SSH's server runs in the host machine, using the host
	// `/etc/passwd`, `/etc/shadow`, redirecting the SSH's connection to the device sdin, stdout and stderr and etc.
	//
	// Check the [modes] package for more information.
	mode modes.Mode
	// authenticator contains methods used by the server to authenticate the user on the device and on the ShellHub
	// server.
	authenticator modes.Authenticator
	// sessioner contains methods used by the server to handle different types of sessions.
	//
	// sessioner also has the [modes.Subsystemer] interface, which contains methods used by the server to handle
	// different types of subsystems.
	sessioner modes.Sessioner
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

// NewServer creates a new server SSH agent server.
func NewServer(api client.Client, authData *models.DeviceAuthResponse, privateKey string, keepAliveInterval int, singleUserPassword string, mode modes.Mode) *Server {
	server := &Server{
		api:                api,
		authData:           authData,
		cmds:               make(map[string]*exec.Cmd),
		Sessions:           make(map[string]net.Conn),
		keepAliveInterval:  keepAliveInterval,
		singleUserPassword: singleUserPassword,
		mode:               mode,
	}

	switch server.mode {
	case modes.HostMode:
		server.authenticator = host.NewAuthenticator(api, authData, singleUserPassword, &server.deviceName)
		server.sessioner = host.NewSessioner(&server.deviceName, server.cmds)
	case modes.ConnectorMode:
		cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
		if err != nil {
			log.Fatal(err)
		}

		server.authenticator = connector.NewAuthenticator(api, cli, authData, &server.containerID)
		server.sessioner = connector.NewSessioner(&server.containerID, cli)
	default:
		log.WithFields(log.Fields{
			"mode": server.mode,
		}).Fatal("Invalid server mode")
	}

	server.sshd = &gliderssh.Server{
		PasswordHandler:        server.passwordHandler,
		PublicKeyHandler:       server.publicKeyHandler,
		Handler:                server.sessionHandler,
		SessionRequestCallback: server.sessionRequestCallback,
		SubsystemHandlers: map[string]gliderssh.SubsystemHandler{
			SFTPSubsystemName: server.sftpSubsystemHandler,
		},
		ConnCallback: func(ctx gliderssh.Context, conn net.Conn) net.Conn {
			closeCallback := func(id string) {
				server.mu.Lock()
				defer server.mu.Unlock()

				if v, ok := server.cmds[id]; ok {
					v.Process.Kill() // nolint:errcheck
					delete(server.cmds, id)
				}
			}

			return &sshConn{conn, closeCallback, ctx}
		},
		LocalPortForwardingCallback: func(ctx gliderssh.Context, destinationHost string, destinationPort uint32) bool {
			return true
		},
		ReversePortForwardingCallback: func(ctx gliderssh.Context, destinationHost string, destinationPort uint32) bool {
			return false
		},
		ChannelHandlers: map[string]gliderssh.ChannelHandler{
			ChannelSession:     gliderssh.DefaultSessionHandler,
			ChannelDirectTcpip: gliderssh.DirectTCPIPHandler,
		},
	}

	err := server.sshd.SetOption(gliderssh.HostKeyFile(privateKey))
	if err != nil {
		log.Warn(err)
	}

	return server
}

// startKeepAlive sends a keep alive message to the server every in keepAliveInterval seconds.
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
				if _, _, err := conn.SendRequest("keepalive", false, nil); err != nil {
					log.Error(err)
				}
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
	// RequestTypeUnknown is the request type for unknown.
	//
	// It is not a valid request type documentated by SSH's RFC, but it can be useful to identify the request type when
	// it is not known.
	RequestTypeUnknown = "unknown"
)

func (s *Server) sessionRequestCallback(session gliderssh.Session, requestType string) bool {
	session.Context().SetValue("request_type", requestType)

	return true
}

func (s *Server) HandleConn(conn net.Conn) {
	s.sshd.HandleConn(conn)
}

func (s *Server) SetDeviceName(name string) {
	s.deviceName = name
}

func (s *Server) SetContainerID(id string) {
	s.containerID = id
}

func (s *Server) CloseSession(id string) {
	if session, ok := s.Sessions[id]; ok {
		session.Close()
		delete(s.Sessions, id)
	}
}

func (s *Server) ListenAndServe() error {
	return s.sshd.ListenAndServe()
}
