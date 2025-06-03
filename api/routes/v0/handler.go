package v0

import (
	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/websocket"
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
