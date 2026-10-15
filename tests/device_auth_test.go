package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/devicekey"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

const (
	otherNamespaceName     = "otherspace"
	otherNamespace         = "00000000-0000-4000-0000-000000000001"
	fastTunnelPingInterval = 2 * time.Second
	tunnelBeatTimeout      = 15 * time.Second
	staleBeatCycles        = 3
	staleBeatWindow        = 2500 * time.Millisecond
)

// TestDeviceAuth drives the endpoint an agent enrolls through with requests built by hand, so each
// case controls exactly which part of the device's identity changes. The provisioning-key paths
// are covered by [TestProvisioningKeyEnrollment].
func TestDeviceAuth(t *testing.T) {
	t.Parallel()

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
		assert.Equal(t, &authorizer.DeviceClaims{UID: res.UID, TenantID: ShellHubNamespace, KeyVerified: true}, claims)
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
	t.Parallel()

	t.Run("an agent's hostname has its dots replaced by underscores", func(t *testing.T) {
		compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

		startAgent(t, t.Context(), compose, NewAgentContainerWithHostname("e2e.agent.example"))

		device := compose.AwaitDeviceWithStatus(t, models.DeviceStatusPending)
		assert.Equal(t, "e2e_agent_example", device.Name)
	})

	t.Run("the tunnel ping keeps advancing last_seen", func(t *testing.T) {
		compose := newFastTunnelPingEnvironment(t, t.Context())

		_, device := startAcceptedAgent(t, t.Context(), compose)

		seen := device.LastSeen
		for range 2 {
			seen = awaitLastSeenPast(t, compose, device.UID, seen)
		}
	})

	t.Run("an agent stopped soon after it connects stays offline", func(t *testing.T) {
		compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

		agent, device := startAcceptedAgent(t, t.Context(), compose)
		require.NoError(t, agent.Stop(t.Context(), nil))
		compose.AwaitDeviceOffline(t, device.UID)

		for cycle := range staleBeatCycles {
			require.NoError(t, agent.Start(t.Context()))
			awaitAgentConnections(t, agent, cycle+2)

			require.NoError(t, agent.Stop(t.Context(), nil))
			compose.AwaitDeviceOffline(t, device.UID)

			require.Never(t, func() bool {
				current, _, err := compose.GetDevice(t.Context(), device.UID)

				return err == nil && current.Online
			}, staleBeatWindow, 100*time.Millisecond, "cycle %d: the device came back online after its agent stopped", cycle)
		}
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

func newFastTunnelPingEnvironment(t *testing.T, ctx context.Context) *environment.DockerCompose {
	t.Helper()

	configurator := environment.New(t, run).WithEnv("SHELLHUB_TUNNEL_PING_INTERVAL", fastTunnelPingInterval.String())

	return newConfiguredSSHEnvironment(t, ctx, configurator, models.SSHAccessModeLegacy)
}

func awaitLastSeenPast(t *testing.T, compose *environment.DockerCompose, uid string, seen time.Time) time.Time {
	t.Helper()

	var current *models.Device

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		device, resp, err := compose.GetDevice(t.Context(), uid)
		if !assert.NoError(tt, err) {
			return
		}

		if !assert.Equal(tt, http.StatusOK, resp.StatusCode(), resp.String()) {
			return
		}

		assert.True(tt, device.LastSeen.After(seen), "last_seen %s has not moved past %s", device.LastSeen, seen)

		current = device
	}, 5*fastTunnelPingInterval, fastTunnelPingInterval/4)

	return current.LastSeen
}

func awaitAgentConnections(t *testing.T, agent testcontainers.Container, connections int) {
	t.Helper()

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		logs, err := agentLogs(t.Context(), agent)
		if !assert.NoError(tt, err) {
			return
		}

		assert.Len(tt, agentConnectedLine.FindAll(logs, -1), connections, "the agent has not connected to the server")
	}, 30*time.Second, 100*time.Millisecond)
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

var devicePrivateKeys sync.Map

func newDevicePublicKey(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	publicKey := string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
	}))
	devicePrivateKeys.Store(publicKey, key)

	return publicKey
}

func signDeviceAuth(ctx context.Context, compose *environment.DockerCompose, req requests.DeviceAuth) (requests.DeviceAuth, error) {
	if req.Challenge != "" || req.Signature != "" {
		return req, nil
	}

	key, ok := devicePrivateKeys.Load(req.PublicKey)
	if !ok {
		return req, errors.New("the request's public key was not made by newDevicePublicKey, so nothing can sign it")
	}

	challenge, err := requestDeviceAuthChallenge(ctx, compose)
	if err != nil {
		return req, err
	}

	return signDeviceAuthWith(key.(*rsa.PrivateKey), req, challenge) //nolint:forcetypeassert // devicePrivateKeys holds nothing but device keys
}

func requestDeviceAuthChallenge(ctx context.Context, compose *environment.DockerCompose) (string, error) {
	challenge := new(models.DeviceAuthChallenge)

	resp, err := compose.Anonymous(ctx).SetResult(challenge).Post("/api/devices/auth/challenge")
	if err != nil {
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("issuing a device auth challenge: %s", resp.String())
	}

	return challenge.Challenge, nil
}

func signDeviceAuthWith(key *rsa.PrivateKey, req requests.DeviceAuth, challenge string) (requests.DeviceAuth, error) {
	signature, err := devicekey.Sign(key, devicekey.Statement{
		Challenge:       challenge,
		TenantID:        req.TenantID,
		ProvisioningKey: req.ProvisioningKey,
		PublicKey:       req.PublicKey,
	})
	if err != nil {
		return req, err
	}

	req.Challenge, req.Signature = challenge, signature

	return req, nil
}

func authDevice(t *testing.T, compose *environment.DockerCompose, req requests.DeviceAuth) *models.DeviceAuthResponse {
	t.Helper()

	req, err := signDeviceAuth(t.Context(), compose, req)
	require.NoError(t, err)

	res := new(models.DeviceAuthResponse)

	resp, err := compose.Anonymous(t.Context()).SetBody(req).SetResult(res).Post("/api/devices/auth")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return res
}
