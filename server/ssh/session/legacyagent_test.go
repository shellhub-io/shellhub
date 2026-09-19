package session

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

const versionLinePrefix = "SSH-"

type heldConn struct {
	net.Conn

	mu       sync.Mutex
	ready    chan struct{}
	buf      []byte
	wake     chan struct{}
	deadline time.Time
}

func newHeldConn(peer net.Conn) *heldConn {
	return &heldConn{Conn: peer, ready: make(chan struct{}), wake: make(chan struct{}, 1), buf: nil, deadline: time.Time{}} //nolint:exhaustruct // sync.Mutex's zero value is the usable one
}

func (c *heldConn) feed(p []byte) {
	c.mu.Lock()
	c.buf = append(c.buf, p...)
	c.mu.Unlock()

	select {
	case c.wake <- struct{}{}:
	default:
	}
}

func (c *heldConn) release() { close(c.ready) }

func (c *heldConn) SetDeadline(t time.Time) error { return c.SetReadDeadline(t) }

func (c *heldConn) SetReadDeadline(t time.Time) error {
	c.mu.Lock()
	c.deadline = t
	c.mu.Unlock()

	select {
	case c.wake <- struct{}{}:
	default:
	}

	return nil
}

func (c *heldConn) SetWriteDeadline(time.Time) error { return nil }

func (c *heldConn) awaitBuffered(marker string) bool {
	for range 10000 {
		c.mu.Lock()
		got := bytes.Contains(c.buf, []byte(marker))
		c.mu.Unlock()

		if got {
			return true
		}

		time.Sleep(time.Millisecond)
	}

	return false
}

func (c *heldConn) Read(p []byte) (int, error) {
	<-c.ready

	for {
		c.mu.Lock()
		if len(c.buf) > 0 {
			n := copy(p, c.buf)
			c.buf = c.buf[n:]
			c.mu.Unlock()

			return n, nil
		}

		deadline := c.deadline
		c.mu.Unlock()

		var expiry <-chan time.Time

		if !deadline.IsZero() {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				return 0, os.ErrDeadlineExceeded
			}

			expiry = time.After(remaining)
		}

		select {
		case <-c.wake:
		case <-expiry:
		}
	}
}

type oneShotListener struct {
	conn    net.Conn
	once    sync.Once
	closing sync.Once
	done    chan struct{}
}

func (l *oneShotListener) Accept() (net.Conn, error) { //nolint:ireturn // net.Listener returns the interface
	var conn net.Conn

	l.once.Do(func() { conn = l.conn })

	if conn == nil {
		<-l.done

		return nil, io.EOF
	}

	return conn, nil
}

func (l *oneShotListener) Close() error {
	l.closing.Do(func() { close(l.done) })

	return nil
}

func (l *oneShotListener) Addr() net.Addr { return l.conn.LocalAddr() }

type device struct {
	name string
	open func(net.Conn) error
	run  func(conn net.Conn, held chan<- string, handshook chan<- error, stop <-chan struct{})
}

func serveSSH(conn net.Conn, handshook chan<- error) {
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		handshook <- err

		return
	}

	signer, err := gossh.NewSignerFromKey(key)
	if err != nil {
		handshook <- err

		return
	}

	config := &gossh.ServerConfig{NoClientAuth: true} //nolint:exhaustruct // the host key is added below and nobody is authenticated
	config.AddHostKey(signer)

	served, chans, reqs, err := gossh.NewServerConn(conn, config)
	if err != nil {
		handshook <- err

		return
	}

	handshook <- nil

	go gossh.DiscardRequests(reqs)

	go func() {
		for newChannel := range chans {
			_, _, _ = newChannel.Accept() //nolint:dogsled // the test never uses the channel
		}
	}()

	_ = served.Wait()
}

func report(held chan<- string, what []byte) {
	select {
	case held <- string(what):
	default:
	}
}

func runV2(conn net.Conn, held chan<- string, handshook chan<- error, _ <-chan struct{}) {
	decoder := json.NewDecoder(conn)

	var header map[string]string
	if err := decoder.Decode(&header); err != nil {
		report(held, nil)

		return
	}

	kept, _ := io.ReadAll(decoder.Buffered())
	report(held, kept)

	serveSSH(conn, handshook)
}

func runV1(conn net.Conn, held chan<- string, handshook chan<- error, stop <-chan struct{}) {
	type rawConnKey struct{}

	listener := &oneShotListener{conn: conn, done: make(chan struct{})} //nolint:exhaustruct // sync.Once's zero value is the usable one

	srv := &http.Server{ //nolint:exhaustruct,gosec // a test server serving one connection
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			return context.WithValue(ctx, rawConnKey{}, c)
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hijacker, ok := w.(http.Hijacker)
			if !ok {
				return
			}

			hijacked, buffered, err := hijacker.Hijack()
			if err != nil {
				return
			}

			report(held, peek(buffered))

			raw, ok := r.Context().Value(rawConnKey{}).(net.Conn)
			if !ok {
				return
			}

			serveSSH(raw, handshook)

			_ = hijacked.Close()
		}),
	}

	go func() {
		<-stop

		_ = listener.Close()
	}()

	_ = srv.Serve(listener)
}

func peek(buffered *bufio.ReadWriter) []byte {
	kept, err := buffered.Peek(buffered.Reader.Buffered())
	if err != nil {
		return nil
	}

	return kept
}

func writeOpenV2(conn net.Conn) error {
	return json.NewEncoder(conn).Encode(map[string]string{"id": "session-uid"})
}

func writeOpenV1(conn net.Conn) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/ssh/session-uid", nil)
	if err != nil {
		return err
	}

	req.Host = "device"

	return req.Write(conn)
}

func devices() []device {
	return []device{
		{name: "v1", open: writeOpenV1, run: runV1},
		{name: "v2", open: writeOpenV2, run: runV2},
	}
}

func startDevice(t *testing.T, dev device) (net.Conn, *heldConn, chan string, chan error) {
	t.Helper()

	gateway, peer := net.Pipe()

	t.Cleanup(func() {
		_ = gateway.Close()
		_ = peer.Close()
	})

	held := newHeldConn(peer)
	kept := make(chan string, 1)
	handshook := make(chan error, 1)

	go func() {
		for {
			buf := make([]byte, 4096)

			n, err := peer.Read(buf)
			if n > 0 {
				held.feed(buf[:n])
			}

			if err != nil {
				return
			}
		}
	}()

	stop := make(chan struct{})
	t.Cleanup(func() { close(stop) })

	go dev.run(held, kept, handshook, stop)

	return gateway, held, kept, handshook
}

// TestADeviceKeepsAVersionLineWrittenBeforeItReadTheHeader is the defect this works around,
// stated as what the device is left holding rather than as a handshake that does not finish.
// The reader that parses the header reads in blocks, so a version line that arrived with the
// header is parsed into that reader and never reaches the connection the handler serves.
func TestADeviceKeepsAVersionLineWrittenBeforeItReadTheHeader(t *testing.T) {
	for _, dev := range devices() {
		t.Run(dev.name, func(t *testing.T) {
			gateway, held, kept, _ := startDevice(t, dev)

			require.NoError(t, dev.open(gateway))

			go gossh.NewClientConn(gateway, "tcp", &gossh.ClientConfig{ //nolint:errcheck,exhaustruct // the handshake cannot finish here; the device's side is what is asserted
				User:            "test",
				HostKeyCallback: gossh.InsecureIgnoreHostKey(), //nolint:gosec // a test host key is generated per run
			})

			require.True(t, held.awaitBuffered(versionLinePrefix), "the gateway never wrote its version line")
			held.release()

			assert.Contains(t, <-kept, versionLinePrefix,
				"the version line must be the thing the device's header reader swallowed")
		})
	}
}

// TestHearingTheAgentFirstSurvivesSuchADevice is the workaround against that same device on
// both transports: nothing is written until the device has spoken, so its header reader has
// only the header to read and the handshake it serves completes.
func TestHearingTheAgentFirstSurvivesSuchADevice(t *testing.T) {
	for _, dev := range devices() {
		t.Run(dev.name, func(t *testing.T) {
			gateway, held, kept, handshook := startDevice(t, dev)

			require.NoError(t, dev.open(gateway))

			held.release()

			greeted, err := awaitAgentGreeting(gateway)
			require.NoError(t, err, "the device must speak first, so that nothing is written before it reads")

			assert.NotContains(t, <-kept, versionLinePrefix,
				"the device's header reader must have had only the header to read")

			client, chans, reqs, err := gossh.NewClientConn(greeted, "tcp", &gossh.ClientConfig{ //nolint:exhaustruct // a test dial authenticates nobody
				User:            "test",
				HostKeyCallback: gossh.InsecureIgnoreHostKey(), //nolint:gosec // a test host key is generated per run
			})
			require.NoError(t, err)

			go gossh.DiscardRequests(reqs)

			go func() {
				for range chans {
				}
			}()

			require.NoError(t, <-handshook, "the device's handshake must complete once nothing was written before it spoke")
			require.NoError(t, client.Close())
		})
	}
}
