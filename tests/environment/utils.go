package environment

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/testport"
	"github.com/stretchr/testify/require"
)

// Service names a container in the test stack.
type Service string

// The services a test may reach through [DockerCompose.Service].
const (
	ServiceGateway Service = "gateway"
	ServiceServer  Service = "server"
)

// GetFreePort returns a port from [testport.Free], free on 127.0.0.1 and below the kernel's
// ephemeral range, and fails t when there is none.
func GetFreePort(t *testing.T) string {
	t.Helper()

	port, err := testport.Free(t.Context())
	require.NoError(t, err)

	return port
}
