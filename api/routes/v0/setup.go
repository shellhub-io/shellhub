package v0

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

const (
	SetupEndpoint  = "/setup"
	SetupSignQuery = "sign"
)

func (h *Handler) Setup(c gateway.Context) error {
	sign := c.QueryParam(SetupSignQuery)
	if sign == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	var req requests.Setup
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.SetupVerify(c.Ctx(), sign); err != nil {
		return err
	}

	if err := h.service.Setup(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
