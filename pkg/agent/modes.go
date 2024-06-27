package agent

import (
	"context"
	"os/exec"

	dockerclient "github.com/docker/docker/client"
	"github.com/shellhub-io/shellhub/pkg/agent/pkg/sysinfo"
	"github.com/shellhub-io/shellhub/pkg/agent/server"
	"github.com/shellhub-io/shellhub/pkg/agent/server/modes/connector"
	"github.com/shellhub-io/shellhub/pkg/agent/server/modes/host"
)

type Info struct {
	ID   string
	Name string
}

// Mode is the Agent execution mode.
//
// Check [HostMode] and [ConnectorMode] for more information.
type Mode interface {
	// ConfigureSSHServer prepares the Agent for listening, setting up the SSH server, its modes and values.
	ConfigureSSHServer(agent *Agent)
	// GetInfo gets information about Agent according to Agent's mode.
	GetInfo() (*Info, error)
}

// ModeHost is the Agent execution mode for `Host`.
//
// The host mode is the default mode one, and turns the host machine into a ShellHub's Agent. The host is
// responsible for the SSH server, authentication and authorization, `/etc/passwd`, `/etc/shadow`, and etc.
type HostMode struct{}

var _ Mode = new(HostMode)

func (m *HostMode) ConfigureSSHServer(agent *Agent) {
	agent.sshd = server.NewServer(
		agent.httpc,
		&host.Mode{
			Authenticator: *host.NewAuthenticator(
				agent.httpc,
				agent.data.Auth,
				agent.config.SingleUserPassword,
				&agent.data.Auth.Name,
			),
			Sessioner: *host.NewSessioner(&agent.data.Auth.Name, make(map[string]*exec.Cmd)),
		},
		&server.Config{
			PrivateKey:        agent.config.PrivateKey,
			KeepAliveInterval: agent.config.KeepAliveInterval,
		},
	)

	agent.sshd.SetDeviceName(agent.data.Auth.Name)
}

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

// ModeConnector is the Agent execution mode for `Connector`.
//
// The `Connector` mode is used to turn a container inside a host into a single device ShellHub's Agent. The host is
// responsible for the SSH server, but the authentication and authorization is made by either the conainer
// internals, `passwd` or `shadow`, or by the ShellHub API.
type ConnectorMode struct {
	cli      *dockerclient.Client
	identity string
}

func NewConnectorMode(cli *dockerclient.Client, identity string) (Mode, error) {
	return &ConnectorMode{
		cli:      cli,
		identity: identity,
	}, nil
}

var _ Mode = new(ConnectorMode)

func (m *ConnectorMode) ConfigureSSHServer(agent *Agent) {
	// NOTICE: When the agent is running in `Connector` mode, we need to identify the container ID to maintain the
	// communication between the server and the agent when the container name on the host changes.  This information is
	// saved inside the device's identity, avoiding significant changes in the current state of the agent.
	// TODO: Evaluate if we can use another field than "MAC" to store the container ID.
	agent.sshd = server.NewServer(
		agent.httpc,
		&connector.Mode{
			Authenticator: *connector.NewAuthenticator(
				agent.httpc,
				m.cli,
				agent.data.Auth,
				&agent.data.Identity.MAC,
			),
			Sessioner: *connector.NewSessioner(&agent.data.Identity.MAC, m.cli),
		},
		&server.Config{
			PrivateKey:        agent.config.PrivateKey,
			KeepAliveInterval: agent.config.KeepAliveInterval,
		},
	)

	agent.sshd.SetContainerID(agent.data.Identity.MAC)
	agent.sshd.SetDeviceName(agent.data.Auth.Name)
}

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
