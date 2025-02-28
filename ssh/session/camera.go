package session

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// Camera is used to record session's events.
type Camera struct {
	mutex *sync.Mutex
	conn  *websocket.Conn
}

func (c *Camera) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if err := c.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "session record connection done"),
		time.Now().Add(time.Minute),
	); err != nil {
		return err
	}

	return c.conn.Close()
}

func (c *Camera) WriteEvent(event *models.SessionEvent) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.conn.WriteJSON(event)
}

func NewCamera(conn *websocket.Conn) *Camera {
	return &Camera{
		mutex: new(sync.Mutex),
		conn:  conn,
	}
}
