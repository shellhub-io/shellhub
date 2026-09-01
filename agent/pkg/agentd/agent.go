// Package agentd provides packages and functions to create a new ShellHub Agent instance.
//
// The ShellHub Agent is a lightweight software component that runs the device and provide communication between the
// device and ShellHub's server. Its main role is to provide a reserve SSH server always connected to the ShellHub
// server, allowing SSH connections to be established to the device even when it is behind a firewall or NAT.
//
// This package provides a simple API to create a new agent instance and start the communication with the server. The
// agent will automatically connect to the server and start listening for incoming connections. Once connected, the
// agent will also automatically reconnect to the server if the connection is lost.
//
// The update process isn't handled by this package. This feature is provided by its main implementation in
// [ShellHub Agent]. Check the [ShellHub Agent] documentation for more information.
//
// # Example:
//
// Creates the agent configuration with the minimum required fields:
//
//	func main() {
//	    cfg := Config{
//	        ServerAddress: "http://localhost:80",
//	        TenantID:      "00000000-0000-4000-0000-000000000000",
//	        PrivateKey:    "/tmp/shellhub.key",
//	    }
//
//	    ctx := context.Background()
//	    ag, err := NewAgentWithConfig(&cfg, new(HostMode))
//	    if err != nil {
//	        panic(err)
//	    }
//
//	    if err := ag.Initialize(); err != nil {
//	        panic(err)
//	    }
//
//	    ag.Listen(ctx)
//	}
//
// # Embedding the agent
//
// This package is meant to be importable so the agent can run in-process inside another Go
// program (for example, ShellHub Desktop). Note that the agent's go.mod replaces
// github.com/gliderlabs/ssh with github.com/shellhub-io/ssh (a fork). Go ignores replace
// directives from non-main modules, so any program embedding this package must replicate that
// same replace directive in its own go.mod.
//
// [ShellHub Agent]: https://github.com/shellhub-io/shellhub/tree/master/agent
package agentd

import (
	"context"
	"crypto/rsa"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"github.com/shellhub-io/shellhub/agent/pkg/keygen"
	"github.com/shellhub-io/shellhub/agent/pkg/sysinfo"
	"github.com/shellhub-io/shellhub/agent/pkg/tunnel"
	"github.com/shellhub-io/shellhub/agent/server"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
	log "github.com/sirupsen/logrus"
)

// Config provides the configuration for the agent service.
type Config struct {
	// Set the ShellHub Cloud server address the agent will use to connect.
	// This is required.
	ServerAddress string `env:"SERVER_ADDRESS,required" validate:"required"`

	// Specify the path to the device private key.
	// If not provided, the agent will generate a new one.
	// This is required.
	PrivateKey string `env:"PRIVATE_KEY,required" validate:"required"`

	// Sets the account tenant id used during communication to associate the
	// device to a specific tenant.
	//
	// It is optional: when empty (and no tenant was persisted from a previous
	// pairing), the agent boots into pairing mode and waits for a user to
	// accept it into a namespace, learning the tenant from the server.
	TenantID string `env:"TENANT_ID"`

	// PairingCode is a pre-authorized pairing code handed to the agent at install
	// time (minted from the console's Add Device page). When set and no tenant is
	// configured, the agent claims it: the server accepts the device into the
	// code's namespace automatically, so it never lands in the pending list.
	PairingCode string `env:"PAIRING_CODE"`

	// InstallKey is a reusable install key handed to the agent at install time (minted from the
	// console's Install Keys page). The key is namespace-scoped, so it enrolls the device on its own:
	// with no TenantID configured the server resolves the namespace from the key, applying the key's
	// mode, tags and ephemeral flag. It may also ride alongside TenantID, which enrolls the same way.
	InstallKey string `env:"INSTALL_KEY"`

	// Determine the interval to send the keep alive message to the server. This
	// has a direct impact of the bandwidth used by the device when in idle
	// state. Default is 30 seconds.
	KeepAliveInterval uint32 `env:"KEEPALIVE_INTERVAL,overwrite,default=30"`

	// Set the device preferred hostname. This provides a hint to the server to
	// use this as hostname if it is available.
	PreferredHostname string `env:"PREFERRED_HOSTNAME"`

	// Set the device preferred identity. This provides a hint to the server to
	// use this identity if it is available.
	PreferredIdentity string `env:"PREFERRED_IDENTITY,default="`

	// Stores the password for single-user mode (without root privileges). If not
	// provided, multi-user mode (with root privileges) is enabled by default.
	// NOTE: The password hash could be generated by ```openssl passwd```.
	SingleUserPassword string `env:"SINGLE_USER_PASSWORD,default=$SIMPLE_USER_PASSWORD"`

	// SimpleUserPassword exists due to a typo on the environmental variable that stores the password for single user
	// mode that was wrongly named `SIMPLE_USER_PASSWORD` instead of `SINGLE_USER_PASSWORD`, and willing to keep the
	// compatibility, this new variable was created.
	SimpleUserPassword string `env:"SIMPLE_USER_PASSWORD"`

	// MaxRetryConnectionTimeout specifies the maximum time, in seconds, that an agent will wait
	// before attempting to reconnect to the ShellHub server. Default is 60 seconds.
	MaxRetryConnectionTimeout int `env:"MAX_RETRY_CONNECTION_TIMEOUT,default=60" validate:"min=10,max=120"`

	// TransportVersion specifies the version of the agent transport protocol to use.
	// Version 1 uses HTTP-based revdial, version 2 uses yamux multiplexing with multistream.
	// Supported values are 1 and 2. Default is 2.
	TransportVersion int `env:"TRANSPORT_VERSION,default=2"`

	// Version is the agent version reported to the server and embedded in the device info.
	// The CLI injects the value set at build time via `-ldflags -X main.AgentVersion=...`.
	// Embedders must set it explicitly.
	Version string

	// Platform identifies the platform the agent is running on (e.g. "native", "docker",
	// "connector"). The CLI injects the value detected at build time; embedders set it
	// explicitly.
	Platform string

	// SFTPServerCommand builds the command used to start the SFTP server subprocess. When nil,
	// the agent re-executes its own binary (/proc/self/exe) with the "sftp" subcommand. An
	// embedding program (where /proc/self/exe is not the agent binary) must set this to point
	// at a binary/subcommand that runs the SFTP server.
	SFTPServerCommand func() *exec.Cmd
}

// HasNamespaceCredential reports whether the configuration carries something naming the namespace
// the device enrolls into: a tenant ID, or an install key, which is namespace-scoped and so resolves
// one on its own. A configuration with neither has to pair, the only path that waits on a user.
// PairingCode does not count, as the pairing flow claims it and returns the tenant it resolved.
func (c *Config) HasNamespaceCredential() bool {
	return c.TenantID != "" || c.InstallKey != ""
}

// LoadConfigFromEnv reads the agent's configuration from SHELLHUB_-prefixed environment
// variables, falling back to the .env file next to the binary when one is present.
//
// The second return value carries the environment as parsed, for callers that log it.
func LoadConfigFromEnv() (*Config, map[string]any, error) {
	applyEnvFileFallback(defaultEnvFilePath)

	cfg, err := envs.ParseWithPrefix[Config]("SHELLHUB_")
	if err != nil {
		log.Error("failed to parse the configuration")

		return nil, nil, err
	}

	if ok, fields, err := validator.New().StructWithFields(cfg); err != nil || !ok {
		log.WithFields(fields).Error("failed to validate the configuration loaded from envs")

		return nil, fields, err
	}

	if persisted, err := ReadPersistedTenant(TenantFilePath(cfg.PrivateKey)); err == nil && persisted != "" {
		switch {
		case cfg.TenantID == "":
			cfg.TenantID = persisted
		case cfg.TenantID != persisted:
			log.WithFields(log.Fields{
				"env_tenant":       cfg.TenantID,
				"persisted_tenant": persisted,
			}).Warn("tenant from environment overrides the tenant persisted by pairing")
		}
	}

	return cfg, nil, nil
}

// Agent is a device's connection to a ShellHub server: it authenticates, keeps the device
// record current, and serves the SSH sessions the server routes to it.
//
// Use [NewAgentWithConfig] to build one, [Agent.Initialize] to authenticate, and
// [Agent.Listen] to serve.
type Agent struct {
	config     *Config
	pubKey     *rsa.PublicKey
	Identity   *models.DeviceIdentity
	Info       *models.DeviceInfo
	authData   *models.DeviceAuthResponse
	cli        client.Client
	serverInfo *models.Info
	server     *server.Server
	listening  chan bool
	closed     atomic.Bool
	mode       Mode
	listener   atomic.Pointer[net.Listener]
	logger     *log.Entry
}

// NewAgent creates a new agent instance, requiring the ShellHub server's address to connect to, the namespace's tenant
// where device own and the path to the private key on the file system.
//
// It builds a minimal [Config], leaving optional fields (including Version and Platform) unset. As a result the device
// is registered with an empty version. Embedders that need to report a version or override other defaults should use
// [NewAgentWithConfig] with a fully populated [Config].
func NewAgent(address string, tenantID string, privateKey string, mode Mode) (*Agent, error) {
	return NewAgentWithConfig(&Config{
		ServerAddress:    address,
		TenantID:         tenantID,
		PrivateKey:       privateKey,
		TransportVersion: TransportV2,
	}, mode)
}

// Reasons [NewAgentWithConfig] rejects a configuration, and why enrolment cannot start.
var (
	ErrNewAgentWithConfigEmptyServerAddress   = errors.New("address is empty")
	ErrNewAgentWithConfigInvalidServerAddress = errors.New("address is invalid")
	ErrNewAgentWithConfigEmptyPrivateKey      = errors.New("private key is empty")
	ErrNewAgentWithConfigNilMode              = errors.New("agent's mode is nil")

	ErrNewAgentWithConfigUnsupportedTransportVersion = errors.New("transport version is unsupported")

	ErrAuthorizeNoNamespaceCredential = errors.New("no tenant or install key to enroll with")
)

// NewAgentWithConfig creates a new agent instance with all configurations.
//
// The tenant may be empty at this point: a tenant-less agent can run [Agent.Setup]
// and pair interactively, but [Agent.Authorize] requires a tenant.
//
// Check [Config] for more information.
func NewAgentWithConfig(config *Config, mode Mode) (*Agent, error) {
	if config.ServerAddress == "" {
		return nil, ErrNewAgentWithConfigEmptyServerAddress
	}

	if _, err := url.ParseRequestURI(config.ServerAddress); err != nil {
		return nil, ErrNewAgentWithConfigInvalidServerAddress
	}

	if config.PrivateKey == "" {
		return nil, ErrNewAgentWithConfigEmptyPrivateKey
	}

	if mode == nil {
		return nil, ErrNewAgentWithConfigNilMode
	}

	switch config.TransportVersion {
	case TransportV1, TransportV2:
	default:
		return nil, ErrNewAgentWithConfigUnsupportedTransportVersion
	}

	return &Agent{
		config: config,
		mode:   mode,
	}, nil
}

// Initialize initializes the ShellHub Agent, generating device identity, loading device information, generating private
// key, reading public key, probing server information and authorizing device on ShellHub server.
//
// When any of the steps fails, the agent will return an error, and the agent will not be able to start.
func (a *Agent) Initialize() error {
	if err := a.Setup(); err != nil {
		return err
	}

	return a.Authorize()
}

// Setup prepares the agent for server communication without authorizing it:
// HTTP client, device identity, device info, key pair and server probe. None
// of these require a tenant, so a tenant-less agent can run Setup and pair.
func (a *Agent) Setup() error {
	var err error

	a.cli, err = client.NewClient(a.config.ServerAddress, client.WithVersion(a.config.Version))
	if err != nil {
		return errors.Wrap(err, "failed to create the HTTP client")
	}

	if err := a.generateDeviceIdentity(); err != nil {
		return errors.Wrap(err, "failed to generate device identity")
	}

	if err := a.loadDeviceInfo(); err != nil {
		return errors.Wrap(err, "failed to load device info")
	}

	if err := a.generatePrivateKey(); err != nil {
		return errors.Wrap(err, "failed to generate private key")
	}

	if err := a.readPublicKey(); err != nil {
		return errors.Wrap(err, "failed to read public key")
	}

	if err := a.probeServerInfo(); err != nil {
		return errors.Wrap(err, "failed to probe server info")
	}

	return nil
}

// Authorize registers the device on the ShellHub server within its namespace.
// [Agent.Setup] must have been run first, and the device must carry something
// naming a namespace: a tenant (from configuration, or injected with
// [Agent.SetTenantID] after a pairing) or an install key, which is
// namespace-scoped and so enrolls on its own. The tenant an install key
// resolved to is adopted from the server's response.
func (a *Agent) Authorize() error {
	if !a.config.HasNamespaceCredential() {
		return ErrAuthorizeNoNamespaceCredential
	}

	if err := a.authorize(); err != nil {
		return errors.Wrap(err, "failed to authorize device")
	}

	if a.config.TenantID == "" {
		a.config.TenantID = a.authData.TenantID
	}

	a.closed.Store(false)

	a.logger = log.WithFields(log.Fields{
		"version":           a.config.Version,
		"tenant_id":         a.config.TenantID,
		"namespace":         a.authData.Namespace,
		"server_address":    a.config.ServerAddress,
		"ssh_endpoint":      a.serverInfo.Endpoints.SSH,
		"api_endpoint":      a.serverInfo.Endpoints.API,
		"transport_version": a.config.TransportVersion,
		"sshid":             fmt.Sprintf("%s.%s@%s", a.authData.Namespace, a.authData.Name, strings.Split(a.serverInfo.Endpoints.SSH, ":")[0]),
	})

	return nil
}

// SetTenantID injects the tenant learned from a pairing so the agent can be
// authorized.
func (a *Agent) SetTenantID(tenant string) {
	a.config.TenantID = tenant
}

// ClearPairingCode drops a pre-authorized pairing code after the server rejected
// it, so a retry falls back to a normal (user-accepted) pairing instead of
// re-sending the dead code.
func (a *Agent) ClearPairingCode() {
	a.config.PairingCode = ""
}

func cleanKeyPath(raw string) (string, error) {
	cleaned := filepath.Clean(raw)
	if cleaned != raw {
		return "", keygen.ErrPathTraversal
	}

	return cleaned, nil
}

func (a *Agent) generatePrivateKey() error {
	keyPath, err := cleanKeyPath(a.config.PrivateKey)
	if err != nil {
		return err
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		if err := keygen.GeneratePrivateKey(keyPath); err != nil {
			return err
		}
	}

	return nil
}

func (a *Agent) readPublicKey() error {
	keyPath, err := cleanKeyPath(a.config.PrivateKey)
	if err != nil {
		a.pubKey = nil

		return err
	}

	key, err := keygen.ReadPublicKey(keyPath)
	a.pubKey = key

	return err
}

func (a *Agent) generateDeviceIdentity() error {
	if id := a.config.PreferredIdentity; id != "" {
		a.Identity = &models.DeviceIdentity{
			MAC: id,
		}

		return nil
	}

	iface, err := sysinfo.PrimaryInterface()
	if err != nil {
		return err
	}

	a.Identity = &models.DeviceIdentity{
		MAC: iface.HardwareAddr.String(),
	}

	return nil
}

func (a *Agent) loadDeviceInfo() error {
	info, err := a.mode.GetInfo()
	if err != nil {
		return err
	}

	a.Info = &models.DeviceInfo{
		ID:         info.ID,
		PrettyName: info.Name,
		Version:    a.config.Version,
		Platform:   a.config.Platform,
		Arch:       runtime.GOARCH,
	}

	return nil
}

func (a *Agent) probeServerInfo() error {
	info, err := a.cli.GetInfo(a.config.Version)
	if err != nil {
		return err
	}

	a.serverInfo = info

	return nil
}

// ErrNoIdentityAndHostname is returned when the device can name itself neither by MAC nor by
// hostname, so the server would have nothing to identify it by.
var ErrNoIdentityAndHostname = errors.New("the device doesn't have a valid hostname and identity. Set PREFERRED_IDENTITY or PREFERRED_HOSTNAME to specify the device's name and identity")

func (a *Agent) buildDeviceAuth() (*models.DeviceAuth, error) {
	auth := &models.DeviceAuth{
		Hostname:   a.config.PreferredHostname,
		Identity:   a.Identity,
		TenantID:   a.config.TenantID,
		PublicKey:  string(keygen.EncodePublicKeyToPem(a.pubKey)),
		InstallKey: a.config.InstallKey,
	}

	if auth.Hostname == "" && (auth.Identity == nil || auth.Identity.MAC == "") {
		return nil, ErrNoIdentityAndHostname
	}

	return auth, nil
}

func (a *Agent) authorize() error {
	auth, err := a.buildDeviceAuth()
	if err != nil {
		return err
	}

	req := &models.DeviceAuthRequest{
		Info:       a.Info,
		DeviceAuth: auth,
	}

	data, err := a.cli.AuthDevice(req)
	if err != nil {
		return err
	}

	a.authData = data

	return err
}

// CreatePairing submits this tenant-less agent's identity to the server and
// returns a short-lived pairing code. [Agent.Setup] must have been run first.
//
// When the agent was configured with a pre-authorized PairingCode, it is sent
// along so the server claims it and accepts the device immediately, returning
// an "accepted" pairing with the tenant instead of a code to poll.
func (a *Agent) CreatePairing() (*models.DevicePairing, error) {
	auth, err := a.buildDeviceAuth()
	if err != nil {
		return nil, err
	}

	return a.cli.CreateDevicePairing(&models.DevicePairingRequest{
		Hostname:  auth.Hostname,
		Identity:  auth.Identity,
		Info:      a.Info,
		PublicKey: auth.PublicKey,
		Code:      a.config.PairingCode,
	})
}

// GetPairingStatus polls the outcome of a pairing code.
func (a *Agent) GetPairingStatus(code string) (*models.DevicePairingStatus, error) {
	return a.cli.GetDevicePairingStatus(code)
}

// CreateDeviceLoginCode requests a short-lived code that deep-links this device into the
// console's accept page. The agent must be initialized first.
func (a *Agent) CreateDeviceLoginCode() (*models.DeviceLoginCode, error) {
	return a.cli.CreateDeviceLoginCode(a.authData.Token)
}

// DeviceStatus returns the device's current status on the server. The agent must be
// initialized first.
func (a *Agent) DeviceStatus() (models.DeviceStatus, error) {
	res, err := a.cli.GetDeviceAuthStatus(a.authData.Token)
	if err != nil {
		return models.DeviceStatusEmpty, err
	}

	return res.Status, nil
}

// Namespace returns the name of the namespace the device belongs to. The agent must be
// initialized first.
func (a *Agent) Namespace() string {
	return a.authData.Namespace
}

func (a *Agent) isClosed() bool {
	return a.closed.Load()
}

// Close closes the ShellHub Agent's listening, stoping it from receive new connection requests.
func (a *Agent) Close() error {
	a.closed.Store(true)

	l := a.listener.Load()
	if l == nil {
		return nil
	}

	return (*l).Close()
}

// The transport protocols an agent can speak to the server. V1 tunnels HTTP over the
// connection; V2 multiplexes streams by protocol name.
const (
	TransportV1 = 1
	TransportV2 = 2
)

// Listen serves connections until ctx is cancelled, using the transport named by the
// configuration.
func (a *Agent) Listen(ctx context.Context) error {
	a.mode.Serve(a)

	switch a.config.TransportVersion {
	case TransportV1:
		return a.listenV1(ctx)
	case TransportV2:
		return a.listenV2(ctx)
	default:
		return fmt.Errorf("unsupported transport version: %d", a.config.TransportVersion)
	}
}

func (a *Agent) listenV1(ctx context.Context) error {
	tun := tunnel.NewTunnelV1()

	tun.Handle(HandleSSHOpenV1, sshHandlerV1(a))
	tun.Handle(HandleSSHCloseV1, sshCloseHandlerV1(a))
	tun.Handle(HandleHTTPProxyV1, httpProxyHandlerV1(a))

	go a.ping(ctx, AgentPingDefaultInterval) //nolint:errcheck

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		for {
			if a.isClosed() {
				a.logger.Info("Stopped listening for connections")

				cancel()

				return
			}

			ShellHubConnectV1Path := "/ssh/connection"

			a.logger.Debug("Using tunnel version 1")

			listener, err := a.cli.NewReverseListenerV1(
				ctx,
				a.authData.Token,
				ShellHubConnectV1Path,
			)
			if err != nil {
				a.logger.Error("Failed to connect to server through reverse tunnel. Retry in 10 seconds")

				time.Sleep(time.Second * 10)

				continue
			}
			a.listener.Store(&listener)

			a.logger.Info("Server connection established")

			a.listening <- true

			if err := tun.Listen(ctx, listener); err != nil {
				a.logger.WithError(err).Error("Tunnel listener exited with error")
			}

			a.listening <- false
		}
	}()

	<-ctx.Done()

	return a.Close()
}

func (a *Agent) listenV2(ctx context.Context) error {
	tun := tunnel.NewTunnelV2(a.cli)

	tun.Handle(HandleSSHOpenV2, sshHandlerV2(a))
	tun.Handle(HandleSSHCloseV2, sshCloseHandlerV2(a))
	tun.Handle(HandleHTTPProxyV2, httpProxyHandlerV2(a))

	go a.ping(ctx, AgentPingDefaultInterval) //nolint:errcheck

	ctx, cancel := context.WithCancel(ctx)
	go func() {
		for {
			if a.isClosed() {
				a.logger.Info("Stopped listening for connections")

				cancel()

				return
			}

			ShellHubConnectV2Path := "/agent/connection"

			a.logger.Debug("Using tunnel version 2")

			listener, err := a.cli.NewReverseListenerV2(
				ctx,
				a.authData.Token,
				ShellHubConnectV2Path,
				client.NewReverseV2ConfigFromMap(a.authData.Config),
			)
			if err != nil {
				a.logger.Error("Failed to connect to server through reverse tunnel. Retry in 10 seconds")

				time.Sleep(time.Second * 10)

				continue
			}
			a.listener.Store(&listener)

			a.logger.Info("Server connection established")

			a.listening <- true

			if err := tun.Listen(ctx, listener); err != nil {
				a.logger.WithError(err).Error("Tunnel listener exited with error")
			}

			a.listening <- false
		}
	}()

	<-ctx.Done()

	return a.Close()
}

// AgentPingDefaultInterval is the default time interval between ping on agent.
const AgentPingDefaultInterval = 10 * time.Minute

func (a *Agent) ping(ctx context.Context, interval time.Duration) error {
	a.listening = make(chan bool)

	if interval == 0 {
		interval = AgentPingDefaultInterval
	}

	<-a.listening // NOTE: wait for the first connection to start to ping the server.
	ticker := time.NewTicker(interval)

	for {
		if a.isClosed() {
			return nil
		}

		select {
		case <-ctx.Done():
			log.WithFields(log.Fields{
				"version":        a.config.Version,
				"tenant_id":      a.authData.Namespace,
				"server_address": a.config.ServerAddress,
			}).Debug("stopped pinging server due to context cancellation")

			return nil
		case ok := <-a.listening:
			if ok {
				log.WithFields(log.Fields{
					"version":        a.config.Version,
					"tenant_id":      a.authData.Namespace,
					"server_address": a.config.ServerAddress,
					"timestamp":      clock.Now(),
				}).Debug("Starting the ping interval to server")

				ticker.Reset(interval)
			} else {
				log.WithFields(log.Fields{
					"version":        a.config.Version,
					"tenant_id":      a.authData.Namespace,
					"server_address": a.config.ServerAddress,
					"timestamp":      clock.Now(),
				}).Debug("Stopped pinging server due listener status")

				ticker.Stop()
			}
		case <-ticker.C:
			if err := a.authorize(); err != nil {
				a.server.SetDeviceName(a.authData.Name)
			}

			log.WithFields(log.Fields{
				"version":        a.config.Version,
				"tenant_id":      a.authData.Namespace,
				"server_address": a.config.ServerAddress,
				"name":           a.authData.Name,
				"hostname":       a.config.PreferredHostname,
				"identity":       a.config.PreferredIdentity,
				"timestamp":      clock.Now(),
			}).Info("Ping")

			randTimeout := time.Duration(rand.Intn(a.config.MaxRetryConnectionTimeout-10)+10) * time.Second //nolint:gosec
			ticker.Reset(interval + randTimeout)
		}
	}
}

// CheckUpdate gets the ShellHub's server version.
func (a *Agent) CheckUpdate() (*semver.Version, error) {
	info, err := a.cli.GetInfo(a.config.Version)
	if err != nil {
		return nil, err
	}

	return semver.NewVersion(info.Version)
}

// GetInfo gets the ShellHub's server information like version and endpoints, and updates the Agent's server's info.
func (a *Agent) GetInfo() (*models.Info, error) {
	if a.serverInfo != nil {
		return a.serverInfo, nil
	}

	info, err := a.cli.GetInfo(a.config.Version)
	if err != nil {
		return nil, err
	}

	a.serverInfo = info

	return info, nil
}

// GetInfo gets information like the version and the enpoints for HTTP and SSH to ShellHub server.
func GetInfo(cfg *Config) (*models.Info, error) {
	cli, err := client.NewClient(cfg.ServerAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create the HTTP client")
	}

	info, err := cli.GetInfo(cfg.Version)
	if err != nil {
		return nil, err
	}

	return info, nil
}
