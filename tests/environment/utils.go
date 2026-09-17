package environment

import (
	"bytes"
	"errors"
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

func freePort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return "", err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", err
	}

	defer l.Close() //nolint:errcheck // port already read; close is best-effort

	tcpAddr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return "", errors.New("listener address is not TCP")
	}

	port := strconv.Itoa(tcpAddr.Port)
	if slices.Contains(freePortController, port) {
		return freePort()
	}

	freePortController = append(freePortController, port)

	return port, nil
}

// GetFreePort returns a randomly available TCP port. It can be used to avoid
// network conflicts in Docker Compose.
func GetFreePort(t *testing.T) string {
	t.Helper()

	port, err := freePort()
	require.NoError(t, err)

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
