// Package publicurl resolves the URL users reach a ShellHub instance at.
package publicurl

import (
	"net/url"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/errors"
)

// ErrLayer is the layer that publicurl errors are reported from.
const ErrLayer = "publicurl"

// ErrCodeInvalidDomain is the code reported when the domain does not make a URL.
const ErrCodeInvalidDomain = iota + 1

// ErrInvalidDomain is returned by [Resolve] when the domain does not make a URL with a host. It is
// joined with the parse error when the domain does not parse.
var ErrInvalidDomain = errors.New("invalid SHELLHUB_DOMAIN", ErrLayer, ErrCodeInvalidDomain)

// Resolve returns the URL users reach the instance at, built from domain, with autoSSL deciding
// the scheme when domain carries none. The URL never ends in a slash, so a path can be appended to
// it. It returns [ErrInvalidDomain] when the result does not parse as a URL with a host.
func Resolve(domain string, autoSSL bool) (*url.URL, error) {
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
