package websocket

import (
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Conn represents a WebSocket connection.
type Conn interface {
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	UnderlyingConn() net.Conn

	Subprotocol() string

	NextWriter(messageType int) (io.WriteCloser, error)
	WriteMessage(messageType int, data []byte) error
	WriteControl(messageType int, data []byte, deadline time.Time) error
	WritePreparedMessage(pm *websocket.PreparedMessage) error
	SetWriteDeadline(t time.Time) error
	EnableWriteCompression(enable bool)
	SetCompressionLevel(level int) error

	NextReader() (messageType int, r io.Reader, err error)
	ReadMessage() (messageType int, p []byte, err error)
	SetReadDeadline(t time.Time) error
	SetReadLimit(limit int64)

	SetCloseHandler(h func(code int, text string) error)
	CloseHandler() func(code int, text string) error
	SetPingHandler(h func(appData string) error)
	PingHandler() func(appData string) error
	SetPongHandler(h func(appData string) error)
	PongHandler() func(appData string) error

	ReadJSON(any) error
}

// Upgrader should be implemented by structures that want to be able to convert an HTTP request into WebSocket connection.
type Upgrader interface {
	Upgrade(res http.ResponseWriter, req *http.Request) (Conn, error)
}

// GorillaUpgrader is the gorilla/websocket implementation of [Upgrader]. The embedded upgrader is
// unexported, so its settings are fixed at construction: build one with
// [NewGorillaWebSocketUpgrader] and reach it through the interface.
type GorillaUpgrader struct {
	upgrader *websocket.Upgrader
}

// Upgrade turns an HTTP request into a websocket connection. It writes the failure response itself
// when the handshake fails, so a caller that gets an error must not write one of its own.
func (u *GorillaUpgrader) Upgrade(res http.ResponseWriter, req *http.Request) (Conn, error) {
	return u.upgrader.Upgrade(res, req, nil)
}

// NewGorillaWebSocketUpgrader returns an upgrader with gorilla's defaults. Leaving CheckOrigin nil
// selects gorilla's same-origin check: a request carrying no Origin header is accepted, and one
// carrying it is accepted only when its host equals the request's Host. A browser page on another
// origin is therefore rejected with 403, while a non-browser client is not.
func NewGorillaWebSocketUpgrader() Upgrader {
	return &GorillaUpgrader{
		upgrader: new(websocket.Upgrader),
	}
}
