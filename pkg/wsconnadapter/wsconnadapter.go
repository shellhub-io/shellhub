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

// an adapter for representing WebSocket connection as a net.Conn
// some caveats apply: https://github.com/gorilla/websocket/issues/441

var ErrUnexpectedMessageType = errors.New("unexpected websocket message type")

const (
	pongTimeout  = time.Second * 35
	pingInterval = time.Second * 30
)

type Adapter struct {
	UUID       string
	conn       *websocket.Conn
	readMutex  sync.Mutex
	writeMutex sync.Mutex
	reader     io.Reader
	stopPingCh chan struct{}
	pongCh     chan bool
	Logger     *log.Entry
	CreatedAt  time.Time
}

type Option func(*Adapter)

func WithID(id string) Option {
	return func(a *Adapter) {
		a.UUID = id
	}
}

func WithDevice(tenant string, device string) Option {
	return func(a *Adapter) {
		a.Logger = a.Logger.WithFields(log.Fields{
			"tenant": tenant,
			"device": device,
		})
	}
}

func New(conn *websocket.Conn, options ...Option) *Adapter {
	adapter := &Adapter{
		conn: conn,
		Logger: log.NewEntry(&log.Logger{
			Out:       os.Stderr,
			Formatter: log.StandardLogger().Formatter,
			Hooks:     log.StandardLogger().Hooks,
			Level:     log.StandardLogger().Level,
		}),
		CreatedAt: clock.Now(),
	}

	for _, option := range options {
		option(adapter)
	}

	return adapter
}

func (a *Adapter) Ping() chan bool {
	if a.pongCh != nil {
		a.Logger.Debug("pong channel is not null")

		return a.pongCh
	}

	a.stopPingCh = make(chan struct{})
	a.pongCh = make(chan bool)

	timeout := time.AfterFunc(pongTimeout, func() {
		a.Logger.Debug("close connection due pong timeout")

		_ = a.Close()
	})

	a.conn.SetPongHandler(func(_ string) error {
		timeout.Reset(pongTimeout)
		a.Logger.Trace("pong timeout")

		// non-blocking channel write
		select {
		case a.pongCh <- true:
			a.Logger.Trace("write true to pong channel")
		default:
		}

		return nil
	})

	// ping loop
	go func() {
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := a.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err != nil {
					a.Logger.WithError(err).Error("failed to write ping message")
				}
			case <-a.stopPingCh:
				a.Logger.Debug("stop ping message received")

				return
			}
		}
	}()

	return a.pongCh
}

func (a *Adapter) Read(b []byte) (int, error) {
	// Read() can be called concurrently, and we mutate some internal state here
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

		// EOF for the current Websocket frame, more will probably come so..
		if errors.Is(err, io.EOF) {
			// .. we must hide this from the caller since our semantics are a
			// stream of bytes across many frames
			err = nil
		}
	}

	a.Logger.WithError(err).
		WithField("bytes", bytesRead).
		Trace("bytes read from wsconnadapter")

	return bytesRead, err
}

func (a *Adapter) Write(b []byte) (int, error) {
	a.writeMutex.Lock()
	defer a.writeMutex.Unlock()

	nextWriter, err := a.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		a.Logger.WithError(err).Trace("failed to get the next writer")

		return 0, err
	}

	bytesWritten, err := nextWriter.Write(b)
	nextWriter.Close()

	a.Logger.WithError(err).
		WithField("bytes", bytesWritten).
		Trace("bytes written from wsconnadapter")

	return bytesWritten, err
}

func (a *Adapter) Close() error {
	select {
	case <-a.stopPingCh:
		a.Logger.Debug("stop ping message received")
	default:
		if a.stopPingCh != nil {
			a.stopPingCh <- struct{}{}
			close(a.stopPingCh)

			a.Logger.Debug("stop ping channel closed")
		}
	}

	return a.conn.Close()
}

func (a *Adapter) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *Adapter) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *Adapter) SetDeadline(t time.Time) error {
	if err := a.SetReadDeadline(t); err != nil {
		a.Logger.WithError(err).Trace("failed to set the deadline")

		return err
	}

	return a.SetWriteDeadline(t)
}

func (a *Adapter) SetReadDeadline(t time.Time) error {
	return a.conn.SetReadDeadline(t)
}

func (a *Adapter) SetWriteDeadline(t time.Time) error {
	a.writeMutex.Lock()
	defer a.writeMutex.Unlock()

	return a.conn.SetWriteDeadline(t)
}
