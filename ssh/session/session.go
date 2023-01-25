package session

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/ssh/pkg/host"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

// Errors returned by the NewSession to the client.
var (
	ErrBillingBlock       = fmt.Errorf("you cannot connect to this device because the namespace is not eligible for the free plan.\\nPlease contact the namespace owner's to upgrade the plan.\\nSee our pricing plans on https://www.shellhub.io/pricing to estimate the cost of your use cases on ShellHub Cloud or go to https://cloud.shellhub.io/settings/billing to upgrade the plan")
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
	Term    = "term"    // iterative pty.
	Exec    = "exec"    // non-iterative pty.
	HereDoc = "heredoc" // heredoc pty.
	SCP     = "scp"     // scp.
	SFTP    = "sftp"    // sftp subsystem.
	Unk     = "unknown" // unknown.
)

// handlePty sets the connection`s type to session.
//
// Connection types possible are: web, term, exec, heredoc, scp, sftp, unknown.
func handlePty(s *Session) {
	ctx := s.Client.Context()

	// TODO: improve and clean.
	pty, _, isPty := s.Client.Pty()
	if isPty {
		s.Term = pty.Term
		s.Type = Unk
	}

	s.Pty = isPty

	env := loadEnv(s.Client.Environ())

	if value, ok := env["WS"]; ok && value == "true" {
		env["WS"] = "false"
		s.Type = Web

		return
	}

	commands := s.Client.Command()

	var cmd string

	if len(commands) != 0 {
		cmd = commands[0]
	}

	if s.Client.Subsystem() == SFTP {
		s.Type = SFTP

		return
	}

	switch {
	case !isPty && strings.HasPrefix(cmd, SCP):
		s.Type = SCP
	case !isPty && cmd != "":
		s.Type = Exec
	case !isPty && metadata.RestoreRequest(ctx) == "shell":
		s.Type = HereDoc
	case isPty:
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

	device := metadata.RestoreDevice(client.Context())
	tag := metadata.RestoreTarget(client.Context())
	api := metadata.RestoreAPI(client.Context())
	lookup := metadata.RestoreLookup(client.Context())
	lookup["username"] = tag.Username
	lookup["ip_address"] = hos.Host

	if envs.IsCloud() || envs.IsEnterprise() {
		if err := api.FirewallEvaluate(lookup); err != nil {
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
			return nil, ErrFindDevice
		}

		if _, status, _ := api.BillingEvaluate(device.TenantID); status != 200 && status != 402 {
			return nil, ErrBillingBlock
		}
	}

	dialed, err := tunnel.Dial(client.Context(), device.UID)
	if err != nil {
		return nil, ErrDial
	}

	uid := client.Context().Value(gliderssh.ContextKeySessionID).(string) //nolint:forcetypeassert

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/ssh/%s", uid), nil)
	if err = req.Write(dialed); err != nil {
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

	handlePty(session)

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
			return nil, nil, err
		}
	}

	cli, chans, reqs, err := gossh.NewClientConn(s.Dialed, Addr, config)
	if err != nil {
		return nil, nil, err
	}

	if config.Timeout > 0 {
		if err := s.Dialed.SetReadDeadline(time.Time{}); err != nil {
			return nil, nil, err
		}
	}

	ch := make(chan *gossh.Request)
	close(ch)

	return gossh.NewClient(cli, chans, ch), reqs, nil
}

// Register registers a new Client at the api.
func (s *Session) Register(_ gliderssh.Session) error {
	if _, err := resty.New().R().
		SetBody(*s).
		Post("http://api:8080/internal/sessions"); err != nil {
		return err
	}

	return nil
}

func (s *Session) Finish() error {
	if s.Dialed != nil {
		request, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/ssh/close/%s", s.UID), nil)

		if err := request.Write(s.Dialed); err != nil {
			log.WithFields(log.Fields{
				"session": s.UID,
			}).Error(err)
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

func HandleRequests(ctx context.Context, reqs <-chan *gossh.Request, c internalclient.Client) {
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
		case <-ctx.Done():
			return
		}
	}
}
