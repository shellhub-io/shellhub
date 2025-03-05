package websocket

import (
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Conn represents a WebSocket connection.
//
//go:generate mockery --name=Conn --filename=conn.go
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
//
//go:generate mockery --name=Upgrader --filename=upgrader.go
type Upgrader interface {
	Upgrade(res http.ResponseWriter, req *http.Request) (Conn, error)
}

// GorillaUpgrader implements [Upgrader] using Gorilla's WebSocket implementation.
type GorillaUpgrader struct {
	upgrader *websocket.Upgrader
}

func (u *GorillaUpgrader) Upgrade(res http.ResponseWriter, req *http.Request) (Conn, error) {
	return u.upgrader.Upgrade(res, req, nil)
}

func NewGorillaWebSocketUpgrader() Upgrader {
	return &GorillaUpgrader{
		upgrader: new(websocket.Upgrader),
	}
}
