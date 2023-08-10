package server

import (
	"net"
	"os/exec"
	"sync"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/agent/server/modes"
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
	mu                 sync.Mutex
	keepAliveInterval  int
	singleUserPassword string
	// mode is the mode of the server.
	//
	// mode is used to identify where and how the SSH's server is running. For example, the modes.HostMode means
	// that the SSH's server runs in the host machine, using the host `/etc/passwd`, `/etc/shadow`, redirecting the
	// SSH's connection to the device sdin, stdout and stderr and etc.
	mode modes.Mode
	// authenticator contains methods by the server to authenticate the user on the device and on the ShellHub server.
	authenticator modes.Authenticator
	// sessioner contains methods used by the server to handle different types of sessions.
	//
	// sessioner also has the subsystemer interface, which contains methods used by the server to handle different
	// types of subsystems.
	sessioner modes.Sessioner
	// features is the list of features enabled by server mode.
	features modes.Features
}

// NewServer creates a new server SSH agent server.
func NewServer(api client.Client, authData *models.DeviceAuthResponse, privateKey string, keepAliveInterval int, singleUserPassword string) *Server {
	server := &Server{
		api:                api,
		authData:           authData,
		cmds:               make(map[string]*exec.Cmd),
		Sessions:           make(map[string]net.Conn),
		keepAliveInterval:  keepAliveInterval,
		singleUserPassword: singleUserPassword,
		mode:               modes.HostMode,
	}

	switch server.mode {
	case modes.HostMode:
		server.authenticator = host.NewAuthenticator(api, authData, singleUserPassword, &server.deviceName)
		server.sessioner = host.NewSessioner(&server.deviceName, server.cmds)
		server.features = host.Features
	}

	server.sshd = &gliderssh.Server{
		PasswordHandler:        server.passwordHandler,
		PublicKeyHandler:       server.publicKeyHandler,
		Handler:                server.sessionHandler,
		SessionRequestCallback: server.sessionRequestCallback,
		RequestHandlers:        gliderssh.DefaultRequestHandlers,
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
			return server.features.IsFeatureEnabled(modes.FeatureLocalPortForwarding)
		},
		ReversePortForwardingCallback: func(ctx gliderssh.Context, destinationHost string, destinationPort uint32) bool {
			return server.features.IsFeatureEnabled(modes.FeatureReversePortForwarding)
		},
		ChannelHandlers: map[string]gliderssh.ChannelHandler{
			"session":       gliderssh.DefaultSessionHandler,
			"direct-tcpip":  gliderssh.DirectTCPIPHandler,
			"dynamic-tcpip": gliderssh.DirectTCPIPHandler,
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

func (s *Server) CloseSession(id string) {
	if session, ok := s.Sessions[id]; ok {
		session.Close()
		delete(s.Sessions, id)
	}
}

func (s *Server) ListenAndServe() error {
	return s.sshd.ListenAndServe()
}
