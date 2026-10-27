package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

// The webhook routes, relative to the API's base path.
const (
	CreateWebhookURL         = "/namespaces/webhooks"
	ListWebhooksURL          = "/namespaces/webhooks"
	GetWebhookURL            = "/namespaces/webhooks/:id"
	UpdateWebhookURL         = "/namespaces/webhooks/:id"
	DeleteWebhookURL         = "/namespaces/webhooks/:id"
	RevealWebhookSecretURL   = "/namespaces/webhooks/:id/reveal" //nolint:gosec // G101: a route path, not a credential
	ListWebhookDeliveriesURL = "/namespaces/webhooks/:id/deliveries"
)

// CreateWebhook subscribes a URL to the namespace's events, returning the webhook and its secret.
func (h *Handler) CreateWebhook(c *gateway.Context) error {
	req := new(requests.CreateWebhook)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, err := h.service.CreateWebhook(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// ListWebhooks serves a page of the namespace's webhooks, without their secrets.
func (h *Handler) ListWebhooks(c *gateway.Context) error {
	req := new(requests.ListWebhooks)

	if err := c.Bind(req); err != nil {
		return err
	}

	req.Paginator.Normalize()

	if err := c.Validate(req); err != nil {
		return err
	}

	res, count, err := h.service.ListWebhooks(c.Ctx(), req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.FormatInt(count, 10))

	return c.JSON(http.StatusOK, res)
}

// GetWebhook serves one of the namespace's webhooks, without its secret.
func (h *Handler) GetWebhook(c *gateway.Context) error {
	req := new(requests.WebhookID)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, err := h.service.GetWebhook(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// UpdateWebhook changes a webhook's URL, event types or enabled flag.
func (h *Handler) UpdateWebhook(c *gateway.Context) error {
	req := new(requests.UpdateWebhook)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, err := h.service.UpdateWebhook(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// DeleteWebhook deletes a webhook and its deliveries.
func (h *Handler) DeleteWebhook(c *gateway.Context) error {
	req := new(requests.WebhookID)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.DeleteWebhook(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// RevealWebhookSecret returns a webhook's signing secret.
func (h *Handler) RevealWebhookSecret(c *gateway.Context) error {
	req := new(requests.WebhookID)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	secret, err := h.service.RevealWebhookSecret(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &responses.RevealWebhookSecret{Secret: secret})
}

// ListWebhookDeliveries serves a page of a webhook's deliveries, newest first.
func (h *Handler) ListWebhookDeliveries(c *gateway.Context) error {
	req := new(requests.ListWebhookDeliveries)

	if err := c.Bind(req); err != nil {
		return err
	}

	req.Paginator.Normalize()

	if err := c.Validate(req); err != nil {
		return err
	}

	res, count, err := h.service.ListWebhookDeliveries(c.Ctx(), req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.FormatInt(count, 10))

	return c.JSON(http.StatusOK, res)
}
