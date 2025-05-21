package environment

import (
	"bytes"
	"io"
	"net"
	"strconv"
	"testing"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

type Service string

const (
	ServiceGateway Service = "gateway"
	ServiceAgent   Service = "agent"
	ServiceServer  Service = "server"
	ServiceUI      Service = "ui"
)

var freePortController []string

// GetFreePort returns a randomly available TCP port. It can be used to avoid
// network conflicts in Docker Compose.
func GetFreePort(t *testing.T) string {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	require.NoError(t, err)

	l, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err)

	defer l.Close()

	port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	if slices.Contains(freePortController, port) {
		return GetFreePort(t)
	}

	freePortController = append(freePortController, port)

	return port
}

func ReaderToString(t *testing.T, reader io.Reader) string {
	buffer := bytes.NewBuffer(make([]byte, 1024))

	_, err := stdcopy.StdCopy(buffer, io.Discard, reader)
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	return buffer.String()
}
