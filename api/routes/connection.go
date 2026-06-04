package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

const (
	CreateConnectionURL = "/connections"
	ListConnectionsURL  = "/connections"
	GetConnectionURL    = "/connections/:id"
	ConnectionStatusURL = "/connections/:id/status"
	UpdateConnectionURL = "/connections/:id"
	DeleteConnectionURL = "/connections/:id"
)

const ParamConnectionID = "id"

func (h *Handler) CreateConnection(c gateway.Context) error {
	var req requests.ConnectionCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// For direct connections, probe reachability at save time. If unreachable and
	// the user hasn't chosen to save anyway, return 422 so the UI can surface the
	// NAT/firewall hint and funnel toward installing the agent.
	if req.Kind == "direct" && !req.Force {
		reachable, err := h.service.ProbeReachability(c.Ctx(), &requests.ConnectionProbe{Host: req.Host, Port: req.Port})
		if err != nil {
			return err
		}

		if !reachable {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "unreachable"})
		}
	}

	connection, err := h.service.CreateConnection(c.Ctx(), &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, connection)
}

func (h *Handler) ListConnections(c gateway.Context) error {
	req := new(requests.ConnectionList)
	if err := c.Bind(req); err != nil {
		return err
	}

	req.Paginator.Normalize()
	req.Sorter.Normalize()

	if err := c.Validate(req); err != nil {
		return err
	}

	connections, count, err := h.service.ListConnections(c.Ctx(), req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, connections)
}

func (h *Handler) GetConnection(c gateway.Context) error {
	var req requests.ConnectionGet
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	connection, err := h.service.GetConnection(c.Ctx(), &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, connection)
}

func (h *Handler) ConnectionStatus(c gateway.Context) error {
	var req requests.ConnectionGet
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	online, err := h.service.ConnectionStatus(c.Ctx(), &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]bool{"online": online})
}

func (h *Handler) UpdateConnection(c gateway.Context) error {
	var req requests.ConnectionUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	connection, err := h.service.UpdateConnection(c.Ctx(), &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, connection)
}

func (h *Handler) DeleteConnection(c gateway.Context) error {
	var req requests.ConnectionDelete
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.DeleteConnection(c.Ctx(), &req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
