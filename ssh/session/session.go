package session

import (
	"context"
	"errors"
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
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/ssh/pkg/host"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// Errors returned by the NewSession to the client.
var (
	ErrBillingBlock       = fmt.Errorf("Connection to this device is not available as your current namespace doesn't qualify for the free plan. To gain access, you'll need to contact the namespace owner to initiate an upgrade.\n\nFor a detailed estimate of costs based on your use-cases with ShellHub Cloud, visit our pricing page at https://www.shellhub.io/pricing. If you wish to upgrade immediately, navigate to https://cloud.shellhub.io/settings/billing. Your cooperation is appreciated.") //nolint:all
	ErrFirewallBlock      = fmt.Errorf("you cannot connect to this device because a firewall rule block your connection")
	ErrFirewallConnection = fmt.Errorf("failed to communicate to the firewall")
	ErrFirewallUnknown    = fmt.Errorf("failed to evaluate the firewall rule")
	ErrHost               = fmt.Errorf("failed to get the device address")
	ErrFindDevice         = fmt.Errorf("failed to find the device")
	ErrDial               = fmt.Errorf("failed to connect to device agent, please check the device connection")
)

type Session struct {
	Client gliderssh.Session
	// Username is the user that is trying to connect to the device; user on device.
	Username string `json:"username"`
	Device   string `json:"device_uid"` // nolint: tagliatelle
	// UID is the device's UID.
	UID           string `json:"uid"`
	IPAddress     string `json:"ip_address"` // nolint: tagliatelle
	Type          string `json:"type"`
	Term          string `json:"term"`
	Authenticated bool   `json:"authenticated"`
	Lookup        map[string]string
	Pty           bool
	Dialed        net.Conn
}

const (
	Web     = "web"     // web terminal.
	Term    = "term"    // interactive session
	Exec    = "exec"    // command execution
	HereDoc = "heredoc" // heredoc pty.
	SCP     = "scp"     // scp.
	SFTP    = "sftp"    // sftp subsystem.
	Unk     = "unknown" // unknown.
)

// setPty sets the connection's pty.
func (s *Session) setPty() {
	pty, _, isPty := s.Client.Pty()
	if isPty {
		s.Term = pty.Term
	}

	s.Pty = isPty
}

// setType sets the connection`s type to session.
//
// Connection types possible are: Web, SFTP, SCP, Exec, HereDoc, Term, Unk (unknown)
func (s *Session) setType() {
	ctx := s.Client.Context()

	env := loadEnv(s.Client.Environ())
	if value, ok := env["WS"]; ok && value == "true" {
		env["WS"] = "false"
		s.Type = Web

		return
	}

	if s.Client.Subsystem() == SFTP {
		s.Type = SFTP

		return
	}

	var cmd string
	commands := s.Client.Command()
	if len(commands) != 0 {
		cmd = commands[0]
	}

	switch {
	case !s.Pty && strings.HasPrefix(cmd, SCP):
		s.Type = SCP
	case !s.Pty && metadata.RestoreRequest(ctx) == "shell":
		s.Type = HereDoc
	case cmd != "":
		s.Type = Exec
	case s.Pty:
		s.Type = Term
	default:
		s.Type = Unk
	}
}

// NewSession creates a new Client from a client to agent, validating data, instance and payment.
func NewSession(client gliderssh.Session, tunnel *httptunnel.Tunnel) (*Session, error) {
	hos, err := host.NewHost(client.RemoteAddr().String())
	if err != nil {
		return nil, ErrHost
	}

	if hos.IsLocalhost() {
		env := loadEnv(client.Environ())
		if value, ok := env["IP_ADDRESS"]; ok {
			hos.Host = value
		}
	}

	clientCtx := client.Context()

	uid := clientCtx.Value(gliderssh.ContextKeySessionID).(string) //nolint:forcetypeassert
	device := metadata.RestoreDevice(clientCtx)
	tag := metadata.RestoreTarget(clientCtx)
	api := metadata.RestoreAPI(clientCtx)
	lookup := metadata.RestoreLookup(clientCtx)

	lookup["username"] = tag.Username
	lookup["ip_address"] = hos.Host

	if envs.IsCloud() || envs.IsEnterprise() {
		if err := api.FirewallEvaluate(lookup); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": uid, "sshid": client.User()}).
				Error("Error when trying to evaluate firewall rules")

			switch {
			case errors.Is(err, internalclient.ErrFirewallConnection):
				return nil, ErrFirewallConnection
			case errors.Is(err, internalclient.ErrFirewallBlock):
				return nil, ErrFirewallBlock
			default:
				return nil, ErrFirewallUnknown
			}
		}
	}

	if envs.IsCloud() && envs.HasBilling() {
		device, err := api.GetDevice(device.UID)
		if err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": uid, "sshid": client.User()}).
				Error("Error when trying to get device")

			return nil, ErrFindDevice
		}

		if evaluatation, status, _ := api.BillingEvaluate(device.TenantID); status != 402 && !evaluatation.CanConnect {
			log.WithError(err).
				WithFields(log.Fields{"session": uid, "sshid": client.User()}).
				Error("Error when trying to evaluate billing")

			return nil, ErrBillingBlock
		}
	}

	dialed, err := tunnel.Dial(client.Context(), device.UID)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("Error when trying to dial")

		return nil, ErrDial
	}

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/ssh/%s", uid), nil)
	if err = req.Write(dialed); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": client.User()}).
			Error("Error when trying to write the request")

		return nil, err
	}

	session := &Session{ //nolint:exhaustruct
		Client:    client,
		UID:       uid,
		Username:  tag.Username,
		IPAddress: hos.Host,
		Device:    device.UID,
		Lookup:    lookup,
		Dialed:    dialed,
	}

	session.setPty()
	session.setType()

	session.Register(client) // nolint:errcheck

	return session, nil
}

func (s *Session) GetType() string {
	return s.Type
}

// NewClientConnWithDeadline creates a new connection to the agent.
func (s *Session) NewClientConnWithDeadline(config *gossh.ClientConfig) (*gossh.Client, <-chan *gossh.Request, error) {
	const Addr = "tcp"

	if config.Timeout > 0 {
		if err := s.Dialed.SetReadDeadline(clock.Now().Add(config.Timeout)); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": s.UID, "sshid": s.Client.User()}).
				Error("Error when trying to set dial deadline")

			return nil, nil, err
		}
	}

	cli, chans, reqs, err := gossh.NewClientConn(s.Dialed, Addr, config)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": s.UID, "sshid": s.Client.User()}).
			Error("Error when trying to create the client's connection")

		return nil, nil, err
	}

	if config.Timeout > 0 {
		if err := s.Dialed.SetReadDeadline(time.Time{}); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": s.UID, "sshid": s.Client.User()}).
				Error("Error when trying to set dial deadline with Time{}")

			return nil, nil, err
		}
	}

	ch := make(chan *gossh.Request)
	close(ch)

	return gossh.NewClient(cli, chans, ch), reqs, nil
}

// Register registers a new Client at the api.
func (s *Session) Register(_ gliderssh.Session) error {
	_, err := resty.New().R().SetBody(*s).Post("http://api:8080/internal/sessions")
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": s.UID, "sshid": s.Client.User()}).
			Error("Error when trying to register the client on API")

		return err
	}

	return nil
}

func (s *Session) Finish() error {
	if s.Dialed != nil {
		request, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/ssh/close/%s", s.UID), nil)

		if err := request.Write(s.Dialed); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": s.UID, "sshid": s.Client.User()}).
				Warning("Error when trying write the request to /ssh/close")
		}
	}

	if errs := internalclient.NewClient().FinishSession(s.UID); len(errs) > 0 {
		log.WithError(errs[0]).
			WithFields(log.Fields{"session": s.UID, "sshid": s.Client.User()}).
			Error("Error when trying to finish the session")

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

func HandleRequests(ctx context.Context, reqs <-chan *gossh.Request, c internalclient.Client, done <-chan struct{}) {
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

				if err := req.Reply(false, nil); err != nil {
					log.Error(err)
				}
			default:
				if req.WantReply {
					if err := req.Reply(false, nil); err != nil {
						log.Error(err)
					}
				}
			}
		case <-done:
			return
		}
	}
}
