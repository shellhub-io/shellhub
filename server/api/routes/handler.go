package routes

import (
	"github.com/shellhub-io/shellhub/pkg/websocket"
	svc "github.com/shellhub-io/shellhub/server/api/services"
)

type Handler struct {
	service svc.Service
	// WebSocketUpgrader is used to turns a HTTP request into WebSocketUpgrader connection.
	WebSocketUpgrader websocket.Upgrader
}

func NewHandler(s svc.Service, w websocket.Upgrader) *Handler {
	return &Handler{
		service:           s,
		WebSocketUpgrader: w,
	}
}
