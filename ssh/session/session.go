package session

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/httptunnel"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/ssh/pkg/host"
	"github.com/shellhub-io/shellhub/ssh/pkg/metadata"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

type Dimensions struct {
	Columns uint32
	Rows    uint32
	Width   uint32
	Height  uint32
}

// NOTICE: [Pty] cannot use [Dimensions] inside itself ude [ssh.Unmarshal] issues.
type Pty struct {
	Term     string
	Columns  uint32
	Rows     uint32
	Width    uint32
	Height   uint32
	Modelist string
}

type Data struct {
	// Username is the user on the device.
	Username string
	// SSHID is the combination of device's name and namespace name.
	SSHID string
	// Device is the device connected.
	Device    *models.Device
	IPAddress string
	// Type is the connection type.
	Type string
	// Term is the terminal used for the client.
	Term string
	// TODO:
	Lookup map[string]string
	// Pty is the PTY dimension.
	Pty Pty
}

// TODO: implement [io.Read] and [io.Write] on session to simplify the data piping.
type Session struct {
	// UID is the session's UID.
	UID string

	// Dialed is websocket connection established between the Agent and the ShellHub SSH server.
	Dialed net.Conn

	// Agent is a [gossh.Client] connected and authenticated to the agent, waiting for a open sesssion request.
	Agent *gossh.Client
	// AgentGlobalReqs is the channel to handle global request like "keepalive".
	AgentGlobalReqs <-chan *gossh.Request

	api internalclient.Client

	Data
}

func (s *Session) checkFirewall() (bool, error) {
	if envs.IsCloud() || envs.IsEnterprise() {
		if err := s.api.FirewallEvaluate(s.Data.Lookup); err != nil {
			switch {
			case errors.Is(err, internalclient.ErrFirewallConnection):
				return false, errors.Join(ErrFirewallConnection, err)
			case errors.Is(err, internalclient.ErrFirewallBlock):
				return false, errors.Join(ErrFirewallBlock, err)
			default:
				return false, errors.Join(ErrFirewallUnknown, err)
			}
		}
	}

	return true, nil
}

func (s *Session) checkBilling() (bool, error) {
	if envs.IsCloud() && envs.HasBilling() {
		device, err := s.api.GetDevice(s.Device.UID)
		if err != nil {
			return false, errors.Join(ErrFindDevice, err)
		}

		if evaluatation, status, _ := s.api.BillingEvaluate(device.TenantID); status != 402 && !evaluatation.CanConnect {
			return false, errors.Join(ErrBillingBlock, err)
		}
	}

	return true, nil
}

func (s *Session) dial(ctx gliderssh.Context, tunnel *httptunnel.Tunnel, device string, session string) (net.Conn, error) {
	dialed, err := tunnel.Dial(ctx, device)
	if err != nil {
		return nil, errors.Join(ErrDial, err)
	}

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/ssh/%s", session), nil)
	if err = req.Write(dialed); err != nil {
		return nil, err
	}

	return dialed, nil
}

// NewSession creates a new session to connect the agent, validating data, instance and payment.
//
// This function is used to create a new session when the client is not available, what is true when the SSH client
// indicate that the request type is `none` or in the case of a port forwarding
func NewSession(ctx gliderssh.Context, tunnel *httptunnel.Tunnel) (*Session, error) {
	hos, err := host.NewHost(ctx.RemoteAddr().String())
	if err != nil {
		return nil, ErrHost
	}

	uid := ctx.Value(gliderssh.ContextKeySessionID).(string) //nolint:forcetypeassert
	session := &Session{
		UID: uid,
		api: metadata.RestoreAPI(ctx),
		Data: Data{
			Username:  metadata.RestoreTarget(ctx).Username,
			IPAddress: hos.Host,
			Device:    metadata.RestoreDevice(ctx),
			Lookup:    metadata.RestoreLookup(ctx),
			SSHID:     metadata.MaybeStoreSSHID(ctx, ctx.User()),
		},
	}

	session.Data.Lookup["username"] = session.Data.Username
	session.Data.Lookup["ip_address"] = session.IPAddress

	if ok, err := session.checkFirewall(); err != nil || !ok {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": session.Data.Username}).
			Error("Error when trying to evaluate firewall rules")

		return nil, err
	}

	if ok, err := session.checkBilling(); err != nil || !ok {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": session.Data.Username}).
			Error("Error when trying to evaluate billing")

		return nil, err
	}

	session.Dialed, err = session.dial(ctx, tunnel, session.Device.UID, uid)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": uid, "sshid": session.Data.Username}).
			Error("Error when trying to dial")

		return nil, ErrDial
	}

	return session, nil
}

// NewAgentConnection creates a new connection to the agent.
func (s *Session) NewAgentConnection(config *gossh.ClientConfig) error {
	const Addr = "tcp"

	if config.Timeout > 0 {
		if err := s.Dialed.SetReadDeadline(clock.Now().Add(config.Timeout)); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
				Error("Error when trying to set dial deadline")

			return err
		}
	}

	conn, chans, reqs, err := gossh.NewClientConn(s.Dialed, Addr, config)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": s.UID}).
			Error("Error when trying to create the client's connection")

		return err
	}

	if err := s.Register(); err != nil {
		return err
	}

	if config.Timeout > 0 {
		if err := s.Dialed.SetReadDeadline(time.Time{}); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
				Error("Error when trying to set dial deadline with Time{}")

			return err
		}
	}

	ch := make(chan *gossh.Request)
	close(ch)

	s.Agent = gossh.NewClient(conn, chans, ch)
	s.AgentGlobalReqs = reqs

	return s.Authenticate()
}

// registerAPISession registers a new session on the API.
func (s *Session) Register() error {
	err := s.api.SessionCreate(requests.SessionCreate{
		UID:       s.UID,
		DeviceUID: s.Device.UID,
		Username:  s.Username,
		IPAddress: s.IPAddress,
		Type:      "none",
		Term:      "none",
	})
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
			Error("Error when trying to register the client on API")

		return err
	}

	return nil
}

// Authenticate marks the session as authenticated on the API.
//
// It returns an error if authentication fails.
func (s *Session) Authenticate() error {
	if errs := s.api.SessionAsAuthenticated(s.UID); len(errs) > 0 {
		return errs[0]
	}

	return nil
}

// Record records the current session state.
//
// It returns an error if any.
func (s *Session) Record(req *models.SessionRecorded, url string) error {
	return s.api.RecordSession(req, url)
}

func (s *Session) KeepAlive() error {
	if errs := s.api.KeepAliveSession(s.UID); len(errs) > 0 {
		log.Error(errs[0])

		return errs[0]
	}

	return nil
}

// Announce is a custom message provided by the end user that can be printed when a new connection within the namespace
// is established.
//
// Returns the announcement or an error, if any. If no announcement is set, it returns an empty string.
func (s *Session) Announce(client gossh.Channel) error {
	if _, err := client.Write([]byte(
		"Connected to " + s.SSHID + " via ShellHub.",
	)); err != nil {
		return err
	}

	namespace, errs := s.api.
		NamespaceLookup(s.Device.TenantID)
	if len(errs) > 0 {
		log.WithError(errs[0]).Warn("unable to retrieve the namespace's connection announcement")

		return errs[0]
	}

	announcement := namespace.Settings.ConnectionAnnouncement

	if announcement == "" {
		return nil
	}

	if _, err := client.Write([]byte("Announcement:\n")); err != nil {
		return err
	}

	// Remove whitespaces and new lines at end
	announcement = strings.TrimRightFunc(announcement, func(r rune) bool {
		return r == ' ' || r == '\n' || r == '\t'
	})

	if _, err := client.Write([]byte("    " + strings.ReplaceAll(announcement, "\n", "\n    ") + "\n")); err != nil {
		return err
	}

	return nil
}

func (s *Session) Finish() error {
	if s.Dialed != nil {
		request, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/ssh/close/%s", s.UID), nil)

		if err := request.Write(s.Dialed); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
				Warning("Error when trying write the request to /ssh/close")
		}
	}

	if errs := s.api.FinishSession(s.UID); len(errs) > 0 {
		log.WithError(errs[0]).
			WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
			Error("Error when trying to finish the session")

		return errs[0]
	}

	return nil
}

type ClientConfigurationAuthentication func(gliderssh.Context, *gossh.ClientConfig) error

func ClientConfigurationAuthenticationPassword(password string) ClientConfigurationAuthentication {
	return func(ctx gliderssh.Context, config *gossh.ClientConfig) error {
		config.Auth = []gossh.AuthMethod{
			gossh.Password(password),
		}

		return nil
	}
}

func ClientConfigurationAuthenticationPublicKey() ClientConfigurationAuthentication {
	return func(ctx gliderssh.Context, config *gossh.ClientConfig) error {
		api := metadata.RestoreAPI(ctx)
		if api == nil {
			return errors.New("failed to get the API from context")
		}

		privateKey, err := api.CreatePrivateKey()
		if err != nil {
			return err
		}

		block, _ := pem.Decode(privateKey.Data)

		parsed, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return err
		}

		signer, err := gossh.NewSignerFromKey(parsed)
		if err != nil {
			return err
		}

		config.Auth = []gossh.AuthMethod{
			gossh.PublicKeys(signer),
		}

		return nil
	}
}

type AgentConfigurationOptions struct {
	Auth ClientConfigurationAuthentication
}

// NewAgentConnectionConfiguration creates a [gossh.ClientConfig] with the default configuration required by ShellHub
// to connect to the device agent that are inside the [gliderssh.Context].
func NewAgentConnectionConfiguration(ctx gliderssh.Context, opts AgentConfigurationOptions) (*gossh.ClientConfig, error) {
	target := metadata.RestoreTarget(ctx)
	if target == nil {
		return nil, errors.New("failed to get the target from context")
	}

	config := &gossh.ClientConfig{
		User:            target.Username,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), // nolint: gosec
	}

	if err := opts.Auth(ctx, config); err != nil {
		return nil, errors.New("failed to generate the authentication information")
	}

	return config, nil
}
