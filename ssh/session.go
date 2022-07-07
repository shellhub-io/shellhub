package main

import (
	"bytes"
	"context"
	"crypto/rsa"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	sshserver "github.com/gliderlabs/ssh"
	"github.com/go-resty/resty/v2"
	client "github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type Session struct {
	session       sshserver.Session
	User          string `json:"username"`
	Target        string `json:"device_uid"`
	UID           string `json:"uid"`
	IPAddress     string `json:"ip_address"`
	Type          string `json:"type"`
	Term          string `json:"term"`
	Authenticated bool   `json:"authenticated"`
	Lookup        map[string]string
	Pty           bool
}

type ConfigOptions struct {
	RecordURL string `envconfig:"record_url"`
}

const (
	Web  = "web"     // webterminal
	Term = "term"    // iterative pty
	Exec = "exec"    // non iterative pty
	SCP  = "scp"     // scp
	Unk  = "unknown" // unknown
)

func handlePty(s *Session) {
	pty, _, isPty := s.session.Pty()
	if isPty {
		s.Term = pty.Term
		s.Type = Unk
	}

	s.Pty = isPty

	env := loadEnv(s.session.Environ())

	if value, ok := env["WS"]; ok && value == "true" {
		env["WS"] = "false"
		s.Type = Web

		return
	}

	commands := s.session.Command()

	var cmd string

	if len(commands) != 0 {
		cmd = commands[0]
	}

	switch {
	case !isPty && strings.HasPrefix(cmd, "scp"):
		s.Type = SCP
	case !isPty && cmd != "":
		s.Type = Exec
	case isPty:
		s.Type = Term
	}
}

type Target struct {
	Username string
	Data     string
}

func NewTarget(target string) (*Target, error) {
	// Target could be a either device id or a SSHID.
	//
	// Example: namespace.00-00-00-00-00-00
	// 'namespace' is the user's namespace in ShellHub.
	// '00-00-00-00-00' is the device's hostname in ShellHub.
	//
	// Example: username@namespace.00-00-00-00-00-00@localhost.
	// 'username' is the user on the device.
	// 'namespace' is the user's namespace in ShellHub.
	// '00-00-00-00-00' is the device's hostname in ShellHub.
	// 'localhost' is the server's address.
	const USERNAME = 0
	const DATA = 1

	parts := strings.SplitN(target, "@", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("cloud not split the target into two parts")
	}

	return &Target{Username: parts[USERNAME], Data: parts[DATA]}, nil
}

// isSSHID checks if target is a SSHID.
func (t *Target) isSSHID() bool {
	return strings.Contains(t.Data, ".")
}

// splitSSHID splits the SSHID target into namespace and hostname into lower strings.
// Namespace is the device's namespace and hostname is the device's name.
func (t *Target) splitSSHID() (string, string, error) {
	if !t.isSSHID() {
		return "", "", fmt.Errorf("target is not from SSHID type")
	}

	const NAMESPACE = 0
	const HOSTNAME = 1

	parts := strings.SplitN(t.Data, ".", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("cloud not split the target into two parts")
	}

	return strings.ToLower(parts[NAMESPACE]), strings.ToLower(parts[HOSTNAME]), nil
}

type Host struct {
	Host string
}

func NewHost(address string) (*Host, error) {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	return &Host{Host: host}, nil
}

// isLocalhost checks if host address is local.
func (h *Host) isLocalhost() bool {
	return h.Host == "127.0.0.1" || h.Host == "::1"
}

// NewSession creates a new session from a client to device, validating data, instance and payment.
// It receives a target, a device which it is desirable to connect, what could be either a device id or a SSHID,
// and a instance of sshserver.Session.
func NewSession(target string, session sshserver.Session) (*Session, error) {
	tag, err := NewTarget(target)
	if err != nil {
		return nil, ErrInvalidSessionTarget
	}

	hos, err := NewHost(session.RemoteAddr().String())
	if err != nil {
		return nil, ErrHost
	}

	if hos.isLocalhost() {
		env := loadEnv(session.Environ())
		if value, ok := env["IP_ADDRESS"]; ok {
			hos.Host = value
		}
	}

	cli := client.NewClient()

	// When session's target doesn't has a dot, it is a connection from web terminal, but it has, session's
	// target is the `SSHID`, what has that dot.
	var namespace string
	var hostname string
	if tag.isSSHID() {
		namespace, hostname, err = tag.splitSSHID()
		if err != nil {
			return nil, ErrInvalidSessionTarget
		}
	} else {
		device, err := cli.GetDevice(tag.Data)
		if err != nil {
			return nil, ErrFindDevice
		}

		namespace = device.Namespace
		hostname = device.Name
	}

	lookup := map[string]string{
		"domain":     namespace,
		"name":       hostname,
		"username":   tag.Username,
		"ip_address": hos.Host,
	}

	uid, errs := cli.Lookup(lookup)
	if len(errs) > 0 || uid == "" {
		return nil, ErrLookupDevice
	}

	if envs.IsCloud() || envs.IsEnterprise() {
		if err := cli.FirewallEvaluate(lookup); err != nil {
			return nil, ErrFirewallBlock
		}
	}

	if envs.IsCloud() && envs.HasBilling() {
		device, err := cli.GetDevice(uid)
		if err != nil {
			return nil, ErrFindDevice
		}

		if _, status, _ := cli.BillingEvaluate(device.TenantID); status != 200 && status != 402 {
			return nil, ErrBillingBlock
		}
	}

	sess := &Session{
		session: session,
		UID:     session.Context().Value(sshserver.ContextKeySessionID).(string),
		User:    tag.Username,
		Target:  uid,
		Lookup:  lookup,
	}

	handlePty(sess)

	return sess, nil
}

const (
	SHELL   = 1
	EXEC    = 2
	HEREDOC = 3
)

type Kind struct {
	Kind int
}

func NewKind(ctx sshserver.Context, isPty bool) *Kind {
	requestType := ctx.Value("request_type").(string)

	var kind int
	switch {
	case isPty:
		kind = SHELL
	case !isPty && requestType == "exec":
		kind = EXEC
	case !isPty && requestType == "shell":
		kind = HEREDOC
	default:
		kind = -1
	}

	return &Kind{Kind: kind}
}

// Get gets the connection's kind.
func (k *Kind) Get() int {
	return k.Kind
}

type Flow struct {
	Stdin  io.WriteCloser
	Stdout io.Reader
	Stderr io.Reader
}

func NewFlow(client *ssh.Session) (*Flow, error) {
	stdin, err := client.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := client.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := client.StderrPipe()
	if err != nil {
		return nil, err
	}

	return &Flow{Stdin: stdin, Stdout: stdout, Stderr: stderr}, nil
}

// PipeIn pipes the session's user stdin to the agent's stdin.
func (f *Flow) PipeIn(session io.Reader) {
	if _, err := io.Copy(f.Stdin, session); err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to copy to from session to agent in raw session")
	}

	f.Stdin.Close()
}

// PipeOut pipes the agent's stdout and stderr to the session's user.
func (f *Flow) PipeOut(session sshserver.Session) {
	if _, err := io.Copy(session, io.MultiReader(f.Stdout, f.Stderr)); err != nil && err != io.EOF {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("Failed to copy to from stdout and stderr to client in raw session")
	}
}

func (s *Session) connect(passwd string, key *rsa.PrivateKey, session sshserver.Session, conn net.Conn, c client.Client, opts ConfigOptions) error { //nolint: gocyclo
	ctx, cancel := context.WithCancel(session.Context())
	defer cancel()

	config := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	if key != nil {
		signer, err := ssh.NewSignerFromKey(key)
		if err != nil {
			return err
		}

		config.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		}
	} else {
		config.Auth = []ssh.AuthMethod{
			ssh.Password(passwd),
		}
	}

	connection, reqs, err := NewClientConnWithDeadline(conn, "tcp", config)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"session": s.UID,
			"err":     err,
		}).Warning("Failed to connect to forwarding")

		return err
	}

	client, err := connection.NewSession()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"session": s.UID,
			"err":     err,
		}).Error("Failed to create session for SSH Client")
	}

	go handleRequests(ctx, reqs, c)

	pty, winCh, isPty := session.Pty()

	kind := NewKind(session.Context(), isPty)

	// status gets the exit status from the client when a error happens. If error is nil, the status is zero
	// meaing there is not error. If none exit code is returned, it return 255.
	status := func(err error) int {
		if err == nil {
			return 0
		}

		fault, ok := err.(*ssh.ExitError)
		if !ok {
			return 255
		}

		return fault.ExitStatus()
	}

	// Gets the ssh's server connection from the context to kill the process initialized by the session.
	server, ok := session.Context().Value(sshserver.ContextKeyConn).(*ssh.ServerConn)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"session": s.UID,
		}).Warning("Type assertion failed")

		return fmt.Errorf("type assertion failed")
	}

	go func() {
		// Waits until the connection send no more data, and so kill the process opened by this connection.
		server.Wait() // nolint:errcheck
		client.Close()
	}()

	flow, err := NewFlow(client)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"session": s.UID,
			"err":     err,
		}).Error("Failed to create a flow of data from client to agent")

		return err
	}

	go flow.PipeIn(session)
	go flow.PipeOut(session)

	switch kind.Get() {
	case SHELL:
		if errs := c.PatchSessions(s.UID); len(errs) > 0 {
			return errs[0]
		}

		if err = client.RequestPty(pty.Term, pty.Window.Height, pty.Window.Width, ssh.TerminalModes{}); err != nil {
			return err
		}

		go func() {
			for win := range winCh {
				if err = client.WindowChange(win.Height, win.Width); err != nil {
					logrus.WithFields(logrus.Fields{
						"session": s.UID,
						"err":     err,
					}).Error("Failed to send WindowChange")
				}
			}
		}()

		go func() {
			buf := make([]byte, 1024)
			n, err := flow.Stdout.Read(buf)
			waitingString := ""
			if err == nil {
				waitingString = string(buf[:n])
				if envs.IsEnterprise() || envs.IsCloud() {
					c.RecordSession(&models.SessionRecorded{
						UID:     s.UID,
						Message: waitingString,
						Width:   pty.Window.Height,
						Height:  pty.Window.Width,
					}, opts.RecordURL)
				}
				waitingString = ""
			}
			for {
				bufReader := bytes.NewReader(buf[:n])
				if _, err = io.Copy(s.session, bufReader); err != nil {
					logrus.WithFields(logrus.Fields{
						"session": s.UID,
						"err":     err,
					}).Error("Failed to copy from stdout in pty session")
				}
				n, err = flow.Stdout.Read(buf)
				if err != nil {
					break
				}
				waitingString += string(buf[:n])
				if envs.IsEnterprise() || envs.IsCloud() {
					c.RecordSession(&models.SessionRecorded{
						UID:     s.UID,
						Message: waitingString,
						Width:   pty.Window.Height,
						Height:  pty.Window.Width,
					}, opts.RecordURL)
				}
				waitingString = ""
			}
		}()

		if err = client.Shell(); err != nil {
			logrus.WithFields(logrus.Fields{
				"session": s.UID,
				"err":     err,
			}).Error("Failed to start a new shell")

			return err
		}

		err = client.Wait()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"session": s.UID,
				"err":     err,
			}).Warning("Client remote command returned a error")
		}

		session.Exit(0) // nolint:errcheck
	case EXEC:
		if errs := c.PatchSessions(s.UID); len(errs) > 0 {
			return errs[0]
		}

		if err = client.Start(session.RawCommand()); err != nil {
			logrus.WithFields(logrus.Fields{
				"session": s.UID,
				"err":     err,
			}).Error("Failed to start session raw command")

			return err
		}

		err = client.Wait()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"session": s.UID,
				"err":     err,
			}).Warning("Client remote command returned a error")
		}

		session.Exit(status(err)) // nolint:errcheck
	case HEREDOC:
		if errs := c.PatchSessions(s.UID); len(errs) > 0 {
			return errs[0]
		}

		if err = client.Shell(); err != nil {
			logrus.WithFields(logrus.Fields{
				"session": s.UID,
				"err":     err,
			}).Error("Failed to start a new shell")

			return err
		}

		err = client.Wait()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"session": s.UID,
				"err":     err,
			}).Warning("Client remote command returned a error")
		}

		session.Exit(status(err)) // nolint:errcheck
	default:
		logrus.Errorln("Kind of connection isn't supported")

		session.Exit(0) // nolint:errcheck //TODO: Exit with the right exit code.
	}

	err = conn.Close()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"session": s.UID,
			"err":     err,
		}).Error("Failed to close the connection")

		return err
	}

	return nil
}

func (s *Session) register(_ sshserver.Session) error {
	if _, err := resty.New().R().
		SetBody(*s).
		Post("http://api:8080/internal/sessions"); err != nil {
		return err
	}

	return nil
}

func (s *Session) finish(conn net.Conn) error {
	if conn != nil {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/ssh/close/%s", s.UID), nil)
		if err := req.Write(conn); err != nil {
			logrus.WithFields(logrus.Fields{
				"err":     err,
				"session": s.session.Context().Value(sshserver.ContextKeySessionID),
			}).Error("Failed to write")
		}
	}

	if errs := client.NewClient().FinishSession(s.UID); len(errs) > 0 {
		return errs[0]
	}

	return nil
}

func loadEnv(env []string) map[string]string {
	m := make(map[string]string, cap(env))

	for _, s := range env {
		sp := strings.Split(s, "=")
		if len(sp) == 2 {
			k := sp[0]
			v := sp[1]
			m[k] = v
		}
	}

	return m
}

func handleRequests(ctx context.Context, reqs <-chan *ssh.Request, c client.Client) {
	for {
		select {
		case req := <-reqs:
			if req == nil {
				break
			}

			switch req.Type {
			case "keepalive":
				if id, ok := ctx.Value(sshserver.ContextKeySessionID).(string); ok {
					if errs := c.KeepAliveSession(id); len(errs) > 0 {
						logrus.Error(errs[0])
					}
				}
			default:
				if req.WantReply {
					if err := req.Reply(false, nil); err != nil {
						logrus.Error(err)
					}
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func NewClientConnWithDeadline(conn net.Conn, addr string, config *ssh.ClientConfig) (*ssh.Client, <-chan *ssh.Request, error) {
	if config.Timeout > 0 {
		if err := conn.SetReadDeadline(clock.Now().Add(config.Timeout)); err != nil {
			return nil, nil, err
		}
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, nil, err
	}

	if config.Timeout > 0 {
		if err := conn.SetReadDeadline(time.Time{}); err != nil {
			return nil, nil, err
		}
	}

	emptyCh := make(chan *ssh.Request)
	close(emptyCh)

	return ssh.NewClient(c, chans, emptyCh), reqs, nil
}
