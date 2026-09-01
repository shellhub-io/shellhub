package agentd

import (
	"context"

	dockerclient "github.com/docker/docker/client"
	"github.com/shellhub-io/shellhub/agent/pkg/sysinfo"
	"github.com/shellhub-io/shellhub/agent/server"
	"github.com/shellhub-io/shellhub/agent/server/modes/connector"
	"github.com/shellhub-io/shellhub/agent/server/modes/host"
)

// Info identifies the operating system, or the container image, the agent runs on.
// It is reported to the server so the device can be shown with the right platform.
type Info struct {
	ID   string
	Name string
}

// Mode is the Agent execution mode.
//
// Check [HostMode] and [ConnectorMode] for more information.
type Mode interface {
	// Serve prepares the Agent for listening, setting up the SSH server, its modes and values on Agent's.
	Serve(agent *Agent)
	// GetInfo gets information about Agent according to Agent's mode.
	//
	// When Agent is running on [HostMode], the info got is from the system where the Agent is running, but when running
	// in [ConnectorMode], the data is retrieved from Docker Engine.
	GetInfo() (*Info, error)
}

// HostMode is the Agent execution mode for `Host`.
//
// The host mode is the default mode one, and turns the host machine into a ShellHub's Agent. The host is
// responsible for the SSH server, authentication and authorization, `/etc/passwd`, `/etc/shadow`, and etc.
type HostMode struct{}

var _ Mode = new(HostMode)

// Serve attaches an SSH server that authenticates against the host's own user database.
func (m *HostMode) Serve(agent *Agent) {
	agent.server = server.NewServer(
		agent.cli,
		&host.Mode{
			Authenticator: *host.NewAuthenticator(agent.cli, agent.authData, agent.config.SingleUserPassword, &agent.authData.Name),
			Sessioner:     *host.NewSessioner(&agent.authData.Name, agent.config.SFTPServerCommand),
		},
		&server.Config{
			PrivateKey:        agent.config.PrivateKey,
			KeepAliveInterval: agent.config.KeepAliveInterval,
			Features:          server.LocalPortForwardFeature,
		},
	)

	agent.server.SetDeviceName(agent.authData.Name)
}

// GetInfo reports the host's distribution, read from its os-release file.
func (m *HostMode) GetInfo() (*Info, error) {
	osrelease, err := sysinfo.GetOSRelease()
	if err != nil {
		return nil, err
	}

	return &Info{
		ID:   osrelease.ID,
		Name: osrelease.Name,
	}, nil
}

// ConnectorMode is the Agent execution mode for `Connector`.
//
// The `Connector` mode is used to turn a container inside a host into a single device ShellHub's Agent. The host is
// responsible for the SSH server, but the authentication and authorization is made by either the conainer
// internals, `passwd` or `shadow`, or by the ShellHub API.
type ConnectorMode struct {
	cli      *dockerclient.Client
	identity string
}

// NewConnectorMode returns a [Mode] that exposes the Docker container named by identity as a
// device, using cli to inspect and exec into it.
func NewConnectorMode(cli *dockerclient.Client, identity string) (Mode, error) {
	return &ConnectorMode{
		cli:      cli,
		identity: identity,
	}, nil
}

var _ Mode = new(ConnectorMode)

// Serve attaches an SSH server whose sessions exec into the container rather than the host.
func (m *ConnectorMode) Serve(agent *Agent) {
	agent.server = server.NewServer(
		agent.cli,
		&connector.Mode{
			Authenticator: *connector.NewAuthenticator(agent.cli, m.cli, agent.authData, &agent.Identity.MAC),
			Sessioner:     *connector.NewSessioner(&agent.Identity.MAC, m.cli),
		},
		&server.Config{
			PrivateKey:        agent.config.PrivateKey,
			KeepAliveInterval: agent.config.KeepAliveInterval,
			Features:          server.NoFeature,
		},
	)

	agent.server.SetContainerID(agent.Identity.MAC)
	agent.server.SetDeviceName(agent.authData.Name)
}

// GetInfo reports the container's image, which stands in for the device's platform.
func (m *ConnectorMode) GetInfo() (*Info, error) {
	info, err := m.cli.ContainerInspect(context.Background(), m.identity)
	if err != nil {
		return nil, err
	}

	return &Info{
		ID:   "docker",
		Name: info.Config.Image,
	}, nil
}
