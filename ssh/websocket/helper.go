package websocket

import (
	"golang.org/x/net/websocket"
)

// GetFromQuery gets a query value from websocket.
func GetFromQuery(ws *websocket.Conn, query string) string {
	return ws.Request().URL.Query().Get(query)
}

// GetFromHeader gets a header value from a websocket.
func GetFromHeader(ws *websocket.Conn, header string) string {
	return ws.Request().Header.Get(header)
}
