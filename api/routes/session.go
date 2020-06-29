package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/sessionmngr"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	GetSessionsURL             = "/sessions"
	GetSessionURL              = "/sessions/:uid"
	SetSessionAuthenticatedURL = "/sessions/:uid"
	CreateSessionURL           = "/sessions"
	FinishSessionURL           = "/sessions/:uid/finish"
)

func GetSessionList(c apicontext.Context) error {
	svc := sessionmngr.NewService(c.Store())

	query := paginator.NewQuery()
	c.Bind(query)

	// TODO: normalize is not required when request is privileged
	query.Normalize()

	sessions, count, err := svc.ListSessions(c.Ctx(), *query)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, sessions)
}

func GetSession(c apicontext.Context) error {
	svc := sessionmngr.NewService(c.Store())

	session, err := svc.GetSession(c.Ctx(), models.UID(c.Param("uid")))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, session)
}

func SetSessionAuthenticated(c apicontext.Context) error {
	var req struct {
		Authenticated bool `json:"authenticated"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	svc := sessionmngr.NewService(c.Store())

	return svc.SetSessionAuthenticated(c.Ctx(), models.UID(c.Param("uid")), req.Authenticated)
}

func CreateSession(c apicontext.Context) error {
	session := new(models.Session)

	if err := c.Bind(&session); err != nil {
		return err
	}

	svc := sessionmngr.NewService(c.Store())

	session, err := svc.CreateSession(c.Ctx(), *session)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, session)
}

func FinishSession(c apicontext.Context) error {
	svc := sessionmngr.NewService(c.Store())

	return svc.DeactivateSession(c.Ctx(), models.UID(c.Param("uid")))
}
