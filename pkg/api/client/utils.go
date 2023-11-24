package client

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	resty "github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
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

// DialContext creates a websocket connection to ShellHub's SSH server.
//
// It receivees the endpoint to connect and the necessary headers for authentication on the server. If the server
// redirect the connection with status [http.StatusTemporaryRedirect] or [http.StatusPermanentRedirect], the DialContext
// method will follow. Any other response from the server will result in an error as result of this function.
func DialContext(ctx context.Context, address string, header http.Header) (*websocket.Conn, *http.Response, error) {
	parseToWS := func(uri string) string {
		return regexp.MustCompile(`^http`).ReplaceAllString(uri, "ws")
	}

	uri, err := url.QueryUnescape(address)
	if err != nil {
		return nil, nil, err
	}

	conn, res, err := websocket.DefaultDialer.DialContext(ctx, parseToWS(uri), header)
	if err != nil {
		if res == nil {
			return nil, nil, err
		}

		switch res.StatusCode {
		case http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
			location, err := res.Location()
			if err != nil {
				return nil, nil, err
			}

			return DialContext(ctx, parseToWS(location.String()), header)
		default:
			return nil, nil, err
		}
	}

	return conn, res, nil
}
