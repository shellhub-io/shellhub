package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
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

const heartbeatTimeout = 60 * time.Second

func NewAgentContainerWithHostname(hostname string) NewAgentContainerOption {
	return func(envs map[string]string) {
		envs["SHELLHUB_PREFERRED_HOSTNAME"] = hostname
	}
}

// TestDeviceAuth drives the endpoint an agent enrolls through with requests built by hand, so each
// case controls exactly which part of the device's identity changes. The provisioning-key paths
// are covered by [TestProvisioningKeyEnrollment].
func TestDeviceAuth(t *testing.T) {
	ctx := context.Background()
	compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

	t.Run("a tenant id earns a token signed for the device", func(t *testing.T) {
		req := newDeviceAuthRequest(t, "token", "02:00:00:00:01:00")

		res := authDevice(t, compose, req)

		assert.NotEmpty(t, res.UID)
		assert.Equal(t, ShellHubNamespace, res.TenantID)
		assert.Equal(t, ShellHubNamespaceName, res.Namespace)
		assert.Equal(t, models.DeviceStatusPending, res.Status)

		claims, err := jwttoken.ClaimsFromBearerToken(apiPublicKey(t), res.Token)
		require.NoError(t, err)
		assert.Equal(t, &authorizer.DeviceClaims{UID: res.UID, TenantID: ShellHubNamespace}, claims)
	})

	t.Run("the uid is derived from the identity", func(t *testing.T) {
		cases := []struct {
			description string
			change      func(t *testing.T, req *requests.DeviceAuth)
			sameUID     bool
		}{
			{
				description: "authenticating again keeps the uid",
				change:      func(*testing.T, *requests.DeviceAuth) {},
				sameUID:     true,
			},
			{
				description: "a new hostname keeps the uid",
				change: func(_ *testing.T, req *requests.DeviceAuth) {
					req.Hostname = "renamed"
				},
				sameUID: true,
			},
			{
				description: "a new mac address changes the uid",
				change: func(_ *testing.T, req *requests.DeviceAuth) {
					req.Identity = &requests.DeviceIdentity{MAC: "02:00:00:00:03:ff"}
				},
				sameUID: false,
			},
			{
				description: "a new public key changes the uid",
				change: func(t *testing.T, req *requests.DeviceAuth) {
					t.Helper()

					req.PublicKey = newDevicePublicKey(t)
				},
				sameUID: false,
			},
		}

		for i, tc := range cases {
			t.Run(tc.description, func(t *testing.T) {
				req := newDeviceAuthRequest(t, fmt.Sprintf("uid-%d", i), fmt.Sprintf("02:00:00:00:03:%02x", i))

				first := authDevice(t, compose, req)

				tc.change(t, &req)
				second := authDevice(t, compose, req)

				if tc.sameUID {
					assert.Equal(t, first.UID, second.UID)
				} else {
					assert.NotEqual(t, first.UID, second.UID)
				}
			})
		}
	})
}

// TestDeviceAuthAgent runs a real agent against the endpoint: the name it enrolls under, the
// heartbeat that keeps last_seen moving, and the re-registration that sends a removed
// device back to pending under the same uid.
func TestDeviceAuthAgent(t *testing.T) {
	t.Run("an agent's hostname has its dots replaced by underscores", func(t *testing.T) {
		ctx := context.Background()
		compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

		startAgent(t, ctx, compose, NewAgentContainerWithHostname("e2e.agent.example"))

		devices := []models.Device{}

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			resp, err := compose.R(ctx).SetResult(&devices).Get("/api/devices?status=pending")
			if !assert.NoError(tt, err) {
				return
			}

			assert.Equal(tt, http.StatusOK, resp.StatusCode(), resp.String())
			if assert.Len(tt, devices, 1) {
				assert.Equal(tt, "e2e_agent_example", devices[0].Name)
			}
		}, 30*time.Second, 1*time.Second)
	})

	t.Run("the agent's heartbeat advances last_seen", func(t *testing.T) {
		ctx := context.Background()
		compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

		_, device := startAcceptedAgent(t, ctx, compose)

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			current := models.Device{}

			resp, err := compose.R(ctx).SetResult(&current).Get("/api/devices/" + device.UID)
			if !assert.NoError(tt, err) {
				return
			}

			assert.Equal(tt, http.StatusOK, resp.StatusCode(), resp.String())
			assert.True(tt, current.LastSeen.After(device.LastSeen), "last_seen %s has not moved past %s", current.LastSeen, device.LastSeen)
		}, heartbeatTimeout, 2*time.Second)
	})

	t.Run("a removed device re-registers as pending under the same uid", func(t *testing.T) {
		ctx := context.Background()
		compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

		_, device := startAcceptedAgent(t, ctx, compose)

		resp, err := compose.R(ctx).Delete("/api/devices/" + device.UID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		devices := []models.Device{}

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			resp, err := compose.R(ctx).SetResult(&devices).Get("/api/devices?status=pending")
			if !assert.NoError(tt, err) {
				return
			}

			assert.Equal(tt, http.StatusOK, resp.StatusCode(), resp.String())
			if assert.Len(tt, devices, 1) {
				assert.Equal(tt, device.UID, devices[0].UID)
			}
		}, 30*time.Second, 1*time.Second)
	})
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

func apiPublicKey(t *testing.T) *rsa.PublicKey {
	t.Helper()

	data, err := os.ReadFile("../api_public_key")
	require.NoError(t, err)

	block, _ := pem.Decode(data)
	require.NotNil(t, block)

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	require.NoError(t, err)

	public, ok := key.(*rsa.PublicKey)
	require.True(t, ok)

	return public
}
