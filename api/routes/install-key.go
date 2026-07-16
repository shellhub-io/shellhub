package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/errors"
)

const (
	CreateInstallKeyURL  = "/namespaces/install-key"
	ListInstallKeysURL   = "/namespaces/install-key"
	UpdateInstallKeyURL  = "/namespaces/install-key/:name"
	RevealInstallKeyURL  = "/namespaces/install-key/:name/reveal"
	HistoryInstallKeyURL = "/namespaces/install-key/:id/history"

	// EnrollmentCallbackURL is the public, token-authenticated endpoint a webhook integrator POSTs its
	// deferred decision to. The token in the path is the credential (no API key/JWT).
	EnrollmentCallbackURL = "/devices/enroll/callback/:token"
)

func (h *Handler) CreateInstallKey(c gateway.Context) error {
	req := new(requests.CreateInstallKey)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, err := h.service.CreateInstallKey(c.Ctx(), req)
	if err != nil {
		// Surface an invalid mode field (webhook/allowlist config) with its per-field body instead of a
		// bare 400, mirroring the update handler.
		var e errors.Error
		if errors.As(err, &e) {
			if data, ok := e.Data.(services.ErrDataInvalidFields); ok {
				return c.JSON(http.StatusBadRequest, data)
			}
		}

		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) ListInstallKeys(c gateway.Context) error {
	req := new(requests.ListInstallKey)

	if err := c.Bind(req); err != nil {
		return err
	}

	req.Paginator.Normalize()

	if req.Sorter.By == "" {
		req.Sorter.By = "created_at"
	}

	if req.Sorter.Order == "" {
		req.Sorter.Order = "desc"
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, count, err := h.service.ListInstallKeys(c.Ctx(), req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) UpdateInstallKey(c gateway.Context) error {
	req := new(requests.UpdateInstallKey)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.UpdateInstallKey(c.Ctx(), req); err != nil {
		// The global error handler answers with a bare status. For an invalid field we escape it and
		// return the offending field(s) and why, so a caller learns what to fix instead of a raw 400.
		var e errors.Error
		if errors.As(err, &e) {
			if data, ok := e.Data.(services.ErrDataInvalidFields); ok {
				return c.JSON(http.StatusBadRequest, data)
			}
		}

		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) RevealInstallKey(c gateway.Context) error {
	req := new(requests.RevealInstallKey)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	key, err := h.service.RevealInstallKey(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &responses.RevealInstallKey{Key: key})
}

func (h *Handler) EnrollmentCallback(c gateway.Context) error {
	req := new(requests.EnrollmentCallback)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.ResolveEnrollmentCallback(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) HistoryInstallKey(c gateway.Context) error {
	req := new(requests.ListInstallKeyEvents)

	if err := c.Bind(req); err != nil {
		return err
	}

	req.Paginator.Normalize()

	if req.Sorter.By == "" {
		req.Sorter.By = "created_at"
	}

	if req.Sorter.Order == "" {
		req.Sorter.Order = "desc"
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	events, count, err := h.service.ListInstallKeyEvents(c.Ctx(), req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, events)
}
