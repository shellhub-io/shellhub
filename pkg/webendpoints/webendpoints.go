// Package webendpoints provides helpers for composing web-endpoint hostnames.
//
// A web endpoint is a named HTTP tunnel that ShellHub exposes on behalf of a
// device.  Each endpoint is reachable at an address that is formed by joining
// an address token (typically a short random slug) and a domain:
//
//	<address>.<domain>   e.g. abc123.tunnel.example.com
//
// The domain itself can be configured at two levels: a per-namespace
// "preferred" override and a system-wide fallback.  The helpers in this
// package centralise that logic so that every caller (nginx template
// generation, API responses, agent-side URL construction) derives the hostname
// the same way.
//
// Note: the nginx-template duplication is by design — keeping the template and
// the Go logic in sync is easier than introducing an indirection layer.
package webendpoints

import (
	"net"
	"regexp"
	"strings"
)

// Domain returns the effective domain for a web endpoint.
//
// When preferred is non-empty it is returned as-is, giving namespace
// administrators the ability to customise the domain without changing the
// system-wide default.  Otherwise fallback (the system-level default) is
// returned.  Both values may be empty strings; callers must handle that case.
func Domain(preferred, fallback string) string {
	if preferred != "" {
		return preferred
	}

	return fallback
}

// address is the hex-encoded MD5 sum that models.WebEndpoint.GenerateAddress
// produces, and the same shape the edge proxy's server_name matches.
var address = regexp.MustCompile(`^[a-f0-9]{32}$`)

// AddressFromHost is the inverse of Host: it takes the Host of an inbound
// request and reports the address it names, or false when the request is not
// addressed to a web endpoint at all.
//
// Rejecting a host that only looks close enough is the point. Whoever the
// address resolves to gets the request proxied into their device, so a caller
// must be able to tell "this is a web endpoint" apart from "this is the
// console" without consulting the database first.
func AddressFromHost(host, domain string) (string, bool) {
	if domain == "" {
		return "", false
	}

	if hostname, _, err := net.SplitHostPort(host); err == nil {
		host = hostname
	}

	// A Host may arrive as a fully qualified name, with the root label spelled
	// out, and still be the same name.
	host = strings.ToLower(strings.TrimSuffix(host, "."))

	candidate, found := strings.CutSuffix(host, "."+strings.ToLower(domain))
	if !found || !address.MatchString(candidate) {
		return "", false
	}

	return candidate, true
}

// Host builds the full hostname for a web endpoint.
//
// The result is "<address>.<domain>" when domain is non-empty, or just
// "<address>" when domain is empty.  The trailing-dot regression guard
// ensures that an empty domain never produces a string ending in ".".
func Host(address, domain string) string {
	if domain == "" {
		return address
	}

	return address + "." + domain
}
