package session

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/host"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/target"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/webhandoff"
	log "github.com/sirupsen/logrus"
	gossh "golang.org/x/crypto/ssh"
)

type Data struct {
	Target *target.Target
	// SSHID is the combination of device's name and namespace name.
	SSHID string
	// Device is the device connected.
	Device *models.Device
	// Namespace is the namespace where device is located.
	Namespace *models.Namespace
	IPAddress string
	// Web reports whether the session originated from the web terminal.
	Web bool
	// UserID is the ShellHub account bound to this session: resolved from the
	// presented key in identity mode, or set by the enrollment/step-up approval.
	// Empty in legacy mode.
	UserID string
	// ApprovalCode is the JIT code minted in identity mode; the gateway polls
	// its decision for enrollment or step-up. Empty otherwise. The enrollment URL
	// derived from it is sent as a mid-handshake banner only once the presented
	// key turns out to be unenrolled, so an enrolled key never sees it.
	ApprovalCode string
	// Fingerprint is the presented SSH public key's fingerprint ("SHA256:…") in
	// identity mode; it is the identity lookup key.
	Fingerprint string
	// KeyData is the presented SSH public key in OpenSSH authorized_keys form,
	// carried so an enrollment can persist it.
	KeyData []byte
	// LastReauthAt is when the resolved identity last re-authenticated, read at
	// key resolution and compared against a policy's reauth_period to decide
	// whether a fresh re-auth is still needed. Nil when it never re-authed.
	LastReauthAt *time.Time
	// SingleUse marks a single-use identity (service accounts only), read at key
	// resolution. When set, the key is burned once this session establishes, so a
	// second connection with it is rejected.
	SingleUse bool
}

// AgentChannel represents a channel open between agent and server.
type AgentChannel struct {
	// Channel is an open channel for communication between the agent and the server.
	Channel gossh.Channel
	// Requests is the channel to handle SSH requests.
	Requests <-chan *gossh.Request
}

// Close closes the underlying agent channel connection.
func (a *AgentChannel) Close() error {
	return a.Channel.Close()
}

// Agent represents a connection to an agent.
type Agent struct {
	conn     net.Conn
	client   *gossh.Client
	mu       sync.Mutex
	channels map[int]*AgentChannel
}

// Close closes the underlying ssh client connection.
func (a *Agent) Close() error {
	a.mu.Lock()
	for _, channel := range a.channels {
		channel.Close() //nolint:errcheck
	}
	a.mu.Unlock()

	return a.client.Close()
}

// ClientChannel represents a channel open between client and server.
type ClientChannel struct {
	// Channel is an open channel for communication between the client and the server.
	Channel gossh.Channel
	// Requests is the channel to handle SSH requests.
	Requests <-chan *gossh.Request
}

// Close closes the underlying client channel connection.
func (c *ClientChannel) Close() error {
	return c.Channel.Close()
}

// Client represents a connection to a client.
type Client struct {
	mu       sync.Mutex
	channels map[int]*ClientChannel
}

// Close closes a connection to client and all its channels.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, channel := range c.channels {
		if err := channel.Close(); err != nil {
			return err
		}
	}

	return nil
}

// TODO: implement [io.Read] and [io.Write] on session to simplify the data piping.
type Session struct {
	// UID is the session's UID.
	UID string

	agent  *Agent
	client *Client

	service services.Service
	dialer  *dialer.Dialer
	// Events is a connection to the endpoint to save session's events.
	Events *Events

	once *sync.Once

	seats seats

	Data
}

// Seat represent a passenger in a session.
type Seat struct {
	// HasPty is the status of pty on the seat.
	HasPty bool
	// Type is the connection type requested on this seat ("exec", "subsystem",
	// …). It is per-seat and not per-session: one seat may run a shell while
	// another runs an exec on the same multiplexed connection.
	Type string
}

type seats struct {
	mu      *sync.Mutex
	counter *atomic.Int32
	items   *sync.Map
}

func newSeats() seats {
	return seats{
		mu:      new(sync.Mutex),
		counter: new(atomic.Int32),
		items:   new(sync.Map),
	}
}

// NewSeat creates a new seat inside seats.
func (s *seats) NewSeat() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := int(s.counter.Load())
	defer s.counter.Add(1)

	s.items.Store(id, &Seat{
		HasPty: false,
		Type:   "",
	})

	return id, nil
}

// Get returns a copy of the seat with the given id.
//
// It is a copy because the stored value is mutated in place by the setters: a
// caller holding the pointer would read fields without synchronization, and the
// pipe goroutine does exactly that.
func (s *seats) Get(seat int) (Seat, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.load(seat)
	if !ok {
		return Seat{}, false //nolint:exhaustruct
	}

	return *item, true
}

func (s *seats) load(seat int) (*Seat, bool) {
	loaded, ok := s.items.Load(seat)
	if !ok {
		return nil, false
	}

	item, ok := loaded.(*Seat)
	if !ok {
		return nil, false
	}

	return item, true
}

// SetPty sets a pty status to a seat from their id.
func (s *seats) SetPty(seat int, status bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.load(seat)
	if !ok {
		log.Warn("failed to set pty because no seat was created before")

		return
	}

	item.HasPty = status

	s.items.Store(seat, item)
}

// SetType sets the connection type on a seat from their id.
func (s *seats) SetType(seat int, kind string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.load(seat)
	if !ok {
		log.Warn("failed to set type because no seat was created before")

		return
	}

	item.Type = kind

	s.items.Store(seat, item)
}

// NewSession creates a new Session but differs from [New] as it only creates
// the session without registering, connecting to the agent, etc.
//
// It's designed to be used within New.
func NewSession(ctx gliderssh.Context, dialer *dialer.Dialer, service services.Service, handoff *webhandoff.Store) (*Session, error) {
	sshid := ctx.User()

	hos, err := host.NewHost(ctx.RemoteAddr().String())
	if err != nil {
		log.WithError(err).
			Error("failed to create a new host")

		return nil, ErrHost
	}

	target, err := target.NewTarget(sshid)
	if err != nil {
		return nil, err
	}

	var namespaceName, deviceName string
	var webUserID string
	web := false
	if target.IsSSHID() {
		namespaceName, deviceName, err = target.SplitSSHID()
		if err != nil {
			return nil, err
		}
	} else {
		if hos.IsLocalhost() {
			web = true

			data, ok := handoff.Take(sshid)
			if !ok {
				log.WithField("sshid", sshid).
					Error("no web session handoff for this connection")

				return nil, ErrWebData
			}

			if data.Device == "" || data.IP == "" {
				log.WithField("sshid", sshid).
					Error("web session handoff is incomplete")

				return nil, ErrWebData
			}

			target.Data = data.Device
			hos.Host = data.IP

			webUserID = data.UserID
		}

		device, err := service.GetDevice(ctx, scope.NewUnbounded(reasonSSHIDDeviceResolve), models.UID(target.Data))
		if err != nil {
			return nil, err
		}

		namespaceName = device.Namespace
		deviceName = device.Name
	}

	lookupDevice, err := service.LookupDevice(ctx, namespaceName, deviceName)
	if err != nil {
		return nil, err
	}

	namespace, err := service.GetNamespace(ctx, lookupDevice.TenantID)
	if err != nil {
		return nil, err
	}

	session := &Session{
		UID:     ctx.SessionID(),
		service: service,
		dialer:  dialer,
		Events:  NewEvents(ctx.SessionID(), service),
		Data: Data{
			IPAddress: hos.Host,
			Target:    target,
			Device:    lookupDevice,
			Namespace: namespace,
			Web:       web,
			UserID:    webUserID,
			SSHID:     fmt.Sprintf("%s@%s.%s", target.Username, namespaceName, deviceName),
		},
		once:  new(sync.Once),
		seats: newSeats(),
		agent: &Agent{
			channels: make(map[int]*AgentChannel),
		},
		client: &Client{
			channels: make(map[int]*ClientChannel),
		},
	}

	advance(ctx, session, StateCreated)

	return session, nil
}

// NewClientChannel accepts a new channel from a client and set a seat for it.
func (s *Session) NewClientChannel(newChannel gossh.NewChannel, seat int) (*ClientChannel, error) {
	s.client.mu.Lock()
	defer s.client.mu.Unlock()

	if _, ok := s.client.channels[seat]; ok {
		return nil, ErrSeatAlreadySet
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		return nil, err
	}

	c := &ClientChannel{
		Channel:  channel,
		Requests: requests,
	}

	s.client.channels[seat] = c

	return c, nil
}

// NewAgentChannel opens a new channel to agent and set a seat for it.
func (s *Session) NewAgentChannel(name string, seat int) (*AgentChannel, error) {
	s.agent.mu.Lock()
	defer s.agent.mu.Unlock()

	if _, ok := s.agent.channels[seat]; ok {
		return nil, ErrSeatAlreadySet
	}

	channel, requests, err := s.agent.client.OpenChannel(name, nil)
	if err != nil {
		return nil, err
	}

	a := &AgentChannel{
		Channel:  channel,
		Requests: requests,
	}

	s.agent.channels[seat] = a

	return a, nil
}

// DropAgentChannel closes the agent channel on a seat and forgets it, so a seat
// whose client side failed to open does not keep a live channel behind it.
func (s *Session) DropAgentChannel(seat int) {
	s.agent.mu.Lock()
	defer s.agent.mu.Unlock()

	channel, ok := s.agent.channels[seat]
	if !ok {
		return
	}

	channel.Close() //nolint:errcheck
	delete(s.agent.channels, seat)
}

func (s *Session) checkFirewall(ctx context.Context) error {
	err := s.service.EvaluateFirewall(ctx, models.FirewallConnection{
		Namespace: s.Namespace.Name,
		Hostname:  s.Device.Name,
		Username:  s.Target.Username,
		IPAddress: s.IPAddress,
	})
	if err == nil {
		return nil
	}

	log.WithError(err).WithFields(log.Fields{
		"uid":    s.UID,
		"sshid":  s.SSHID,
		"tenant": s.Namespace.TenantID,
	}).Info("an error or a firewall rule block this connection")

	if errors.Is(err, services.ErrFirewallBlocked) {
		return ErrFirewallBlock
	}

	return ErrFirewallUnknown
}

func (s *Session) checkBilling(ctx context.Context) error {
	err := s.service.EvaluateBilling(ctx, s.Device.TenantID)
	if errors.Is(err, services.ErrBillingBlocked) {
		return ErrBillingBlock
	}

	return err
}

func (s *Session) register(ctx context.Context) error {
	_, err := s.service.CreateSession(ctx, requests.SessionCreate{
		UID:       s.UID,
		DeviceUID: s.Device.UID,
		Username:  s.Target.Username,
		UserID:    s.UserID,
		IPAddress: s.IPAddress,
		Type:      "none",
		Term:      "none",
		Web:       s.Web,
	})
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
			Error("Error when trying to register the client on API")

		return err
	}

	return nil
}

func (s *Session) authenticate(ctx context.Context) error {
	value := true

	return s.service.UpdateSession(ctx, models.UID(s.UID), models.SessionUpdate{ //nolint:exhaustruct
		Authenticated: &value,
	})
}

// Errors reported by [Session.Recorded] when the session was never going to be recorded. They
// are expected conditions, not failures, and callers should not report them as such. The
// specific ones wrap [ErrRecordingSkipped] so a caller that only wants to know whether anything
// went wrong needs a single comparison.
var (
	ErrRecordingSkipped  = errors.New("session recording skipped")
	ErrRecordingDisabled = fmt.Errorf("%w: disabled for this namespace", ErrRecordingSkipped)
	ErrRecordingNoPty    = fmt.Errorf("%w: session has no pty", ErrRecordingSkipped)
)

// Recorded marks the session as recorded on the API.
//
// It returns an error wrapping [ErrRecordingSkipped] when the session is not meant to be
// recorded; any other error means marking it failed.
func (s *Session) Recorded(seat int) error {
	value := true

	if !s.Namespace.Settings.SessionRecord {
		return ErrRecordingDisabled
	}

	if seat, ok := s.seats.Get(seat); !ok || !seat.HasPty {
		return ErrRecordingNoPty
	}

	return s.service.UpdateSession(context.Background(), models.UID(s.UID), models.SessionUpdate{ //nolint:exhaustruct
		Recorded: &value,
	})
}

func (s *Session) connect(ctx gliderssh.Context, authOpt authFunc) error {
	config := &gossh.ClientConfig{
		User:            s.Target.Username,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), //nolint:gosec
		Timeout:         sshconf.ConnectTimeout,
	}

	if err := authOpt(s, config); err != nil {
		return errors.New("fail to generate the authentication information")
	}

	const Addr = "tcp"

	if s.agent.conn == nil {
		if err := s.Dial(ctx); err != nil {
			return err
		}
	}

	if config.Timeout > 0 {
		if err := s.agent.conn.SetDeadline(clock.Now().Add(config.Timeout)); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
				Error("Error when trying to set dial deadline")

			return err
		}
	}

	conn, chans, reqs, err := gossh.NewClientConn(s.agent.conn, Addr, config)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": s.UID}).
			Error("Error when trying to create the client's connection")

		s.agent.conn = nil

		return err
	}

	if config.Timeout > 0 {
		if err := s.agent.conn.SetDeadline(time.Time{}); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
				Error("Error when trying to set dial deadline with Time{}")

			return err
		}
	}

	ch := make(chan *gossh.Request)
	close(ch)

	s.agent.client = gossh.NewClient(conn, chans, ch)

	go s.drainAgentRequests(ctx, reqs)

	return nil
}

const keepAliveRequestPrefix = "keepalive"

// KeepAliveRequestType is the keepalive the gateway forwards to the client.
const KeepAliveRequestType = keepAliveRequestPrefix + "@shellhub.io"

func (s *Session) drainAgentRequests(ctx gliderssh.Context, reqs <-chan *gossh.Request) {
	logger := log.WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID})

	defer logger.Trace("agent global requests drained")

	for req := range reqs {
		if !strings.HasPrefix(req.Type, keepAliveRequestPrefix) {
			if req.WantReply {
				if err := req.Reply(false, nil); err != nil {
					logger.WithError(err).Warn("failed to refuse a global request from the agent")
				}
			}

			continue
		}

		if req.WantReply {
			if err := req.Reply(true, nil); err != nil {
				logger.WithError(err).Warn("failed to reply to the keepalive from the agent")
			}
		}

		if err := s.KeepAlive(ctx); err != nil {
			logger.WithError(err).Warn("failed to record the session as alive")
		}

		if conn, ok := ctx.Value(gliderssh.ContextKeyConn).(gossh.Conn); ok && conn != nil {
			if _, _, err := conn.SendRequest(KeepAliveRequestType, false, req.Payload); err != nil {
				logger.WithError(err).Warn("failed to forward the keepalive to the client")
			}
		}
	}
}

var ErrDialUnknown = errors.New("unknown protocol version")

// Dial establishes the underlying transport to the target device. For V1
// transports an HTTP GET request is issued (legacy reverse tunnel). For
// V2 transports a multistream protocol selection is performed using the
// ProtoSSHOpen identifier followed by a JSON envelope with the session
// id. After this method returns s.agent.conn is a raw channel ready for
// SSH key exchange and channel opens.
func (s *Session) Dial(ctx gliderssh.Context) error {
	var err error

	ctx.Lock()
	defer ctx.Unlock()

	conn, err := s.dialer.DialTo(ctx, s.Device.TenantID, s.Device.UID, dialer.SSHOpenTarget{SessionID: s.UID})
	if err != nil {
		log.WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).WithError(err).Error("failed to open ssh session")

		return errors.Join(ErrDial, err)
	}

	s.agent.conn = conn

	return nil
}

func (s *Session) checkLicense(ctx context.Context) error {
	err := s.service.EvaluateLicense(ctx)
	if errors.Is(err, services.ErrLicenseBlocked) {
		return ErrLicenseBlock
	}

	return err
}

func (s *Session) Evaluate(ctx gliderssh.Context) error {
	if envs.IsEnterprise() {
		if err := s.checkLicense(ctx); err != nil {
			return err
		}
	}

	if envs.IsEnterpriseOrCloud() && !s.Namespace.Settings.IsIdentityAccess() {
		if err := s.checkFirewall(ctx); err != nil {
			return err
		}
	}

	if envs.IsCloud() {
		if err := s.checkBilling(ctx); err != nil {
			return err
		}
	}

	if s.Namespace.Settings.IsIdentityAccess() {
		has, err := s.service.NamespaceHasAccessPolicies(ctx, s.Namespace.TenantID)
		if err != nil {
			return err
		}

		if !has {
			return ErrAccessDenied
		}
	}

	advance(ctx, s, StateEvaluated)

	return nil
}

func (s *Session) openApproval(ctx context.Context, kind models.SSHApprovalKind, reauthPeriod *int) error {
	approval, err := s.service.CreateSSHApproval(ctx, &requests.SSHApprovalCreate{
		SessionUID:   s.UID,
		SSHID:        s.SSHID,
		TenantID:     s.Namespace.TenantID,
		DeviceUID:    s.Device.UID,
		DeviceName:   s.Device.Name,
		Username:     s.Target.Username,
		IPAddress:    s.IPAddress,
		Kind:         kind,
		Fingerprint:  s.Fingerprint,
		Data:         s.KeyData,
		ReauthPeriod: reauthPeriod,
	})
	if err != nil {
		return err
	}

	s.ApprovalCode = approval.Code

	return nil
}

// IsIdentityMode reports whether the session's namespace uses the identity-based
// SSH access mode, where the presented key is the identity.
func (s *Session) IsIdentityMode() bool {
	return s.Namespace.Settings.IsIdentityAccess()
}

func consoleURL(domain string, autoSSL bool, path string) string {
	scheme := "http"
	if autoSSL {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s%s", scheme, domain, path)
}

func groupCode(code string) string {
	if len(code) != pairingcode.DeviceCodeLength {
		return code
	}

	half := len(code) / 2

	return code[:half] + " " + code[half:]
}

func buildAddKeyBanner(domain string, autoSSL bool, code, fingerprint string) string {
	lines := []string{
		"",
		"  ShellHub doesn't know this SSH key yet.",
		"",
		"  Open the link to add it to your identities. This login",
		"  continues once you do:",
		"",
		"    " + consoleURL(domain, autoSSL, "/ssh-identities/new/"+code),
		"",
		"  Security code:  " + groupCode(code),
		"  Key:            " + fingerprint,
		"",
		"  Waiting...",
		"",
	}

	return strings.Join(lines, "\r\n")
}

func buildReauthBanner(domain string, autoSSL bool, code string) string {
	lines := []string{
		"",
		"  An access policy asks you to re-authenticate.",
		"",
		"  Open the link to do it in the console. This login",
		"  continues once you do:",
		"",
		"    " + consoleURL(domain, autoSSL, "/ssh-identities/confirm/"+code),
		"",
		"  Security code:  " + groupCode(code),
		"",
		"  Waiting...",
		"",
	}

	return strings.Join(lines, "\r\n")
}

// Auth authenticate a [Session] based on the provided context.
//
// As a client may try to create N sessions with the same context, a [snapshot] is used
// to save/retrieve the current session state. To illustrate a practical use of this
// pattern you can imagine a client that wants to connect to a specified device. It first
// calls the `PublicKeyVerified` handler with a specified context. At this stage, there are no
// sessions associated with the provided context, and a new one will be created. If it
// fails, the same client (and consequently the same context) will call the
// `PasswordHandler`, which also calls `session.New`. Since we have already created a
// session in the previous authentication attempt, instead of repeating all operations,
// we can safely retrieve the same session again but attempt authentication with a
// password this time.
//
// Next steps can use the context's snapshot to retrieve the created session. An error is
// returned if any occurs.
func (s *Session) Auth(ctx gliderssh.Context, auth Auth) error {
	sess, state := ObtainSession(ctx)
	switch state {
	case StateEvaluated:
		if err := auth.Evaluate(sess); err != nil {
			return err
		}

		if err := sess.register(ctx); err != nil {
			return err
		}

		advance(ctx, sess, StateRegistered)

		fallthrough
	case StateRegistered:
		if err := sess.connect(ctx, auth.Auth()); err != nil {
			return err
		}

		if sess.SingleUse {
			won, err := sess.service.ConsumeSSHIdentity(ctx, sess.Namespace.TenantID, sess.Fingerprint)
			if err != nil {
				return err
			}

			if !won {
				return ErrAccessDenied
			}
		}

		if err := sess.authenticate(ctx); err != nil {
			return err
		}
	default:
		return errors.New("invalid session state")
	}

	advance(ctx, sess, StateFinished)

	return nil
}

func (s *Session) NewSeat() (int, error) {
	return s.seats.NewSeat()
}

// Seat returns a copy of the seat with the given id.
func (s *Session) Seat(seat int) (Seat, bool) {
	return s.seats.Get(seat)
}

// SetSeatPty records whether the seat has a pty.
func (s *Session) SetSeatPty(seat int, status bool) {
	s.seats.SetPty(seat, status)
}

// SetSeatType records the connection type requested on the seat, such as
// [SeatTypeExec].
func (s *Session) SetSeatType(seat int, kind string) {
	s.seats.SetType(seat, kind)
}

// LogFields returns the fields that identify this session in a log line.
//
// It tolerates a partially built session: log calls sit on failure paths, where
// the session may not have got far enough to have a device.
func (s *Session) LogFields() log.Fields {
	fields := log.Fields{
		"session": s.UID,
		"sshid":   s.SSHID,
		"ip":      s.IPAddress,
	}

	if s.Device != nil {
		fields["device"] = s.Device.UID
	}

	if s.Target != nil {
		fields["username"] = s.Target.Username
	}

	return fields
}

// Event registers an event to the session.
func (s *Session) Event(t string, data any, seat int) {
	s.Events.Write(models.SessionEvent{
		Session:   s.UID,
		Type:      models.SessionEventType(t),
		Timestamp: clock.Now(),
		Data:      data,
		Seat:      seat,
	})
}

// EventWriter registers events against a session. [Session] implements it; it
// exists so the channel handlers can record events through a fake.
type EventWriter interface {
	Event(t string, data any, seat int)
}

// Event registers an event whose payload is an SSH request body, decoded into D.
func Event[D any](sess EventWriter, t string, data []byte, seat int) {
	d := new(D)
	if err := gossh.Unmarshal(data, d); err != nil {
		return
	}

	sess.Event(t, d, seat)
}

func (s *Session) KeepAlive(ctx context.Context) error {
	if err := s.service.KeepAliveSession(ctx, models.UID(s.UID)); err != nil {
		log.WithError(err).
			WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
			Error("Error when trying to keep alive the session")

		return err
	}

	return nil
}

// Announce is a custom message provided by the end user that can be printed when a new connection within the namespace
// is established.
//
// Returns the announcement or an error, if any. If no announcement is set, it returns an empty string.
func (s *Session) Announce(client gossh.Channel) error {
	if _, err := client.Write([]byte(
		"Connected to " + s.SSHID + " via ShellHub.\n\r",
	)); err != nil {
		return err
	}

	announcement := s.Namespace.Settings.ConnectionAnnouncement

	if announcement == "" {
		return nil
	}

	announcement = strings.TrimRightFunc(announcement, func(r rune) bool {
		return r == ' ' || r == '\n' || r == '\t'
	})

	if _, err := client.Write([]byte(strings.ReplaceAll(announcement, "\n", "\n\r") + "\n\r")); err != nil {
		return err
	}

	return nil
}

// Finish terminates the session between Agent and Client, sending a request to Agent to closes it.
func (s *Session) Finish() (err error) {
	s.once.Do(func() {
		log.WithFields(log.Fields{
			"uid": s.UID,
		}).Trace("session finish called")

		s.Events.Close() //nolint:errcheck

		if s.agent.conn != nil {
			request, _ := http.NewRequestWithContext(context.Background(), http.MethodDelete, "/ssh/close/"+s.UID, nil)

			if err = request.Write(s.agent.conn); err != nil {
				log.WithError(err).
					WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
					Warning("Error when trying write the request to /ssh/close")
			}
		}

		if err := s.service.DeactivateSession(context.Background(), models.UID(s.UID)); err != nil {
			log.WithError(err).
				WithFields(log.Fields{"session": s.UID, "sshid": s.SSHID}).
				Error("Error when trying to finish the session")
		}

		log.WithFields(
			log.Fields{
				"uid":      s.UID,
				"device":   s.Device.UID,
				"username": s.Target.Username,
				"ip":       s.IPAddress,
			}).Info("session finished")
	})

	return nil
}
