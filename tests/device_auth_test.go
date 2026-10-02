package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	otherNamespaceName = "otherspace"
	otherNamespace     = "00000000-0000-4000-0000-000000000001"
	lastSeenQuietPolls = 15
	tunnelPingTimeout  = 60 * time.Second
)

// TestDeviceAuth drives the endpoint an agent enrolls through with requests built by hand, so each
// case controls exactly which part of the device's identity changes. The provisioning-key paths
// are covered by [TestProvisioningKeyEnrollment].
func TestDeviceAuth(t *testing.T) {
	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)
	compose.NewNamespace(t, ShellHubUsername, otherNamespaceName, otherNamespace, "")

	t.Run("a tenant id earns a token signed for the device", func(t *testing.T) {
		res := authDevice(t, compose, newDeviceAuthRequest(t, "token", "02:00:00:00:01:00"))

		assert.NotEmpty(t, res.UID)
		assert.Equal(t, ShellHubNamespace, res.TenantID)
		assert.Equal(t, ShellHubNamespaceName, res.Namespace)
		assert.Equal(t, models.DeviceStatusPending, res.Status)

		claims, err := jwttoken.ClaimsFromBearerToken(compose.APIPublicKey(t), res.Token)
		require.NoError(t, err)
		assert.Equal(t, &authorizer.DeviceClaims{UID: res.UID, TenantID: ShellHubNamespace}, claims)
	})

	t.Run("the uid is derived from the identity", func(t *testing.T) {
		otherPublicKey := newDevicePublicKey(t)
		provisioningKey := compose.CreateProvisioningKey(t, &requests.CreateProvisioningKey{
			Name: "uid",
			Mode: string(models.ProvisioningKeyModeManual),
		})

		cases := []struct {
			description string
			change      func(req *requests.DeviceAuth)
			compare     assert.ComparisonAssertionFunc
		}{
			{
				description: "authenticating again keeps the uid",
				change:      func(*requests.DeviceAuth) {},
				compare:     assert.Equal,
			},
			{
				description: "a new hostname keeps the uid",
				change: func(req *requests.DeviceAuth) {
					req.Hostname = "renamed"
				},
				compare: assert.Equal,
			},
			{
				description: "presenting a provisioning key keeps the uid",
				change: func(req *requests.DeviceAuth) {
					req.ProvisioningKey = provisioningKey.Key
				},
				compare: assert.Equal,
			},
			{
				description: "a new mac address changes the uid",
				change: func(req *requests.DeviceAuth) {
					req.Identity = &requests.DeviceIdentity{MAC: "02:00:00:00:03:ff"}
				},
				compare: assert.NotEqual,
			},
			{
				description: "a new public key changes the uid",
				change: func(req *requests.DeviceAuth) {
					req.PublicKey = otherPublicKey
				},
				compare: assert.NotEqual,
			},
			{
				description: "another namespace changes the uid",
				change: func(req *requests.DeviceAuth) {
					req.TenantID = otherNamespace
				},
				compare: assert.NotEqual,
			},
		}

		for i, tc := range cases {
			t.Run(tc.description, func(t *testing.T) {
				req := newDeviceAuthRequest(t, fmt.Sprintf("uid-%d", i), fmt.Sprintf("02:00:00:00:03:%02x", i))

				first := authDevice(t, compose, req)

				tc.change(&req)
				second := authDevice(t, compose, req)

				tc.compare(t, first.UID, second.UID)
			})
		}
	})
}

// TestDeviceAuthAgent runs a real agent, for what the server does with an agent beyond a single
// request: the hostname the agent reports, the tunnel it keeps open, and the enrollment it repeats
// after the device is removed.
func TestDeviceAuthAgent(t *testing.T) {
	t.Run("an agent's hostname has its dots replaced by underscores", func(t *testing.T) {
		compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

		startAgent(t, t.Context(), compose, NewAgentContainerWithHostname("e2e.agent.example"))

		device := compose.AwaitDeviceWithStatus(t, models.DeviceStatusPending)
		assert.Equal(t, "e2e_agent_example", device.Name)
	})

	t.Run("the tunnel ping advances last_seen once the agent has settled", func(t *testing.T) {
		compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

		_, device := startAcceptedAgent(t, t.Context(), compose)

		settled := awaitSettledLastSeen(t, compose, device.UID)

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			current, resp, err := compose.GetDevice(t.Context(), device.UID)
			if !assert.NoError(tt, err) {
				return
			}

			assert.Equal(tt, http.StatusOK, resp.StatusCode(), resp.String())
			assert.True(tt, current.LastSeen.After(settled), "last_seen %s has not moved past %s", current.LastSeen, settled)
		}, tunnelPingTimeout, 2*time.Second)
	})

	t.Run("a removed device re-registers as pending under the same uid", func(t *testing.T) {
		compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

		agent, device := startAcceptedAgent(t, t.Context(), compose)
		require.NoError(t, agent.Stop(t.Context(), nil))

		compose.DeleteDevice(t, device.UID)

		removed := compose.AwaitDeviceWithStatus(t, models.DeviceStatusRemoved)
		require.Equal(t, device.UID, removed.UID)

		require.NoError(t, agent.Start(t.Context()))

		pending := compose.AwaitDeviceWithStatus(t, models.DeviceStatusPending)
		assert.Equal(t, device.UID, pending.UID)
	})

	t.Run("a removed device's running agent re-registers as pending without a restart", func(t *testing.T) {
		compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

		_, device := startAcceptedAgent(t, t.Context(), compose)

		compose.DeleteDevice(t, device.UID)

		pending := compose.AwaitDeviceWithStatus(t, models.DeviceStatusPending)
		assert.Equal(t, device.UID, pending.UID)
	})
}

func awaitSettledLastSeen(t *testing.T, compose *environment.DockerCompose, uid string) time.Time {
	t.Helper()

	var settled time.Time

	unchanged := 0

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		current, resp, err := compose.GetDevice(t.Context(), uid)
		if !assert.NoError(tt, err) {
			return
		}

		if !assert.Equal(tt, http.StatusOK, resp.StatusCode(), resp.String()) {
			return
		}

		if current.LastSeen.Equal(settled) {
			unchanged++
		} else {
			settled = current.LastSeen
			unchanged = 0
		}

		assert.GreaterOrEqual(tt, unchanged, lastSeenQuietPolls, "last_seen still moving: %s", settled)
	}, 60*time.Second, 1*time.Second)

	return settled
}

func newDeviceAuthRequest(t *testing.T, hostname, mac string) requests.DeviceAuth {
	t.Helper()

	return requests.DeviceAuth{
		Info: &requests.DeviceInfo{
			ID:         "e2e",
			PrettyName: "e2e",
			Version:    "e2e",
			Arch:       "amd64",
			Platform:   "native",
		},
		Hostname:  hostname,
		Identity:  &requests.DeviceIdentity{MAC: mac},
		PublicKey: newDevicePublicKey(t),
		TenantID:  ShellHubNamespace,
	}
}

func newDevicePublicKey(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
	}))
}

func authDevice(t *testing.T, compose *environment.DockerCompose, req requests.DeviceAuth) *models.DeviceAuthResponse {
	t.Helper()

	res := new(models.DeviceAuthResponse)

	resp, err := compose.Anonymous(t.Context()).SetBody(req).SetResult(res).Post("/api/devices/auth")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return res
}
