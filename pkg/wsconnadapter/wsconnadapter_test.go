package wsconnadapter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestPair(t *testing.T) (client *Adapter, server *Adapter) {
	t.Helper()

	serverReady := make(chan *Adapter, 1)

	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if !assert.NoError(t, err) {
			return
		}

		serverReady <- New(conn)
	}))
	t.Cleanup(srv.Close)

	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil) //nolint:bodyclose // gorilla/websocket documents that the handshake response body need not be closed.
	require.NoError(t, err)

	client = New(ws)
	server = <-serverReady

	return client, server
}

func TestConcurrentClose(t *testing.T) {
	client, server := newTestPair(t)
	defer server.Close() //nolint:errcheck

	client.Ping()

	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()
			_ = client.Close()
		}()
	}

	wg.Wait()
}

func TestCloseReturnsSameError(t *testing.T) {
	client, server := newTestPair(t)
	defer server.Close() //nolint:errcheck

	err1 := client.Close()
	err2 := client.Close()
	err3 := client.Close()

	assert.Equal(t, err1, err2)
	assert.Equal(t, err2, err3)
}

func TestConcurrentPing(t *testing.T) {
	client, server := newTestPair(t)
	defer server.Close() //nolint:errcheck
	defer client.Close() //nolint:errcheck

	const goroutines = 50

	channels := make([]chan bool, goroutines)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := range goroutines {
		go func() {
			defer wg.Done()
			channels[i] = client.Ping()
		}()
	}

	wg.Wait()

	for i := 1; i < goroutines; i++ {
		assert.Equal(t, channels[0], channels[i],
			"Ping() returned different channels on concurrent calls")
	}
}

func TestCloseWithoutPing(t *testing.T) {
	client, server := newTestPair(t)
	defer server.Close() //nolint:errcheck

	err := client.Close()
	assert.NoError(t, err)
}

// TestPingFailureClosesConnection verifies the keep-alive contract: when a ping
// write fails the loop closes the adapter and stops, so a dead connection does
// not stay registered.
func TestPingFailureClosesConnection(t *testing.T) {
	client, server := newTestPair(t)

	client.pingInterval = 5 * time.Millisecond
	client.pongTimeout = time.Hour // keep the pong timeout out of the picture

	client.Ping()

	_ = server.Close()

	require.Eventually(t, func() bool {
		select {
		case <-client.stopPingCh:
			return true
		default:
			return false
		}
	}, 2*time.Second, 5*time.Millisecond, "ping failure did not tear down the connection")

	assert.NotPanics(t, func() { _ = client.Close() })
}
