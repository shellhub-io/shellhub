package session

import (
	"time"

	"github.com/gorilla/websocket"
)

type Camera struct {
	conn *websocket.Conn
}

// Close closes the camera's WebSocket connections normally.
func (c *Camera) Close() error {
	if err := c.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "session record connection done"),
		time.Now().Add(time.Minute),
	); err != nil {
		return err
	}

	return c.conn.Close()
}

// WriteFrame writes a session's frame into the WebSocket connection.
func (c *Camera) WriteFrame(frame any) error {
	return c.conn.WriteJSON(frame)
}

// NewCamera creates a new camera, using a WebSocket connections.
func NewCamera(conn *websocket.Conn) *Camera {
	return &Camera{
		conn: conn,
	}
}
