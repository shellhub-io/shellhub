package sshsrvtest

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net"
	"os"

	gliderssh "github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

// Conn represents a test SSH connection.
type Conn struct {
	l      net.Listener
	config *gossh.ClientConfig
	client *gossh.Client

	Agent  *gossh.Session
	Server *gliderssh.Server
}

// exit panics the execution with `err` and writes `msg` to Stderr
func exit(err error, msg string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", msg, err.Error())
	panic(err)
}

// setupListener sets up a new TCP listener.
func setupListener() net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		exit(err, fmt.Sprintf("failed to listen on port: %v", err))
	}

	return l
}

// New creates and initializes a new SSH test connection.
func New(srv *gliderssh.Server, cfg *gossh.ClientConfig) *Conn {
	conn := &Conn{
		l:      setupListener(),
		Server: srv,
		config: cfg,
	}

	return conn
}

func (c *Conn) Start() {
	go c.startServer()
	c.startConnection()
}

// Teardown terminate the connection and server.
func (c *Conn) Teardown() {
	c.Server.Close()
	c.Agent.Close()
	c.client.Close()
	c.l.Close()
}

// setup simulates the `gliderssh.Server.ensureHandlers` and `gliderssh.Server.ensureHostSigners` without
// lock.
func (c *Conn) setup() {
	// setup Handlers
	if c.Server.RequestHandlers == nil {
		c.Server.RequestHandlers = make(map[string]gliderssh.RequestHandler)
		for k, v := range gliderssh.DefaultRequestHandlers {
			c.Server.RequestHandlers[k] = v
		}
	}
	if c.Server.ChannelHandlers == nil {
		c.Server.ChannelHandlers = make(map[string]gliderssh.ChannelHandler)
		for k, v := range gliderssh.DefaultChannelHandlers {
			c.Server.ChannelHandlers[k] = v
		}
	}
	if c.Server.SubsystemHandlers == nil {
		c.Server.SubsystemHandlers = make(map[string]gliderssh.SubsystemHandler)
		for k, v := range gliderssh.DefaultSubsystemHandlers {
			c.Server.SubsystemHandlers[k] = v
		}
	}

	// setup HostSigners
	if len(c.Server.HostSigners) == 0 {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			exit(err, "failed to generate RSA key")
		}
		signer, err := gossh.NewSignerFromKey(key)
		if err != nil {
			exit(err, "failed to generate signer")
		}
		c.Server.HostSigners = append(c.Server.HostSigners, signer)
	}
}

// startServer starts the server and blocks until waiting for a single connection.
func (c *Conn) startServer() {
	c.setup()

	conn, err := c.l.Accept()
	if err != nil {
		exit(err, "failed to accept the connection")
	}

	c.Server.HandleConn(conn)
}

// startConnection starts a new SSH connection, sets up the client and session.
func (c *Conn) startConnection() {
	client, err := gossh.Dial("tcp", c.l.Addr().String(), c.config)
	if err != nil {
		exit(err, "failed to dial")
	}
	c.client = client

	session, err := client.NewSession()
	if err != nil {
		exit(err, "failed to create the session")
	}
	c.Agent = session
}
