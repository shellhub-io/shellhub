package main

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"golang.org/x/crypto/ssh"
)

func TestDeviceKeyVerification(t *testing.T) {
	t.Parallel()

	compose := newSSHEnvironment(t, t.Context(), legacyNamespace)
	requireVerifiedDeviceKeys(t, compose, true)

	unsigned := newDeviceAuthRequest(t, "verified", "02:00:00:00:02:00")
	stored, ok := devicePrivateKeys.Load(unsigned.PublicKey)
	require.True(t, ok)

	key, ok := stored.(*rsa.PrivateKey)
	require.True(t, ok)

	challenge := func(t *testing.T) string {
		t.Helper()

		issued, err := requestDeviceAuthChallenge(t.Context(), compose)
		require.NoError(t, err)

		return issued
	}

	signed := func(t *testing.T, challenge string) requests.DeviceAuth {
		t.Helper()

		req, err := signDeviceAuthWith(key, unsigned, challenge)
		require.NoError(t, err)

		return req
	}

	status := func(t *testing.T, req requests.DeviceAuth) int {
		t.Helper()

		resp, err := compose.Anonymous(t.Context()).SetBody(req).Post("/api/devices/auth")
		require.NoError(t, err)

		return resp.StatusCode()
	}

	t.Run("a namespace that requires verified device keys refuses the device's public metadata alone", func(t *testing.T) {
		assert.Equal(t, http.StatusForbidden, status(t, unsigned))
	})

	issued := challenge(t)

	t.Run("a signature by the device key enrolls the device", func(t *testing.T) {
		verified := authDevice(t, compose, signed(t, issued))
		assert.Equal(t, models.DeviceStatusPending, requireDeviceAuthStatus(t, compose, verified.Token))
	})

	t.Run("the same signed request is honoured again within the window", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, status(t, signed(t, issued)))
	})

	t.Run("a challenge with a tampered tag is refused", func(t *testing.T) {
		raw, err := base64.RawURLEncoding.DecodeString(challenge(t))
		require.NoError(t, err)
		raw[len(raw)-1] ^= 0x01

		assert.Equal(t, http.StatusForbidden, status(t, signed(t, base64.RawURLEncoding.EncodeToString(raw))))
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

	return runAgent(t, ctx, agent)
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

func requireVerifiedDeviceKeys(t *testing.T, compose *environment.DockerCompose, required bool) {
	t.Helper()

	resp, err := compose.R(t.Context()).
		SetBody(map[string]any{"settings": map[string]bool{"require_verified_device_key": required}}).
		Put("/api/namespaces/" + ShellHubNamespace)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
}

func TestDeviceKeyVerificationWithAReleasedAgent(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	compose := newSSHEnvironment(t, ctx, legacyNamespace)
	requireVerifiedDeviceKeys(t, compose, true)

	_, current := startAcceptedAgent(t, ctx, compose)
	requireSession(t, ctx, compose, current)

	released := startReleasedAgent(t, ctx, compose, NewAgentContainerWithHostname("released"))
	environment.AwaitLogContains(t, released, "status_code=403")

	t.Run("a namespace that requires verified device keys never enrolls an agent that cannot sign", func(t *testing.T) {
		assert.NotContains(t, deviceNames(compose.ListDevices(t, models.DeviceStatusEmpty)), "released")
	})

	t.Run("the refused agent stays a single line in the tenant ID history while it retries", func(t *testing.T) {
		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			var events []models.ProvisioningKeyEvent
			resp, err := compose.R(ctx).SetResult(&events).Get("/api/devices/tenant-registrations")
			require.NoError(tt, err)
			require.Equal(tt, http.StatusOK, resp.StatusCode(), resp.String())

			var refused []models.ProvisioningKeyEvent
			for _, event := range events {
				if event.Hostname == "released" {
					refused = append(refused, event)
				}
			}

			require.Len(tt, refused, 1)
			assert.Equal(tt, models.ProvisioningKeyEventRefusalUnverifiedDeviceKey, refused[0].Refusal)
			assert.GreaterOrEqual(tt, refused[0].Attempts, 2)
			assert.Empty(tt, refused[0].DeviceStatus)
		}, 3*time.Minute, 2*time.Second)
	})

	t.Run("the refused agent lands pending once the namespace stops requiring verified device keys", func(t *testing.T) {
		requireVerifiedDeviceKeys(t, compose, false)

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			var devices []models.Device
			resp, err := compose.R(ctx).SetQueryParam("status", string(models.DeviceStatusPending)).SetResult(&devices).Get("/api/devices")
			require.NoError(tt, err)
			require.Equal(tt, http.StatusOK, resp.StatusCode(), resp.String())
			assert.Contains(tt, deviceNames(devices), "released")
		}, 3*time.Minute, 2*time.Second)
	})
}
