package client

import (
	"errors"
	"net"
	"net/http"
	"strings"

	resty "github.com/go-resty/resty/v2"
)

func getHostname(host string) (hostname string) {
	if strings.Index(host, ":") > 0 {
		host, _, _ = net.SplitHostPort(host)
	}
	hostname = strings.ToLower(host)

	return
}

func getDomain(host string) string {
	host = getHostname(host)
	ss := strings.Split(host, ".")
	if len(ss) < 3 {
		return host
	}
	ss = ss[1:]

	return strings.Join(ss, ".")
}

// SameDomainRedirectPolicy allows redirect only if the redirected domain
// is the same as original domain, e.g. redirect to "www.imroc.cc" from
// "imroc.cc" is allowed, but redirect to "google.com" is not allowed.
func SameDomainRedirectPolicy() resty.RedirectPolicyFunc {
	return func(req *http.Request, via []*http.Request) error {
		if getDomain(req.URL.Host) != getDomain(via[0].URL.Host) {
			return errors.New("different domain name is not allowed")
		}

		return nil
	}
}
