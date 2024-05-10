package routes

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	ParamSessionID = "uid"
)

func (h *Handler) createSession() *Route {
	return &Route{
		endpoint:              "/sessions",
		method:                MethodPost,
		group:                 GroupInternal,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			return c.NoContent(200)

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
		},
	}
}

func (h *Handler) getSession() *Route {
	return &Route{
		endpoint:              "/sessions/:uid",
		method:                MethodGet,
		group:                 GroupPublic,
		requiresAuthorization: true,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			return c.NoContent(200)

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
		},
	}
}

func (h *Handler) listSessions() *Route {
	return &Route{
		endpoint:              "/sessions",
		method:                MethodGet,
		group:                 GroupPublic,
		requiresAuthorization: true,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			return c.NoContent(200)

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
		},
	}
}

// TODO:
// authenticateSession, finishSession and keepAliveSession can be a single route.

func (h *Handler) authenticateSession() *Route {
	return &Route{
		endpoint:              "/sessions/:uid",
		method:                MethodPatch,
		group:                 GroupInternal,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			return c.NoContent(200)

			var req requests.SessionAuthenticatedSet
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			return h.service.SetSessionAuthenticated(c.Ctx(), models.UID(req.UID), req.Authenticated)
		},
	}
}
func (h *Handler) finishSession() *Route {
	return &Route{
		endpoint:              "/sessions/:uid/finish",
		method:                MethodPost,
		group:                 GroupInternal,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			return c.NoContent(200)

			var req requests.SessionFinish
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			return h.service.DeactivateSession(c.Ctx(), models.UID(req.UID))
		},
	}
}

func (h *Handler) keepAliveSession() *Route {
	return &Route{
		endpoint:              "/sessions/:uid/keepalive",
		method:                MethodPost,
		group:                 GroupInternal,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			return c.NoContent(200)

			var req requests.SessionKeepAlive
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			return h.service.KeepAliveSession(c.Ctx(), models.UID(req.UID))
		},
	}
}

func (h *Handler) playRecordedSession() *Route {
	return &Route{
		endpoint:              "/sessions/:uid/play",
		method:                MethodGet,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			return c.NoContent(http.StatusOK)
		},
	}
}

func (h *Handler) recordSession() *Route {
	return &Route{
		endpoint:              "/sessions/:uid/record",
		method:                MethodPost,
		group:                 GroupInternal,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			return c.NoContent(http.StatusOK)
		},
	}
}

func (h *Handler) deleteRecordedSession() *Route {
	return &Route{
		endpoint:              "/sessions/:uid/record",
		method:                MethodDelete,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			return c.NoContent(http.StatusOK)
		},
	}
}
