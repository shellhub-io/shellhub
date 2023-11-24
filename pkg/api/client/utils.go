package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	resty "github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
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

//go:generate mockery --name=IReverser --filename=reverser.go
type IReverser interface {
	Auth(ctx context.Context, token string) error
	NewListener() (*revdial.Listener, error)
}

type Reverser struct {
	conn *websocket.Conn
	// host is the ShellHub's server address.
	//
	// It is used to create the websocket connection to the ShellHub's server.
	host string
}

var _ IReverser = new(Reverser)

func NewReverser(host string) *Reverser {
	return &Reverser{
		host: host,
	}
}

// Auth creates a initial connection to the ShellHub SSH's server and authenticate it with the token received.
func (r *Reverser) Auth(ctx context.Context, token string) error {
	uri, err := url.JoinPath(r.host, "/ssh/connection")
	if err != nil {
		return err
	}

	header := http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", token)},
	}

	conn, _, err := DialContext(ctx, uri, header)
	if err != nil {
		return err
	}

	r.conn = conn

	return nil
}

// NewListener creates a new reverse listener to be used by the Agent to receive connections from the ShellHub's server.
//
// It uses the authenticated connection generate by the [Auth] method to create a new reverse listener. Through this
// connection, the Agent will be able to receive connections from the ShellHub's server. This connections are,
// essentially, the SSH operations requested by the user.
func (r *Reverser) NewListener() (*revdial.Listener, error) {
	if r.conn == nil {
		return nil, errors.New("listener is not authenticated")
	}

	return revdial.NewListener(wsconnadapter.New(r.conn), func(ctx context.Context, path string) (*websocket.Conn, *http.Response, error) {
		uri, err := url.JoinPath(r.host, path)
		if err != nil {
			return nil, nil, err
		}

		return DialContext(ctx, uri, nil)
	}), nil
}
