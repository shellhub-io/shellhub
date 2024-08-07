package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
)

//go:generate mockery --name=IReverser --filename=reverser.go
type IReverser interface {
	Auth(ctx context.Context, token string, connPath string) error
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
func (r *Reverser) Auth(ctx context.Context, token string, connPath string) error {
	uri, err := url.JoinPath(r.host, connPath)
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
