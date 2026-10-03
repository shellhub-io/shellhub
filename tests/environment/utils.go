package environment

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Service names a container in the test stack.
type Service string

// The services a test may reach through [DockerCompose.Service].
const (
	ServiceGateway Service = "gateway"
	ServiceServer  Service = "server"
)

const (
	lowestRandomPort       = 10000
	defaultEphemeralStart  = 32768
	ephemeralPortRangePath = "/proc/sys/net/ipv4/ip_local_port_range"
)

var freePortController []string

func ephemeralPortStart() int {
	data, err := os.ReadFile(ephemeralPortRangePath)
	if err != nil {
		return defaultEphemeralStart
	}

	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return defaultEphemeralStart
	}

	start, err := strconv.Atoi(fields[0])
	if err != nil || start <= lowestRandomPort {
		return defaultEphemeralStart
	}

	return start
}

func freePort(ctx context.Context) (string, error) {
	limit := ephemeralPortStart()
	span := big.NewInt(int64(limit - lowestRandomPort))

	for range limit - lowestRandomPort {
		offset, err := rand.Int(rand.Reader, span)
		if err != nil {
			return "", err
		}

		port := strconv.Itoa(lowestRandomPort + int(offset.Int64()))
		if slices.Contains(freePortController, port) {
			continue
		}

		l, err := new(net.ListenConfig).Listen(ctx, "tcp", net.JoinHostPort("127.0.0.1", port))
		if err == nil {
			l.Close() //nolint:errcheck // the probe listener is discarded; a failed close changes nothing
		}

		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		if err != nil {
			continue
		}

		freePortController = append(freePortController, port)

		return port, nil
	}

	return "", fmt.Errorf("no free port between %d and %d", lowestRandomPort, limit)
}

// GetFreePort returns a random TCP port free on 127.0.0.1, below the kernel's ephemeral range so
// that rootless Docker under pasta, which forwards only non-ephemeral ports, can publish it.
func GetFreePort(t *testing.T) string {
	t.Helper()

	port, err := freePort(t.Context())
	require.NoError(t, err)

	return port
}
