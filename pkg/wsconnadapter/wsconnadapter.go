package wsconnadapter

import (
	"errors"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
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
	conn       *websocket.Conn
	readMutex  sync.Mutex
	writeMutex sync.Mutex
	reader     io.Reader
	stopPingCh chan struct{}
	pongCh     chan bool
	Logger     *log.Entry
}

func (a *Adapter) WithID(requestID string) *Adapter {
	a.Logger.Info("Setting request ID for WebSocket adapter", log.Fields{
		"request-id": requestID,
	})

	a.Logger = a.Logger.WithFields(log.Fields{
		"request-id": requestID,
	})

	return a
}

func (a *Adapter) WithDevice(tenant string, device string) *Adapter {
	a.Logger.Info("Setting tenant and device for WebSocket adapter", log.Fields{
		"tenant": tenant,
		"device": device,
	})

	a.Logger = a.Logger.WithFields(log.Fields{
		"tenant": tenant,
		"device": device,
	})

	return a
}

func New(conn *websocket.Conn) *Adapter {
	logger := log.NewEntry(&log.Logger{
		Out:       os.Stderr,
		Formatter: log.StandardLogger().Formatter,
		Hooks:     log.StandardLogger().Hooks,
		Level:     log.StandardLogger().Level,
	})

	logger.Info("Creating new WebSocket connection adapter", log.Fields{
		"local_addr":  conn.LocalAddr().String(),
		"remote_addr": conn.RemoteAddr().String(),
	})

	adapter := &Adapter{
		conn:   conn,
		Logger: logger,
	}

	return adapter
}

func (a *Adapter) Ping() chan bool {
	a.Logger.Info("Setting up ping/pong mechanism")

	if a.pongCh != nil {
		a.Logger.Debug("Pong channel already exists, reusing existing channel")

		return a.pongCh
	}

	a.stopPingCh = make(chan struct{})
	a.pongCh = make(chan bool)

	a.Logger.Debug("Created ping/pong channels", log.Fields{
		"ping_interval": pingInterval,
		"pong_timeout":  pongTimeout,
	})

	timeout := time.AfterFunc(pongTimeout, func() {
		a.Logger.Warn("Pong timeout reached, closing connection", log.Fields{
			"timeout_duration": pongTimeout,
		})

		if err := a.Close(); err != nil {
			a.Logger.WithError(err).Error("Failed to close connection after pong timeout")
		}
	})

	a.conn.SetPongHandler(func(_ string) error {
		prevTimeout := timeout.Reset(pongTimeout)
		a.Logger.Debug("Pong received, reset timeout", log.Fields{
			"prev_timeout_active": prevTimeout,
			"new_timeout":         pongTimeout,
		})

		// non-blocking channel write
		select {
		case a.pongCh <- true:
			a.Logger.Trace("Successfully wrote true to pong channel")
		default:
			a.Logger.Debug("Pong channel buffer full, skipping notification")
		}

		return nil
	})

	// ping loop
	go func() {
		a.Logger.Info("Starting ping loop goroutine")
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()

		pingCount := 0

		for {
			select {
			case <-ticker.C:
				pingCount++
				a.Logger.Debug("Sending ping message", log.Fields{
					"ping_count": pingCount,
				})

				if err := a.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err != nil {
					a.Logger.WithError(err).Error("Failed to write ping message", log.Fields{
						"ping_count": pingCount,
					})
				} else {
					a.Logger.Debug("Successfully sent ping message", log.Fields{
						"ping_count": pingCount,
					})
				}
			case <-a.stopPingCh:
				a.Logger.Info("Stopping ping loop", log.Fields{
					"total_pings_sent": pingCount,
				})

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
	a.Logger.Trace("Acquiring write mutex")
	a.writeMutex.Lock()
	defer func() {
		a.writeMutex.Unlock()
		a.Logger.Trace("Released write mutex")
	}()

	a.Logger.Debug("Getting next writer for binary message", log.Fields{
		"bytes_to_write": len(b),
	})

	nextWriter, err := a.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		a.Logger.WithError(err).Error("Failed to get the next writer")

		return 0, err
	}

	a.Logger.Trace("Writing bytes to WebSocket connection")
	bytesWritten, err := nextWriter.Write(b)
	if err != nil {
		a.Logger.WithError(err).Error("Error writing to WebSocket connection", log.Fields{
			"bytes_written":  bytesWritten,
			"bytes_to_write": len(b),
		})
	}

	closeErr := nextWriter.Close()
	if closeErr != nil {
		a.Logger.WithError(closeErr).Error("Error closing WebSocket writer")
		// If we already have an error, keep it, otherwise return the close error
		if err == nil {
			err = closeErr
		}
	}

	a.Logger.WithError(err).
		WithField("bytes", bytesWritten).
		Debug("Completed write operation")

	return bytesWritten, err
}

func (a *Adapter) Close() error {
	a.Logger.Info("Closing WebSocket connection adapter")

	if a.stopPingCh != nil {
		select {
		case <-a.stopPingCh:
			a.Logger.Debug("Stop ping channel already closed")
		default:
			a.Logger.Debug("Sending stop signal to ping goroutine")
			a.stopPingCh <- struct{}{}
			a.Logger.Debug("Closing stop ping channel")
			close(a.stopPingCh)
		}
	} else {
		a.Logger.Debug("No ping loop to stop")
	}

	a.Logger.Debug("Closing underlying WebSocket connection")
	err := a.conn.Close()
	if err != nil {
		a.Logger.WithError(err).Error("Error closing WebSocket connection")
	} else {
		a.Logger.Info("WebSocket connection closed successfully")
	}

	return err
}

func (a *Adapter) LocalAddr() net.Addr {
	addr := a.conn.LocalAddr()
	a.Logger.Trace("Local address requested", log.Fields{
		"addr":    addr.String(),
		"network": addr.Network(),
	})

	return addr
}

func (a *Adapter) RemoteAddr() net.Addr {
	addr := a.conn.RemoteAddr()
	a.Logger.Trace("Remote address requested", log.Fields{
		"addr":    addr.String(),
		"network": addr.Network(),
	})

	return addr
}

func (a *Adapter) SetDeadline(t time.Time) error {
	a.Logger.Debug("Setting read and write deadlines", log.Fields{
		"deadline": t.String(),
	})

	if err := a.SetReadDeadline(t); err != nil {
		a.Logger.WithError(err).Error("Failed to set read deadline")

		return err
	}

	return a.SetWriteDeadline(t)
}

func (a *Adapter) SetReadDeadline(t time.Time) error {
	a.Logger.Debug("Setting read deadline", log.Fields{
		"deadline": t.String(),
	})

	err := a.conn.SetReadDeadline(t)
	if err != nil {
		a.Logger.WithError(err).Error("Failed to set read deadline")
	}

	return err
}

func (a *Adapter) SetWriteDeadline(t time.Time) error {
	a.Logger.Debug("Setting write deadline", log.Fields{
		"deadline": t.String(),
	})

	a.writeMutex.Lock()
	defer func() {
		a.writeMutex.Unlock()
		a.Logger.Trace("Released write mutex after setting write deadline")
	}()

	err := a.conn.SetWriteDeadline(t)
	if err != nil {
		a.Logger.WithError(err).Error("Failed to set write deadline")
	}

	return err
}
