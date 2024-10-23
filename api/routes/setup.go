package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

const SetupEndpoint = "/setup"

func (h *Handler) Setup(c gateway.Context) error {
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
}

func (h *Handler) SetupCheck(c gateway.Context) error {
	if err := h.service.SetupCheck(c.Ctx()); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
