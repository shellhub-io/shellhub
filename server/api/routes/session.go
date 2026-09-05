package routes

import (
	"context"
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

// The session routes, relative to the API's base path.
const (
	GetSessionsURL = "/sessions"
	GetSessionURL  = "/sessions/:uid"
)

// The path parameter name these routes bind by.
const (
	ParamSessionID = "uid"
)

// GetSessionList serves the namespace's sessions, filtered and paginated as requested.
func (h *Handler) GetSessionList(ctx context.Context, sc scope.Scope, _ gateway.Actor, req *requests.ListSessions) ([]models.Session, int, error) {
	return h.service.ListSessions(ctx, sc, req)
}

// GetSession serves one session by UID.
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
