package testport_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/shellhub-io/shellhub/pkg/testport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

type fakeContainer struct {
	terminated   bool
	terminateErr error
}

func (f *fakeContainer) Terminate(context.Context, ...testcontainers.TerminateOption) error {
	f.terminated = true

	return f.terminateErr
}

func boundHostPort(t *testing.T, bind testcontainers.ContainerCustomizer) string {
	t.Helper()

	req := testcontainers.GenericContainerRequest{}
	require.NoError(t, bind.Customize(&req))

	hostConfig := &container.HostConfig{}
	req.HostConfigModifier(hostConfig)

	bindings := hostConfig.PortBindings[network.MustParsePort("5432/tcp")]
	require.Len(t, bindings, 1)

	return bindings[0].HostPort
}

func TestRun(t *testing.T) {
	t.Run("a collision is retried on another port", func(t *testing.T) {
		var containers []*fakeContainer
		var ports []string

		got, err := testport.Run(t.Context(), "5432/tcp", func(_ context.Context, bind testcontainers.ContainerCustomizer) (*fakeContainer, error) {
			c := &fakeContainer{}
			containers = append(containers, c)
			ports = append(ports, boundHostPort(t, bind))

			if len(containers) == 1 {
				return c, errors.New("Bind for 0.0.0.0:" + ports[0] + " failed: port is already allocated")
			}

			return c, nil
		})
		require.NoError(t, err)

		require.Len(t, containers, 2)
		assert.Same(t, containers[1], got)
		assert.True(t, containers[0].terminated)
		assert.False(t, containers[1].terminated)
		assert.NotEqual(t, ports[0], ports[1])
	})

	t.Run("an error that is not a collision is not retried", func(t *testing.T) {
		calls := 0
		failure := errors.New("pull access denied")

		_, err := testport.Run(t.Context(), "5432/tcp", func(context.Context, testcontainers.ContainerCustomizer) (*fakeContainer, error) {
			calls++

			return &fakeContainer{}, failure
		})

		require.ErrorIs(t, err, failure)
		assert.Equal(t, 1, calls)
	})

	t.Run("a collision without a container is retried", func(t *testing.T) {
		calls := 0
		second := &fakeContainer{}

		got, err := testport.Run(t.Context(), "5432/tcp", func(context.Context, testcontainers.ContainerCustomizer) (*fakeContainer, error) {
			calls++
			if calls == 1 {
				return nil, errors.New("port is already allocated")
			}

			return second, nil
		})
		require.NoError(t, err)

		assert.Equal(t, 2, calls)
		assert.Same(t, second, got)
	})

	t.Run("a failed termination stops the retry", func(t *testing.T) {
		calls := 0
		collision := errors.New("port is already allocated")
		terminateErr := errors.New("container is stuck")

		_, err := testport.Run(t.Context(), "5432/tcp", func(context.Context, testcontainers.ContainerCustomizer) (*fakeContainer, error) {
			calls++

			return &fakeContainer{terminateErr: terminateErr}, collision
		})

		require.ErrorIs(t, err, testport.ErrTerminate)
		require.ErrorIs(t, err, terminateErr)
		require.ErrorIs(t, err, collision)
		assert.Equal(t, 1, calls)
	})

	t.Run("repeated collisions stop at the attempt limit", func(t *testing.T) {
		var containers []*fakeContainer
		var failures []error

		_, err := testport.Run(t.Context(), "5432/tcp", func(context.Context, testcontainers.ContainerCustomizer) (*fakeContainer, error) {
			c := &fakeContainer{}
			containers = append(containers, c)
			failures = append(failures, fmt.Errorf("attempt %d: listen tcp4 0.0.0.0:1: bind: address already in use", len(containers)))

			return c, failures[len(failures)-1]
		})

		require.Len(t, containers, testport.Attempts)
		require.ErrorIs(t, err, failures[len(failures)-1])
		for _, c := range containers {
			assert.True(t, c.terminated)
		}
	})
}

func TestFree(t *testing.T) {
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
			path := filepath.Join(t.TempDir(), "ip_local_port_range")
			if tc.portRange != nil {
				require.NoError(t, os.WriteFile(path, []byte(*tc.portRange), 0o600))
			}

			t.Cleanup(testport.SetEphemeralPortRangePath(path))

			require.Equal(t, tc.start, testport.EphemeralPortStart())

			seen := map[string]bool{}

			for range 20 {
				port, err := testport.Free(t.Context())
				require.NoError(t, err)

				number, err := strconv.Atoi(port)
				require.NoError(t, err)

				assert.GreaterOrEqual(t, number, 10000)
				assert.Less(t, number, tc.start)
				assert.False(t, seen[port], "port %s handed out twice", port)

				seen[port] = true
			}
		})
	}
}

func TestFreeErrors(t *testing.T) {
	t.Run("returns ErrNoFreePort once the range is exhausted", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "ip_local_port_range")
		require.NoError(t, os.WriteFile(path, []byte("10001\t60999\n"), 0o600))
		t.Cleanup(testport.SetEphemeralPortRangePath(path))

		_, err := testport.Free(t.Context())
		if err == nil {
			_, err = testport.Free(t.Context())
		}

		require.ErrorIs(t, err, testport.ErrNoFreePort)
	})

	t.Run("returns the context error when ctx is done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		_, err := testport.Free(ctx)
		require.ErrorIs(t, err, context.Canceled)
	})
}
