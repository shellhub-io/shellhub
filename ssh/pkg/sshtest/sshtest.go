// Package sshtest provides utilities for setting up SSH servers and clients for testing purposes.
package sshtest

import (
	"crypto/rand"
	"crypto/rsa"
	"net"
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
)

// SSHTest represents a test SSH connection.
type SSHTest struct {
	t       *testing.T        // t represents the testing instance.
	l       net.Listener      // l is the listener for the SSH server.
	Server  *gliderssh.Server // Server is the SSH server instance.
	Session *gossh.Session    // Session represents the gossh session.
	Client  *gossh.Client     // Client is the SSH client instance.
}

// setupListener sets up a new TCP listener.
func setupListener(t *testing.T) net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on a port: %v", err)
	}

	return l
}

// Start creates and initializes a new SSH test connection.
func Start(t *testing.T, cfg *gossh.ClientConfig, srv *gliderssh.Server) *SSHTest {
	conn := &SSHTest{
		t:      t,
		l:      setupListener(t),
		Server: srv,
	}

	go conn.startServer() // nolint:errcheck
	conn.startConnection(cfg)

	return conn
}

// Teardown terminate the connection and server.
func (ssh *SSHTest) Teardown() {
	ssh.Server.Close()
	ssh.Session.Close()
	ssh.Client.Close()
	ssh.l.Close()
}

// setup simulates the `gliderssh.Server.ensureHandlers` and `gliderssh.Server.ensureHostSigners` without
// lock.
func (ssh *SSHTest) setup() {
	// setup Handlers
	if ssh.Server.RequestHandlers == nil {
		ssh.Server.RequestHandlers = make(map[string]gliderssh.RequestHandler)
		for k, v := range gliderssh.DefaultRequestHandlers {
			ssh.Server.RequestHandlers[k] = v
		}
	}
	if ssh.Server.ChannelHandlers == nil {
		ssh.Server.ChannelHandlers = make(map[string]gliderssh.ChannelHandler)
		for k, v := range gliderssh.DefaultChannelHandlers {
			ssh.Server.ChannelHandlers[k] = v
		}
	}
	if ssh.Server.SubsystemHandlers == nil {
		ssh.Server.SubsystemHandlers = make(map[string]gliderssh.SubsystemHandler)
		for k, v := range gliderssh.DefaultSubsystemHandlers {
			ssh.Server.SubsystemHandlers[k] = v
		}
	}

	// setup HostSigners
	if len(ssh.Server.HostSigners) == 0 {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			ssh.t.Fatal(err)
		}
		signer, err := gossh.NewSignerFromKey(key)
		if err != nil {
			ssh.t.Fatal(err)
		}
		ssh.Server.HostSigners = append(ssh.Server.HostSigners, signer)
	}
}

// startServer serves a single SSH connection.
func (ssh *SSHTest) startServer() {
	ssh.setup()
	conn, err := ssh.l.Accept()
	if err != nil {
		// since serveOnce is called in a goroutine, we use panic instead of `t.Fatal`
		panic(err)
	}

	ssh.Server.ChannelHandlers = map[string]gliderssh.ChannelHandler{
		"session":      gliderssh.DefaultSessionHandler,
		"direct-tcpip": gliderssh.DirectTCPIPHandler,
	}
	ssh.Server.HandleConn(conn)
}

// startConnection starts a new SSH connection, sets up the client and session.
func (ssh *SSHTest) startConnection(config *gossh.ClientConfig) {
	if config == nil {
		config = &gossh.ClientConfig{
			User: "user",
			Auth: []gossh.AuthMethod{
				gossh.Password("pass"),
			},
		}
	}

	if config.HostKeyCallback == nil {
		config.HostKeyCallback = gossh.InsecureIgnoreHostKey()
	}

	client, err := gossh.Dial("tcp", ssh.l.Addr().String(), config)
	if err != nil {
		ssh.t.Fatal(err)
	}
	ssh.Client = client

	session, err := client.NewSession()
	if err != nil {
		ssh.t.Fatal(err)
	}
	ssh.Session = session
}
