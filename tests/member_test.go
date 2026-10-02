package main

import (
	"io"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

const (
	departingUsername = "departing"
	departingEmail    = "departing@ossystems.com.br"
	departingPassword = "password"
)

var agentPairingLink = regexp.MustCompile(`accept-device\?code=([` + pairingcode.Alphabet + `]+)`)

// TestRemoveMemberThroughTheAdminCLI removes a member with the admin CLI, a process apart from the
// server that caches the member's API keys and holds the tunnels of the devices they paired.
func TestRemoveMemberThroughTheAdminCLI(t *testing.T) {
	t.Run("refuses their token and API keys at once", func(t *testing.T) {
		compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)
		member := signInDepartingMember(t, compose)

		key := new(responses.CreateAPIKey)
		resp, err := asBearer(t, compose, member.Token).
			SetBody(&requests.CreateAPIKey{Name: "departing", ExpiresAt: -1}).
			SetResult(key).
			Post("/api/namespaces/api-key")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		withToken := func() *resty.Request { return asBearer(t, compose, member.Token) }
		withKey := func() *resty.Request { return compose.Anonymous(t.Context()).SetHeader("X-API-Key", key.Key) }

		for _, request := range []func() *resty.Request{withToken, withKey} {
			resp, err := request().Get("/api/devices")
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		}

		compose.RemoveMember(t, departingUsername, ShellHubNamespaceName)

		for _, request := range []func() *resty.Request{withToken, withKey} {
			resp, err := request().Get("/api/devices")
			require.NoError(t, err)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode(), resp.String())
		}
	})

	t.Run("closes their paired devices' tunnels within a heartbeat", func(t *testing.T) {
		compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)
		member := signInDepartingMember(t, compose)

		agent := startAgent(t, t.Context(), compose, NewAgentContainerUnpaired())
		uid := pairAgent(t, compose, agent, member.Token)
		compose.AwaitDeviceOnline(t, uid)

		compose.RemoveMember(t, departingUsername, ShellHubNamespaceName)

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			current, resp, err := compose.GetDevice(t.Context(), uid)
			if !assert.NoError(tt, err) {
				return
			}

			if !assert.Equal(tt, http.StatusOK, resp.StatusCode(), resp.String()) {
				return
			}

			assert.Equal(tt, models.DeviceStatusRemoved, current.Status)
			assert.False(tt, current.Online, "the removed device still holds its tunnel")
		}, tunnelPingTimeout, time.Second)
	})
}

func signInDepartingMember(t *testing.T, compose *environment.DockerCompose) *models.UserAuthResponse {
	t.Helper()

	compose.NewUser(t, departingUsername, departingEmail, departingPassword)
	compose.NewMember(t, departingUsername, ShellHubNamespaceName, string(authorizer.RoleAdministrator))

	login := compose.AuthUser(t, departingUsername, departingPassword)

	signedInto := new(models.UserAuthResponse)
	resp, err := asBearer(t, compose, login.Token).
		SetResult(signedInto).
		Get("/api/auth/token/" + ShellHubNamespace)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return signedInto
}

func asBearer(t *testing.T, compose *environment.DockerCompose, token string) *resty.Request {
	t.Helper()

	return compose.Anonymous(t.Context()).SetAuthToken(token)
}

// pairAgent reads the pairing code an unpaired agent prints and accepts it into the test namespace
// as the bearer of token, who becomes the device's owner. It returns the device's uid.
func pairAgent(t *testing.T, compose *environment.DockerCompose, agent testcontainers.Container, token string) string {
	t.Helper()

	var code string

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		reader, err := agent.Logs(t.Context())
		if !assert.NoError(tt, err) {
			return
		}

		defer func() { _ = reader.Close() }()

		logs, err := io.ReadAll(reader)
		if !assert.NoError(tt, err) {
			return
		}

		match := agentPairingLink.FindSubmatch(logs)
		if assert.NotNil(tt, match, "the agent has not printed a pairing code") {
			code = string(match[1])
		}
	}, 30*time.Second, time.Second)

	accepted := new(models.DevicePairingAccepted)
	resp, err := asBearer(t, compose, token).
		SetBody(map[string]string{"tenant_id": ShellHubNamespace}).
		SetResult(accepted).
		Post("/api/devices/pairing/" + code + "/accept")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	require.NotEmpty(t, accepted.UID)

	return accepted.UID
}
