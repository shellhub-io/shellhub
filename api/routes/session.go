package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	GetSessionsURL      = "/sessions"
	GetSessionURL       = "/sessions/:uid"
	UpdateSessionURL    = "/sessions/:uid"
	CreateSessionURL    = "/sessions"
	FinishSessionURL    = "/sessions/:uid/finish"
	KeepAliveSessionURL = "/sessions/:uid/keepalive"
	EventsSessionsURL   = "/sessions/:uid/events"
)

const (
	ParamSessionID = "uid"
)

func (h *Handler) GetSessionList(c gateway.Context) error {
	paginator := query.NewPaginator()
	if err := c.Bind(paginator); err != nil {
		return err
	}

	// TODO: normalize is not required when request is privileged
	paginator.Normalize()

	sessions, count, err := h.service.ListSessions(c.Ctx(), *paginator)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, sessions)
}

func (h *Handler) GetSession(c gateway.Context) error {
	var req requests.SessionGet
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	session, err := h.service.GetSession(c.Ctx(), models.UID(req.UID))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, session)
}

func (h *Handler) UpdateSession(c gateway.Context) error {
	var req requests.SessionUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	return h.service.UpdateSession(c.Ctx(), models.UID(req.UID), models.SessionUpdate{
		Recorded:      req.Recorded,
		Authenticated: req.Authenticated,
		Type:          req.Type,
	})
}

func (h *Handler) CreateSession(c gateway.Context) error {
	var req requests.SessionCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	session, err := h.service.CreateSession(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, session)
}

func (h *Handler) FinishSession(c gateway.Context) error {
	var req requests.SessionFinish
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	return h.service.DeactivateSession(c.Ctx(), models.UID(req.UID))
}

func (h *Handler) KeepAliveSession(c gateway.Context) error {
	var req requests.SessionKeepAlive
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	return h.service.KeepAliveSession(c.Ctx(), models.UID(req.UID))
}

func (h *Handler) EventSession(c gateway.Context) error {
	var req requests.SessionIDParam

	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if !c.IsWebSocket() {
		return c.NoContent(http.StatusBadRequest)
	}

	connection, err := h.WebSocketUpgrader.Upgrade(c.Response(), c.Request())
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	defer connection.Close()

	var r requests.SessionEvent
	for {
		if err := connection.ReadJSON(&r); err != nil {
			if websocket.IsErrorCloseNormal(err) || websocket.IsUnexpectedCloseError(err) {
				log.WithError(err).WithFields(log.Fields{
					"uid": req.UID,
				}).Debug("events websocket closed with a ignored error")

				return nil
			}

			return err
		}

		if err := c.Validate(&r); err != nil {
			return err
		}

		if err := h.service.EventSession(c.Ctx(), models.UID(req.UID), &models.SessionEvent{
			Session:   req.UID,
			Type:      models.SessionEventType(r.Type),
			Timestamp: r.Timestamp,
			Data:      r.Data,
			Seat:      r.Seat,
		}); err != nil {
			return err
		}
	}
}
