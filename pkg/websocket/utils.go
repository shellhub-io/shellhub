package websocket

import "github.com/gorilla/websocket"

// IsErrorCloseNormal returns true if it the error received contains a 1000 as its code, as specified by in RFC 6455,
// section 11.7 for a normal close message.
func IsErrorCloseNormal(err error) bool {
	return websocket.IsCloseError(err, websocket.CloseNormalClosure)
}
