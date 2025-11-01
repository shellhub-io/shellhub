package srdp

import (
	"fmt"
	"io"

	"github.com/shellhub-io/shellhub/pkg/srdp/displays"
	log "github.com/sirupsen/logrus"
)

type Auth interface {
	SecurityType(c *Connection) byte
}

// NoAuth implements no authentication.
// https://www.rfc-editor.org/rfc/rfc6143.html#section-7.2.1
type NoAuth struct{}

var _ Auth = (*NoAuth)(nil)

func NewNoAuth() Auth {
	return &NoAuth{}
}

func (a *NoAuth) SecurityType(c *Connection) byte {
	return SecurityTypeNone
}

// SRDPAuth implements SRDP authentication.
// https://www.rfc-editor.org/rfc/rfc6143.html#section-7.2.2
type SRDPAuth struct {
	Password string
}

var _ Auth = (*SRDPAuth)(nil)

func NewSRDPAuth(password string) Auth {
	return &SRDPAuth{Password: password}
}

func (a *SRDPAuth) SecurityType(c *Connection) byte {
	return SecurityTypeSRDPAuth
}

type Config struct {
	Name   string
	Auths  []Auth
	Logger *log.Entry
}

type SRDPServer struct {
	config  Config
	clients map[io.ReadWriteCloser]*Connection
	done    chan struct{}
	logger  *log.Entry
}

func NewSRDPServer(config *Config) *SRDPServer {
	var logger *log.Entry
	if config != nil && config.Logger != nil {
		logger = config.Logger
	} else {
		logger = log.NewEntry(log.StandardLogger())
	}

	return &SRDPServer{
		config:  *config,
		clients: make(map[io.ReadWriteCloser]*Connection),
		done:    make(chan struct{}),
		logger:  logger,
	}
}

func (s *SRDPServer) Handle(conn io.ReadWriteCloser, display displays.Display) error {
	c, err := NewConnection(conn, display, s.logger)
	if err != nil {
		return fmt.Errorf("failed to create SRDP connection: %v", err)
	}

	s.clients[conn] = c

	defer func() {
		c.Close() // Cleanup encoding resources on disconnect
		conn.Close()

		delete(s.clients, conn)
	}()

	logger := s.logger.WithFields(log.Fields{
		"conn_id": c.connID,
		"width":   c.width,
		"height":  c.height,
		"fps":     c.fps,
	})

	logger.Trace("Starting SRDP handshake")

	if err := c.exhangeVersion(ProtocolVersion); err != nil {
		return fmt.Errorf("protocol version handshake failed: %v", err)
	}

	logger.Trace("Protocol version handshake completed")

	if err := c.securityHandshake(s.config.Auths); err != nil {
		return fmt.Errorf("security handshake failed: %v", err)
	}

	logger.Trace("Security handshake completed")

	if err := c.clientInit(); err != nil {
		return fmt.Errorf("client initialization failed: %v", err)
	}

	logger.Trace("Client initialization completed")

	if err := c.serverInit(); err != nil {
		return fmt.Errorf("server initialization failed: %v", err)
	}

	logger.Trace("Server initialization completed, entering message loop")

	return c.loop()
}
