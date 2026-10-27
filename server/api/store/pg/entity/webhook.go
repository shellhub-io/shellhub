package entity

import (
	"encoding/json"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

// Webhook is a row of webhooks, one namespace's subscription to its events.
type Webhook struct {
	bun.BaseModel `bun:"table:webhooks,alias:webhook"`

	ID              string    `bun:"id,pk,type:uuid,nullzero,default:gen_random_uuid()"`
	NamespaceID     string    `bun:"namespace_id,type:uuid"`
	URL             string    `bun:"url"`
	SecretEncrypted string    `bun:"secret_encrypted"`
	Events          []string  `bun:"events,array"`
	Enabled         bool      `bun:"enabled"`
	CreatedAt       time.Time `bun:"created_at"`
	UpdatedAt       time.Time `bun:"updated_at"`
}

// WebhookFromModel projects a webhook into its row form.
func WebhookFromModel(model *models.Webhook) *Webhook {
	events := make([]string, len(model.Events))
	for i, event := range model.Events {
		events[i] = string(event)
	}

	return &Webhook{
		ID:              model.ID,
		NamespaceID:     model.TenantID,
		URL:             model.URL,
		SecretEncrypted: model.SecretEncrypted,
		Events:          events,
		Enabled:         model.Enabled,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}

// WebhookToModel rebuilds a webhook from its row.
func WebhookToModel(e *Webhook) *models.Webhook {
	events := make([]models.WebhookEventType, len(e.Events))
	for i, event := range e.Events {
		events[i] = models.WebhookEventType(event)
	}

	return &models.Webhook{
		ID:              e.ID,
		TenantID:        e.NamespaceID,
		URL:             e.URL,
		SecretEncrypted: e.SecretEncrypted,
		Events:          events,
		Enabled:         e.Enabled,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}

// WebhookEvent is a row of webhook_events, one change waiting to reach its webhooks.
type WebhookEvent struct {
	bun.BaseModel `bun:"table:webhook_events,alias:webhook_event"`

	ID          string          `bun:"id,pk,type:uuid"`
	NamespaceID string          `bun:"namespace_id,type:uuid"`
	Type        string          `bun:"type"`
	Data        json.RawMessage `bun:"data,type:jsonb"`
	CreatedAt   time.Time       `bun:"created_at"`
}

// WebhookDelivery is a row of webhook_deliveries, one event's sending to one webhook. EventType is
// not a column: it is read from the event the row points at.
type WebhookDelivery struct {
	bun.BaseModel `bun:"table:webhook_deliveries,alias:webhook_delivery"`

	ID             string     `bun:"id,pk,type:uuid,nullzero,default:gen_random_uuid()"`
	EventID        string     `bun:"event_id,type:uuid"`
	EventType      string     `bun:"event_type,scanonly"`
	WebhookID      string     `bun:"webhook_id,type:uuid"`
	NamespaceID    string     `bun:"namespace_id,type:uuid"`
	Status         string     `bun:"status"`
	Attempts       int        `bun:"attempts"`
	NextAttemptAt  time.Time  `bun:"next_attempt_at"`
	LeaseUntil     *time.Time `bun:"lease_until"`
	LastStatusCode *int       `bun:"last_status_code"`
	LastError      *string    `bun:"last_error"`
	DeliveredAt    *time.Time `bun:"delivered_at"`
	CreatedAt      time.Time  `bun:"created_at"`
	UpdatedAt      time.Time  `bun:"updated_at"`
}

// WebhookDeliveryFromModel projects a delivery into its row form, with its lease released.
func WebhookDeliveryFromModel(model *models.WebhookDelivery) *WebhookDelivery {
	delivery := &WebhookDelivery{
		ID:            model.ID,
		EventID:       model.EventID,
		WebhookID:     model.WebhookID,
		NamespaceID:   model.TenantID,
		Status:        string(model.Status),
		Attempts:      model.Attempts,
		NextAttemptAt: model.NextAttemptAt,
		DeliveredAt:   model.DeliveredAt,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}

	if model.LastStatusCode != 0 {
		delivery.LastStatusCode = &model.LastStatusCode
	}

	if model.LastError != "" {
		delivery.LastError = &model.LastError
	}

	return delivery
}

// WebhookDeliveryToModel rebuilds a delivery from its row.
func WebhookDeliveryToModel(e *WebhookDelivery) *models.WebhookDelivery {
	delivery := &models.WebhookDelivery{
		ID:            e.ID,
		EventID:       e.EventID,
		EventType:     models.WebhookEventType(e.EventType),
		WebhookID:     e.WebhookID,
		TenantID:      e.NamespaceID,
		Status:        models.WebhookDeliveryStatus(e.Status),
		Attempts:      e.Attempts,
		NextAttemptAt: e.NextAttemptAt,
		DeliveredAt:   e.DeliveredAt,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}

	if e.LastStatusCode != nil {
		delivery.LastStatusCode = *e.LastStatusCode
	}

	if e.LastError != nil {
		delivery.LastError = *e.LastError
	}

	return delivery
}
