package routes

import (
	"errors"
	"net/http"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

const (
	ScanKnownHostURL   = "/connections/host-key/scan"
	AcceptKnownHostURL = "/connections/host-key/accept"
	GetKnownHostURL    = "/connections/host-key"
	DeleteKnownHostURL = "/connections/host-key"
)

func (h *Handler) ScanKnownHost(c gateway.Context) error {
	var req requests.KnownHostScan
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	result, err := h.service.ScanKnownHost(c.Ctx(), &req)
	if err != nil {
		// Surface a target we can't reach/read as a 422 (not a 500), and let the
		// UI tell a blocked address apart from an unreachable host.
		switch {
		case errors.Is(err, svc.ErrEgressBlocked):
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "blocked"})
		case errors.Is(err, svc.ErrKnownHostUnreachable):
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "unreachable"})
		}

		return err
	}

	return c.JSON(http.StatusOK, result)
}

func (h *Handler) AcceptKnownHost(c gateway.Context) error {
	var req requests.KnownHostAccept
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	knownHost, err := h.service.AcceptKnownHost(c.Ctx(), &req)
	if err != nil {
		if errors.Is(err, svc.ErrKnownHostInvalidKey) {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": "invalid_key"})
		}

		return err
	}

	return c.JSON(http.StatusOK, knownHost)
}

func (h *Handler) GetKnownHost(c gateway.Context) error {
	var req requests.KnownHostGet
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	knownHost, err := h.service.GetKnownHost(c.Ctx(), &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, knownHost)
}

func (h *Handler) DeleteKnownHost(c gateway.Context) error {
	var req requests.KnownHostDelete
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.DeleteKnownHost(c.Ctx(), &req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
