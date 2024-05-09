package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

const SetupEndpoint = "/setup"

func (h *Handler) setup() *Route { //nolint:unused
	return &Route{
		endpoint:    "/setup",
		method:      MethodPost,
		group:       GroupDisable,
		blockAPIKey: false,
		middlewares: []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.Setup
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			if err := h.service.Setup(c.Ctx(), req); err != nil {
				return err
			}

			return c.NoContent(http.StatusOK)
		},
	}
}
