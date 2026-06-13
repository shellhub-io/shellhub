package share

import (
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
)

// Register wires the shareable-terminal routes into the given Echo router.
//
//   - GET    /ssh/shares                 list active shares for the namespace (user-authenticated)
//   - POST   /ssh/shares                 create a share (agent-authenticated via the gateway)
//   - DELETE /ssh/shares/:token          revoke a share (namespace owner, user-authenticated)
//   - GET    /ssh/shares/:token/stream   producer stream pushed by the agent (authenticated)
//   - GET    /ws/share/:token            public guest viewer (no authentication)
func Register(router *echo.Echo, registry *Registry, cli internalclient.Client) {
	h := &Handlers{registry: registry, cli: cli}

	router.GET("/ssh/shares", h.HandleList)
	router.POST("/ssh/shares", h.HandleCreate)
	router.DELETE("/ssh/shares/:token", h.HandleDelete)
	router.GET("/ssh/shares/:token/stream", h.HandleStream)
	router.GET("/ws/share/:token", h.HandleView)
}
