package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/request"
)

const SetupEndpoint = "/setup"

func (h *Handler) Setup(c gateway.Context) error {
	var req request.Setup
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
