package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

const connectorLabel = "io.shellhub.e2e.connector"

// TestConnectorContainerLifecycle covers a container a connector serves going through the states a
// device goes through, in order: it asks to join as pending, comes online once accepted, goes
// offline while stopped and back online under the same uid when started again, and once removed is
// kept as removed until it starts again and asks to join as pending under the same uid.
func TestConnectorContainerLifecycle(t *testing.T) {
	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

	name := "connected-" + randomHex(t)
	target := startLabelledContainer(t, ctx, name)
	startConnector(t, ctx, compose, connectorLabel+"="+name)

	pending := awaitContainerWithStatus(t, compose, models.DeviceStatusPending)
	require.Equal(t, name, pending.Name, "the container asked to join under another name")

	uid := pending.UID

	compose.UpdateDeviceStatus(t, uid, environment.DeviceActionAccept)
	compose.AwaitDeviceOnline(t, uid)

	require.NoError(t, target.Stop(ctx, nil))
	compose.AwaitDeviceOffline(t, uid)

	require.NoError(t, target.Start(ctx))
	compose.AwaitDeviceOnline(t, uid)
	require.Equal(t, uid, awaitContainerWithStatus(t, compose, models.DeviceStatusAccepted).UID, "the restarted container came back under another uid")

	require.NoError(t, target.Stop(ctx, nil))
	compose.AwaitDeviceOffline(t, uid)
	compose.DeleteDevice(t, uid)

	removed := requireDevice(t, compose, uid)
	require.Equal(t, models.DeviceStatusRemoved, removed.Status, "the removed container was not kept as removed")
	require.NotNil(t, removed.RemovedAt)

	require.NoError(t, target.Start(ctx))
	require.Equal(t, uid, awaitContainerWithStatus(t, compose, models.DeviceStatusPending).UID, "the removed container asked to join again under another uid")
}

func startLabelledContainer(t *testing.T, ctx context.Context, name string) testcontainers.Container {
	t.Helper()

	image, err := buildAgentImage(ctx, "")
	require.NoError(t, err)

	labels := run.Labels()
	labels[connectorLabel] = name

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:      image,
			Name:       name,
			Labels:     labels,
			Entrypoint: []string{"sleep"},
			Cmd:        []string{"infinity"},
		},
		Started: true,
		Logger:  log.New(io.Discard, "", log.LstdFlags),
	})
	require.NoError(t, err)

	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	return c
}

func startConnector(t *testing.T, ctx context.Context, compose *environment.DockerCompose, label string) testcontainers.Container {
	t.Helper()

	image, err := buildAgentImage(ctx, "")
	require.NoError(t, err)

	gateway := compose.Service(environment.ServiceGateway).GetContainerID()

	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: image,
			Cmd:   []string{"connector"},
			Env: map[string]string{
				"SHELLHUB_SERVER_ADDRESS":  "http://localhost",
				"SHELLHUB_TENANT_ID":       ShellHubNamespace,
				"SHELLHUB_PRIVATE_KEYS":    "/tmp/keys",
				"SHELLHUB_CONNECTOR_LABEL": label,
			},
			Labels: run.Labels(),
			HostConfigModifier: func(hc *container.HostConfig) {
				hc.NetworkMode = container.NetworkMode("container:" + gateway)
				hc.Mounts = append(hc.Mounts, mount.Mount{
					Type:   mount.TypeBind,
					Source: testcontainers.MustExtractDockerSocket(ctx),
					Target: "/var/run/docker.sock",
				})
			},
		},
		Started: true,
		Logger:  log.New(io.Discard, "", log.LstdFlags),
	})
	require.NoError(t, err)

	t.Cleanup(func() { _ = c.Terminate(context.Background()) })

	return c
}

func awaitContainerWithStatus(t *testing.T, compose *environment.DockerCompose, status models.DeviceStatus) models.Device {
	t.Helper()

	containers := []models.Device{}

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		resp, err := compose.R(t.Context()).
			SetQueryParam("status", string(status)).
			SetResult(&containers).
			Get("/api/containers")
		assert.NoError(tt, err)
		assert.Equal(tt, http.StatusOK, resp.StatusCode())
		assert.Len(tt, containers, 1)
	}, 30*time.Second, time.Second)

	return containers[0]
}

func randomHex(t *testing.T) string {
	t.Helper()

	raw := make([]byte, 6)
	_, err := rand.Read(raw)
	require.NoError(t, err)

	return hex.EncodeToString(raw)
}
