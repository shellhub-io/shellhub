package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	errs "github.com/shellhub-io/shellhub/server/api/routes/errors"
	"github.com/shellhub-io/shellhub/server/api/services"
	log "github.com/sirupsen/logrus"
)

// The session routes, relative to the API's base path.
const (
	GetSessionsURL = "/sessions"
	GetSessionURL  = "/sessions/:uid"
)

// GetSessionList serves the namespace's sessions, filtered and paginated as requested.
func (h *Handler) GetSessionList(c *gateway.Context) error {
	req := new(requests.ListSessions)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	req.Paginator.Normalize()

	if err := req.Filters.Unmarshal(); err != nil {
		log.WithError(err).WithField("filter", req.Filters.Raw).Warn("failed to decode session list filter")

		return errs.NewErrInvalidEntity(map[string]string{"filter": err.Error()})
	}

	if err := query.ValidateFilters(&req.Filters, services.SessionFilterFields); err != nil {
		return errs.NewErrInvalidEntity(map[string]string{"filter": err.Error()})
	}

	sc, err := c.AdminOrScope()
	if err != nil {
		return err
	}

	sessions, count, err := h.service.ListSessions(c.Ctx(), sc, req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.FormatInt(count, 10))

	return c.JSON(http.StatusOK, sessions)
}

// GetSession serves one session by UID. The session carries its event timeline only for a caller
// who may replay recordings, or one on the instance admin console; a failed timeline read serves the
// session without it. It answers 422 on a body that fails to bind (413 or 415 on an oversized or XML
// body), 400 on a body it cannot read or a request that fails validation, 403 when the caller
// carries no namespace, and 404 when the session is not in the caller's scope or the store fails to
// resolve it.
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

	req.Timeline = !sc.IsBounded() || c.Role().HasPermission(authorizer.SessionPlay)

	session, err := h.service.GetSession(c.Ctx(), sc, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, session)
}
