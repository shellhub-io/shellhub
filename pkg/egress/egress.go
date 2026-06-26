// Package egress dials connection targets through an SSRF guardian so the server
// can't be used as a pivot to reach internal, reserved, or metadata addresses. It
// is the one place this logic lives, shared by every path that opens a connection
// to a user-supplied host: the API connection service, the SSH direct-connect path,
// and the cloud team-connection service. Keeping it here stops the guardian policy
// from drifting between them.
package egress

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strconv"
	"time"

	"code.dny.dev/ssrf"
	"github.com/shellhub-io/shellhub/pkg/envs"
)

// devAllowedV4Prefixes are private ranges let through the guardian on top of the
// public-only default. They are applied ONLY in development, so external connects
// can be exercised against the docker host or a host-machine sshd. In production
// the guardian is public-only with no extra configuration and no code to remove.
var devAllowedV4Prefixes = []netip.Prefix{
	netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.168.0.0/16"),
}

// ErrBlocked means the target isn't a permitted connection endpoint: the guardian
// rejected it (loopback, link-local/metadata, reserved, or a private address that
// isn't allowlisted). Distinct from a host that is simply unreachable.
var ErrBlocked = errors.New("the address is not a permitted connection target")

const dialTimeout = 4 * time.Second

// GuardedDialer returns a dialer whose Control validates the real resolved IP at
// the socket layer (closing the DNS-rebind window a resolve-then-dial would leave)
// and permits only the given port. Private ranges are allowed only in development;
// otherwise the guardian is public-only. Callers needing a different timeout (a
// host-key scan, a full connect) set Timeout on the returned dialer.
func GuardedDialer(port int) *net.Dialer {
	opts := []ssrf.Option{
		ssrf.WithPorts(uint16(port)), //nolint:gosec // port is validated to 1-65535 by the request layer.
	}
	if envs.IsDevelopment() {
		opts = append(opts, ssrf.WithAllowedV4Prefixes(devAllowedV4Prefixes...))
	}

	return &net.Dialer{
		Timeout: dialTimeout,
		Control: ssrf.New(opts...).Safe,
	}
}

// IsBlocked reports whether err is the guardian rejecting the target by policy, as
// opposed to the host being unreachable for another reason.
func IsBlocked(err error) bool {
	return errors.Is(err, ssrf.ErrProhibitedIP) ||
		errors.Is(err, ssrf.ErrProhibitedPort) ||
		errors.Is(err, ssrf.ErrProhibitedNetwork)
}

// Probe dials host:port through the guardian and reports whether it is reachable
// and, separately, whether the guardian blocked it (policy) versus the host being
// down. The dial honors ctx, so it is cancelled when the caller's request ends.
func Probe(ctx context.Context, host string, port int) (reachable, blocked bool) {
	conn, err := GuardedDialer(port).DialContext(ctx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return false, IsBlocked(err)
	}
	conn.Close() //nolint:errcheck

	return true, false
}

// Reachable reports plain reachability, collapsing the blocked distinction (used
// for connection status, where a blocked target reads the same as unreachable).
func Reachable(ctx context.Context, host string, port int) bool {
	reachable, _ := Probe(ctx, host, port)

	return reachable
}
