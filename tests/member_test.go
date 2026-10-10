package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
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
		awaitAgentConnected(t, agent)
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

func pairAgent(t *testing.T, compose *environment.DockerCompose, agent testcontainers.Container, token string) string {
	t.Helper()

	return acceptPairing(t, asBearer(t, compose, token), awaitAgentPairingCode(t, agent)).UID
}
