// Package webhook holds what every outbound webhook the server sends shares: the HTTP client that
// refuses to reach the server's own network, the allowlist that opts prefixes back in, and the
// HMAC that lets a receiver tell the server's requests from forged ones.
package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"net/netip"
	"strings"

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

// Sign returns the hex-encoded HMAC-SHA256 of body under secret, the X-ShellHub-Signature an
// enrollment webhook receiver checks.
func Sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)

	return hex.EncodeToString(mac.Sum(nil))
}
