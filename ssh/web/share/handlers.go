package share

import (
	"net/http"
	"sort"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	Subprotocols:    []string{"binary"},
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

// Handlers exposes the HTTP handlers for the shareable terminal feature.
type Handlers struct {
	registry *Registry
	cli      internalclient.Client
}

// HandleCreate registers a new shareable terminal session for an authenticated agent/device.
//
// The gateway authenticates the request and injects the X-Device-UID and X-Tenant-ID headers,
// which this handler trusts (same model as the reverse-tunnel connection handlers).
func (h *Handlers) HandleCreate(c echo.Context) error {
	deviceUID := c.Request().Header.Get("X-Device-UID")
	tenantID := c.Request().Header.Get("X-Tenant-ID")

	if deviceUID == "" || tenantID == "" {
		return c.NoContent(http.StatusUnauthorized)
	}

	var req models.ShareCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	token, e := h.registry.create(deviceUID, tenantID, req)

	if req.Cols > 0 && req.Rows > 0 {
		e.hub.Resize(Dimensions{Cols: req.Cols, Rows: req.Rows})
	}

	url := c.Scheme() + "://" + c.Request().Host + "/share/" + token

	return c.JSON(http.StatusOK, models.ShareCreateResponse{
		Token:     token,
		URL:       url,
		ExpiresAt: e.expiresAt,
	})
}

// HandleList returns the active shares owned by the authenticated namespace, including how many
// guests are currently watching each one. The gateway authenticates the user and injects the
// X-Tenant-ID header.
func (h *Handlers) HandleList(c echo.Context) error {
	tenantID := c.Request().Header.Get("X-Tenant-ID")
	if tenantID == "" {
		return c.NoContent(http.StatusUnauthorized)
	}

	entries := h.registry.list(tenantID)

	shares := make([]models.ShareInfo, 0, len(entries))
	for token, e := range entries {
		info := models.ShareInfo{
			Token:      token,
			URL:        c.Scheme() + "://" + c.Request().Host + "/share/" + token,
			Name:       e.name,
			Command:    e.command,
			Writable:   e.writable,
			DeviceUID:  e.deviceUID,
			DeviceName: e.deviceUID,
			Viewers:    e.hub.Viewers(),
			CreatedAt:  e.createdAt,
			ExpiresAt:  e.expiresAt,
		}

		if device, err := h.cli.GetDevice(c.Request().Context(), e.deviceUID); err == nil && device != nil {
			if device.Name != "" {
				info.DeviceName = device.Name
			}
			info.DeviceOnline = device.Online
			if device.Info != nil {
				info.DeviceOS = device.Info.ID
			}
		}

		shares = append(shares, info)
	}

	sort.Slice(shares, func(i, j int) bool {
		return shares[i].CreatedAt.After(shares[j].CreatedAt)
	})

	return c.JSON(http.StatusOK, shares)
}

// HandleStream binds the producer (the agent) to a share's hub. It reads binary frames as raw
// PTY output and JSON text frames as control events (resize), broadcasting both to guests. When
// the producer disconnects, the share is torn down.
func (h *Handlers) HandleStream(c echo.Context) error {
	token := c.Param("token")

	e, ok := h.registry.get(token)
	if !ok {
		return c.NoContent(http.StatusNotFound)
	}

	// Only the device that created the share may push its stream.
	if uid := c.Request().Header.Get("X-Device-UID"); uid == "" || uid != e.deviceUID {
		return c.NoContent(http.StatusForbidden)
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	defer h.registry.remove(token)

	// In collaborative mode, forward guest keystrokes (drained from the hub) down to the agent as
	// binary frames. This goroutine is the only writer on the producer connection.
	if e.writable {
		go func() {
			for {
				select {
				case <-e.hub.Done():
					return
				case data := <-e.hub.Input():
					if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
						return
					}
				}
			}
		}()
	}

	for {
		typ, data, err := conn.ReadMessage()
		if err != nil {
			return nil
		}

		switch typ {
		case websocket.BinaryMessage:
			e.hub.Output(data)
		case websocket.TextMessage:
			ctrl, err := decodeControl(data)
			if err != nil {
				continue
			}

			if ctrl.Kind == controlKindResize {
				e.hub.Resize(Dimensions{Cols: ctrl.Cols, Rows: ctrl.Rows})
			}
		}
	}
}

// HandleDelete revokes a share, immediately disconnecting all guests. The gateway authenticates the
// namespace owner; only shares belonging to their tenant can be revoked.
func (h *Handlers) HandleDelete(c echo.Context) error {
	tenantID := c.Request().Header.Get("X-Tenant-ID")
	if tenantID == "" {
		return c.NoContent(http.StatusUnauthorized)
	}

	token := c.Param("token")

	e, ok := h.registry.get(token)
	if !ok || e.tenantID != tenantID {
		return c.NoContent(http.StatusNotFound)
	}

	h.registry.remove(token)

	return c.NoContent(http.StatusNoContent)
}

// HandleView serves a public, read-only guest viewer. It subscribes to the share's hub and writes
// every broadcast frame to the websocket. Inbound frames from the guest are discarded (read-only).
func (h *Handlers) HandleView(c echo.Context) error {
	token := c.Param("token")

	e, ok := h.registry.get(token)
	if !ok {
		return c.NoContent(http.StatusNotFound)
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Handshake: tell the guest whether the share is collaborative so it can enable input. This is
	// written before the subscribe loop starts, so it is the only writer at this point.
	if init, err := encodeInit(e.name, e.writable); err == nil {
		_ = conn.WriteMessage(websocket.TextMessage, init)
	}

	sub := e.hub.Subscribe()
	defer e.hub.Unsubscribe(sub)

	// Read frames from the guest: in collaborative mode binary frames are keystrokes forwarded to
	// the host; otherwise everything is discarded (read-only).
	go func() {
		for {
			typ, data, err := conn.ReadMessage()
			if err != nil {
				conn.Close()

				return
			}

			if e.writable && typ == websocket.BinaryMessage {
				e.hub.SendInput(data)
			}
		}
	}()

	for {
		select {
		case <-e.hub.Done():
			return nil
		case msg, open := <-sub.out:
			if !open {
				return nil
			}

			if err := conn.WriteMessage(msg.typ, msg.data); err != nil {
				log.WithError(err).Debug("failed to write share frame to guest")

				return nil
			}
		}
	}
}
