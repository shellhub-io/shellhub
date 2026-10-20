package app

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startHTTPServer(t *testing.T, handler http.Handler, readHeaderTimeout time.Duration) *httptest.Server {
	t.Helper()

	srv := httptest.NewUnstartedServer(handler)
	srv.Config = newHTTPServer(handler, readHeaderTimeout)
	srv.Start()
	t.Cleanup(srv.Close)

	return srv
}

func TestHTTPServerReadHeaderTimeout(t *testing.T) {
	const readHeaderTimeout = 100 * time.Millisecond

	t.Run("closes a connection whose headers never finish", func(t *testing.T) {
		srv := startHTTPServer(t, http.NotFoundHandler(), readHeaderTimeout)

		conn, err := new(net.Dialer).DialContext(t.Context(), "tcp", srv.Listener.Addr().String())
		require.NoError(t, err)
		t.Cleanup(func() { conn.Close() }) //nolint:errcheck // the test is over, so a failed close changes nothing

		_, err = conn.Write([]byte("GET / HTTP/1.1\r\nHost: shellhub\r\n"))
		require.NoError(t, err)

		require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second))) //nolint:forbidigo // a socket deadline, which a clock.Clock cannot drive

		_, err = conn.Read(make([]byte, 1))
		assert.ErrorIs(t, err, io.EOF)
	})

	t.Run("keeps an upgraded WebSocket open past the timeout", func(t *testing.T) {
		upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

		srv := startHTTPServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close() //nolint:errcheck // the peer already left, so a failed close changes nothing

			for {
				kind, message, err := conn.ReadMessage()
				if err != nil {
					return
				}

				if err := conn.WriteMessage(kind, message); err != nil {
					return
				}
			}
		}), readHeaderTimeout)

		conn, res, err := websocket.DefaultDialer.DialContext(t.Context(), "ws"+strings.TrimPrefix(srv.URL, "http"), nil) //nolint:bodyclose // gorilla already consumed the 101 handshake response, which has no body
		require.NoError(t, err)
		require.Equal(t, http.StatusSwitchingProtocols, res.StatusCode)
		t.Cleanup(func() { conn.Close() }) //nolint:errcheck // the test is over, so a failed close changes nothing

		time.Sleep(3 * readHeaderTimeout)

		require.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte("still open")))

		_, message, err := conn.ReadMessage()
		require.NoError(t, err)
		assert.Equal(t, "still open", string(message))
	})
}
