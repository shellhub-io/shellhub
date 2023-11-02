package session

import (
	"errors"
	"fmt"
	"net"
	"net/http"
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
