// Package origin resolves how a client reached the server: the address it connected from, and the
// host, scheme and port it addressed. Only a peer the operator lists in SHELLHUB_PROXY_TRUSTED_IPS,
// or loopback, may speak for a client through X-Forwarded-*; every other peer is taken to be the
// client, whatever its headers say. The client is the right-most X-Forwarded-For hop outside that
// list, the rule the gateway gives Caddy, so a client cannot choose its address by prepending hops.
// X-Real-IP names no one. It is the one place the server reads those headers.
package origin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"slices"
	"strconv"
	"strings"
)

// ErrInvalidTrust is returned, wrapped, by [ParsePrefixes] and [ParseTrust] for an entry that is
// not a CIDR prefix.
var ErrInvalidTrust = errors.New("invalid SHELLHUB_PROXY_TRUSTED_IPS: want space-separated CIDR prefixes")

// Origin is how a client reached the server.
type Origin struct {
	// IP is the address the client connected from.
	IP string
	// Host is the host the client addressed, with the port when the client named one.
	Host string
	// Scheme is http or https.
	Scheme string
	// Port is the port the client connected to, empty when a proxy speaking for the client did not
	// name it.
	Port string
}

// Trust is the set of peers allowed to speak for a client: loopback, and the prefixes it was
// parsed from.
type Trust struct {
	prefixes []netip.Prefix
}

var loopback = []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8"), netip.MustParsePrefix("::1/128")}

// ParseTrust parses list, CIDR prefixes separated by white space, into the peers allowed to speak
// for a client besides loopback. An empty list trusts loopback alone. It returns an error wrapping
// [ErrInvalidTrust] for an entry that is not a CIDR prefix, a bare address included.
func ParseTrust(list string) (*Trust, error) {
	listed, err := ParsePrefixes(list)
	if err != nil {
		return nil, err
	}

	return &Trust{prefixes: append(append([]netip.Prefix(nil), loopback...), listed...)}, nil
}

// ParsePrefixes parses list, CIDR prefixes separated by white space, into the prefixes [ParseTrust]
// trusts besides loopback: masked, and an IPv4 prefix mapped into IPv6 written as IPv4. An empty
// list returns none. It returns an error wrapping [ErrInvalidTrust] for an entry that is not a
// CIDR prefix, a bare address included.
func ParsePrefixes(list string) ([]netip.Prefix, error) {
	var prefixes []netip.Prefix

	for entry := range strings.FieldsSeq(list) {
		prefix, err := netip.ParsePrefix(entry)
		if err != nil {
			return nil, fmt.Errorf("%w: %q", ErrInvalidTrust, entry)
		}

		if prefix.Addr().Is4In6() && prefix.Bits() >= 96 {
			prefix = netip.PrefixFrom(prefix.Addr().Unmap(), prefix.Bits()-96)
		}

		prefixes = append(prefixes, prefix.Masked())
	}

	return prefixes, nil
}

// Loopback returns the trust of an empty list: loopback alone may speak for a client.
func Loopback() *Trust {
	return &Trust{prefixes: loopback}
}

// Trusts reports whether a peer connecting from addr may speak for a client. An addr that does not
// parse is not trusted.
func (t *Trust) Trusts(addr string) bool {
	ip, err := netip.ParseAddr(addr)
	if err != nil {
		return false
	}

	ip = ip.Unmap()
	for _, prefix := range t.prefixes {
		if prefix.Contains(ip) {
			return true
		}
	}

	return false
}

// Resolve returns how the client behind r reached the server. When the peer is trusted, the client
// is the right-most X-Forwarded-For hop that is not trusted, or the peer when every hop is trusted
// or a hop to the right of that hop does not parse as an IP; X-Real-IP is never read. An
// X-Forwarded-Proto of http or https, an X-Forwarded-Port in 1-65535 and any non-empty
// X-Forwarded-Host, taken as is, each replace what the connection says, and the port is only what
// X-Forwarded-Port names. Otherwise the peer is the client, and the host, scheme and port are the
// request's own: Host, TLS, and the port on Host or else the local port.
func (t *Trust) Resolve(r *http.Request) Origin {
	peer := peerAddress(r.RemoteAddr)
	resolved := Origin{IP: peer, Host: r.Host, Scheme: "http"}
	if r.TLS != nil {
		resolved.Scheme = "https"
	}

	if !t.Trusts(peer) {
		resolved.Port = connectedPort(r)

		return resolved
	}

	if client, ok := t.untrustedHop(r.Header.Values("X-Forwarded-For")); ok {
		resolved.IP = client
	}

	if host := r.Header.Get("X-Forwarded-Host"); host != "" {
		resolved.Host = host
	}

	if scheme := r.Header.Get("X-Forwarded-Proto"); scheme == "http" || scheme == "https" {
		resolved.Scheme = scheme
	}

	if port := r.Header.Get("X-Forwarded-Port"); validPort(port) {
		resolved.Port = port
	}

	return resolved
}

// ClientIP returns the address of the client behind r, as [Trust.Resolve] does. It has the shape
// of echo's IPExtractor.
func (t *Trust) ClientIP(r *http.Request) string {
	return t.Resolve(r).IP
}

type contextKey struct{}

// NewContext returns a copy of ctx carrying o.
func NewContext(ctx context.Context, o Origin) context.Context {
	return context.WithValue(ctx, contextKey{}, o)
}

// FromContext returns the origin [NewContext] stored on ctx, and false when there is none.
func FromContext(ctx context.Context) (Origin, bool) {
	o, ok := ctx.Value(contextKey{}).(Origin)

	return o, ok
}

func (t *Trust) untrustedHop(fields []string) (string, bool) {
	hops := strings.Split(strings.Join(fields, ","), ",")
	for _, hop := range slices.Backward(hops) {
		ip, ok := hopAddress(hop)
		if !ok {
			return "", false
		}

		if !t.Trusts(ip.String()) {
			return ip.String(), true
		}
	}

	return "", false
}

func hopAddress(hop string) (netip.Addr, bool) {
	hop = strings.TrimSpace(hop)
	if host, _, err := net.SplitHostPort(hop); err == nil {
		hop = host
	}

	hop = strings.TrimSuffix(strings.TrimPrefix(hop, "["), "]")
	hop, _, _ = strings.Cut(hop, "%")

	ip, err := netip.ParseAddr(hop)
	if err != nil {
		return netip.Addr{}, false
	}

	return ip.Unmap(), true
}

func peerAddress(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}

	return host
}

func connectedPort(r *http.Request) string {
	if _, port, err := net.SplitHostPort(r.Host); err == nil && validPort(port) {
		return port
	}

	if local, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr); ok {
		if _, port, err := net.SplitHostPort(local.String()); err == nil {
			return port
		}
	}

	return ""
}

func validPort(port string) bool {
	n, err := strconv.Atoi(port)

	return err == nil && n >= 1 && n <= 65535
}
