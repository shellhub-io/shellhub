package dialer

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newAgentConn returns the server side of a live WebSocket, as the connection
// handler hands it to Bind. The peer is left connected and silent, so the
// session stays up until the test closes it.
func newAgentConn(t *testing.T) *wsconnadapter.Adapter {
	t.Helper()

	ready := make(chan *wsconnadapter.Adapter, 1)

	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)

		ready <- wsconnadapter.New(conn)
	}))
	t.Cleanup(srv.Close)

	peer, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http"), nil)
	require.NoError(t, err)
	t.Cleanup(func() { peer.Close() })

	return <-ready
}

func TestManagerDialFailsWhenTheStreamCannotBeOpened(t *testing.T) {
	server, agent := net.Pipe()

	t.Cleanup(func() {
		server.Close()
		agent.Close()
	})

	session, err := yamux.Client(server, nil)
	require.NoError(t, err)
	require.NoError(t, session.Close())

	m := NewManager()
	m.Connections.Store("tenant:uid", session)

	conn, version, err := m.Dial(context.Background(), "tenant:uid")

	assert.Error(t, err, "a stream that could not be opened must be reported to the caller")
	assert.Nil(t, conn)
	assert.Equal(t, TransportVersionUnknown, version)
}

func TestManagerBindClosesTheConnectionItDisplaced(t *testing.T) {
	key := NewKey("tenant", "uid")

	offline := make(chan string, 1)

	m := NewManager()
	m.DialerDoneCallback = func(key string) { offline <- key }

	require.NoError(t, m.Bind("tenant", "uid", newAgentConn(t)))

	stored, ok := m.Connections.Load(key)
	require.True(t, ok)
	displaced := stored.(*yamux.Session)

	require.NoError(t, m.Bind("tenant", "uid", newAgentConn(t)))

	assert.Eventually(t, displaced.IsClosed, time.Second, 10*time.Millisecond,
		"a reconnect must close the connection it replaced instead of leaving it resident")

	assert.Eventually(t, func() bool { return m.Connections.Size(key) == 1 },
		time.Second, 10*time.Millisecond, "only the live connection may remain registered")

	stored, ok = m.Connections.Load(key)
	require.True(t, ok)
	assert.NotSame(t, displaced, stored.(*yamux.Session), "the live connection is the one that is dialed")

	select {
	case key := <-offline:
		t.Fatalf("device %q was marked offline while a live connection remained", key)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestManagerReportsTheDeviceOfflineWhenItsLastConnectionGoes(t *testing.T) {
	key := NewKey("tenant", "uid")

	offline := make(chan string, 1)

	m := NewManager()
	m.DialerDoneCallback = func(key string) { offline <- key }

	require.NoError(t, m.Bind("tenant", "uid", newAgentConn(t)))

	stored, ok := m.Connections.Load(key)
	require.True(t, ok)
	require.NoError(t, stored.(*yamux.Session).Close())

	select {
	case reported := <-offline:
		assert.Equal(t, key, reported)
	case <-time.After(time.Second):
		t.Fatal("the device was never reported offline")
	}

	assert.Equal(t, 0, m.Connections.Size(key))
}
