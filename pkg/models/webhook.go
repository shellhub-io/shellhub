package models

import (
	"encoding/json"
	"time"
)

// WebhookEventType names a change a webhook can subscribe to.
type WebhookEventType string

// The device events a webhook can subscribe to, one for each status a device can enter.
const (
	WebhookEventDevicePending  WebhookEventType = "device.pending"
	WebhookEventDeviceAccepted WebhookEventType = "device.accepted"
	WebhookEventDeviceRejected WebhookEventType = "device.rejected"
	WebhookEventDeviceRemoved  WebhookEventType = "device.removed"
)

// WebhookEventTypes lists every event type a webhook can subscribe to.
var WebhookEventTypes = []WebhookEventType{
	WebhookEventDevicePending,
	WebhookEventDeviceAccepted,
	WebhookEventDeviceRejected,
	WebhookEventDeviceRemoved,
}

// Webhook is a namespace's subscription to events: where to send them, which ones, and the secret
// that signs each delivery. A webhook that is not Enabled receives no new events. The secret is
// stored encrypted and never returned with the webhook.
type Webhook struct {
	ID              string             `json:"id"`
	TenantID        string             `json:"tenant_id"`
	URL             string             `json:"url"`
	Events          []WebhookEventType `json:"events"`
	Enabled         bool               `json:"enabled"`
	SecretEncrypted string             `json:"-"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// WebhookEvent is one change to send to every webhook subscribed to its type. Its ID stays the same
// across every delivery and retry, so a receiver can drop a duplicate.
type WebhookEvent struct {
	ID        string           `json:"id"`
	TenantID  string           `json:"tenant_id"`
	Type      WebhookEventType `json:"type"`
	Data      json.RawMessage  `json:"data"`
	CreatedAt time.Time        `json:"created_at"`
}

// WebhookDeliveryStatus is where a delivery stands: still being tried, received, or given up on.
type WebhookDeliveryStatus string

// The statuses a delivery moves through. A pending delivery becomes delivered on a 2xx answer and
// failed once its attempts run out.
const (
	WebhookDeliveryPending   WebhookDeliveryStatus = "pending"
	WebhookDeliveryDelivered WebhookDeliveryStatus = "delivered"
	WebhookDeliveryFailed    WebhookDeliveryStatus = "failed"
)

// WebhookDelivery is the sending of one event to one webhook, with the outcome of its last attempt.
type WebhookDelivery struct {
	ID             string                `json:"id"`
	EventID        string                `json:"event_id"`
	EventType      WebhookEventType      `json:"event_type"`
	WebhookID      string                `json:"webhook_id"`
	TenantID       string                `json:"tenant_id"`
	Status         WebhookDeliveryStatus `json:"status"`
	Attempts       int                   `json:"attempts"`
	NextAttemptAt  time.Time             `json:"next_attempt_at"`
	LastStatusCode int                   `json:"last_status_code,omitempty"`
	LastError      string                `json:"last_error,omitempty"`
	DeliveredAt    *time.Time            `json:"delivered_at,omitempty"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}
