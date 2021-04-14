package main

import (
	"crypto/rsa"
	"net/url"
	"os"
	"runtime"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	"github.com/shellhub-io/shellhub/agent/pkg/keygen"
	"github.com/shellhub-io/shellhub/agent/pkg/sysinfo"
	"github.com/shellhub-io/shellhub/pkg/api/client"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/revdial"
)

type Agent struct {
	opts          *ConfigOptions
	pubKey        *rsa.PublicKey
	Identity      *models.DeviceIdentity
	Info          *models.DeviceInfo
	authData      *models.DeviceAuthResponse
	cli           client.Client
	serverInfo    *models.Info
	serverAddress *url.URL
	sessions      []string
}

func NewAgent(opts *ConfigOptions) (*Agent, error) {
	a := &Agent{}

	serverAddress, err := url.Parse(opts.ServerAddress)
	if err != nil {
		return nil, err
	}

	a.serverAddress = serverAddress

	return &Agent{
		opts: opts,
		cli:  client.NewClient(client.WithURL(serverAddress)),
	}, nil
}

// initialize initializes agent
func (a *Agent) initialize() error {
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

	if err := a.authorize(); err != nil {
		return errors.Wrap(err, "failed to authorize device")
	}

	return nil
}

func (a *Agent) generatePrivateKey() error {
	if _, err := os.Stat(a.opts.PrivateKey); os.IsNotExist(err) {
		err := keygen.GeneratePrivateKey(a.opts.PrivateKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Agent) readPublicKey() error {
	key, err := keygen.ReadPublicKey(a.opts.PrivateKey)
	a.pubKey = key
	return err
}

// generateDeviceIdentity generates device identity
func (a *Agent) generateDeviceIdentity() error {
	iface, err := sysinfo.PrimaryInterface()
	if err != nil {
		return err
	}

	a.Identity = &models.DeviceIdentity{
		MAC: iface.HardwareAddr.String(),
	}

	return nil
}

// loadDeviceInfo load some device information
func (a *Agent) loadDeviceInfo() error {
	osrelease, err := sysinfo.GetOSRelease()
	if err != nil {
		return nil
	}

	a.Info = &models.DeviceInfo{
		ID:         osrelease.ID,
		PrettyName: osrelease.Name,
		Version:    AgentVersion,
		Arch:       runtime.GOARCH,
		Platform:   AgentPlatform,
	}

	return nil
}

// checkUpdate check for agent updates
func (a *Agent) checkUpdate() (*semver.Version, error) {
	info, err := a.cli.GetInfo(AgentVersion)
	if err != nil {
		return nil, err
	}

	return semver.NewVersion(info.Version)
}

// probeServerInfo probe server information
func (a *Agent) probeServerInfo() error {
	info, err := a.cli.GetInfo(AgentVersion)
	a.serverInfo = info
	return err
}

// authorize send auth request to the server
func (a *Agent) authorize() error {
	authData, err := a.cli.AuthDevice(&models.DeviceAuthRequest{
		Info:     a.Info,
		Sessions: a.sessions,
		DeviceAuth: &models.DeviceAuth{
			Hostname:  a.opts.PreferredHostname,
			Identity:  a.Identity,
			TenantID:  a.opts.TenantID,
			PublicKey: string(keygen.EncodePublicKeyToPem(a.pubKey)),
		},
	})

	a.authData = authData

	return err
}

func (a *Agent) newReverseListener() (*revdial.Listener, error) {
	return a.cli.NewReverseListener(a.authData.Token)
}
