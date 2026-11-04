package gateway

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTheServerReceivesTheClientAndItsConnection(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ssh/connection" {
			w.Header().Set("X-Seen-Real-IP", r.Header.Get("X-Real-IP"))
			w.Header().Set("X-Seen-Request-ID", r.Header.Get("X-Request-ID"))
			w.Header().Set("X-Seen-Forwarded-For", r.Header.Get("X-Forwarded-For"))
			w.Header().Set("X-Seen-Forwarded-Host", r.Header.Get("X-Forwarded-Host"))
			w.Header().Set("X-Seen-Forwarded-Proto", r.Header.Get("X-Forwarded-Proto"))
			w.WriteHeader(http.StatusNoContent)

			return
		}

		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "not a hijacker", http.StatusInternalServerError)

			return
		}

		conn, buffered, err := hijacker.Hijack()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()

		if _, err := io.WriteString(conn, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n"); err != nil {
			return
		}

		echoed := make([]byte, 4)
		if _, err := io.ReadFull(buffered, echoed); err == nil {
			conn.Write(echoed) //nolint:errcheck // a failed echo fails the client's read
		}
	})

	cfg := configurations()["community"]
	address := startServing(t, cfg, handler)

	t.Run("the client address is the peer's, not the one it claims", func(t *testing.T) {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+address+"/api/info", nil)
		require.NoError(t, err)
		req.Host = cfg.Domain + ":8443"
		req.Header.Set("X-Real-IP", "203.0.113.9")
		req.Header.Set("X-Forwarded-For", "203.0.113.9")
		req.Header.Set("X-Forwarded-Host", "attacker.example")
		req.Header.Set("X-Forwarded-Proto", "https")

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NoError(t, res.Body.Close())

		require.Equal(t, http.StatusNoContent, res.StatusCode)
		assert.Equal(t, "127.0.0.1", res.Header.Get("X-Seen-Real-IP"))
		assert.NotEmpty(t, res.Header.Get("X-Seen-Request-ID"))
		assert.Equal(t, "127.0.0.1", res.Header.Get("X-Seen-Forwarded-For"))
		assert.Equal(t, cfg.Domain+":8443", res.Header.Get("X-Seen-Forwarded-Host"))
		assert.Equal(t, "http", res.Header.Get("X-Seen-Forwarded-Proto"))
	})

	t.Run("a tunnel takes over the connection", func(t *testing.T) {
		conn, err := new(net.Dialer).DialContext(t.Context(), "tcp", address)
		require.NoError(t, err)
		t.Cleanup(func() { _ = conn.Close() })
		require.NoError(t, conn.SetDeadline(clock.New().Now().Add(10*time.Second)))

		_, err = io.WriteString(conn, "GET /ssh/connection HTTP/1.1\r\nHost: "+cfg.Domain+"\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n")
		require.NoError(t, err)

		reader := bufio.NewReader(conn)
		status, err := reader.ReadString('\n')
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(status, "HTTP/1.1 101"), "the server answered %q", status)

		for {
			line, err := reader.ReadString('\n')
			require.NoError(t, err)

			if line == "\r\n" {
				break
			}
		}

		_, err = io.WriteString(conn, "ping")
		require.NoError(t, err)

		echoed := make([]byte, 4)
		_, err = io.ReadFull(reader, echoed)
		require.NoError(t, err)
		assert.Equal(t, "ping", string(echoed))
	})
}

func TestTheServerReceivesTheClientNamedByTheLoadBalancer(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Seen-Real-IP", r.Header.Get("X-Real-IP"))
		w.Header().Set("X-Seen-Remote-Addr", r.RemoteAddr)
		w.WriteHeader(http.StatusNoContent)
	})

	cfg := configurations()["proxy protocol"]
	address := startServing(t, cfg, handler)

	conn, err := new(net.Dialer).DialContext(t.Context(), "tcp", address)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	require.NoError(t, conn.SetDeadline(clock.New().Now().Add(10*time.Second)))

	_, err = io.WriteString(conn, "PROXY TCP4 198.51.100.7 127.0.0.1 40000 80\r\nGET /api/info HTTP/1.1\r\nHost: "+cfg.Domain+"\r\nX-Real-IP: 203.0.113.9\r\nConnection: close\r\n\r\n")
	require.NoError(t, err)

	res, err := http.ReadResponse(bufio.NewReader(conn), nil)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	require.Equal(t, http.StatusNoContent, res.StatusCode)
	assert.Equal(t, "198.51.100.7", res.Header.Get("X-Seen-Real-IP"))
	assert.True(t, strings.HasPrefix(res.Header.Get("X-Seen-Remote-Addr"), "198.51.100.7:"), res.Header.Get("X-Seen-Remote-Addr"))
}

func startServing(t *testing.T, cfg *GatewayConfig, handler http.Handler) string {
	t.Helper()

	cfg.upstreams = taggedUpstreams(t)
	cfg.dataDir = t.TempDir()

	ports := freePorts(t, 2)
	cfg.httpPort, cfg.httpsPort = ports[0], ports[1]

	require.NoError(t, Start(cfg, handler))
	t.Cleanup(func() { _ = Stop() })

	return net.JoinHostPort("127.0.0.1", strconv.Itoa(cfg.httpPort))
}

func freePorts(t *testing.T, count int) []int {
	t.Helper()

	ports := make([]int, 0, count)
	listeners := make([]net.Listener, 0, count)

	for range count {
		listener, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
		require.NoError(t, err)

		listeners = append(listeners, listener)
		ports = append(ports, listener.Addr().(*net.TCPAddr).Port) //nolint:forcetypeassert // a tcp listener's address is a *net.TCPAddr
	}

	for _, listener := range listeners {
		require.NoError(t, listener.Close())
	}

	return ports
}
