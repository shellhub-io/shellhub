package web

import (
	"bufio"
	"net"
	"net/http"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/websocket"
)

func TestConnKeepAlive(t *testing.T) {
	t.Run("an idle terminal stays open while the pings flow", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			browser, frames := openTerminal(t, func(*Conn) {})

			time.Sleep(90 * time.Second)
			synctest.Wait()

			require.NoError(t, websocket.Message.Send(browser, "echo after idle"), "the terminal closed while idle")
			assert.Equal(t, "echo after idle", <-frames, "the terminal closed while idle")

			require.NoError(t, browser.Close())
		})
	})

	t.Run("an idle terminal closes at the deadline once the pings stop", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			browser, frames := openTerminal(t, func(conn *Conn) { conn.Pinger.Stop() })

			time.Sleep(59 * time.Second)
			synctest.Wait()
			assert.True(t, open(frames), "the terminal closed before the deadline")

			time.Sleep(2 * time.Second)
			synctest.Wait()
			assert.False(t, open(frames), "the terminal outlived the deadline with no pings")

			browser.Close() //nolint:errcheck // the server already closed its end
		})
	})
}

func TestConnCloseUnblocksAWriteTheBrowserStoppedReading(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		conns := make(chan *Conn)
		released := make(chan struct{})

		browser := dialTerminal(t, func(socket *websocket.Conn) {
			conns <- NewConn(socket)
			<-released
		})

		conn := <-conns

		stalled := make(chan error)

		go func() {
			_, err := conn.WriteBinary([]byte("output the browser never reads"))
			stalled <- err
		}()

		synctest.Wait()

		conn.Close() //nolint:errcheck // the close frame cannot reach a browser that stopped reading
		require.Error(t, <-stalled, "the stalled write outlived the close")

		close(released)
		browser.Close() //nolint:errcheck // the server already closed its end
	})
}

func openTerminal(t *testing.T, started func(*Conn)) (*websocket.Conn, <-chan string) {
	t.Helper()

	browser := dialTerminal(t, func(socket *websocket.Conn) {
		conn := NewConn(socket)
		defer conn.Close() //nolint:errcheck // the browser end is what the test observes

		go conn.KeepAlive()
		started(conn)

		buffer := make([]byte, 1024)

		for {
			read, err := conn.Read(buffer)
			if err != nil {
				return
			}

			if _, err := conn.Write(buffer[:read]); err != nil {
				return
			}
		}
	})

	frames := make(chan string)

	go func() {
		defer close(frames)

		for {
			var frame string
			if err := websocket.Message.Receive(browser, &frame); err != nil {
				return
			}

			frames <- frame
		}
	}()

	return browser, frames
}

func dialTerminal(t *testing.T, handler websocket.Handler) *websocket.Conn {
	t.Helper()

	browserEnd, serverEnd := net.Pipe()

	go serveTerminal(serverEnd, handler)

	config, err := websocket.NewConfig("ws://shellhub.test/ws/ssh", "http://shellhub.test")
	require.NoError(t, err)

	browser, err := websocket.NewClient(config, browserEnd)
	require.NoError(t, err)

	return browser
}

func serveTerminal(end net.Conn, handler websocket.Handler) {
	defer end.Close() //nolint:errcheck // net.Pipe's Close always returns nil

	reader := bufio.NewReader(end)

	request, err := http.ReadRequest(reader)
	if err != nil {
		return
	}

	handler.ServeHTTP(hijackedResponse{conn: end, buffer: bufio.NewReadWriter(reader, bufio.NewWriter(end))}, request)
}

type hijackedResponse struct {
	http.ResponseWriter

	conn   net.Conn
	buffer *bufio.ReadWriter
}

func (h hijackedResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return h.conn, h.buffer, nil
}

func open(frames <-chan string) bool {
	select {
	case _, received := <-frames:
		return received
	default:
		return true
	}
}
