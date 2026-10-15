// Package wsconnadaptertest connects a [wsconnadapter.Adapter] to a WebSocket peer over an
// in-memory pipe, so a test can drive the adapter's keep-alive inside a testing/synctest bubble
// without a socket or the real clock.
package wsconnadaptertest

import (
	"context"
	"net"
	"net/http"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
	"github.com/stretchr/testify/require"
)

// NewPair upgrades one end of a pipe into an [wsconnadapter.Adapter] built with options, the end
// the server holds, and returns it with the raw WebSocket at the other end, the end the agent
// holds. Neither end reads until the caller does, and both are closed when t ends.
func NewPair(t *testing.T, options ...wsconnadapter.Option) (*wsconnadapter.Adapter, *websocket.Conn) {
	t.Helper()

	listener := newPipeListener()
	accepted := make(chan *websocket.Conn, 1)

	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return r.Header.Get("Origin") == "" }}
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { //nolint:gosec // only this test's peer reaches a pipe
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		accepted <- conn
	})}

	served := make(chan struct{})

	go func() {
		defer close(served)

		_ = server.Serve(listener)
	}()

	dialer := websocket.Dialer{NetDialContext: listener.dial}

	peer, _, err := dialer.Dial("ws://pipe/", nil) //nolint:bodyclose // gorilla/websocket documents that the handshake response body need not be closed.
	require.NoError(t, err)

	adapter := wsconnadapter.New(<-accepted, options...)

	t.Cleanup(func() {
		_ = adapter.Close()
		_ = peer.Close()
		_ = server.Close()
		<-served
	})

	return adapter, peer
}

type pipeListener struct {
	conns chan net.Conn
	done  chan struct{}
	once  sync.Once
}

func newPipeListener() *pipeListener {
	return &pipeListener{conns: make(chan net.Conn), done: make(chan struct{})}
}

func (l *pipeListener) dial(ctx context.Context, _, _ string) (net.Conn, error) {
	client, server := net.Pipe()

	select {
	case l.conns <- server:
		return client, nil
	case <-l.done:
		return nil, net.ErrClosed
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (l *pipeListener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.conns:
		return conn, nil
	case <-l.done:
		return nil, net.ErrClosed
	}
}

func (l *pipeListener) Close() error {
	l.once.Do(func() { close(l.done) })

	return nil
}

func (l *pipeListener) Addr() net.Addr {
	return pipeAddr{}
}

type pipeAddr struct{}

func (pipeAddr) Network() string { return "pipe" }

func (pipeAddr) String() string { return "pipe" }
