package requests

import (
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/models"
)

// CreateWebhook is the request to subscribe a URL to a namespace's events. TenantID comes from the
// caller's header, so a webhook cannot be created in another namespace.
type CreateWebhook struct {
	TenantID string                    `header:"X-Tenant-ID"`
	URL      string                    `json:"url" validate:"required,url"`
	Events   []models.WebhookEventType `json:"events" validate:"required,min=1,dive,oneof=device.pending device.accepted device.rejected device.removed"`
}

// ListWebhooks is the request to page through a namespace's webhooks.
type ListWebhooks struct {
	TenantID string `header:"X-Tenant-ID"`
	query.Paginator
}

// WebhookID addresses one of the namespace's webhooks.
type WebhookID struct {
	TenantID string `header:"X-Tenant-ID"`
	ID       string `param:"id" validate:"required,uuid"`
}

// UpdateWebhook is a partial update of a webhook: a nil field is left unchanged.
type UpdateWebhook struct {
	TenantID string                    `header:"X-Tenant-ID"`
	ID       string                    `param:"id" validate:"required,uuid"`
	URL      *string                   `json:"url" validate:"omitempty,url"`
	Events   []models.WebhookEventType `json:"events" validate:"omitempty,min=1,dive,oneof=device.pending device.accepted device.rejected device.removed"`
	Enabled  *bool                     `json:"enabled"`
}

// ListWebhookDeliveries is the request to page through one webhook's deliveries, newest first.
type ListWebhookDeliveries struct {
	TenantID string `header:"X-Tenant-ID"`
	ID       string `param:"id" validate:"required,uuid"`
	query.Paginator
}
