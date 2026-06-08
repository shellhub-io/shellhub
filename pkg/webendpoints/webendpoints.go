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
