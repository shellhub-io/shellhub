package revdial_test

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	connPath    = "/ssh/revdial"
	controlPath = "/control"
)

type controlMessage struct {
	Command  string `json:"command,omitempty"`
	ConnPath string `json:"connPath,omitempty"`
	Err      string `json:"err,omitempty"`
}

type agent struct {
	t     *testing.T
	ctl   net.Conn
	wsURL string
	ready chan string
}

type control func(t *testing.T, mux *http.ServeMux, wsURL string) (serverConn net.Conn, agentConn net.Conn)

func websocketControl(t *testing.T, mux *http.ServeMux, wsURL string) (net.Conn, net.Conn) {
	t.Helper()

	accepted := make(chan net.Conn, 1)

	mux.HandleFunc(controlPath, func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("the control connection could not be upgraded: %v", err)

			return
		}

		accepted <- wsconnadapter.New(conn)
	})

	peer, _, err := websocket.DefaultDialer.DialContext(t.Context(), wsURL+controlPath, nil) //nolint:bodyclose // gorilla/websocket documents that the handshake response body need not be closed.
	require.NoError(t, err)

	return <-accepted, wsconnadapter.New(peer)
}

func pipeControl(*testing.T, *http.ServeMux, string) (net.Conn, net.Conn) {
	return net.Pipe()
}

func newDialer(t *testing.T) (*revdial.Dialer, *agent) {
	t.Helper()

	return newDialerOver(t, websocketControl)
}

func newDialerOver(t *testing.T, ctl control) (*revdial.Dialer, *agent) {
	t.Helper()

	mux := http.NewServeMux()
	mux.Handle(connPath, revdial.ConnHandler(websocket.Upgrader{}))

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	serverCtl, agentCtl := ctl(t, mux, wsURL)

	t.Cleanup(func() {
		_ = serverCtl.Close()
		_ = agentCtl.Close()
	})

	logger := log.New()
	logger.SetOutput(io.Discard)

	d := revdial.NewDialer(log.NewEntry(logger), serverCtl, connPath, clock.New())
	t.Cleanup(func() { _ = d.Close() })

	a := &agent{
		t:     t,
		ctl:   agentCtl,
		wsURL: wsURL,
		ready: make(chan string, 1024),
	}

	go forwardControlCommand(a.ctl, "conn-ready", func(msg controlMessage) { a.ready <- msg.ConnPath })

	return d, a
}

func forwardControlCommand(conn net.Conn, command string, deliver func(controlMessage)) {
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return
		}

		var msg controlMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			return
		}

		if msg.Command == command {
			deliver(msg)
		}
	}
}

func (a *agent) awaitConnReady() (string, bool) {
	select {
	case path := <-a.ready:
		return path, true
	case <-time.After(5 * time.Second):
		return "", assert.Fail(a.t, "the server never announced a connection to pick up")
	}
}

func (a *agent) dialBack(ctx context.Context, path string) (net.Conn, bool) {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, a.wsURL+path, nil) //nolint:bodyclose // gorilla/websocket documents that the handshake response body need not be closed.
	if err != nil {
		a.t.Errorf("the agent could not dial back %s: %v", path, err)

		return nil, false
	}

	adapter := wsconnadapter.New(conn)
	a.t.Cleanup(func() { _ = adapter.Close() })

	return adapter, true
}

func (a *agent) failPickup(path string, reason string) bool {
	line, err := json.Marshal(controlMessage{Command: "pickup-failed", ConnPath: path, Err: reason})
	if err == nil {
		_, err = a.ctl.Write(append(line, '\n'))
	}

	if err != nil {
		a.t.Errorf("the agent could not report a failed pick-up: %v", err)

		return false
	}

	return true
}

func dialWithin(t *testing.T, d *revdial.Dialer, timeout time.Duration) (net.Conn, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), timeout)
	defer cancel()

	return d.Dial(ctx)
}

func TestDialServesEveryConcurrentDialItsOwnConnection(t *testing.T) {
	t.Parallel()

	d, a := newDialer(t)

	const dials = 16

	go func() {
		for range dials {
			path, ok := a.awaitConnReady()
			if !ok {
				return
			}

			conn, ok := a.dialBack(t.Context(), path)
			if !ok {
				return
			}

			go func() {
				_, _ = conn.Write([]byte(path))
			}()
		}
	}()

	var wg sync.WaitGroup

	received := make(chan string, dials)
	failures := make(chan error, dials)

	for range dials {
		wg.Go(func() {
			conn, err := dialWithin(t, d, 5*time.Second)
			if err != nil {
				failures <- err

				return
			}

			defer func() { _ = conn.Close() }()

			if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil { //nolint:forbidigo // a deadline needs the wall clock
				failures <- err

				return
			}

			buf := make([]byte, 256)
			n, err := conn.Read(buf)
			if err != nil {
				failures <- err

				return
			}

			received <- string(buf[:n])
		})
	}

	wg.Wait()
	close(failures)
	close(received)

	for err := range failures {
		require.NoError(t, err, "every dial must get a connection or its own error")
	}

	paths := make(map[string]bool)
	for path := range received {
		assert.False(t, paths[path], "two dials received the same connection: %s", path)
		paths[path] = true
	}

	assert.Len(t, paths, dials, "every dial must end up holding a distinct connection")
}

func TestDialReportsItsOwnPickupFailure(t *testing.T) {
	t.Parallel()

	d, a := newDialer(t)

	go func() {
		if path, ok := a.awaitConnReady(); ok {
			a.failPickup(path, "the agent could not reach the server")
		}
	}()

	conn, err := dialWithin(t, d, 5*time.Second)

	require.ErrorContains(t, err, "the agent could not reach the server")
	assert.Nil(t, conn)
}

func TestDialIgnoresAPickupFailureForADialNoLongerWaiting(t *testing.T) {
	t.Parallel()

	d, a := newDialerOver(t, pipeControl)

	gaveUp, cancel := context.WithCancel(t.Context())

	abandoned := make(chan error, 1)
	go func() {
		_, err := d.Dial(gaveUp)
		abandoned <- err
	}()

	stalePath, ok := a.awaitConnReady()
	require.True(t, ok)

	cancel()
	require.ErrorIs(t, <-abandoned, context.Canceled)

	agentDone := make(chan struct{})
	go func() {
		defer close(agentDone)

		path, ok := a.awaitConnReady()
		if !ok || !a.failPickup(stalePath, "stale") {
			return
		}

		a.dialBack(t.Context(), path)
	}()

	conn, err := dialWithin(t, d, 5*time.Second)
	<-agentDone

	require.NoError(t, err, "a failure for a dial that gave up must not fail the one still waiting")
	require.NotNil(t, conn)
	_ = conn.Close()
}

func TestDialClosesTheConnectionOfADialThatGaveUp(t *testing.T) {
	t.Parallel()

	d, a := newDialer(t)

	ctx, cancel := context.WithCancel(t.Context())

	dialed := make(chan error, 1)
	go func() {
		_, err := d.Dial(ctx)
		dialed <- err
	}()

	path, ok := a.awaitConnReady()
	require.True(t, ok)

	cancel()

	select {
	case err := <-dialed:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		require.FailNow(t, "a cancelled dial must return")
	}

	late, ok := a.dialBack(t.Context(), path)
	require.True(t, ok)

	requireClosedByPeer(t, late)
}

func TestCloseReturnsWhenNobodyIsWatchingDone(t *testing.T) {
	t.Parallel()

	serverCtl, agentCtl := net.Pipe()
	t.Cleanup(func() { _ = agentCtl.Close() })

	logger := log.New()
	logger.SetOutput(io.Discard)

	d := revdial.NewDialer(log.NewEntry(logger), serverCtl, connPath, clock.New())

	closed := make(chan struct{})
	go func() {
		_ = d.Close()
		close(closed)
	}()

	select {
	case <-closed:
	case <-time.After(5 * time.Second):
		require.FailNow(t, "Close must return even when nothing is receiving on Done")
	}

	select {
	case <-d.Done():
	default:
		require.FailNow(t, "Done must be closed once Close returned")
	}
}

func TestConnHandlerAnswersANonUpgradePickupWithOnlyTheUpgradersResponse(t *testing.T) {
	t.Parallel()

	d, a := newDialer(t)

	dialed := make(chan error, 1)
	go func() {
		_, err := dialWithin(t, d, 5*time.Second)
		dialed <- err
	}()

	path, ok := a.awaitConnReady()
	require.True(t, ok)

	httpURL := "http" + strings.TrimPrefix(a.wsURL, "ws")
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, httpURL+path, nil)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, "Bad Request\n", string(body))

	require.True(t, a.failPickup(path, "the agent gave up"))
	require.Error(t, <-dialed)
}

type dialFunc func(context.Context, string) (*websocket.Conn, *http.Response, error)

type pickup struct {
	path string
	conn net.Conn
}

type pickupServer struct {
	url      string
	accepted chan pickup
}

func newPickupServer(t *testing.T) *pickupServer {
	t.Helper()

	accepted := make(chan pickup, 64)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("the pick-up could not be upgraded: %v", err)

			return
		}

		adapter := wsconnadapter.New(conn)
		t.Cleanup(func() { _ = adapter.Close() })

		accepted <- pickup{path: r.URL.Path, conn: adapter}
	}))
	t.Cleanup(srv.Close)

	return &pickupServer{url: "ws" + strings.TrimPrefix(srv.URL, "http"), accepted: accepted}
}

func (s *pickupServer) dial(ctx context.Context, path string) (*websocket.Conn, *http.Response, error) {
	return websocket.DefaultDialer.DialContext(ctx, s.url+path, nil)
}

type controlPeer struct {
	t        *testing.T
	ctl      net.Conn
	failures chan controlMessage
}

func newListener(t *testing.T, dial dialFunc) (*revdial.Listener, *controlPeer) {
	t.Helper()

	serverCtl, agentCtl := net.Pipe()
	t.Cleanup(func() {
		_ = serverCtl.Close()
		_ = agentCtl.Close()
	})

	s := &controlPeer{t: t, ctl: serverCtl, failures: make(chan controlMessage, 64)}
	go forwardControlCommand(s.ctl, "pickup-failed", func(msg controlMessage) { s.failures <- msg })

	ln := revdial.NewListener(agentCtl, dial)
	t.Cleanup(func() { _ = ln.Close() })

	return ln, s
}

func (s *controlPeer) announce(path string) {
	line, err := json.Marshal(controlMessage{Command: "conn-ready", ConnPath: path})
	require.NoError(s.t, err)

	_, err = s.ctl.Write(append(line, '\n'))
	require.NoError(s.t, err)
}

func receiveWithin[T any](t *testing.T, ch <-chan T, what string) T {
	t.Helper()

	select {
	case v := <-ch:
		return v
	case <-time.After(5 * time.Second):
		require.FailNow(t, what)

		var zero T

		return zero
	}
}

func requireClosedByPeer(t *testing.T, conn net.Conn) {
	t.Helper()

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second))) //nolint:forbidigo // a deadline needs the wall clock

	_, err := conn.Read(make([]byte, 1))
	require.Error(t, err)

	if netErr, ok := errors.AsType[net.Error](err); ok {
		require.False(t, netErr.Timeout(), "a connection nobody will claim must be closed, not left open")
	}
}

func TestListenerClosesAPickupThatFinishesAfterClose(t *testing.T) {
	t.Parallel()

	for range 50 {
		pickups := newPickupServer(t)

		dialing := make(chan struct{})
		release := make(chan struct{})

		ln, peer := newListener(t, func(ctx context.Context, path string) (*websocket.Conn, *http.Response, error) {
			close(dialing)
			<-release

			return pickups.dial(ctx, path)
		})

		peer.announce("/pickup")
		receiveWithin(t, dialing, "the listener never started the pick-up")

		require.NoError(t, ln.Close())
		close(release)

		requireClosedByPeer(t, receiveWithin(t, pickups.accepted, "the pick-up never reached the server").conn)

		_, err := ln.Accept()
		require.ErrorIs(t, err, revdial.ErrListenerClosed)
	}
}

func overflowAcceptQueue(t *testing.T, pickups *pickupServer, peer *controlPeer) (map[string]net.Conn, controlMessage) {
	t.Helper()

	announced := revdial.AcceptQueueSize + 1

	for i := range announced {
		peer.announce(fmt.Sprintf("/pickup/%d", i))
	}

	byPath := make(map[string]net.Conn, announced)
	for range announced {
		p := receiveWithin(t, pickups.accepted, "a pick-up never reached the server")
		byPath[p.path] = p.conn
	}

	failure := receiveWithin(t, peer.failures, "the pick-up beyond the accept queue was never reported as failed")
	require.Contains(t, byPath, failure.ConnPath)

	return byPath, failure
}

func TestListenerCloseClosesThePickupsWaitingToBeAccepted(t *testing.T) {
	t.Parallel()

	pickups := newPickupServer(t)
	ln, peer := newListener(t, pickups.dial)

	queued, failure := overflowAcceptQueue(t, pickups, peer)
	delete(queued, failure.ConnPath)

	require.NoError(t, ln.Close())

	for _, conn := range queued {
		requireClosedByPeer(t, conn)
	}
}

func TestListenerFailsAPickupThatFindsTheAcceptQueueFull(t *testing.T) {
	t.Parallel()

	pickups := newPickupServer(t)
	_, peer := newListener(t, pickups.dial)

	byPath, failure := overflowAcceptQueue(t, pickups, peer)

	requireClosedByPeer(t, byPath[failure.ConnPath])

	select {
	case extra := <-peer.failures:
		require.FailNow(t, "only the pick-up beyond the accept queue may fail", "also failed: %s", extra.ConnPath)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestListenerLeavesNoGoroutineBehindAfterClose(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		pending := revdial.ControlWriteQueueSize + 4

		dialing := make(chan struct{}, pending)
		release := make(chan struct{})

		ln, peer := newListener(t, func(context.Context, string) (*websocket.Conn, *http.Response, error) {
			dialing <- struct{}{}
			<-release

			return nil, nil, errors.New("the server is unreachable")
		})

		for i := range pending {
			peer.announce(fmt.Sprintf("/pickup/%d", i))
		}

		for range pending {
			receiveWithin(t, dialing, "a pick-up never started")
		}

		require.NoError(t, ln.Close())
		close(release)
	})
}
