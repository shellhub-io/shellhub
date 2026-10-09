package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/devicekey"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"golang.org/x/crypto/ssh"
)

func TestDeviceKeyProof(t *testing.T) {
	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	unsigned := newDeviceAuthRequest(t, "proof", "02:00:00:00:02:00")
	unsigned.PublicKey = string(pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey)}))

	challenge := func(t *testing.T) string {
		t.Helper()

		res := new(models.DeviceAuthChallenge)

		resp, err := compose.Anonymous(t.Context()).SetResult(res).Post("/api/devices/auth/challenge")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		return res.Challenge
	}

	signed := func(t *testing.T, challenge string) requests.DeviceAuth {
		t.Helper()

		signature, err := devicekey.Sign(key, devicekey.Statement{Challenge: challenge, TenantID: ShellHubNamespace, PublicKey: unsigned.PublicKey})
		require.NoError(t, err)

		req := unsigned
		req.Challenge = challenge
		req.Signature = signature

		return req
	}

	status := func(t *testing.T, req requests.DeviceAuth) int {
		t.Helper()

		resp, err := compose.Anonymous(t.Context()).SetBody(req).Post("/api/devices/auth")
		require.NoError(t, err)

		return resp.StatusCode()
	}

	legacy := authDevice(t, compose, unsigned)

	issued := challenge(t)
	proven := authDevice(t, compose, signed(t, issued))
	require.Equal(t, legacy.UID, proven.UID, "the proof authenticates the device the legacy request enrolled")

	t.Run("a replayed proof is refused", func(t *testing.T) {
		assert.Equal(t, http.StatusForbidden, status(t, signed(t, issued)))
	})

	t.Run("the device's public metadata alone no longer earns its token", func(t *testing.T) {
		assert.Equal(t, http.StatusForbidden, status(t, unsigned))
	})

	t.Run("a token issued before the proof is refused", func(t *testing.T) {
		resp, err := asBearer(t, compose, legacy.Token).Get("/api/devices/auth/status")
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode(), resp.String())
	})

	t.Run("the token issued with the proof is honoured", func(t *testing.T) {
		assert.Equal(t, models.DeviceStatusPending, requireDeviceAuthStatus(t, compose, proven.Token))
	})
}

const releasedAgentVersion = "v0.26.0"

func buildReleasedAgentImage(ctx context.Context) (string, error) {
	return cachedAgentImage(ctx, "released-"+releasedAgentVersion, environment.AgentBuild{
		Repository: "shellhub-e2e/agent-released",
		Context:    "../agent",
		Dockerfile: "Dockerfile.released",
		Version:    releasedAgentVersion,
		Output:     agentBuildLog(),
	})
}

func startReleasedAgent(t *testing.T, ctx context.Context, compose *environment.DockerCompose, opts ...NewAgentContainerOption) testcontainers.Container {
	t.Helper()

	image, err := buildReleasedAgentImage(ctx)
	require.NoError(t, err)

	agent, err := newAgentContainer(ctx, compose.Service(environment.ServiceGateway).GetContainerID(), image, opts...)
	require.NoError(t, err)

	require.NoError(t, agent.Start(ctx))

	t.Cleanup(func() {
		_ = agent.Terminate(context.Background())
	})

	return agent
}

func reachDevice(ctx context.Context, addr string, device *models.Device) error {
	client, err := dialClientOnce(ctx, addr, &ssh.ClientConfig{
		User:            deviceSSHID(device),
		Auth:            []ssh.AuthMethod{ssh.Password(ShellHubAgentPassword)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // the gateway's own host key is not what these tests are about
		Timeout:         10 * time.Second,
	})
	if err != nil {
		return err
	}

	defer client.Close() //nolint:errcheck // the command already ran; closing reports nothing about the device

	_, err = deviceOutput(client, "true")

	return err
}

func requireSession(t *testing.T, ctx context.Context, compose *environment.DockerCompose, device *models.Device) {
	t.Helper()

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		assert.NoError(tt, reachDevice(ctx, compose.SSHAddress(), device))
	}, 60*time.Second, 2*time.Second, "no session reached %s", device.Name)
}

func TestDeviceKeyProofWithAReleasedAgent(t *testing.T) {
	t.Run("an agent released before the proof keeps connecting until the instance requires it", func(t *testing.T) {
		ctx := t.Context()
		compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

		startReleasedAgent(t, ctx, compose)
		pending := compose.AwaitDeviceWithStatus(t, models.DeviceStatusPending)
		compose.UpdateDeviceStatus(t, pending.UID, environment.DeviceActionAccept)
		released := compose.AwaitDeviceOnline(t, pending.UID)

		_, current := startAcceptedAgent(t, ctx, compose)

		requireSession(t, ctx, compose, &released)
		requireSession(t, ctx, compose, current)

		compose.RestartServer(t)
		requireSession(t, ctx, compose, &released)

		compose.RestartServerWith(t, "SHELLHUB_REQUIRE_DEVICE_KEY_PROOF", "true")

		requireSession(t, ctx, compose, current)
		assert.Never(t, func() bool {
			return reachDevice(ctx, compose.SSHAddress(), &released) == nil
		}, 30*time.Second, 2*time.Second, "the released agent still serves sessions once the proof is required")
	})

	t.Run("an instance that requires the proof never enrolls an agent released before it", func(t *testing.T) {
		ctx := t.Context()
		compose := newConfiguredSSHEnvironment(t, ctx, environment.New(t, run).WithEnv("SHELLHUB_REQUIRE_DEVICE_KEY_PROOF", "true"), models.SSHAccessModeLegacy)

		_, current := startAcceptedAgent(t, ctx, compose)
		requireSession(t, ctx, compose, current)

		released := startReleasedAgent(t, ctx, compose, NewAgentContainerWithHostname("released"))
		environment.AwaitLogContains(t, released, "status_code=403")

		assert.NotContains(t, deviceNames(compose.ListDevices(t, models.DeviceStatusEmpty)), "released")
	})
}
