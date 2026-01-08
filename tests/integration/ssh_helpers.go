package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"golang.org/x/crypto/ssh"
)

// SSHTestEnvironment encapsulates all resources needed for SSH tests.
type SSHTestEnvironment struct {
	Compose *environment.DockerCompose
	Agent   tc.Container
	Device  *models.Device
}

// NewSSHTestEnvironment creates a new SSH test environment with agent in compose network.
func NewSSHTestEnvironment(ctx context.Context, t *testing.T, compose *environment.DockerCompose, opts ...AgentOption) *SSHTestEnvironment {
	config := &agentConfig{
		username: environment.DefaultAgentUsername,
		password: environment.DefaultAgentPassword,
	}

	for _, opt := range opts {
		opt(config)
	}

	// Create agent in the same network as compose services
	agent, err := environment.NewAgentContainer(ctx, environment.AgentContainerOptions{
		ServerAddress: "http://gateway:80", // Use service discovery
		TenantID:      environment.DefaultNamespace,
		Identity:      config.identity,
		Username:      config.username,
		Password:      config.password,
		Networks:      []string{compose.Env("SHELLHUB_NETWORK")},
		NetworkAlias:  "test-agent",
	})
	require.NoError(t, err)

	// Register agent for automatic cleanup
	compose.RegisterAgent(agent)

	// Start agent
	err = agent.Start(ctx)
	require.NoError(t, err)

	// Wait for device to appear and accept it
	devices := []models.Device{}
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		resp, err := compose.R(ctx).SetResult(&devices).
			Get("/api/devices?status=pending")
		assert.Equal(tt, 200, resp.StatusCode())
		assert.NoError(tt, err)
		assert.Len(tt, devices, 1)
	}, environment.EventuallyTimeout, environment.EventuallyInterval)

	// Accept device
	resp, err := compose.R(ctx).
		Patch(fmt.Sprintf("/api/devices/%s/accept", devices[0].UID))
	require.Equal(t, 200, resp.StatusCode())
	require.NoError(t, err)

	// Wait for device to come online
	device := models.Device{}
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		resp, err := compose.R(ctx).
			SetResult(&device).
			Get(fmt.Sprintf("/api/devices/%s", devices[0].UID))
		assert.Equal(tt, 200, resp.StatusCode())
		assert.NoError(tt, err)
		assert.True(tt, device.Online)
	}, environment.EventuallyTimeout, environment.EventuallyInterval)

	return &SSHTestEnvironment{
		Compose: compose,
		Agent:   agent,
		Device:  &device,
	}
}

// NewSSHClient creates a new SSH client connection to the test device.
func (env *SSHTestEnvironment) NewSSHClient(t *testing.T, authMethod ssh.AuthMethod, username, namespace string) *ssh.Client {
	config := &ssh.ClientConfig{
		User:            fmt.Sprintf("%s@%s.%s", username, namespace, env.Device.Name),
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	}

	var conn *ssh.Client
	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		var err error
		conn, err = ssh.Dial("tcp", fmt.Sprintf("localhost:%s", env.Compose.Env("SHELLHUB_SSH_PORT")), config)
		assert.NoError(tt, err)
	}, environment.EventuallyTimeout, environment.EventuallyInterval)

	return conn
}

// NewPasswordAuth creates a password-based SSH auth method.
func NewPasswordAuth(password string) ssh.AuthMethod {
	return ssh.Password(password)
}

// NewPublicKeyAuth creates a public key-based SSH auth method.
func NewPublicKeyAuth(t *testing.T, privateKey interface{}) ssh.AuthMethod {
	signer, err := ssh.NewSignerFromKey(privateKey)
	require.NoError(t, err)

	return ssh.PublicKeys(signer)
}

type agentConfig struct {
	identity string
	username string
	password string
}

// AgentOption is a functional option for configuring agent creation.
type AgentOption func(*agentConfig)

// WithIdentity sets a custom identity for the agent.
func WithIdentity(identity string) AgentOption {
	return func(c *agentConfig) {
		c.identity = identity
	}
}

// WithCredentials sets custom credentials for the agent.
func WithCredentials(username, password string) AgentOption {
	return func(c *agentConfig) {
		c.username = username
		c.password = password
	}
}
