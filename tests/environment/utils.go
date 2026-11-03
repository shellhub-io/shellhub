package environment

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/testport"
	"github.com/stretchr/testify/require"
)

const stackArtifactsDir = ".stack"

// Service names a container in the test stack.
type Service string

// The services a test may reach through [DockerCompose.Service]. ServiceRelay forwards port 80
// of its network namespace to the server's, so an agent that joins that namespace reaches the
// gateway at http://localhost and keeps its route while the server restarts.
// ServiceObjectStorage, the S3-compatible store session recordings are kept in, runs in
// enterprise and cloud stacks only. ServiceTelemetryCollector runs only in a stack started with
// [DockerComposeConfigurator.WithTelemetryCollector].
const (
	ServiceServer             Service = "server"
	ServiceRelay              Service = "relay"
	ServicePostgres           Service = "postgres"
	ServiceRedis              Service = "redis"
	ServiceObjectStorage      Service = "minio"
	ServiceTelemetryCollector Service = "telemetry-collector"
)

// ReservePort reserves a port through [testport.Reserve] and releases it when the test ends.
func ReservePort(t *testing.T) string {
	t.Helper()

	r, err := testport.Reserve()
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, r.Release()) })

	return r.Port()
}
