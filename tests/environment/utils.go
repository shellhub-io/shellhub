package environment

import (
	"bytes"
	"io"
	"net"
	"slices"
	"strconv"
	"testing"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/stretchr/testify/require"
)

// Service names a container in the test stack.
type Service string

// The services a test may reach through [DockerCompose.Service].
const (
	ServiceGateway Service = "gateway"
	ServiceAgent   Service = "agent"
	ServiceServer  Service = "server"
)

var freePortController []string

// GetFreePort returns a randomly available TCP port. It can be used to avoid
// network conflicts in Docker Compose.
func GetFreePort(t *testing.T) string {
	t.Helper()

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	require.NoError(t, err)

	l, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err)

	defer l.Close() //nolint:errcheck

	tcpAddr, ok := l.Addr().(*net.TCPAddr)
	require.True(t, ok, "listener address is not TCP")

	port := strconv.Itoa(tcpAddr.Port)
	if slices.Contains(freePortController, port) {
		return GetFreePort(t)
	}

	freePortController = append(freePortController, port)

	return port
}

// ReaderToString drains a Docker multiplexed stream, returning its stdout and discarding stderr.
func ReaderToString(t *testing.T, reader io.Reader) string {
	t.Helper()

	buffer := bytes.NewBuffer(make([]byte, 1024))

	_, err := stdcopy.StdCopy(buffer, io.Discard, reader)
	require.NoError(t, err)

	return buffer.String()
}
