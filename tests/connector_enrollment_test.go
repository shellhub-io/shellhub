package main

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

// TestConnectorEnrollment covers what a connector reports for the containers it serves when they
// ask to join: the identity and name each one carries, and which containers it serves at all.
func TestConnectorEnrollment(t *testing.T) {
	t.Parallel()

	t.Run("a container asks to join under the first 12 characters of its ID", func(t *testing.T) {
		t.Parallel()

		compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

		name := "identity-" + randomHex(t)
		target := startLabelledContainer(t, t.Context(), name)
		startConnector(t, t.Context(), compose, connectorLabel+"="+name)

		pending := awaitContainerWithStatus(t, compose, models.DeviceStatusPending)
		require.NotNil(t, pending.Identity)
		assert.Equal(t, target.GetContainerID()[:12], pending.Identity.MAC)
	})

	t.Run("a container's name has its dots replaced by underscores and is cut to 64 characters", func(t *testing.T) {
		t.Parallel()

		compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

		value := randomHex(t)
		name := "sanitized." + value + "." + strings.Repeat("x", 64)
		startIdleContainer(t, t.Context(), name, map[string]string{connectorLabel: value})
		connector := startConnector(t, t.Context(), compose, connectorLabel+"="+value)

		sanitized := "sanitized_" + value + "_" + strings.Repeat("x", 41)

		pending := awaitContainerWithStatus(t, compose, models.DeviceStatusPending)
		assert.Equal(t, sanitized, pending.Name)
		assert.Contains(t, containerLogs(t, t.Context(), connector), "hostname="+sanitized, "the connector left the dots for the server to replace")
	})

	t.Run("a container without the connector's label is left out", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		compose := newSSHEnvironment(t, ctx, models.SSHAccessModeLegacy)

		value := randomHex(t)
		runningBefore := startIdleContainer(t, ctx, "unlabelled-before-"+value, nil)
		otherValue := startIdleContainer(t, ctx, "other-value-"+value, map[string]string{connectorLabel: "other-" + value})
		startIdleContainer(t, ctx, "labelled-before-"+value, map[string]string{connectorLabel: value})

		connector := startConnector(t, ctx, compose, connectorLabel+"="+value)
		awaitContainerWithStatus(t, compose, models.DeviceStatusPending)

		startedAfter := startIdleContainer(t, ctx, "unlabelled-after-"+value, nil)
		barrier := startIdleContainer(t, ctx, "labelled-after-"+value, map[string]string{connectorLabel: value})

		names := []string{}
		for _, c := range awaitContainersWithStatus(t, compose, models.DeviceStatusPending, 2) {
			names = append(names, c.Name)
		}

		assert.ElementsMatch(t, []string{"labelled-before-" + value, "labelled-after-" + value}, names)

		logs := containerLogs(t, ctx, connector)
		require.Contains(t, logs, barrier.GetContainerID()[:12], "the connector never started serving the container started after it")

		for _, left := range []testcontainers.Container{runningBefore, otherValue, startedAfter} {
			assert.NotContains(t, logs, left.GetContainerID()[:12], "the connector served a container outside its label")
		}
	})
}

func containerLogs(t *testing.T, ctx context.Context, c testcontainers.Container) string {
	t.Helper()

	reader, err := c.Logs(ctx)
	require.NoError(t, err)

	defer reader.Close() //nolint:errcheck // the logs were read in full

	logs, err := io.ReadAll(reader)
	require.NoError(t, err)

	return string(logs)
}
