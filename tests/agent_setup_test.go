package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/moby/moby/api/types/network"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

func NewAgentContainerWithPrivateKey(path string) NewAgentContainerOption {
	return func(envs map[string]string) {
		envs["SHELLHUB_PRIVATE_KEY"] = path
	}
}

// NewAgentContainerWithoutIdentity drops the preferred identity, so the agent reads the MAC of its
// primary interface.
func NewAgentContainerWithoutIdentity() NewAgentContainerOption {
	return func(envs map[string]string) {
		delete(envs, "SHELLHUB_PREFERRED_IDENTITY")
	}
}

// TestAgentSetup covers what an agent does on its own before the server knows it: the key it
// generates and the identity it reads from its network interface.
func TestAgentSetup(t *testing.T) {
	t.Parallel()

	t.Run("the agent generates its private key on first run and keeps it", func(t *testing.T) {
		compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

		const keyPath = "/var/lib/shellhub/device.key"

		agent, device := startAcceptedAgent(t, t.Context(), compose, NewAgentContainerWithPrivateKey(keyPath))

		generated := readAgentFile(t, agent, keyPath)

		block, _ := pem.Decode([]byte(generated))
		require.NotNil(t, block, "the agent wrote no PEM block to %s", keyPath)
		assert.Equal(t, "RSA PRIVATE KEY", block.Type)

		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		require.NoError(t, err)
		assert.Equal(t, encodeDevicePublicKey(&key.PublicKey), device.PublicKey, "the device enrolled with a key other than the one on disk")

		assert.Equal(t, []string{"600", "700"}, strings.Fields(execInAgent(t, agent, "stat", "-c", "%a", keyPath, "/var/lib/shellhub")), "the key and the directory made for it are readable by others")

		require.NoError(t, agent.Stop(t.Context(), nil))
		compose.AwaitDeviceOffline(t, device.UID)

		require.NoError(t, agent.Start(t.Context()))
		compose.AwaitDeviceOnline(t, device.UID)

		assert.Equal(t, generated, readAgentFile(t, agent, keyPath), "a restart replaced the key")
		assert.Len(t, compose.ListDevices(t, models.DeviceStatusEmpty), 1, "the restarted agent enrolled another device")
	})

	t.Run("the agent derives the device uid from its mac address", func(t *testing.T) {
		compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

		const (
			firstMAC  = "02:5e:00:00:24:01"
			secondMAC = "02:5e:00:00:24:02"
		)

		key := newAgentPrivateKey(t)

		startWithMAC := func(mac string) testcontainers.Container {
			return startCustomAgent(t, t.Context(), compose, onStackNetworkWithKey(t, compose, mac, key), NewAgentContainerWithoutIdentity())
		}

		agent := startWithMAC(firstMAC)

		first := compose.AwaitDeviceWithStatus(t, models.DeviceStatusPending)
		require.NotNil(t, first.Identity)
		assert.Equal(t, firstMAC, first.Identity.MAC)

		compose.UpdateDeviceStatus(t, first.UID, environment.DeviceActionAccept)
		compose.AwaitDeviceOnline(t, first.UID)

		require.NoError(t, agent.Stop(t.Context(), nil))
		compose.AwaitDeviceOffline(t, first.UID)

		agent = startWithMAC(secondMAC)

		second := compose.AwaitDeviceWithStatus(t, models.DeviceStatusPending)
		require.NotNil(t, second.Identity)
		assert.Equal(t, secondMAC, second.Identity.MAC)
		assert.NotEqual(t, first.UID, second.UID, "two mac addresses behind one key enrolled as one device")

		require.NoError(t, agent.Stop(t.Context(), nil))

		startWithMAC(firstMAC)

		compose.AwaitDeviceOnline(t, first.UID)
		assert.Len(t, compose.ListDevices(t, models.DeviceStatusEmpty), 2, "the first mac address came back as another device")
	})
}

func execInAgent(t *testing.T, agent testcontainers.Container, cmd ...string) string {
	t.Helper()

	exit, output, err := agent.Exec(t.Context(), cmd, tcexec.Multiplexed())
	require.NoError(t, err)

	content, err := io.ReadAll(output)
	require.NoError(t, err)
	require.Zero(t, exit, string(content))

	return strings.TrimSpace(string(content))
}

func newAgentPrivateKey(t *testing.T) []byte {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
}

func encodeDevicePublicKey(key *rsa.PublicKey) string {
	return string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(key)}))
}

func stackNetwork(t *testing.T, compose *environment.DockerCompose) string {
	t.Helper()

	networks, err := compose.Service(environment.ServiceGateway).Networks(t.Context())
	require.NoError(t, err)
	require.Len(t, networks, 1, "the gateway is on more than the stack's network")

	return networks[0]
}

func onStackNetworkWithKey(t *testing.T, compose *environment.DockerCompose, mac string, key []byte) func(*testcontainers.ContainerRequest) {
	t.Helper()

	hardware, err := net.ParseMAC(mac)
	require.NoError(t, err)

	stack := stackNetwork(t, compose)

	gatewayIP, err := compose.Service(environment.ServiceGateway).ContainerIP(t.Context())
	require.NoError(t, err)

	return func(req *testcontainers.ContainerRequest) {
		req.HostConfigModifier = nil
		req.Networks = []string{stack}
		req.EndpointSettingsModifier = func(settings map[string]*network.EndpointSettings) {
			settings[stack].MacAddress = network.HardwareAddr(hardware)
		}
		req.Entrypoint = []string{"sh", "-c", `echo "$0 localhost" > /etc/hosts && exec ./agent`, gatewayIP}
		withAgentPrivateKey(key)(req)
	}
}

func withAgentPrivateKey(key []byte) func(*testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.Files = append(req.Files, testcontainers.ContainerFile{
			Reader:            bytes.NewReader(key),
			ContainerFilePath: req.Env["SHELLHUB_PRIVATE_KEY"],
			FileMode:          0o600,
		})
	}
}
