// Package testport publishes test containers on host ports below the kernel's ephemeral range, so
// that rootless Docker under pasta, which forwards only non-ephemeral ports, can reach them.
//
// [Free] picks such a port. [Run] starts a container on one and, when another process takes the
// port between the pick and the start, starts it again on a new one.
package testport

import (
	"context"
	"crypto/rand"
	"math/big"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

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
	// ErrCodeTerminate is the code reported when a container whose port collided cannot be removed.
	ErrCodeTerminate
)

var (
	// ErrNoFreePort is returned by [Free] when every port between 10000 and the start of the
	// ephemeral range is in use or already handed out.
	ErrNoFreePort = errors.New("no free port below the ephemeral range", ErrLayer, ErrCodeNoFreePort)
	// ErrTerminate is returned by [Run], joined with the collision and the termination failure, when
	// it cannot remove a container whose port collided.
	ErrTerminate = errors.New("cannot terminate the container whose host port collided", ErrLayer, ErrCodeTerminate)
)

// Attempts is how many host ports [Run] tries before it returns the last collision.
const Attempts = 5

const (
	lowestRandomPort      = 10000
	defaultEphemeralStart = 32768
)

var (
	ephemeralPortRangePath = "/proc/sys/net/ipv4/ip_local_port_range"
	handedOutMu            sync.Mutex
	handedOut              = map[int]struct{}{}
)

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

func claim(port int) bool {
	handedOutMu.Lock()
	defer handedOutMu.Unlock()

	if _, taken := handedOut[port]; taken {
		return false
	}

	handedOut[port] = struct{}{}

	return true
}

func handedOutAlready(port int) bool {
	handedOutMu.Lock()
	defer handedOutMu.Unlock()

	_, taken := handedOut[port]

	return taken
}

func listenable(ctx context.Context, port int) bool {
	l, err := new(net.ListenConfig).Listen(ctx, "tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return false
	}

	l.Close() //nolint:errcheck // the probe listener is discarded; a failed close changes nothing

	return true
}

// Free returns a TCP port that is free on 127.0.0.1 and lies between 10000 and the start of the
// kernel's ephemeral range, or 32768 when the range cannot be read. It never returns the same port
// twice in one process. The probe answers for the host only when the caller shares the host's
// network namespace, and another process may take the port before the caller binds it; [Run]
// retries that case. Free returns the context's error when ctx is done, and [ErrNoFreePort] when
// every port in the range was busy or already handed out as it walked past.
func Free(ctx context.Context) (string, error) {
	span := ephemeralPortStart() - lowestRandomPort

	first, err := rand.Int(rand.Reader, big.NewInt(int64(span)))
	if err != nil {
		return "", err
	}

	for i := range span {
		port := lowestRandomPort + (int(first.Int64())+i)%span
		if handedOutAlready(port) {
			continue
		}

		free := listenable(ctx, port)
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		if free && claim(port) {
			return strconv.Itoa(port), nil
		}
	}

	return "", ErrNoFreePort
}

// Terminator is the one method [Run] needs from a container, to remove one whose port collided.
type Terminator interface {
	Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error
}

// Run calls run with a customizer that publishes containerPort, such as "5432/tcp", on a host port
// from [Free]. The container must expose containerPort, or testcontainers drops the binding.
//
// When the daemon refuses the host port because another process took it, Run terminates the
// container run returned and tries again on a new port, up to [Attempts] times, after which it
// returns the last error. When that termination fails, Run stops and returns [ErrTerminate] joined
// with both errors. Any other error from run is returned at once, along with the container run
// returned, which the caller then owns.
func Run[C Terminator](ctx context.Context, containerPort string, run func(ctx context.Context, bind testcontainers.ContainerCustomizer) (C, error)) (C, error) {
	var zero C

	port, err := network.ParsePort(containerPort)
	if err != nil {
		return zero, err
	}

	for attempt := 1; ; attempt++ {
		hostPort, err := Free(ctx)
		if err != nil {
			return zero, err
		}

		c, err := run(ctx, bindTo(port, hostPort))
		if err == nil || !isCollision(err) {
			return c, err
		}

		if !isNil(c) {
			if terminateErr := c.Terminate(ctx); terminateErr != nil {
				return zero, errors.Wrap(err, errors.Wrap(ErrTerminate, terminateErr))
			}
		}

		if attempt == Attempts {
			return zero, err
		}
	}
}

func bindTo(port network.Port, hostPort string) testcontainers.ContainerCustomizer {
	return testcontainers.WithHostConfigModifier(func(hostConfig *container.HostConfig) {
		if hostConfig.PortBindings == nil {
			hostConfig.PortBindings = network.PortMap{}
		}

		hostConfig.PortBindings[port] = []network.PortBinding{{HostPort: hostPort}}
	})
}

func isCollision(err error) bool {
	msg := err.Error()

	return strings.Contains(msg, "port is already allocated") || strings.Contains(msg, "address already in use")
}

func isNil(c any) bool {
	if c == nil {
		return true
	}

	v := reflect.ValueOf(c)

	return v.Kind() == reflect.Pointer && v.IsNil()
}
