package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	GetSessionsURL             = "/sessions"
	GetSessionURL              = "/sessions/:uid"
	SetSessionAuthenticatedURL = "/sessions/:uid"
	CreateSessionURL           = "/sessions"
	FinishSessionURL           = "/sessions/:uid/finish"
	RecordSessionURL           = "/sessions/:uid/record"
	PlaySessionURL             = "/sessions/:uid/play"
)

func (h *handler) GetSessionList(c apicontext.Context) error {
	query := paginator.NewQuery()
	if err := c.Bind(query); err != nil {
		return err
	}

	// TODO: normalize is not required when request is privileged
	query.Normalize()

	sessions, count, err := h.service.ListSessions(c.Ctx(), *query)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, sessions)
}

func (h *handler) GetSession(c apicontext.Context) error {
	session, err := h.service.GetSession(c.Ctx(), models.UID(c.Param("uid")))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, session)
}

func (h *handler) SetSessionAuthenticated(c apicontext.Context) error {
	var req struct {
		Authenticated bool `json:"authenticated"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	return h.service.SetSessionAuthenticated(c.Ctx(), models.UID(c.Param("uid")), req.Authenticated)
}

func (h *handler) CreateSession(c apicontext.Context) error {
	session := new(models.Session)

	if err := c.Bind(&session); err != nil {
		return err
	}

	session, err := h.service.CreateSession(c.Ctx(), *session)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, session)
}

func (h *handler) FinishSession(c apicontext.Context) error {
	return h.service.DeactivateSession(c.Ctx(), models.UID(c.Param("uid")))
}

func (h *handler) RecordSession(c apicontext.Context) error {
	return c.JSON(http.StatusOK, nil)
}

func (h *handler) PlaySession(c apicontext.Context) error {
	return c.JSON(http.StatusOK, nil)
}

func (h *handler) DeleteRecordedSession(c apicontext.Context) error {
	return c.JSON(http.StatusOK, nil)
}
