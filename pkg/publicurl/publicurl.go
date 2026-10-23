// Package publicurl resolves the URL users reach a ShellHub instance at.
package publicurl

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/errors"
)

// ErrLayer is the layer that publicurl errors are reported from.
const ErrLayer = "publicurl"

const (
	// ErrCodeInvalidDomain is the code reported when the domain does not make a URL.
	ErrCodeInvalidDomain = iota + 1
	// ErrCodeInvalidOverride is the code reported when the override is not an http or https origin.
	ErrCodeInvalidOverride
)

var (
	// ErrInvalidDomain is returned by [Resolve] when the domain does not make a URL with a host. It
	// is joined with the parse error when the domain does not parse.
	ErrInvalidDomain = errors.New("invalid SHELLHUB_DOMAIN", ErrLayer, ErrCodeInvalidDomain)
	// ErrInvalidOverride is returned by [Resolve] when the override is not an http or https URL made
	// of a host and an optional port from 1 to 65535 alone: no userinfo, and nothing past the host
	// but a trailing slash. It is joined with the parse error when the override does not parse.
	ErrInvalidOverride = errors.New("invalid SHELLHUB_PUBLIC_URL: want an http or https URL with only a host and an optional port", ErrLayer, ErrCodeInvalidOverride)
)

// Resolve returns the URL users reach the instance at. A non-empty override, set when TLS ends
// before the gateway, wins. Otherwise the URL is built from domain, with autoSSL deciding the
// scheme when domain carries none. The URL never ends in a slash, so a path can be appended to it.
// It returns [ErrInvalidOverride] or [ErrInvalidDomain] when the one it uses is unusable.
func Resolve(domain string, autoSSL bool, override string) (*url.URL, error) {
	if override != "" {
		return parseOverride(override)
	}

	if !strings.Contains(domain, "://") {
		scheme := "http"
		if autoSSL {
			scheme = "https"
		}

		domain = scheme + "://" + domain
	}

	parsed, err := url.Parse(domain)
	if err != nil {
		return nil, errors.Wrap(ErrInvalidDomain, err)
	}

	if parsed.Hostname() == "" {
		return nil, ErrInvalidDomain
	}

	parsed.Path = strings.TrimSuffix(parsed.Path, "/")

	return parsed, nil
}

func parseOverride(override string) (*url.URL, error) {
	parsed, err := url.Parse(override)
	if err != nil {
		return nil, errors.Wrap(ErrInvalidOverride, err)
	}

	isHTTP := parsed.Scheme == "http" || parsed.Scheme == "https"
	hasOnlyOrigin := parsed.User == nil && (parsed.Path == "" || parsed.Path == "/") && parsed.RawQuery == "" && parsed.Fragment == ""
	if !isHTTP || parsed.Hostname() == "" || !hasValidPort(parsed) || !hasOnlyOrigin {
		return nil, ErrInvalidOverride
	}

	return &url.URL{Scheme: parsed.Scheme, Host: parsed.Host}, nil
}

func hasValidPort(parsed *url.URL) bool {
	if parsed.Port() == "" {
		return !strings.HasSuffix(parsed.Host, ":")
	}

	port, err := strconv.Atoi(parsed.Port())

	return err == nil && port >= 1 && port <= 65535
}
