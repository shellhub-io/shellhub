// Package webhook holds what every outbound webhook the server sends shares: the HTTP client that
// refuses to reach the server's own network, the allowlist that opts prefixes back in, and the
// HMAC that lets a receiver tell the server's requests from forged ones.
package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"syscall"

	"code.dny.dev/ssrf"
	log "github.com/sirupsen/logrus"
)

// ParseCIDRs parses the comma-separated prefixes a webhook may reach, for NewClient. It logs and
// skips an invalid prefix rather than failing.
func ParseCIDRs(raw string) []netip.Prefix {
	var prefixes []netip.Prefix

	for cidr := range strings.SplitSeq(raw, ",") {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}

		prefix, err := netip.ParsePrefix(cidr)
		if err != nil {
			log.WithError(err).WithField("cidr", cidr).Warn("ignoring invalid webhook allowed CIDR")

			continue
		}

		prefixes = append(prefixes, prefix)
	}

	return prefixes
}

// NewClient returns an HTTP client whose every dial, redirects and DNS rebinding included, refuses
// loopback, private and link-local addresses unless allowed lists the prefix. It sets no timeout:
// the caller bounds each request with its context.
func NewClient(allowed []netip.Prefix) *http.Client {
	opts := []ssrf.Option{ssrf.WithAnyPort()}

	var v4, v6 []netip.Prefix

	for _, prefix := range allowed {
		if prefix.Addr().Is4() {
			v4 = append(v4, prefix)
		} else {
			v6 = append(v6, prefix)
		}
	}

	if len(v4) > 0 {
		opts = append(opts, ssrf.WithAllowedV4Prefixes(v4...))
	}

	if len(v6) > 0 {
		opts = append(opts, ssrf.WithAllowedV6Prefixes(v6...))
	}

	guard := ssrf.New(opts...)

	return &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{Control: guard.Safe}).DialContext,
		},
	}
}

// StatusError is a receiver's answer outside 2xx, carrying its status code.
type StatusError int

func (e StatusError) Error() string {
	return fmt.Sprintf("webhook returned status %d", int(e))
}

// Describe returns why a request from NewClient's client failed, in words fit to show the
// webhook's owner. It names the kind of failure and never the host, address or resolver, so the
// reason tells nobody what the server's network looks like. A StatusError is described as is, and
// a failure it cannot tell apart is "the request failed".
func Describe(err error) string {
	var (
		status StatusError
		dns    *net.DNSError
		cert   *tls.CertificateVerificationError
		netErr net.Error
	)

	switch {
	case errors.As(err, &status):
		return status.Error()
	case errors.Is(err, ssrf.ErrProhibitedIP), errors.Is(err, ssrf.ErrProhibitedNetwork),
		errors.Is(err, ssrf.ErrProhibitedPort), errors.Is(err, ssrf.ErrInvalidHostPort):
		return "the address is not allowed"
	case errors.As(err, &dns):
		return "could not resolve the host"
	case errors.Is(err, context.DeadlineExceeded), errors.As(err, &netErr) && netErr.Timeout():
		return "the request timed out"
	case errors.Is(err, syscall.ECONNREFUSED):
		return "the connection was refused"
	case errors.As(err, &cert):
		return "the certificate could not be verified"
	default:
		return "the request failed"
	}
}

// Sign returns the hex-encoded HMAC-SHA256 of body under secret, the X-ShellHub-Signature an
// enrollment webhook receiver checks.
func Sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)

	return hex.EncodeToString(mac.Sum(nil))
}

// SignV1 returns the X-ShellHub-Signature of an event delivery: "v1=" and the hex-encoded
// HMAC-SHA256 under secret of the Unix timestamp the X-ShellHub-Timestamp header carries, a dot,
// and body. Signing the timestamp lets a receiver refuse a delivery replayed later.
func SignV1(secret string, timestamp int64, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(timestamp, 10) + "."))
	mac.Write(body)

	return "v1=" + hex.EncodeToString(mac.Sum(nil))
}
