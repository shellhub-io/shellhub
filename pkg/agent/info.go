package agent

import (
	"context"

	dockerclient "github.com/docker/docker/client"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/sysinfo"
)

type Info struct {
	ID   string
	Name string
}

// InfoMode is the Agent execution mode.
//
// Check [HostInfoMode] and [ConnectorInfoMode] for more information.
type InfoMode interface {
	// GetInfo gets information about Agent according to Agent's mode.
	//
	// When Agent is running on [HostMode], the info got is from the system where the Agent is running, but when running
	// in [ConnectorMode], the data is retrieved from Docker Engine.
	GetInfo() (*Info, error)
}

// ModeHost is the Agent execution mode for `Host`.
//
// The host mode is the default mode one, and turns the host machine into a ShellHub's Agent. The host is
// responsible for the SSH server, authentication and authorization, `/etc/passwd`, `/etc/shadow`, and etc.
type HostInfoMode struct{}

var _ InfoMode = new(HostInfoMode)

func (m *HostInfoMode) GetInfo() (*Info, error) {
	osrelease, err := sysinfo.GetOSRelease()
	if err != nil {
		return nil, err
	}

	return &Info{
		ID:   osrelease.ID,
		Name: osrelease.Name,
	}, nil
}

// ConnectorInfoMode is the Agent execution mode for `Connector`.
//
// The `Connector` mode is used to turn a container inside a host into a single device ShellHub's Agent. The host is
// responsible for the SSH server, but the authentication and authorization is made by either the conainer
// internals, `passwd` or `shadow`, or by the ShellHub API.
type ConnectorInfoMode struct {
	cli      *dockerclient.Client
	identity string
}

func NewConnectorMode(cli *dockerclient.Client, identity string) (InfoMode, error) {
	return &ConnectorInfoMode{
		cli:      cli,
		identity: identity,
	}, nil
}

var _ InfoMode = new(ConnectorInfoMode)

func (m *ConnectorInfoMode) GetInfo() (*Info, error) {
	info, err := m.cli.ContainerInspect(context.Background(), m.identity)
	if err != nil {
		return nil, err
	}

	return &Info{
		ID:   "docker",
		Name: info.Config.Image,
	}, nil
}
