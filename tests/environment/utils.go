package environment

import (
	"bytes"
	"io"
	"net"
	"strconv"
	"sync"
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
	ServiceAPI     Service = "api"
	ServiceSSH     Service = "ssh"
	ServiceUI      Service = "ui"
)

var (
	freePortController []string
	portMutex          sync.Mutex
)

// PortReservation holds a reserved port and its listener to prevent race conditions.
type PortReservation struct {
	Port     string
	listener *net.TCPListener
}

// Release releases the port reservation, allowing it to be used.
func (pr *PortReservation) Release() {
	if pr.listener != nil {
		pr.listener.Close()
	}
}

// ReservePort reserves a randomly available TCP port and keeps it reserved
// until Release() is called. This prevents race conditions where another
// process might claim the port between allocation and usage.
func ReservePort(t *testing.T) *PortReservation {
	portMutex.Lock()
	defer portMutex.Unlock()

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	require.NoError(t, err)

	l, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err)

	port := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	if slices.Contains(freePortController, port) {
		l.Close()

		return ReservePort(t)
	}

	freePortController = append(freePortController, port)

	return &PortReservation{
		Port:     port,
		listener: l,
	}
}

func ReaderToString(t *testing.T, reader io.Reader) string {
	buffer := bytes.NewBuffer(make([]byte, 1024))

	_, err := stdcopy.StdCopy(buffer, io.Discard, reader)
	if !assert.NoError(t, err) {
		assert.FailNow(t, err.Error())
	}

	return buffer.String()
}
