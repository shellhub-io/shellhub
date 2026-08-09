package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/server/api/services"
	log "github.com/sirupsen/logrus"
)

const (
	GetSessionsURL = "/sessions"
	GetSessionURL  = "/sessions/:uid"
)

const (
	ParamSessionID = "uid"
)

func (h *Handler) GetSessionList(c *gateway.Context) error {
	req := new(requests.ListSessions)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	// TODO: normalize is not required when request is privileged
	req.Paginator.Normalize()

	if err := req.Filters.Unmarshal(); err != nil {
		log.WithError(err).WithField("filter", req.Filters.Raw).Warn("failed to decode session list filter")

		return c.NoContent(http.StatusBadRequest)
	}

	if err := query.ValidateFilters(&req.Filters, services.SessionFilterFields); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	sc, err := c.AdminOrScope()
	if err != nil {
		return err
	}

	sessions, count, err := h.service.ListSessions(c.Ctx(), sc, req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, sessions)
}

func (h *Handler) GetSession(c *gateway.Context) error {
	var req requests.SessionGet
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	sc, err := c.AdminOrScope()
	if err != nil {
		return err
	}

	session, err := h.service.GetSession(c.Ctx(), sc, models.UID(req.UID))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, session)
}
