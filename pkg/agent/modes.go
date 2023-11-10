package agent

import (
	"context"

	dockerclient "github.com/docker/docker/client"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/sysinfo"
	"github.com/shellhub-io/shellhub/pkg/agent/server"
	"github.com/shellhub-io/shellhub/pkg/agent/server/modes"
)

type Info struct {
	ID   string
	Name string
}

// Mode is the Agent execution mode.
//
// The agent can be executed in two different modes: host and connector.
// The host mode is the default mode, where the agent will listen for incoming connections and will be responsible for
// the SSH server. The connector mode is used to turn all containers inside a host into a single device and be
// responsible for the SSH server of all containers.
//
// Check [ModeHost] and [ModeConnector] for more information.
type Mode interface {
	// Configure prepares the agent for listening, setting up the SSH server, its modes and values on Agent's.
	Configure(agent *Agent)
	// GetInfo gets information about agent according to Agent's mode.
	//
	// When Agent is running on [ModeHost], the info got is from the system where the Agent is running, but when running
	// in [ConnectorMode], the data is retrieved from Docker Engine.
	GetInfo() (*Info, error)
}

// ModeHost is the Agent execution mode for `host`.
//
// The host mode is the default mode one, and turns the host machine into a ShellHub's Agent. The host is
// responsible for the SSH server, authentication and authorization, `/etc/passwd`, `/etc/shadow`, and etc.
type HostMode struct{}

var _ Mode = new(HostMode)

func (n *HostMode) Configure(a *Agent) {
	a.server = server.NewServer(
		a.cli,
		a.authData,
		a.config.PrivateKey,
		a.config.KeepAliveInterval,
		a.config.SingleUserPassword,
		modes.HostMode,
	)

	a.server.SetDeviceName(a.authData.Name)
}

func (n *HostMode) GetInfo() (*Info, error) {
	osrelease, err := sysinfo.GetOSRelease()
	if err != nil {
		return nil, err
	}

	return &Info{
		ID:   osrelease.ID,
		Name: osrelease.Name,
	}, nil
}

// ModeConnector is the Agent execution mode for `connector`.
//
// The connector mode is used to turn a container inside a host into a single device ShellHub's Agent. The host is
// responsible for the SSH server, but the authentication and authorization is made by either the conainer
// internals, `passwd` or `shadow`, or by the ShellHub API.
type ConnectorMode struct {
	identity string
}

func NewConnectorMode(identity string) Mode {
	return &ConnectorMode{
		identity: identity,
	}
}

var _ Mode = new(ConnectorMode)

func (c *ConnectorMode) Configure(a *Agent) {
	// NOTICE: When the agent is running in Connector Mode, we need to identify the container ID to maintain the
	// communication between the server and the agent when the container name on the host changes.  This information is
	// saved inside the device's identity, avoiding significant changes in the current state of the agent.
	// TODO: Evaluate if we can use another field than "MAC" to store the container ID.
	a.server = server.NewServer(
		a.cli,
		a.authData,
		a.config.PrivateKey,
		a.config.KeepAliveInterval,
		a.config.SingleUserPassword,
		modes.ConnectorMode,
	)
	a.server.SetContainerID(a.Identity.MAC)
	a.server.SetDeviceName(a.authData.Name)
}

func (c *ConnectorMode) CreateServer(agent *Agent) *server.Server {
	agent.server = server.NewServer(
		agent.cli,
		agent.authData,
		agent.config.PrivateKey,
		agent.config.KeepAliveInterval,
		agent.config.SingleUserPassword,
		modes.ConnectorMode,
	)

	return agent.server
}

func (c *ConnectorMode) GetInfo() (*Info, error) {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	defer cli.Close()

	info, err := cli.ContainerInspect(context.Background(), c.identity)
	if err != nil {
		return nil, err
	}

	return &Info{
		ID:   "docker",
		Name: info.Config.Image,
	}, nil
}
