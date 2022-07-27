package session

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/ssh/pkg/host"
	"github.com/shellhub-io/shellhub/ssh/pkg/kind"
	"github.com/shellhub-io/shellhub/ssh/pkg/target"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var (
	ErrTarget          = fmt.Errorf("invalid session target")
	ErrLookupDevice    = fmt.Errorf("could not lookup for device data")
	ErrBillingBlock    = fmt.Errorf("you cannot connect to this device because the namespace is not eligible for the free plan.\\nPlease contact the namespace owner's to upgrade the plan.\\nSee our pricing plans on https://www.shellhub.io/pricing to estimate the cost of your use cases on ShellHub Cloud or go to https://cloud.shellhub.io/settings/billing to upgrade the plan")
	ErrFirewallBlock   = fmt.Errorf("a firewall rule block this action")
	ErrHost            = fmt.Errorf("it could not get the device address")
	ErrKindShell       = fmt.Errorf("it could not open a shell in the device")
	ErrKindExec        = fmt.Errorf("it could not exec the command in the device")
	ErrKindHeredoc     = fmt.Errorf("it could not exec the commands in the device")
	ErrKindUnsupported = fmt.Errorf("this connection kind does not exist")
	ErrFlow            = fmt.Errorf("it cloud not open a data from client to agent")
	ErrConnect         = fmt.Errorf("it could not connect to device") // NOTICE: this error happens when password wasn't corret.
)

type Session struct {
	session       gliderssh.Session
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

// NewSession creates a new session from a client to device, validating data, instance and payment.
// It receives an SSHID what contains the device's hostname and namespace which it is desirable to
// connect or a device ID, and a instance of sshserver.Session.
func NewSession(sshid string, session gliderssh.Session) (*Session, error) {
	log.WithFields(log.Fields{
		"sshid":   sshid,
		"session": session,
	}).Trace("The creation of a new ShellHub session instance was initialized")

	tag, err := target.NewTarget(sshid)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"sshid": sshid,
		}).Error("Failed to get the session's target")

		return nil, ErrTarget
	}

	hos, err := host.NewHost(session.RemoteAddr().String())
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"sshid": sshid,
		}).Error("Failed to get session's address")

		return nil, ErrHost
	}

	if hos.IsLocalhost() {
		env := loadEnv(session.Environ())
		if value, ok := env["IP_ADDRESS"]; ok {
			hos.Host = value
		}
	}

	api := internalclient.NewClient()

	// When session's target doesn't has a dot, it is a connection from web terminal, but it has, session's
	// target is the `SSHID`, what has that dot.
	var lookup map[string]string

	if tag.IsSSHID() {
		namespace, hostname, err := tag.SplitSSHID()
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"sshid": sshid,
			}).Error("Failed to get the device's hostname and namespace")

			return nil, ErrTarget
		}

		lookup = map[string]string{
			"domain": namespace,
			"name":   hostname,
		}
	} else {
		device, err := api.GetDevice(tag.Data)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"sshid": sshid,
			}).Error("Failed to get the device from API")

			return nil, ErrFindDevice
		}

		lookup = map[string]string{
			"domain": device.Namespace,
			"name":   device.Name,
		}
	}

	lookup["username"] = tag.Username
	lookup["ip_address"] = hos.Host

	log.WithFields(log.Fields{
		"sshid":  sshid,
		"lookup": lookup,
	}).Debug("Device's to lookup at the API")

	uid, errs := api.Lookup(lookup)
	if len(errs) > 0 || uid == "" {
		log.WithError(err).WithFields(log.Fields{
			"sshid": sshid,
		}).Error("Failed to lookup for device in API")

		return nil, ErrLookupDevice
	}

	if envs.IsCloud() || envs.IsEnterprise() {
		if err := api.FirewallEvaluate(lookup); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"sshid": sshid,
			}).Info("A firewall rule blocked this action")

			return nil, ErrFirewallBlock
		}
	}

	if envs.IsCloud() && envs.HasBilling() {
		device, err := api.GetDevice(uid)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"sshid": sshid,
			}).Error("Failed to get the device's data in the API server")

			return nil, ErrFindDevice
		}

		if _, status, _ := api.BillingEvaluate(device.TenantID); status != 200 && status != 402 {
			log.WithError(err).WithFields(log.Fields{
				"sshid":  sshid,
				"device": device.UID,
				"tenant": device.TenantID,
			}).Info("The billing blocked this action")

			return nil, ErrBillingBlock
		}
	}

	sess := &Session{ // nolint:exhaustruct
		session:   session,
		UID:       session.Context().Value(gliderssh.ContextKeySessionID).(string),
		User:      tag.Username,
		IPAddress: hos.Host,
		Target:    uid,
		Lookup:    lookup,
	}

	handlePty(sess)

	log.WithFields(log.Fields{
		"sshid":   sshid,
		"session": session,
	}).Trace("The new ShellHub session instance created")

	return sess, nil
}

// Connect trys to connect a client to the device what he or she wants to access. It receives a possible
// session's password or public key, the SSH session opened from client to the server a network connection,
// our API client for internal routes and configurations about connection's kind.
func (s *Session) Connect(passwd string, key *rsa.PrivateKey, session gliderssh.Session, conn net.Conn, c internalclient.Client, opts kind.ConfigOptions) error {
	log.WithFields(log.Fields{
		"session": s.UID,
	}).Trace("A connection between a client and agent was initialized")

	ctx, cancel := context.WithCancel(session.Context())
	defer cancel()

	config := &ssh.ClientConfig{ // nolint: exhaustruct
		User: s.User,
		Auth: []ssh.AuthMethod{},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	if key != nil {
		signer, err := ssh.NewSignerFromKey(key)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"session": s.UID,
			}).Error("Failed to get the signer from public key")

			return ErrSignerPublicKey
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
		log.WithError(err).WithFields(log.Fields{
			"session": s.UID,
		}).Error("Failed to connect to forwarding")

		return ErrConnect
	}

	client, err := connection.NewSession()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"session": s.UID,
		}).Error("Failed to create session for SSH Client")
	}

	go handleRequests(ctx, reqs, c)

	pty, winCh, isPty := session.Pty()

	kid := kind.NewKind(session.Context(), isPty)

	// Gets the ssh's server connection from the context to kill the process initialized by the session.
	server, ok := session.Context().Value(gliderssh.ContextKeyConn).(*ssh.ServerConn)
	if !ok {
		log.WithFields(log.Fields{
			"session": s.UID,
		}).Warning("Type assertion failed")

		return fmt.Errorf("type assertion failed")
	}

	go func() {
		// Waits until the connection send no more data, and so kill the process opened by this connection.
		server.Wait() // nolint:errcheck
		client.Close()

		log.WithFields(log.Fields{
			"session": s.UID,
		}).Info("Processes opened was closed")
	}()

	switch kid.Get() {
	case kind.SHELL:
		err = kid.Shell(c, s.UID, client, session, pty, winCh, opts)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"session": s.UID,
			}).Error("Failed to create a shell")

			return ErrKindShell
		}
	case kind.EXEC:
		err = kid.Exec(c, s.UID, client, session)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"session": s.UID,
			}).Error("Failed to exec a command")

			return ErrKindExec
		}
	case kind.HEREDOC:
		err = kid.Heredoc(c, s.UID, client, session)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"session": s.UID,
			}).Error("Failed to exec a interative commands")

			return ErrKindHeredoc
		}
	default:
		log.WithError(err).WithFields(log.Fields{
			"session": s.UID,
		}).Error("This connection isn't supported")
		session.Exit(255) // nolint:errcheck

		return ErrKindUnsupported
	}

	err = conn.Close()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"session": s.UID,
		}).Error("Failed to close the connection")

		return err
	}

	log.WithFields(log.Fields{
		"session": s.UID,
	}).Trace("A connection between a client and agent was closed")

	return nil
}

func (s *Session) Register(_ gliderssh.Session) error {
	log.WithFields(log.Fields{
		"session": s.UID,
	}).Trace("Trying to register a new session at the API")

	if _, err := resty.New().R().
		SetBody(*s).
		Post("http://api:8080/internal/sessions"); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"session": s.UID,
		}).Error("Failed to register a new session at the API")

		return err
	}

	log.WithFields(log.Fields{
		"session": s.UID,
	}).Trace("Session registered at the API")

	return nil
}

func (s *Session) Finish(conn net.Conn) error {
	if conn != nil {
		request, err := http.NewRequest("DELETE", fmt.Sprintf("/ssh/close/%s", s.UID), nil)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"session": s.UID,
			}).Warning("Failed to request the session close")
		}

		if err := request.Write(conn); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"session": s.UID,
			}).Warning("Failed to write the request to connection")
		}
	}

	if errs := internalclient.NewClient().FinishSession(s.UID); len(errs) > 0 {
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

func handleRequests(ctx context.Context, reqs <-chan *ssh.Request, c internalclient.Client) {
	for {
		select {
		case req := <-reqs:
			if req == nil {
				break
			}

			switch req.Type {
			case "keepalive":
				if id, ok := ctx.Value(gliderssh.ContextKeySessionID).(string); ok {
					if errs := c.KeepAliveSession(id); len(errs) > 0 {
						log.Error(errs[0])
					}
				}
			default:
				if req.WantReply {
					if err := req.Reply(false, nil); err != nil {
						log.Error(err)
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
			log.WithError(err).WithFields(log.Fields{
				"address": addr,
			}).Error("Failed to read the dealine from the connection")

			return nil, nil, err
		}
	}

	cli, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"address": addr,
		}).Error("Failed to create a new SSHE client connection")

		return nil, nil, err
	}

	if config.Timeout > 0 {
		if err := conn.SetReadDeadline(time.Time{}); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"address": addr,
			}).Error("Failed to read the dealine from the connection")

			return nil, nil, err
		}
	}

	ch := make(chan *ssh.Request)
	close(ch)

	return ssh.NewClient(cli, chans, ch), reqs, nil
}
