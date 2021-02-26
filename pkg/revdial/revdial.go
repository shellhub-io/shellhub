// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package revdial implements a Dialer and Listener which work together
// to turn an accepted connection (for instance, a Hijacked HTTP request) into
// a Dialer which can then create net.Conns connecting back to the original
// dialer, which then gets a net.Listener accepting those conns.
//
// This is basically a very minimal SOCKS5 client & server.
//
// The motivation is that sometimes you want to run a server on a
// machine deep inside a NAT. Rather than connecting to the machine
// directly (which you can't, because of the NAT), you have the
// sequestered machine connect out to a public machine. Both sides
// then use revdial and the public machine can become a client for the
// NATed machine.
package revdial

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
)

// dialerUniqParam is the parameter name of the GET URL form value
// containing the Dialer's random unique ID.
const dialerUniqParam = "revdial.dialer"

// The Dialer can create new connections.
type Dialer struct {
	conn       net.Conn // hijacked client conn
	path       string   // e.g. "/revdial"
	uniqID     string
	pickupPath string // path + uniqID: "/revdial?revdial.dialer="+uniqID

	incomingConn chan net.Conn
	pickupFailed chan error
	connReady    chan bool
	donec        chan struct{}
	closeOnce    sync.Once
	online       chan bool
}

var (
	dmapMu  sync.Mutex
	dialers = map[string]*Dialer{}
)

// NewDialer returns the side of the connection which will initiate
// new connections. This will typically be the side which did the HTTP
// Hijack. The connection is (typically) the hijacked HTTP client
// connection. The connPath is the HTTP path and optional query (but
// without scheme or host) on the dialer where the ConnHandler is
// mounted.
func NewDialer(c net.Conn, connPath string) *Dialer {
	d := &Dialer{
		path:         connPath,
		uniqID:       newUniqID(),
		conn:         c,
		donec:        make(chan struct{}),
		connReady:    make(chan bool),
		incomingConn: make(chan net.Conn),
		pickupFailed: make(chan error),
		online:       make(chan bool),
	}

	join := "?"
	if strings.Contains(connPath, "?") {
		join = "&"
	}
	d.pickupPath = connPath + join + dialerUniqParam + "=" + d.uniqID
	d.register()
	go d.serve() // nolint:errcheck
	return d
}

func newUniqID() string {
	buf := make([]byte, 16)
	rand.Read(buf) // nolint:errcheck
	return fmt.Sprintf("%x", buf)
}

func (d *Dialer) register() {
	dmapMu.Lock()
	defer dmapMu.Unlock()
	dialers[d.uniqID] = d
}

func (d *Dialer) unregister() {
	dmapMu.Lock()
	defer dmapMu.Unlock()
	delete(dialers, d.uniqID)
}

// Done returns a channel which is closed when d is closed (either by
// this process on purpose, by a local error, or close or error from
// the peer).
func (d *Dialer) Done() <-chan struct{} { return d.donec }

// Close closes the Dialer.
func (d *Dialer) Close() error {
	d.online <- false
	d.closeOnce.Do(d.close)
	return nil
}

func (d *Dialer) close() {
	d.unregister()
	d.conn.Close()
	close(d.donec)
}

// Dial creates a new connection back to the Listener.
func (d *Dialer) Dial(ctx context.Context) (net.Conn, error) {
	// First, tell serve that we want a connection:
	select {
	case d.connReady <- true:
	case <-d.donec:
		return nil, errors.New("revdial.Dialer closed")
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Then pick it up:
	select {
	case c := <-d.incomingConn:
		return c, nil
	case err := <-d.pickupFailed:
		return nil, err
	case <-d.donec:
		return nil, errors.New("revdial.Dialer closed")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (d *Dialer) matchConn(c net.Conn) {
	select {
	case d.incomingConn <- c:
	case <-d.donec:
	}
}

// serve blocks and runs the control message loop, keeping the peer
// alive and notifying the peer when new connections are available.
func (d *Dialer) serve() error {
	defer d.Close()
	go func() {
		defer d.Close()
		br := bufio.NewReader(d.conn)
		for {
			line, err := br.ReadSlice('\n')
			if err != nil {
				return
			}
			var msg controlMsg
			if err := json.Unmarshal(line, &msg); err != nil {
				log.Printf("revdial.Dialer read invalid JSON: %q: %v", line, err)
				return
			}
			switch msg.Command {
			case "pickup-failed":
				err := fmt.Errorf("revdial listener failed to pick up connection: %v", msg.Err)
				select {
				case d.pickupFailed <- err:
				case <-d.donec:
					return
				}
			}
		}
	}()
	for {
		if err := d.sendMessage(controlMsg{Command: "keep-alive"}); err != nil {
			return err
		}

		t := time.NewTimer(30 * time.Second)
		select {
		case <-t.C:
			continue
		case <-d.connReady:
			t.Stop()
			if err := d.sendMessage(controlMsg{
				Command:  "conn-ready",
				ConnPath: d.pickupPath,
			}); err != nil {
				return err
			}
		case <-d.donec:
			t.Stop()
			return errors.New("revdial.Dialer closed")
		}
	}
}

func (d *Dialer) sendMessage(m controlMsg) error {
	if err := d.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return err
	}

	j, _ := json.Marshal(m)
	j = append(j, '\n')

	if _, err := d.conn.Write(j); err != nil {
		return err
	}

	if err := d.conn.SetWriteDeadline(time.Time{}); err != nil {
		return err
	}

	return nil
}

// NewListener returns a new Listener, accepting connections which
// arrive from the provided server connection, which should be after
// any necessary authentication (usually after an HTTP exchange).
//
// The provided dialServer func is responsible for connecting back to
// the server and doing TLS setup.
func NewListener(serverConn net.Conn, dialServer func(context.Context, string) (*websocket.Conn, *http.Response, error)) *Listener {
	ln := &Listener{
		sc:    serverConn,
		dial:  dialServer,
		connc: make(chan net.Conn, 8), // arbitrary
		donec: make(chan struct{}),
	}
	go ln.run()
	return ln
}

var _ net.Listener = (*Listener)(nil)

// Listener is a net.Listener, returning new connections which arrive
// from a corresponding Dialer.
type Listener struct {
	sc     net.Conn
	connc  chan net.Conn
	donec  chan struct{}
	dial   func(context.Context, string) (*websocket.Conn, *http.Response, error)
	writec chan<- []byte

	mu      sync.Mutex // guards below, closing connc, and writing to rw
	readErr error
	closed  bool
}

type controlMsg struct {
	Command  string `json:"command,omitempty"`  // "keep-alive", "conn-ready", "pickup-failed"
	ConnPath string `json:"connPath,omitempty"` // conn pick-up URL path for "conn-url", "pickup-failed"
	Err      string `json:"err,omitempty"`
}

// run reads control messages from the public server forever until the connection dies, which
// then closes the listener.
func (ln *Listener) run() {
	defer ln.Close()

	// Write loop
	writec := make(chan []byte, 8)
	ln.writec = writec
	go func() {
		for {
			select {
			case <-ln.donec:
				return
			case msg := <-writec:
				if _, err := ln.sc.Write(msg); err != nil {
					log.Printf("revdial.Listener: error writing message to server: %v", err)
					ln.Close()
					return
				}
			}
		}
	}()

	// Read loop
	br := bufio.NewReader(ln.sc)
	for {
		line, err := br.ReadSlice('\n')
		if err != nil {
			return
		}
		var msg controlMsg
		if err := json.Unmarshal(line, &msg); err != nil {
			log.Printf("revdial.Listener read invalid JSON: %q: %v", line, err)
			return
		}
		switch msg.Command {
		case "keep-alive":
			// Occasional no-op message from server to keep
			// us alive through NAT timeouts.
		case "conn-ready":
			go ln.grabConn(msg.ConnPath)
		default:
			// Ignore unknown messages
		}
	}
}

func (ln *Listener) sendMessage(m controlMsg) {
	j, _ := json.Marshal(m)
	j = append(j, '\n')
	ln.writec <- j
}

func (ln *Listener) grabConn(path string) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	wsConn, resp, err := ln.dial(ctx, path)
	if err != nil {
		ln.sendMessage(controlMsg{Command: "pickup-failed", ConnPath: path, Err: err.Error()})
		return
	}

	failPickup := func(err error) {
		wsConn.Close()
		log.Printf("revdial.Listener: failed to pick up connection to %s: %v", path, err)
		ln.sendMessage(controlMsg{Command: "pickup-failed", ConnPath: path, Err: err.Error()})
	}

	if resp.StatusCode != 101 {
		failPickup(fmt.Errorf("non-101 response %v", resp.Status))
		return
	}

	select {
	case ln.connc <- wsconnadapter.New(wsConn):
	case <-ln.donec:
	}
}

// Closed reports whether the listener has been closed.
func (ln *Listener) Closed() bool {
	ln.mu.Lock()
	defer ln.mu.Unlock()
	return ln.closed
}

// Accept blocks and returns a new connection, or an error.
func (ln *Listener) Accept() (net.Conn, error) {
	c, ok := <-ln.connc
	if !ok {
		ln.mu.Lock()
		err, closed := ln.readErr, ln.closed
		ln.mu.Unlock()
		if err != nil && !closed {
			return nil, fmt.Errorf("revdial: Listener closed; %v", err)
		}
		return nil, ErrListenerClosed
	}
	return c, nil
}

// ErrListenerClosed is returned by Accept after Close has been called.
var ErrListenerClosed = errors.New("revdial: Listener closed")

// Close closes the Listener, making future Accept calls return an
// error.
func (ln *Listener) Close() error {
	ln.mu.Lock()
	defer ln.mu.Unlock()
	if ln.closed {
		return nil
	}
	go ln.sc.Close()
	ln.closed = true
	close(ln.connc)
	close(ln.donec)
	return nil
}

// Addr returns a dummy address. This exists only to conform to the
// net.Listener interface.
func (ln *Listener) Addr() net.Addr { return fakeAddr{} }

type fakeAddr struct{}

func (fakeAddr) Network() string { return "revdial" }
func (fakeAddr) String() string  { return "revdialconn" }

// ConnHandler returns the HTTP handler that needs to be mounted somewhere
// that the Listeners can dial out and get to. A dialer to connect to it
// is given to NewListener and the path to reach it is given to NewDialer
// to use in messages to the listener.
func ConnHandler(upgrader websocket.Upgrader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dialerUniq := r.FormValue(dialerUniqParam)

		dmapMu.Lock()
		d, ok := dialers[dialerUniq]
		dmapMu.Unlock()
		if !ok {
			http.Error(w, "unknown dialer", http.StatusBadRequest)
			return
		}

		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		d.matchConn(wsconnadapter.New(wsConn))
	})
}

func (d *Dialer) IsOnline() bool {
	return <-d.online
}
