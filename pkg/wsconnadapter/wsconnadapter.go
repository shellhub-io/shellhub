// Package wsconnadapter presents a gorilla/websocket connection as a net.Conn.
//
// The websocket connection is message-oriented and the net.Conn it stands in for is
// not, so the adapter stitches reads across frames and serialises writes. The caveats
// that shape it are gorilla/websocket#441.
package wsconnadapter

import (
	"errors"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shellhub-io/shellhub/pkg/clock"
	log "github.com/sirupsen/logrus"
)

// ErrUnexpectedMessageType is returned by Read when the peer sends a frame that is not
// a binary one. The adapter carries a byte stream, so any other type is a protocol error
// rather than something to skip.
var ErrUnexpectedMessageType = errors.New("unexpected websocket message type")

const (
	defaultPongTimeout  = time.Second * 35
	defaultPingInterval = time.Second * 30
)

// Adapter is a net.Conn backed by a websocket connection. The zero value is not usable;
// build one with New. Read and Write are each safe for concurrent use, Close is
// idempotent, and the keep-alive loop starts only once Ping is called.
type Adapter struct {
	UUID       string
	conn       *websocket.Conn
	readMutex  sync.Mutex
	writeMutex sync.Mutex
	reader     io.Reader
	stopPingCh chan struct{}
	pongCh     chan bool
	pingOnce   sync.Once
	closeOnce  sync.Once
	closeErr   error
	Logger     *log.Entry
	CreatedAt  time.Time

	// pingInterval and pongTimeout control the keep-alive loop. They default to
	// the package defaults and are only overridden in tests.
	pingInterval time.Duration
	pongTimeout  time.Duration
}

// Option configures an Adapter during New.
type Option func(*Adapter)

// WithID sets the adapter's UUID, which the caller uses to correlate the connection with
// its own bookkeeping. It is not read by the adapter itself.
func WithID(id string) Option {
	return func(a *Adapter) {
		a.UUID = id
	}
}

// WithDevice tags every log line the adapter emits with the tenant and device it belongs
// to, so a connection can be followed through the logs of a busy server.
func WithDevice(tenant string, device string) Option {
	return func(a *Adapter) {
		a.Logger = a.Logger.WithFields(log.Fields{
			"tenant": tenant,
			"device": device,
		})
	}
}

// New wraps conn in an Adapter. The adapter takes ownership of conn: closing the adapter
// closes it, and the caller must not use conn directly afterwards. Keep-alive is off until
// Ping is called.
func New(conn *websocket.Conn, options ...Option) *Adapter {
	adapter := &Adapter{
		conn: conn,
		Logger: log.NewEntry(&log.Logger{
			Out:       os.Stderr,
			Formatter: log.StandardLogger().Formatter,
			Hooks:     log.StandardLogger().Hooks,
			Level:     log.StandardLogger().Level,
		}),
		CreatedAt:    clock.Now(),
		pingInterval: defaultPingInterval,
		pongTimeout:  defaultPongTimeout,
	}

	for _, option := range options {
		option(adapter)
	}

	return adapter
}

// Ping starts the keep-alive loop and returns the channel each pong is announced on. It is
// idempotent: later calls return the same channel without starting a second loop. The
// channel is not buffered and a pong is dropped rather than blocking the read loop, so a
// caller that stops receiving slows nothing down.
//
// A failed ping write is terminal — a broken pipe or a closed socket — so the adapter closes
// itself, which propagates teardown to whoever is reading. Missing pongs for pongTimeout has
// the same effect.
func (a *Adapter) Ping() chan bool {
	a.pingOnce.Do(func() {
		a.stopPingCh = make(chan struct{})
		a.pongCh = make(chan bool)

		timeout := time.AfterFunc(a.pongTimeout, func() {
			a.Logger.Debug("close connection due pong timeout")

			_ = a.Close()
		})

		a.conn.SetPongHandler(func(_ string) error {
			timeout.Reset(a.pongTimeout)
			a.Logger.Trace("pong timeout")

			select {
			case a.pongCh <- true:
				a.Logger.Trace("write true to pong channel")
			default:
			}

			return nil
		})

		go func() {
			ticker := time.NewTicker(a.pingInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if err := a.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err != nil { //nolint:forbidigo // a deadline or an elapsed-time measurement needs the wall clock
						a.Logger.
							WithError(err).
							WithField("lifetime", clock.Now().Sub(a.CreatedAt).String()).
							Warn("reverse connection ping failed, tearing down connection")

						_ = a.Close()

						return
					}
				case <-a.stopPingCh:
					a.Logger.Debug("stop ping message received")

					return
				}
			}
		}()
	})

	return a.pongCh
}

// Read fills b from the current websocket frame and is safe for concurrent use — it holds a
// mutex because it advances the reader the next call resumes from.
//
// A frame boundary is not the end of the stream: the adapter's semantics are a byte stream
// spread over many frames, so an io.EOF from the frame currently being read is reported as a
// nil error and the next call opens the following frame.
func (a *Adapter) Read(b []byte) (int, error) {
	a.readMutex.Lock()
	defer a.readMutex.Unlock()

	if a.reader == nil {
		messageType, reader, err := a.conn.NextReader()
		if err != nil {
			return 0, err
		}

		if messageType != websocket.BinaryMessage {
			return 0, ErrUnexpectedMessageType
		}

		a.reader = reader
	}

	bytesRead, err := a.reader.Read(b)
	if err != nil {
		a.reader = nil

		if errors.Is(err, io.EOF) {
			err = nil
		}
	}

	a.Logger.WithError(err).
		WithField("bytes", bytesRead).
		Trace("bytes read from wsconnadapter")

	return bytesRead, err
}

// Write sends b as a single binary frame and is safe for concurrent use. A short write is
// reported as it happened: the count returned is what the frame took.
func (a *Adapter) Write(b []byte) (int, error) {
	a.writeMutex.Lock()
	defer a.writeMutex.Unlock()

	nextWriter, err := a.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		a.Logger.WithError(err).Trace("failed to get the next writer")

		return 0, err
	}

	bytesWritten, err := nextWriter.Write(b)
	_ = nextWriter.Close()

	a.Logger.WithError(err).
		WithField("bytes", bytesWritten).
		Trace("bytes written from wsconnadapter")

	return bytesWritten, err
}

// Close stops the keep-alive loop and closes the underlying connection. It is idempotent —
// every call after the first returns the error the first one produced.
func (a *Adapter) Close() error {
	a.closeOnce.Do(func() {
		if a.stopPingCh != nil {
			close(a.stopPingCh)
			a.Logger.Debug("stop ping channel closed")
		}

		a.closeErr = a.conn.Close()
	})

	return a.closeErr
}

// LocalAddr returns the local address of the underlying websocket connection.
func (a *Adapter) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

// RemoteAddr returns the peer address of the underlying websocket connection. Behind a proxy
// this is the proxy, not the device.
func (a *Adapter) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

// SetDeadline applies t to both directions, and returns on the first failure — so a failure
// on the read side leaves the write deadline unchanged.
func (a *Adapter) SetDeadline(t time.Time) error {
	if err := a.SetReadDeadline(t); err != nil {
		a.Logger.WithError(err).Trace("failed to set the deadline")

		return err
	}

	return a.SetWriteDeadline(t)
}

// SetReadDeadline applies t to the read side.
func (a *Adapter) SetReadDeadline(t time.Time) error {
	return a.conn.SetReadDeadline(t)
}

// SetWriteDeadline applies t to the write side. It takes the write mutex, so it waits for an
// in-flight Write rather than racing it.
func (a *Adapter) SetWriteDeadline(t time.Time) error {
	a.writeMutex.Lock()
	defer a.writeMutex.Unlock()

	return a.conn.SetWriteDeadline(t)
}
