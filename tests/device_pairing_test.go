package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

const codeLifetimeLeft = 5 * time.Second

var (
	agentPairingLink   = regexp.MustCompile(`accept-device\?code=([` + pairingcode.Alphabet + `]+)`)
	agentConnectedLine = regexp.MustCompile(`msg="Server connection established".* tenant_id=(\S+)`)
)

// TestDevicePairing drives the codes an agent prints for a person to accept it from a browser with
// requests built by hand, accepting them as the namespace's owner. A pairing code stands for a
// tenant-less agent whose device does not exist yet; a login code stands for a pending device that
// already enrolled with a tenant. The cases share one stack, so every device carries a hostname and
// a MAC address no other case uses. What a running agent makes of a pairing is covered by
// [TestDevicePairingAgent].
func TestDevicePairing(t *testing.T) {
	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	t.Run("pairing", func(t *testing.T) { testPairingCode(t, compose) })
}

// TestDevicePairingAgent runs a real agent with no tenant, for what it does with a pairing: the code
// it asks for, the status it polls, and the tenant it joins once someone accepts the code.
func TestDevicePairingAgent(t *testing.T) {
	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	t.Run("an agent without a tenant starts a pairing and waits on its code", func(t *testing.T) {
		agent := startAgent(t, t.Context(), compose, NewAgentContainerUnpaired(), NewAgentContainerWithHostname("pairing-waits"))
		code := awaitAgentPairingCode(t, agent)

		status, resp, err := getPairingStatus(t.Context(), compose, code)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Equal(t, &models.DevicePairingStatus{Status: models.DeviceStatusPending, Name: "pairing-waits"}, status)

		assert.NotContains(t, deviceNames(compose.ListDevices(t, models.DeviceStatusEmpty)), "pairing-waits", "a pairing created a device before anyone accepted it")

		assert.Never(t, func() bool {
			logs, err := agentLogs(t.Context(), agent)

			return err != nil || agentConnectedLine.Match(logs)
		}, 5*time.Second, time.Second, "the agent connected before anyone accepted its code")
	})

	t.Run("the agent learns of the acceptance by polling its code, without a restart", func(t *testing.T) {
		agent := startAgent(t, t.Context(), compose, NewAgentContainerUnpaired(), NewAgentContainerWithHostname("pairing-polled"))
		code := awaitAgentPairingCode(t, agent)

		accepted := acceptPairing(t, compose.R(t.Context()), code)

		status, resp, err := getPairingStatus(t.Context(), compose, code)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Equal(t, &models.DevicePairingStatus{
			Status:   models.DeviceStatusAccepted,
			TenantID: ShellHubNamespace,
			UID:      accepted.UID,
			Name:     "pairing-polled",
		}, status)

		awaitAgentConnected(t, agent)

		logs, err := agentLogs(t.Context(), agent)
		require.NoError(t, err)
		assert.Equal(t, 1, bytes.Count(logs, []byte(`msg="Starting ShellHub"`)), "the agent restarted to learn of the acceptance")
	})

	t.Run("the agent keeps the tenant it receives and serves the namespace with it", func(t *testing.T) {
		agent := startAgent(t, t.Context(), compose, NewAgentContainerUnpaired())

		accepted := acceptPairing(t, compose.R(t.Context()), awaitAgentPairingCode(t, agent))

		assert.Equal(t, ShellHubNamespace, awaitAgentConnected(t, agent))
		assert.Equal(t, ShellHubNamespace, readAgentFile(t, agent, "/tmp/shellhub.key.tenant"), "the agent did not keep the tenant it learned")
		assert.Equal(t, ShellHubNamespace, compose.AwaitDeviceOnline(t, accepted.UID).TenantID)
	})
}

func testPairingCode(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("pairing again with the same public key returns the pending code", func(t *testing.T) {
		req := newPairingRequest(t, "pairing-twice", "02:00:00:00:90:01")

		first := startPairing(t, compose, req)
		require.Equal(t, models.DeviceStatusPending, first.Status)
		require.NotEmpty(t, first.Code)

		again := startPairing(t, compose, req)
		assert.Equal(t, first.Code, again.Code)
		assert.Equal(t, models.DeviceStatusPending, again.Status)

		anotherKey := startPairing(t, compose, newPairingRequest(t, "pairing-twice", "02:00:00:00:90:01"))
		assert.NotEqual(t, first.Code, anotherKey.Code, "a different public key was handed the same code")
	})

	t.Run("a device already accepted is answered with its tenant instead of a code", func(t *testing.T) {
		req := newPairingRequest(t, "pairing-accepted", "02:00:00:00:90:02")

		acceptPairing(t, compose.R(t.Context()), startPairing(t, compose, req).Code)

		assert.Equal(t, &models.DevicePairing{Status: models.DeviceStatusAccepted, TenantID: ShellHubNamespace}, startPairing(t, compose, req))
	})

	t.Run("a code past its lifetime is not found", func(t *testing.T) {
		req := newPairingRequest(t, "pairing-expired", "02:00:00:00:90:03")
		pairing := startPairing(t, compose, req)

		compose.ExpireCacheEntryIn(t, pairingCodeCacheKey(pairing.Code), codeLifetimeLeft)

		status, resp, err := getPairingStatus(t.Context(), compose, pairing.Code)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), "the code was gone before its lifetime ran out: %s", resp.String())
		require.Equal(t, models.DeviceStatusPending, status.Status)

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			_, resp, err := getPairingStatus(t.Context(), compose, pairing.Code)
			if assert.NoError(tt, err) {
				assert.Equal(tt, http.StatusNotFound, resp.StatusCode(), resp.String())
			}
		}, codeLifetimeLeft+10*time.Second, time.Second)

		_, resp, err = postPairingAccept(compose.R(t.Context()), pairing.Code)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode(), resp.String())

		assert.NotContains(t, deviceNames(compose.ListDevices(t, models.DeviceStatusEmpty)), req.Hostname, "an expired code enrolled a device")
	})
}

func agentLogs(ctx context.Context, agent testcontainers.Container) ([]byte, error) {
	reader, err := agent.Logs(ctx)
	if err != nil {
		return nil, err
	}

	defer func() { _ = reader.Close() }()

	return io.ReadAll(reader)
}

func awaitAgentLog(t *testing.T, agent testcontainers.Container, line *regexp.Regexp, missing string) string {
	t.Helper()

	var captured string

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		logs, err := agentLogs(t.Context(), agent)
		if !assert.NoError(tt, err) {
			return
		}

		match := line.FindSubmatch(logs)
		if assert.NotNil(tt, match, missing) {
			captured = string(match[1])
		}
	}, 30*time.Second, time.Second)

	return captured
}

func awaitAgentPairingCode(t *testing.T, agent testcontainers.Container) string {
	t.Helper()

	return awaitAgentLog(t, agent, agentPairingLink, "the agent has not printed a pairing code")
}

func awaitAgentConnected(t *testing.T, agent testcontainers.Container) string {
	t.Helper()

	return awaitAgentLog(t, agent, agentConnectedLine, "the agent has not connected to the server")
}

func readAgentFile(t *testing.T, agent testcontainers.Container, path string) string {
	t.Helper()

	exit, output, err := agent.Exec(t.Context(), []string{"cat", path}, tcexec.Multiplexed())
	require.NoError(t, err)

	content, err := io.ReadAll(output)
	require.NoError(t, err)
	require.Zero(t, exit, string(content))

	return strings.TrimSpace(string(content))
}

func pairingCodeCacheKey(code string) string {
	return "pairing_code/" + code
}

func newPairingRequest(t *testing.T, hostname, mac string) requests.DevicePairingCreate {
	t.Helper()

	auth := newDeviceAuthRequest(t, hostname, mac)

	return requests.DevicePairingCreate{
		Info:      auth.Info,
		Hostname:  auth.Hostname,
		Identity:  auth.Identity,
		PublicKey: auth.PublicKey,
	}
}

func startPairing(t *testing.T, compose *environment.DockerCompose, req requests.DevicePairingCreate) *models.DevicePairing {
	t.Helper()

	pairing := new(models.DevicePairing)

	resp, err := compose.Anonymous(t.Context()).SetBody(req).SetResult(pairing).Post("/api/devices/pairing")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return pairing
}

func getPairingStatus(ctx context.Context, compose *environment.DockerCompose, code string) (*models.DevicePairingStatus, *resty.Response, error) {
	status := new(models.DevicePairingStatus)

	resp, err := compose.Anonymous(ctx).SetResult(status).Get("/api/devices/pairing/" + code + "/status")

	return status, resp, err
}

func postPairingAccept(req *resty.Request, code string) (*models.DevicePairingAccepted, *resty.Response, error) {
	accepted := new(models.DevicePairingAccepted)

	resp, err := req.
		SetBody(map[string]string{"tenant_id": ShellHubNamespace}).
		SetResult(accepted).
		Post("/api/devices/pairing/" + code + "/accept")

	return accepted, resp, err
}

func acceptPairing(t *testing.T, req *resty.Request, code string) *models.DevicePairingAccepted {
	t.Helper()

	accepted, resp, err := postPairingAccept(req, code)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	require.NotEmpty(t, accepted.UID)

	return accepted
}
