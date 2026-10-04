package testport_test

import (
	"context"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/shellhub-io/shellhub/pkg/testport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

func exclusiveBind(t *testing.T, port string) error {
	t.Helper()

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	require.NoError(t, err)
	defer syscall.Close(fd) //nolint:errcheck // the probe socket is discarded

	num, err := strconv.Atoi(port)
	require.NoError(t, err)

	return syscall.Bind(fd, &syscall.SockaddrInet4{Port: num})
}

func useEphemeralPortRange(t *testing.T, portRange *string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "ip_local_port_range")
	if portRange != nil {
		require.NoError(t, os.WriteFile(path, []byte(*portRange), 0o600))
	}

	t.Cleanup(testport.SetEphemeralPortRangePath(path))
}

func TestReserveHoldsThePortForTheDaemonOnly(t *testing.T) {
	r, err := testport.Reserve()
	require.NoError(t, err)

	require.ErrorIs(t, exclusiveBind(t, r.Port()), syscall.EADDRINUSE)

	l, err := new(net.ListenConfig).Listen(t.Context(), "tcp", net.JoinHostPort("0.0.0.0", r.Port()))
	require.NoError(t, err)
	require.NoError(t, l.Close())

	require.NoError(t, r.Release())
	require.NoError(t, exclusiveBind(t, r.Port()))
	require.NoError(t, r.Release())
}

func TestReserve(t *testing.T) {
	cases := []struct {
		description string
		portRange   *string
		start       int
	}{
		{
			description: "reads the start of the ephemeral range",
			portRange:   new("20000\t60999\n"),
			start:       20000,
		},
		{
			description: "falls back to 32768 when the range starts at or below 10000",
			portRange:   new("9000\t60999\n"),
			start:       32768,
		},
		{
			description: "falls back to 32768 when the range file is empty",
			portRange:   new(""),
			start:       32768,
		},
		{
			description: "falls back to 32768 when the range file is missing",
			start:       32768,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			useEphemeralPortRange(t, tc.portRange)

			require.Equal(t, tc.start, testport.EphemeralPortStart())

			seen := map[string]bool{}

			for range 20 {
				r, err := testport.Reserve()
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, r.Release()) })

				number, err := strconv.Atoi(r.Port())
				require.NoError(t, err)

				assert.GreaterOrEqual(t, number, 10000)
				assert.Less(t, number, tc.start)
				assert.False(t, seen[r.Port()], "port %s reserved twice", r.Port())

				seen[r.Port()] = true
			}
		})
	}
}

func TestReserveReturnsErrNoFreePortOnceTheRangeIsHeld(t *testing.T) {
	useEphemeralPortRange(t, new("10001\t60999\n"))

	r, err := testport.Reserve()
	if err == nil {
		t.Cleanup(func() { require.NoError(t, r.Release()) })
		_, err = testport.Reserve()
	}

	require.ErrorIs(t, err, testport.ErrNoFreePort)
}

func TestBindPublishesOnAReservationReleasedOnTerminate(t *testing.T) {
	req := testcontainers.GenericContainerRequest{}
	require.NoError(t, testport.Bind("5432/tcp").Customize(&req))

	hostConfig := &container.HostConfig{}
	req.HostConfigModifier(hostConfig)

	bindings := hostConfig.PortBindings[network.MustParsePort("5432/tcp")]
	require.Len(t, bindings, 1)

	port := bindings[0].HostPort
	require.ErrorIs(t, exclusiveBind(t, port), syscall.EADDRINUSE)

	for _, hooks := range req.LifecycleHooks {
		for _, hook := range hooks.PostTerminates {
			require.NoError(t, hook(t.Context(), nil))
		}
	}

	require.NoError(t, exclusiveBind(t, port))
}

func TestReleaseFreesThePortWhileAChildProcessRuns(t *testing.T) {
	r, err := testport.Reserve()
	require.NoError(t, err)

	child := exec.CommandContext(context.WithoutCancel(t.Context()), "cat")
	stdin, err := child.StdinPipe()
	require.NoError(t, err)
	require.NoError(t, child.Start())
	t.Cleanup(func() {
		require.NoError(t, stdin.Close())
		require.NoError(t, child.Wait())
	})

	require.NoError(t, r.Release())
	require.NoError(t, exclusiveBind(t, r.Port()))
}
