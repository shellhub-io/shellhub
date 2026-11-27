package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	resty "github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
)

const maxRedirects = 3

var (
	errRedirectOtherHost = errors.New("redirect to another host refused")
	errRedirectDowngrade = errors.New("redirect from a secure to an insecure scheme refused")
	errTooManyRedirects  = errors.New("too many redirects")
)

func isSecureScheme(scheme string) bool {
	return scheme == "https" || scheme == "wss"
}

func checkRedirect(origin, target *url.URL, followed int) error {
	switch {
	case followed >= maxRedirects:
		return fmt.Errorf("%w: stopped after %d", errTooManyRedirects, maxRedirects)
	case !strings.EqualFold(origin.Hostname(), target.Hostname()):
		return fmt.Errorf("%w: %s to %s", errRedirectOtherHost, origin.Hostname(), target.Hostname())
	case isSecureScheme(origin.Scheme) && !isSecureScheme(target.Scheme):
		return fmt.Errorf("%w: %s to %s", errRedirectDowngrade, origin.Scheme, target.Scheme)
	default:
		return nil
	}
}

func sameHostRedirectPolicy() resty.RedirectPolicyFunc {
	return func(req *http.Request, via []*http.Request) error {
		return checkRedirect(via[0].URL, req.URL, len(via)-1)
	}
}

func useWebSocketScheme(u *url.URL) {
	if rest, ok := strings.CutPrefix(u.Scheme, "http"); ok {
		u.Scheme = "ws" + rest
	}
}

// DialContext creates a websocket connection to ShellHub's SSH server, sending header on the handshake.
//
// It follows a [http.StatusTemporaryRedirect] or [http.StatusPermanentRedirect] at most three times, sending header
// to every hop, and only while the target keeps the hostname of address and, when address is secure, a secure
// scheme. It returns an error when address does not unescape or parse, when it refuses a redirect, naming the
// rule that refused it, when a redirect has no Location header, returning [http.ErrNoLocation], and when a Location
// header does not parse. Any other failure returns the dialer's error. The response is nil whenever the error is not.
// ctx bounds every hop.
func DialContext(ctx context.Context, address string, header http.Header) (*websocket.Conn, *http.Response, error) {
	uri, err := url.QueryUnescape(address)
	if err != nil {
		return nil, nil, err
	}

	origin, err := url.Parse(uri)
	if err != nil {
		return nil, nil, err
	}

	useWebSocketScheme(origin)

	target := origin
	for followed := 0; ; followed++ {
		conn, res, err := websocket.DefaultDialer.DialContext(ctx, target.String(), header)
		if err == nil {
			return conn, res, nil
		}

		if res == nil || (res.StatusCode != http.StatusTemporaryRedirect && res.StatusCode != http.StatusPermanentRedirect) {
			return nil, nil, err
		}

		location, err := res.Location()
		if err != nil {
			return nil, nil, err
		}

		useWebSocketScheme(location)
		if err := checkRedirect(origin, location, followed); err != nil {
			return nil, nil, err
		}

		target = location
	}
}
