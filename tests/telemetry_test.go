package main

import (
	"bufio"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/moby/moby/client"
	"github.com/shellhub-io/shellhub/pkg/telemetry"
	"github.com/shellhub-io/shellhub/pkg/testimage"
	"github.com/shellhub-io/shellhub/pkg/testrun"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

const (
	telemetryHeartbeatCron = "30 5 * * *"

	telemetryFlushWindow = 10 * time.Second
)

type telemetryEvent = telemetry.Preview

var telemetryCollectorBuild sync.Mutex

type telemetryCollector struct {
	container testcontainers.Container
	writeKey  string
}

// TestTelemetry runs the daily heartbeat on demand against a collector stub the stack starts before
// the server, so the stub also sees what the server sends as it boots.
func TestTelemetry(t *testing.T) {
	t.Parallel()

	t.Run("an instance with telemetry on sends its heartbeat", func(t *testing.T) {
		t.Parallel()

		compose := newConfiguredSSHEnvironment(t, t.Context(), telemetryEnvironment(t), newNamespace)
		collector := newTelemetryCollector(compose, telemetry.CommunityWriteKey)

		started := collector.awaitEvent(t, "instance started", anyTelemetryEvent)
		assert.Equal(t, "community", started.Properties["edition"])

		compose.RunCron(t, telemetryHeartbeatCron)

		heartbeat := collector.awaitEvent(t, "instance heartbeat", anyTelemetryEvent)
		assert.Equal(t, started.DistinctID, heartbeat.DistinctID)
		assert.NotEmpty(t, heartbeat.DistinctID)
		assert.Equal(t, "community", heartbeat.Properties["edition"])
		assert.Equal(t, map[string]any{"process_person_profile": false}, heartbeat.Options)
		assert.Contains(t, heartbeat.Properties, "devices_accepted")
	})

	t.Run("SHELLHUB_TELEMETRY=false sends nothing", func(t *testing.T) {
		t.Parallel()

		compose := newConfiguredSSHEnvironment(t, t.Context(),
			telemetryEnvironment(t).WithEnv("SHELLHUB_TELEMETRY", "false"), newNamespace)
		collector := newTelemetryCollector(compose, telemetry.CommunityWriteKey)

		compose.RunCron(t, telemetryHeartbeatCron)

		collector.requireSilence(t)
	})
}

// TestEnterpriseTelemetry drives the admin telemetry setting: the preview is what the heartbeat
// sends, and turning it off stops the heartbeat without a restart.
func TestEnterpriseTelemetry(t *testing.T) {
	t.Parallel()

	compose := newEnterpriseEnvironment(t, t.Context(), telemetryEnvironment(t))
	collector := newTelemetryCollector(compose, telemetry.EnterpriseWriteKey)

	t.Run("the preview is the heartbeat the instance sends", func(t *testing.T) {
		status := readTelemetryStatus(t, compose)
		require.True(t, status.Enabled)
		require.False(t, status.DisabledByEnvironment)
		require.Len(t, status.Preview, 1)

		compose.RunCron(t, telemetryHeartbeatCron)

		heartbeat := collector.awaitEvent(t, "instance heartbeat", anyTelemetryEvent)
		assert.Equal(t, status.Preview[0].DistinctID, heartbeat.DistinctID)
		assert.Equal(t, status.Preview[0].Options, heartbeat.Options)

		for name, value := range status.Preview[0].Properties {
			assert.Equal(t, value, heartbeat.Properties[name], "property %s", name)
		}
	})

	t.Run("no other response carries the instance ID", func(t *testing.T) {
		instanceID := readTelemetryStatus(t, compose).Preview[0].DistinctID
		require.NotEmpty(t, instanceID)

		info, err := compose.Anonymous(t.Context()).Get("/info")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, info.StatusCode(), info.String())
		assert.NotContains(t, info.String(), instanceID)

		stats, err := compose.R(t.Context()).Get("/admin/api/stats")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, stats.StatusCode(), stats.String())
		assert.NotContains(t, stats.String(), instanceID)
	})

	t.Run("turning telemetry off stops the heartbeat", func(t *testing.T) {
		resp, err := compose.R(t.Context()).SetBody(map[string]bool{"enabled": false}).Put("/admin/api/telemetry")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		assert.False(t, readTelemetryStatus(t, compose).Enabled)

		sent := len(collector.requireEvents(t))

		compose.ExpireCacheEntryIn(t, "telemetry:heartbeat:"+time.Now().UTC().Format(time.DateOnly), 0) //nolint:forbidigo // the server dates its claim by the wall clock it shares with this host
		compose.RunCron(t, telemetryHeartbeatCron)

		time.Sleep(telemetryFlushWindow)
		assert.Len(t, collector.requireEvents(t), sent, "the instance sent telemetry after it was turned off")
	})
}

type telemetryStatus struct {
	Enabled               bool             `json:"enabled"`
	DisabledByEnvironment bool             `json:"disabled_by_environment"`
	Preview               []telemetryEvent `json:"preview"`
}

func readTelemetryStatus(t *testing.T, compose *environment.DockerCompose) telemetryStatus {
	t.Helper()

	status := telemetryStatus{}

	resp, err := compose.R(t.Context()).SetResult(&status).Get("/admin/api/telemetry")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return status
}

func telemetryEnvironment(t *testing.T) *environment.DockerComposeConfigurator {
	t.Helper()

	return environment.New(t, run).
		WithCronTrigger().
		WithTelemetryCollector(buildTelemetryCollector(t))
}

func buildTelemetryCollector(t *testing.T) string {
	t.Helper()

	buildArgs, err := testimage.BuildArgs("..")
	require.NoError(t, err)

	provider, err := testcontainers.NewDockerProvider()
	require.NoError(t, err)

	defer provider.Close() //nolint:errcheck // the image is built or the build failed; closing the provider changes neither

	telemetryCollectorBuild.Lock()
	defer telemetryCollectorBuild.Unlock()

	image, err := provider.BuildImage(t.Context(), &testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Repo:      "telemetry-collector",
			Tag:       "test",
			Context:   "cmd/telemetry-collector",
			KeepImage: true,
			BuildArgs: buildArgs,
			BuildOptionsModifier: func(options *client.ImageBuildOptions) {
				options.Labels = map[string]string{testrun.RunLabel: run.ID()}
			},
		},
	})
	require.NoError(t, err)

	return image
}

func newTelemetryCollector(compose *environment.DockerCompose, writeKey string) *telemetryCollector {
	return &telemetryCollector{container: compose.Service(environment.ServiceTelemetryCollector), writeKey: writeKey}
}

func (c *telemetryCollector) requireEvents(t *testing.T) []telemetryEvent {
	t.Helper()

	events, err := c.events(t.Context())
	require.NoError(t, err)

	return events
}

func (c *telemetryCollector) events(ctx context.Context) ([]telemetryEvent, error) {
	reader, err := c.container.Logs(ctx)
	if err != nil {
		return nil, err
	}

	defer reader.Close() //nolint:errcheck // a failed Close of a log stream already read in full loses nothing

	var events []telemetryEvent

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		var request struct {
			Authorization string `json:"authorization"`
			Body          struct {
				Batch []telemetryEvent `json:"batch"`
			} `json:"body"`
		}

		if json.Unmarshal(scanner.Bytes(), &request) != nil || request.Authorization == "" {
			continue
		}

		if request.Authorization != "Bearer "+c.writeKey {
			return nil, fmt.Errorf("the collector received the write key %q, want %q", request.Authorization, "Bearer "+c.writeKey)
		}

		events = append(events, request.Body.Batch...)
	}

	return events, scanner.Err()
}

func anyTelemetryEvent(telemetryEvent) bool { return true }

func (c *telemetryCollector) awaitEvent(t *testing.T, name string, match func(telemetryEvent) bool) telemetryEvent {
	t.Helper()

	var found telemetryEvent

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		events, err := c.events(t.Context())
		if !assert.NoError(tt, err) {
			return
		}

		for _, event := range events {
			if event.Event == name && match(event) {
				found = event

				return
			}
		}

		assert.Fail(tt, "no "+name+" event yet")
	}, 30*time.Second, 500*time.Millisecond)

	return found
}

func (c *telemetryCollector) requireSilence(t *testing.T) {
	t.Helper()

	time.Sleep(telemetryFlushWindow)
	require.Empty(t, c.requireEvents(t), "the instance sent telemetry with SHELLHUB_TELEMETRY=false")
}

// TestEnterpriseTelemetryDisabledByEnvironment checks that SHELLHUB_TELEMETRY=false locks the admin
// setting: the console reads it as forced off, and turning the setting on sends nothing.
func TestEnterpriseTelemetryDisabledByEnvironment(t *testing.T) {
	t.Parallel()

	compose := newEnterpriseEnvironment(t, t.Context(), telemetryEnvironment(t).WithEnv("SHELLHUB_TELEMETRY", "false"))
	collector := newTelemetryCollector(compose, telemetry.EnterpriseWriteKey)

	resp, err := compose.R(t.Context()).SetBody(map[string]bool{"enabled": true}).Put("/admin/api/telemetry")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	status := readTelemetryStatus(t, compose)
	assert.True(t, status.Enabled)
	assert.True(t, status.DisabledByEnvironment)

	compose.RunCron(t, telemetryHeartbeatCron)

	collector.requireSilence(t)
}

// TestCloudTelemetry checks that cloud adds a namespace heartbeat per namespace, grouped by the
// HMAC of its tenant ID under the deployment's key, and authenticated with the Cloud write key.
func TestCloudTelemetry(t *testing.T) {
	t.Parallel()

	const (
		tenantKey = "cloud-telemetry-tenant-key"
		writeKey  = "cloud-telemetry-write-key"
	)

	compose := telemetryEnvironment(t).
		WithEdition(environment.EditionCloud).
		WithEnv("SHELLHUB_TELEMETRY_TENANT_KEY", tenantKey).
		WithEnv("SHELLHUB_TELEMETRY_WRITE_KEY", writeKey).
		Up(t.Context())
	t.Cleanup(compose.Down)

	tenant := uuid.Generate()
	compose.NewUser(t, "telemetry-owner", "telemetry-owner"+stripeTestEmailDomain, ShellHubPassword)
	compose.NewNamespace(t, "telemetry-owner", "telemetry-space", tenant)

	collector := newTelemetryCollector(compose, writeKey)

	compose.RunCron(t, telemetryHeartbeatCron)

	mac := hmac.New(sha256.New, []byte(tenantKey))
	mac.Write([]byte(tenant))
	hashed := hex.EncodeToString(mac.Sum(nil))

	namespace := collector.awaitEvent(t, "namespace heartbeat", func(event telemetryEvent) bool {
		groups, _ := event.Properties["$groups"].(map[string]any)

		return groups["namespace"] == hashed
	})

	assert.InDelta(t, 1, namespace.Properties["members"], 0)
	assert.NotContains(t, namespace.Properties, "tenant_id")
}
