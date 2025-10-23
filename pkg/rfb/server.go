package rfb

import (
	"fmt"
	"io"

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

// RFBAuth implements RFB authentication.
// https://www.rfc-editor.org/rfc/rfc6143.html#section-7.2.2
type RFBAuth struct {
	Password string
}

var _ Auth = (*RFBAuth)(nil)

func NewRFBAuth(password string) Auth {
	return &RFBAuth{Password: password}
}

func (a *RFBAuth) SecurityType(c *Connection) byte {
	return SecurityTypeRFBAuth
}

type Config struct {
	Name   string
	Auths  []Auth
	Logger *log.Entry
}

type RFBServer struct {
	config  Config
	clients map[io.ReadWriteCloser]*Connection
	done    chan struct{}
	logger  *log.Entry
}

// NewRFBServer creates a new RFB server
func NewRFBServer(config *Config) (*RFBServer, error) {
	var logger *log.Entry
	if config != nil && config.Logger != nil {
		logger = config.Logger
	} else {
		logger = log.NewEntry(log.StandardLogger())
	}

	return &RFBServer{
		config:  *config,
		clients: make(map[io.ReadWriteCloser]*Connection),
		done:    make(chan struct{}),
		logger:  logger,
	}, nil
}

func (s *RFBServer) Handle(conn io.ReadWriteCloser, display Display) error {
	// Create RFB connection handler
	c := NewConnection(conn, display, s.logger)
	s.clients[conn] = c

	defer func() {
		c.Close() // Cleanup encoding resources on disconnect
		conn.Close()

		delete(s.clients, conn)
	}()

	s.logger.Trace("Starting RFB handshake")

	// Protocol version handshake
	if err := c.protocolVersionHandshake(ProtocolVersion); err != nil {
		return fmt.Errorf("protocol version handshake failed: %v", err)
	}

	s.logger.Trace("Protocol version handshake completed")

	// Security handshake
	if err := c.securityHandshake(s.config.Auths); err != nil {
		return fmt.Errorf("security handshake failed: %v", err)
	}

	s.logger.Trace("Security handshake completed")

	// Client initialization
	if err := c.clientInitialization(); err != nil {
		return fmt.Errorf("client initialization failed: %v", err)
	}
	s.logger.Trace("Client initialization completed")

	// Server initialization
	if err := c.serverInitialization(s.config.Name); err != nil {
		return fmt.Errorf("server initialization failed: %v", err)
	}

	s.logger.Trace("Server initialization completed, entering message loop")

	// Main message loop
	return c.messageLoop()
}
