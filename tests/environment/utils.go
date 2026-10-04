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

// ReservePort reserves a port through [testport.Reserve] and releases it when the test ends.
func ReservePort(t *testing.T) string {
	t.Helper()

	r, err := testport.Reserve()
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, r.Release()) })

	return r.Port()
}
