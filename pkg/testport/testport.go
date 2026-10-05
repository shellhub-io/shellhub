//go:build linux

// Package testport publishes test containers on host ports below the kernel's ephemeral range, so
// that rootless Docker under pasta, which forwards only non-ephemeral ports, can reach them. It
// builds on Linux only, because the reservation depends on how Linux shares a bound port.
//
// [Reserve] picks such a port and holds it, so that no other test process picks it too. [Bind]
// publishes a container on one.
package testport

import (
	"context"
	"crypto/rand"
	"math/big"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
)

// ErrLayer is the layer that testport errors are reported from.
const ErrLayer = "testport"

const (
	// ErrCodeNoFreePort is the code reported when no port below the ephemeral range is free.
	ErrCodeNoFreePort = iota + 1
	// ErrCodeReserve is the code reported when a port cannot be reserved for a reason other than
	// being in use.
	ErrCodeReserve
)

var (
	// ErrNoFreePort is returned by [Reserve] when every port between 10000 and the start of the
	// ephemeral range is in use or reserved.
	ErrNoFreePort = errors.New("no free port below the ephemeral range", ErrLayer, ErrCodeNoFreePort)
	// ErrReserve is returned by [Reserve], joined with the underlying error, when a socket fails
	// for a reason other than the port being in use. The port is its data.
	ErrReserve = errors.New("cannot reserve a port", ErrLayer, ErrCodeReserve)
)

const (
	lowestRandomPort      = 10000
	defaultEphemeralStart = 32768
)

var ephemeralPortRangePath = "/proc/sys/net/ipv4/ip_local_port_range"

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

// Reservation holds a host port until [Reservation.Release]. While held, a bind with
// SO_REUSEADDR that then listens, as the Docker daemon does when it publishes a port, takes the
// port, and a bind without SO_REUSEADDR, as [Reserve] in another process does, fails.
type Reservation struct {
	fd   int
	port string
}

// Port returns the reserved port as a decimal string, the form Docker port bindings and
// connection strings take.
func (r *Reservation) Port() string {
	return r.port
}

// Release frees the port for any process to bind. Calls after the first do nothing and return
// nil.
func (r *Reservation) Release() error {
	if r.fd < 0 {
		return nil
	}

	fd := r.fd
	r.fd = -1

	return syscall.Close(fd)
}

func hold(port int) (int, error) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM|syscall.SOCK_CLOEXEC, 0)
	if err != nil {
		return 0, err
	}

	if err := syscall.Bind(fd, &syscall.SockaddrInet4{Port: port}); err != nil {
		syscall.Close(fd) //nolint:errcheck // the bind error is the one the caller acts on

		return 0, err
	}

	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		syscall.Close(fd) //nolint:errcheck // the setsockopt error is the one the caller acts on

		return 0, err
	}

	return fd, nil
}

// Reserve holds a TCP port between 10000 and the start of the kernel's ephemeral range, or 32768
// when the range cannot be read. It walks the range once from a random port and returns
// [ErrNoFreePort] when every port in it is bound or reserved, and [ErrReserve] when a socket
// fails for another reason. The reservation holds the port in the caller's network namespace,
// which is the host's only when the caller shares it, and child processes do not inherit it.
func Reserve() (*Reservation, error) {
	span := ephemeralPortStart() - lowestRandomPort

	first, _ := rand.Int(rand.Reader, big.NewInt(int64(span)))

	for i := range span {
		port := lowestRandomPort + (int(first.Int64())+i)%span

		fd, err := hold(port)
		if errors.Is(err, syscall.EADDRINUSE) {
			continue
		}

		if err != nil {
			return nil, errors.Wrap(errors.WithData(ErrReserve, port), err)
		}

		return &Reservation{fd: fd, port: strconv.Itoa(port)}, nil
	}

	return nil, ErrNoFreePort
}

// Bind publishes containerPort, such as "5432/tcp", on a port from [Reserve] and releases the
// reservation once the container terminates. A container that fails to create or start is never
// terminated, so its reservation stays held until the process exits. The container must expose
// containerPort, or testcontainers drops the binding. Bind sets a host config modifier, after
// which testcontainers no longer copies the deprecated host fields of ContainerRequest, such as
// Binds and NetworkMode, so set those through their options instead. Read the port back with
// MappedPort.
func Bind(containerPort string) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		port, err := network.ParsePort(containerPort)
		if err != nil {
			return err
		}

		r, err := Reserve()
		if err != nil {
			return err
		}

		if err := testcontainers.WithHostConfigModifier(func(hc *container.HostConfig) {
			if hc.PortBindings == nil {
				hc.PortBindings = network.PortMap{}
			}

			hc.PortBindings[port] = []network.PortBinding{{HostPort: r.Port()}}
		})(req); err != nil {
			return errors.Wrap(err, r.Release())
		}

		return testcontainers.WithAdditionalLifecycleHooks(testcontainers.ContainerLifecycleHooks{
			PostTerminates: []testcontainers.ContainerHook{
				func(context.Context, testcontainers.Container) error { return r.Release() },
			},
		})(req)
	}
}
