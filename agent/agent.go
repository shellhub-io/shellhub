package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/shellhub-io/shellhub/agent/pkg/keygen"
	"github.com/shellhub-io/shellhub/agent/pkg/sysinfo"
	"github.com/shellhub-io/shellhub/pkg/api/openapi"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
)

type Agent struct {
	opts          *ConfigOptions
	pubKey        *rsa.PublicKey
	Identity      *models.DeviceIdentity
	Info          *models.DeviceInfo
	authData      *models.DeviceAuthResponse
	cli           *openapi.APIClient
	serverInfo    *models.Info
	serverAddress *url.URL
	sessions      []string
}

func NewAgent(opts *ConfigOptions) (*Agent, error) {
	a := &Agent{}
	config := openapi.NewConfiguration()

	serverAddress, err := url.Parse(opts.ServerAddress)
	if err != nil {
		return nil, err
	}

	a.serverAddress = serverAddress

	return &Agent{
		opts: opts,
		cli:  openapi.NewAPIClient(config),
	}, nil
}

// initialize initializes agent.
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

// generateDeviceIdentity generates device identity.
func (a *Agent) generateDeviceIdentity() error {
	// priorize identity from env
	if id := a.opts.PreferredIdentity; id != "" {
		a.Identity = &models.DeviceIdentity{
			MAC: id,
		}

		return nil
	}

	// get identity from network interface
	iface, err := sysinfo.PrimaryInterface()
	if err != nil {
		return err
	}

	a.Identity = &models.DeviceIdentity{
		MAC: iface.HardwareAddr.String(),
	}

	return nil
}

// loadDeviceInfo load some device information.
func (a *Agent) loadDeviceInfo() error {
	osrelease, err := sysinfo.GetOSRelease()
	if err != nil {
		return err
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

// checkUpdate check for agent updates.
func (a *Agent) checkUpdate() (*semver.Version, error) {
	ctx := context.Background()

	req := a.cli.DefaultApi.GetInfo(ctx)
	info, _, err := req.Execute()
	if err != nil {
		return nil, err
	}

	return semver.NewVersion(*info.Version)
}

// probeServerInfo probe server information.
func (a *Agent) probeServerInfo() error {
	ctx := context.Background()

	req := a.cli.DefaultApi.GetInfo(ctx)
	info, _, err := req.Execute()
	if err != nil {
		return err
	}

	a.serverInfo = &models.Info{
		Version: *info.Version,
		Endpoints: models.Endpoints{
			API: *info.Endpoints.Api,
			SSH: *info.Endpoints.Ssh,
		},
	}

	return err
}

// authorize send auth request to the server.
func (a *Agent) authorize() error {
	ctx := context.Background()

	data := openapi.PostAuthDeviceRequest{
		Info: openapi.DeviceInfo{
			Id:         &a.Info.ID,
			PrettyName: &a.Info.PrettyName,
			Version:    &a.Info.Version,
			Arch:       &a.Info.Arch,
			Platform:   &a.Info.Platform,
		},
		Hostname: a.opts.PreferredHostname,
		Identity: &openapi.DeviceIdentity{
			Mac: &a.Identity.MAC,
		},
		PublicKey: string(keygen.EncodePublicKeyToPem(a.pubKey)),
		TenantId:  a.opts.TenantID,
	}

	req := a.cli.DevicesApi.PostAuthDevice(ctx)
	auth, _, err := req.PostAuthDeviceRequest(data).Execute()
	if err != nil {
		return err
	}

	a.authData = &models.DeviceAuthResponse{
		UID:       *auth.Uid,
		Token:     *auth.Token,
		Name:      *auth.Name,
		Namespace: *auth.Namespace,
	}

	return err
}

func (a *Agent) newReverseListener() (*revdial.Listener, error) {
	req, _ := http.NewRequest("GET", "", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.authData.Token))

	conn, _, err := websocket.DefaultDialer.Dial(strings.Join([]string{fmt.Sprintf("%s://%s", "ws", a.serverInfo.Endpoints.API), "/ssh/connection"}, ""), req.Header)
	if err != nil {
		return nil, err
	}

	lis := revdial.NewListener(wsconnadapter.New(conn),
		func(ctx context.Context, path string) (*websocket.Conn, *http.Response, error) {
			return websocket.DefaultDialer.DialContext(ctx, strings.Join([]string{fmt.Sprintf("%s://%s", "ws", a.serverInfo.Endpoints.API), path}, ""), nil)
		},
	)

	return lis, nil
}
