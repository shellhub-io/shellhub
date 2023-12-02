package wsconnadapter

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// an adapter for representing WebSocket connection as a net.Conn
// some caveats apply: https://github.com/gorilla/websocket/issues/441

var ErrUnexpectedMessageType = errors.New("unexpected websocket message type")

const (
	pongTimeout  = time.Second * 35
	pingInterval = time.Second * 30
)

type Adapter struct {
	conn       *websocket.Conn
	readMutex  sync.Mutex
	writeMutex sync.Mutex
	reader     io.Reader
	stopPingCh chan struct{}
	pongCh     chan bool
}

func New(conn *websocket.Conn) *Adapter {
	adapter := &Adapter{
		conn: conn,
	}

	return adapter
}

func (a *Adapter) Ping() chan bool {
	if a.pongCh != nil {
		return a.pongCh
	}

	a.stopPingCh = make(chan struct{})
	a.pongCh = make(chan bool)

	timeout := time.AfterFunc(pongTimeout, func() {
		_ = a.Close()
	})

	a.conn.SetPongHandler(func(data string) error {
		timeout.Reset(pongTimeout)

		// non-blocking channel write
		select {
		case a.pongCh <- true:
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
					logrus.WithError(err).Error("Failed to write ping message")
				}
			case <-a.stopPingCh:
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

	return bytesRead, err
}

func (a *Adapter) Write(b []byte) (int, error) {
	a.writeMutex.Lock()
	defer a.writeMutex.Unlock()

	nextWriter, err := a.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}

	bytesWritten, err := nextWriter.Write(b)
	nextWriter.Close()

	return bytesWritten, err
}

func (a *Adapter) Close() error {
	select {
	case <-a.stopPingCh:
	default:
		a.stopPingCh <- struct{}{}
		close(a.stopPingCh)
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
