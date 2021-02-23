package webhook

// Webhook request headers
const (
	// A unique ID that identifies the delivered webhook
	WebhookIDHeader = "X-SHELLHUB-WEBHOOK-ID"
	// Name of the event that has been triggered
	WebhookEventHeader = "X-SHELLHUB-WEBHOOK-EVENT"
	// A signature created using the webhook secret key
	WebhookSignatureHeader = "X-SHELLHUB-WEBHOOK-SIGNATURE"
)

// Webhook event types
const (
	// A new connection was made to the SSH Server
	WebhookIncomingConnectionEvent = "incoming_connection"
)

// IncomingConnectionWebhookRequest is the body payload
type IncomingConnectionWebhookRequest struct {
	Username  string `json:"username"`
	Hostname  string `json:"hostname"`
	Namespace string `json:"namespace"`
	SourceIP  string `json:"source_ip"`
}

// IncommingConnectionWebhookResponse is the expected response body
type IncomingConnectionWebhookResponse struct {
	// Timeout to wait for connection to be established
	Timeout int `json:"timeout"`
}
