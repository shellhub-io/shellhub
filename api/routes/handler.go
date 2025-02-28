package routes

import (
	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/websocket"
)

type Handler struct {
	service svc.Service
	// WebSocketUpgrader is used to turns an HTTP request into WebSocketUpgrader connection.
	WebSocketUpgrader websocket.Upgrader
}

func NewHandler(s svc.Service) *Handler {
	return &Handler{
		service:           s,
		WebSocketUpgrader: websocket.NewGorillaWebSocketUpgrader(),
	}
}
