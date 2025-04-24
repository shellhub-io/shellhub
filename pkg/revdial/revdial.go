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
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
	log "github.com/sirupsen/logrus"
)

var (
	ErrDialerClosed   = errors.New("revdial.Dialer closed")
	ErrDialerTimedout = errors.New("revdial.Dialer timedout")
)

// dialerUniqParam is the parameter name of the GET URL form value
// containing the Dialer's random unique ID.
const dialerUniqParam = "revdial.dialer"

// dialerKeepAliveTimeout represents the duration for the keepalive timeout
const dialerKeepAliveTimeout = 35 * time.Second

// The Dialer can create new connections.
type Dialer struct {
	conn       net.Conn // hijacked client conn
	path       string   // e.g. "/revdial"
	uniqID     string
	pickupPath string // path + uniqID: "/revdial?revdial.dialer="+uniqID

	incomingConn chan net.Conn
	pickupFailed chan error
	connReady    chan string
	donec        chan struct{}
	closeOnce    sync.Once
	logger       *log.Entry

	mu sync.Mutex
}

var dialers = sync.Map{}

// NewDialer returns the side of the connection which will initiate
// new connections. This will typically be the side which did the HTTP
// Hijack. The connection is (typically) the hijacked HTTP client
// connection. The connPath is the HTTP path and optional query (but
// without scheme or host) on the dialer where the ConnHandler is
// mounted.
func NewDialer(logger *log.Entry, c net.Conn, connPath string) *Dialer {
	d := &Dialer{
		path:         connPath,
		uniqID:       newUniqID(),
		conn:         c,
		donec:        make(chan struct{}),
		connReady:    make(chan string, 8),
		incomingConn: make(chan net.Conn),
		pickupFailed: make(chan error),
		logger:       logger,
	}

	join := "?"
	if strings.Contains(connPath, "?") {
		join = "&"
	}
	d.pickupPath = connPath + join + dialerUniqParam + "=" + d.uniqID
	d.register()
	go d.serve() // nolint:errcheck

	d.logger.Debug("new dialer connection")

	return d
}

func newUniqID() string {
	buf := make([]byte, 16)
	rand.Read(buf) // nolint:errcheck

	return fmt.Sprintf("%x", buf)
}

func (d *Dialer) register() {
	dialers.Store(d.uniqID, d)
}

func (d *Dialer) unregister() {
	dialers.Delete(d.uniqID)
}

// Done returns a channel which is closed when d is closed (either by
// this process on purpose, by a local error, or close or error from
// the peer).
func (d *Dialer) Done() <-chan struct{} { return d.donec }

// Close closes the Dialer.
func (d *Dialer) Close() error {
	d.closeOnce.Do(d.close)

	return nil
}

func (d *Dialer) close() {
	d.logger.Debug("dialer connection closed")

	d.unregister()
	d.conn.Close()
	d.donec <- struct{}{}
	close(d.donec)
}

func isEqual(c net.Conn, uuid string) bool {
	adapter, ok := c.(*wsconnadapter.Adapter)
	if !ok {
		return false
	}

	return adapter.UUID == uuid
}

// Dial creates a new connection back to the Listener.
func (d *Dialer) Dial(ctx context.Context) (net.Conn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	uuid := uuid.Generate()

	// First, tell serve that we want a connection:
	select {
	case d.connReady <- uuid:
		d.logger.Debug("message true to conn ready channel")
	case <-d.donec:
		d.logger.Debug("dial done")

		return nil, ErrDialerClosed
	case <-ctx.Done():
		d.logger.Debug("dial done due context cancellation")

		return nil, ctx.Err()
	}

	// Then pick it up:
	for {
		select {
		case c := <-d.incomingConn:
			d.logger.Debug("new incoming connection")

			if !isEqual(c, uuid) {
				d.logger.Debug("skipping unmatch connection")

				_ = c.Close()

				continue
			}

			d.logger.Debug("using fresh connection")

			return c, nil
		case err := <-d.pickupFailed:
			d.logger.Debug("failed to pick-up connection")

			return nil, err
		case <-d.donec:
			d.logger.Debug("dial done on pick-up")

			return nil, ErrDialerClosed
		case <-ctx.Done():
			d.logger.Debug("dial done on pick-up due context cancellation")

			return nil, ctx.Err()
		}
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
		defer d.logger.Debug("dialer serve done")

		br := bufio.NewReader(d.conn)
		for {
			line, err := br.ReadSlice('\n')
			if err != nil {
				d.logger.WithError(err).Trace("failed to read the agent's command")

				unexpectedError := websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure)
				if !errors.Is(err, net.ErrClosed) && unexpectedError {
					d.logger.WithError(err).Error("revdial.Dialer failed to read")
				}

				return
			}
			var msg controlMsg
			if err := json.Unmarshal(line, &msg); err != nil {
				d.logger.WithError(err).WithField("line", line).Printf("revdial.Dialer read invalid JSON")

				return
			}

			switch msg.Command {
			case "pickup-failed":
				err := fmt.Errorf("revdial listener failed to pick up connection: %v", msg.Err)
				select {
				case d.pickupFailed <- err:
				case <-d.donec:
					d.logger.WithError(err).Debug("failed to pick-up connection")

					return
				}
			case "keep-alive":
			default:
				// Ignore unknown messages
				log.WithField("message", msg.Command).Debug("unknown message received")
			}
		}
	}()
	for {
		if err := d.sendMessage(controlMsg{Command: "keep-alive"}); err != nil {
			d.logger.WithError(err).Debug("failed to send keep-alive message to device")

			return err
		}

		t := time.NewTimer(30 * time.Second)
		select {
		case <-t.C:
			continue
		case uuid := <-d.connReady:
			t.Stop()
			if err := d.sendMessage(controlMsg{
				Command:  "conn-ready",
				ConnPath: d.pickupPath + fmt.Sprintf("&uuid=%s", uuid),
			}); err != nil {
				d.logger.WithError(err).Debug("failed to send conn-ready message to device")

				return err
			}
		case <-d.donec:
			t.Stop()

			return ErrDialerClosed
		}
	}
}

func (d *Dialer) sendMessage(m controlMsg) error {
	if err := d.conn.SetWriteDeadline(clock.Now().Add(10 * time.Second)); err != nil {
		d.logger.WithError(err).Debug("failed to set the write dead line to device")

		return err
	}

	j, _ := json.Marshal(m)
	j = append(j, '\n')

	if _, err := d.conn.Write(j); err != nil {
		d.logger.WithError(err).Debug("failed to write on the connection")

		return err
	}

	return d.conn.SetWriteDeadline(time.Time{})
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
	done := func() {
		ln.Close()
	}

	var onceDefer sync.Once
	defer onceDefer.Do(done)

	closeTimer := time.AfterFunc(dialerKeepAliveTimeout, done)

	// Write loop
	writec := make(chan []byte, 8)
	ln.writec = writec
	go func() {
		defer onceDefer.Do(done)

		for {
			select {
			case <-ln.donec:
				return
			case msg := <-writec:
				if _, err := ln.sc.Write(msg); err != nil {
					log.Printf("revdial.Listener: error writing message to server: %v", err)

					return
				}
			}
		}
	}()

	go func() {
		defer onceDefer.Do(done)

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
				closeTimer.Reset(dialerKeepAliveTimeout)
			case "conn-ready":
				go ln.grabConn(msg.ConnPath)
			default:
				// Ignore unknown messages
			}
		}
	}()

	for {
		ln.sendMessage(controlMsg{Command: "keep-alive"})

		t := time.NewTimer(30 * time.Second)
		select {
		case <-t.C:
			continue
		case <-ln.donec:
			t.Stop()

			return
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
		uuid := r.FormValue("uuid")

		d, ok := dialers.Load(dialerUniq)
		if !ok {
			http.Error(w, "unknown dialer", http.StatusBadRequest)

			return
		}

		wsConn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		c := wsconnadapter.New(wsConn)
		c.UUID = uuid

		d.(*Dialer).matchConn(c)
	})
}
