package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

// The instance setup route, relative to the API's base path.
const (
	SetupEndpoint = "/setup"
)

// Setup creates the first user and namespace. It is reachable unauthenticated, and refuses
// once the instance has been set up.
func (h *Handler) Setup(c *gateway.Context) error {
	var req requests.Setup

	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	res, err := h.service.Setup(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}
