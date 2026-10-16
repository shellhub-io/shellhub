package main

import (
	"bytes"
	"context"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

const (
	selfUpdateRepository   = "shellhub/agent"
	selfUpdateFromVersion  = "v0.20.0"
	selfUpdateKeyDirectory = "/var/lib/shellhub"
	selfUpdateTimeout      = 3 * time.Minute
)

// TestAgentDockerSelfUpdate covers an agent running as a Docker container replacing itself with the
// version the server reports. The agent starts from an image in a registry the test runs, at a
// version below the server's, with the Docker socket and its private key on a volume. The server's
// version is pushed to the same repository and its tag removed from the daemon, so only a pull
// from the registry brings it back. The update starts a second container from the new image,
// which stops and removes the first, recreates it under its name from the new image, and removes
// itself.
func TestAgentDockerSelfUpdate(t *testing.T) {
	t.Parallel()

	compose := newSSHEnvironment(t, t.Context(), legacyNamespace)
	next := reportedServerVersion(t, compose)

	registry, err := environment.StartRegistry(t.Context(), run)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, registry.Close(t.Context())) })

	current := pushDockerAgent(t, registry, selfUpdateFromVersion)
	nextImage := pushDockerAgent(t, registry, next)

	docker := newDockerClient(t)

	original := startDockerAgent(t, compose, docker, current)

	replacement := awaitContainerReplaced(t, docker, original)

	t.Run("the agent pulls the server's version and runs it under its own name", func(t *testing.T) {
		assert.Equal(t, nextImage, replacement.Config.Image)

		_, err := docker.ImageInspect(t.Context(), nextImage)
		require.NoError(t, err, "the server's version was not pulled from the registry")

		acceptPendingDevice(t, compose)

		logs := daemonContainerLogs(t, docker, replacement.ID)
		assert.Regexp(t, `msg="Starting ShellHub".* version=`+regexp.QuoteMeta(next)+`\b`, logs, "the replacement does not run the server's version")
		assert.Regexp(t, agentConnectedLine, logs, "the replacement has not connected to the server")
	})

	t.Run("the update leaves no container behind but the one under the agent's name", func(t *testing.T) {
		_, err := docker.ContainerInspect(t.Context(), original.ID, client.ContainerInspectOptions{})
		require.True(t, cerrdefs.IsNotFound(err), "the container the agent ran in before the update is still there: %v", err)

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			ids, err := containersFrom(t.Context(), docker, nextImage)
			if assert.NoError(tt, err) {
				assert.Equal(tt, []string{replacement.ID}, ids, "the container that ran the update is still there")
			}
		}, 30*time.Second, time.Second)
	})
}

// TestAgentDockerSelfUpdateFailure covers a Docker self-update that cannot pull the server's
// version: the registry the agent came from holds only the agent's own version. The agent logs the
// failure, does not report the update as applied, and stays in the container it started in,
// connected to the server.
func TestAgentDockerSelfUpdateFailure(t *testing.T) {
	t.Parallel()

	compose := newSSHEnvironment(t, t.Context(), legacyNamespace)
	next := strings.TrimPrefix(reportedServerVersion(t, compose), "v")

	registry, err := environment.StartRegistry(t.Context(), run)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, registry.Close(t.Context())) })

	current := pushDockerAgent(t, registry, selfUpdateFromVersion)

	docker := newDockerClient(t)

	original := startDockerAgent(t, compose, docker, current)

	var logs string

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		read, err := readContainerLogs(t.Context(), docker, original.ID)
		if !assert.NoError(tt, err) {
			return
		}

		logs = read
		assert.Regexp(tt, agentUpdateWaiting, logs, "the agent has not finished its first update check")
	}, selfUpdateTimeout, time.Second)

	t.Run("the agent logs the failure and does not report the update as applied", func(t *testing.T) {
		assert.Regexp(t, `msg="Failed to apply update".* next_version=`+regexp.QuoteMeta(next)+`\b`, logs)
		assert.NotRegexp(t, agentUpdateApplied, logs)
	})

	t.Run("the agent keeps running in its own container and connects to the server", func(t *testing.T) {
		acceptPendingDevice(t, compose)

		inspected, err := docker.ContainerInspect(t.Context(), original.Name, client.ContainerInspectOptions{})
		require.NoError(t, err)
		assert.Equal(t, original.ID, inspected.Container.ID, "another container took the agent's name")
		assert.True(t, inspected.Container.State.Running, "the agent's container stopped")
		assert.Equal(t, current, inspected.Container.Config.Image)

		assert.Regexp(t, agentConnectedLine, daemonContainerLogs(t, docker, original.ID))
	})
}

func reportedServerVersion(t *testing.T, compose *environment.DockerCompose) string {
	t.Helper()

	info := new(models.Info)

	resp, err := compose.Anonymous(t.Context()).SetResult(info).Get("/info")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	require.NotEmpty(t, info.Version)

	return info.Version
}

func pushDockerAgent(t *testing.T, registry *environment.Registry, version string) string {
	t.Helper()

	image, err := buildAgentImage(t.Context(), version, "docker")
	require.NoError(t, err)

	reference, err := registry.Push(t.Context(), image, selfUpdateRepository, version)
	require.NoError(t, err)

	return reference
}

func startDockerAgent(t *testing.T, compose *environment.DockerCompose, docker *client.Client, image string) container.InspectResponse {
	t.Helper()

	volume := "shellhub-e2e-agent-key-" + randomHex(t)
	t.Cleanup(func() {
		ctx := context.WithoutCancel(t.Context())

		mounting, _ := docker.ContainerList(ctx, client.ContainerListOptions{All: true, Filters: make(client.Filters).Add("volume", volume)})
		for _, c := range mounting.Items {
			_, _ = docker.ContainerRemove(ctx, c.ID, client.ContainerRemoveOptions{Force: true})
		}

		_, _ = docker.VolumeRemove(ctx, volume, client.VolumeRemoveOptions{Force: true})
	})

	req, err := agentContainerRequest("", image, NewAgentContainerWithPrivateKey(selfUpdateKeyDirectory+"/device.key"), func(envs map[string]string) {
		envs["SHELLHUB_SERVER_ADDRESS"] = compose.BaseURL()
	})
	require.NoError(t, err)

	req.HostConfigModifier = func(hc *container.HostConfig) {
		hc.NetworkMode = network.NetworkHost
		hc.Mounts = append(hc.Mounts,
			mount.Mount{Type: mount.TypeBind, Source: testcontainers.MustExtractDockerSocket(t.Context()), Target: "/var/run/docker.sock"},
			mount.Mount{Type: mount.TypeVolume, Source: volume, Target: selfUpdateKeyDirectory},
		)
	}

	agent, err := createAgentContainer(t.Context(), req)
	require.NoError(t, err)
	runAgent(t, t.Context(), agent)

	inspected, err := docker.ContainerInspect(t.Context(), agent.GetContainerID(), client.ContainerInspectOptions{})
	require.NoError(t, err)

	return inspected.Container
}

func awaitContainerReplaced(t *testing.T, docker *client.Client, original container.InspectResponse) container.InspectResponse {
	t.Helper()

	var replacement container.InspectResponse

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		inspected, err := docker.ContainerInspect(t.Context(), original.Name, client.ContainerInspectOptions{})
		if !assert.NoError(tt, err) {
			return
		}

		if !assert.NotEqual(tt, original.ID, inspected.Container.ID, "the agent's name still holds the container it started in") {
			return
		}

		if assert.True(tt, inspected.Container.State.Running, "the replacement has not started yet") {
			replacement = inspected.Container
		}
	}, selfUpdateTimeout, time.Second)

	return replacement
}

func newDockerClient(t *testing.T) *client.Client {
	t.Helper()

	docker, err := client.New(client.FromEnv)
	require.NoError(t, err)
	t.Cleanup(func() { _ = docker.Close() })

	return docker
}

func daemonContainerLogs(t *testing.T, docker *client.Client, id string) string {
	t.Helper()

	logs, err := readContainerLogs(t.Context(), docker, id)
	require.NoError(t, err)

	return logs
}

func readContainerLogs(ctx context.Context, docker *client.Client, id string) (string, error) {
	reader, err := docker.ContainerLogs(ctx, id, client.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", err
	}

	defer reader.Close() //nolint:errcheck // the logs have been read in full

	var stdout, stderr bytes.Buffer

	if _, err := stdcopy.StdCopy(&stdout, &stderr, reader); err != nil {
		return "", err
	}

	return stdout.String() + stderr.String(), nil
}

func containersFrom(ctx context.Context, docker *client.Client, image string) ([]string, error) {
	listed, err := docker.ContainerList(ctx, client.ContainerListOptions{All: true, Filters: make(client.Filters).Add("ancestor", image)})
	if err != nil {
		return nil, err
	}

	ids := []string{}
	for _, c := range listed.Items {
		ids = append(ids, c.ID)
	}

	slices.Sort(ids)

	return ids, nil
}
