package wsconnadapter

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestPair creates a connected pair of WebSocket adapters via
// an in-process HTTP server. The caller must close both adapters.
func newTestPair(t *testing.T) (client *Adapter, server *Adapter) {
	t.Helper()

	serverReady := make(chan *Adapter, 1)

	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		serverReady <- New(conn)
	}))
	t.Cleanup(srv.Close)

	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)

	client = New(ws)
	server = <-serverReady

	return client, server
}

func TestConcurrentClose(t *testing.T) {
	client, server := newTestPair(t)
	defer server.Close()

	// Call Ping so stopPingCh is initialised.
	client.Ping()

	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()
			client.Close()
		}()
	}

	wg.Wait()
}

func TestCloseReturnsSameError(t *testing.T) {
	client, server := newTestPair(t)
	defer server.Close()

	err1 := client.Close()
	err2 := client.Close()
	err3 := client.Close()

	assert.Equal(t, err1, err2)
	assert.Equal(t, err2, err3)
}

func TestConcurrentPing(t *testing.T) {
	client, server := newTestPair(t)
	defer server.Close()
	defer client.Close()

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

	// All goroutines must get back the same channel.
	for i := 1; i < goroutines; i++ {
		assert.Equal(t, channels[0], channels[i],
			"Ping() returned different channels on concurrent calls")
	}
}

func TestCloseWithoutPing(t *testing.T) {
	client, server := newTestPair(t)
	defer server.Close()

	// Close without ever calling Ping — stopPingCh is nil.
	err := client.Close()
	assert.NoError(t, err)
}
