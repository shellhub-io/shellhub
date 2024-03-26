package environment

import (
	"io"
	"net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Service string

const (
	ServiceGateway Service = "gateway"
	ServiceAgent   Service = "agent"
	ServiceAPI     Service = "api"
	ServiceCLI     Service = "cli"
	ServiceSSH     Service = "ssh"
	ServiceUI      Service = "ui"
)

// getFreePort returns a randomly available TCP port. It can be used to avoid
// network conflicts in Docker Compose.
func getFreePort(t *testing.T) string {
	t.Helper()

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	l, err := net.ListenTCP("tcp", addr)
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}
	defer l.Close()

	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
}

func ReaderToString(t *testing.T, reader io.Reader) string {
	t.Helper()

	data, err := io.ReadAll(reader)
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	return string(data)
}
