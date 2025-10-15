package vnc

import (
	"fmt"
	"io"
	"log"
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

// VNCAuth implements VNC authentication.
// https://www.rfc-editor.org/rfc/rfc6143.html#section-7.2.2
type VNCAuth struct {
	Password string
}

var _ Auth = (*VNCAuth)(nil)

func NewVNCAuth(password string) Auth {
	return &VNCAuth{Password: password}
}

func (a *VNCAuth) SecurityType(c *Connection) byte {
	return SecurityTypeVNCAuth
}

type Config struct {
	Name  string
	Auths []Auth
}

type VNCServer struct {
	config  Config
	clients map[io.ReadWriteCloser]*Connection
	done    chan struct{}
}

// NewVNCServer creates a new VNC server
func NewVNCServer(config *Config) (*VNCServer, error) {
	return &VNCServer{
		config:  *config,
		clients: make(map[io.ReadWriteCloser]*Connection),
		done:    make(chan struct{}),
	}, nil
}

func (s *VNCServer) Handle(conn io.ReadWriteCloser, display Display) error {
	defer func() {
		conn.Close()
		delete(s.clients, conn)
		// log.Printf("Client %s disconnected", conn.RemoteAddr())
	}()

	// Create VNC connection handler
	c := NewConnection(conn, display)
	s.clients[conn] = c

	log.Printf("Starting VNC handshake")

	// Protocol version handshake
	if err := c.protocolVersionHandshake(ProtocolVersion); err != nil {
		return fmt.Errorf("protocol version handshake failed: %v", err)
	}

	log.Printf("Protocol version handshake completed")

	// Security handshake
	if err := c.securityHandshake(s.config.Auths); err != nil {
		return fmt.Errorf("security handshake failed: %v", err)
	}

	log.Printf("Security handshake completed")

	// Client initialization
	if err := c.clientInitialization(); err != nil {
		return fmt.Errorf("client initialization failed: %v", err)
	}
	log.Printf("Client initialization completed")

	// Server initialization
	if err := c.serverInitialization(s.config.Name); err != nil {
		return fmt.Errorf("server initialization failed: %v", err)
	}

	log.Printf("Server initialization completed, entering message loop")

	// Main message loop
	return c.messageLoop()
}
