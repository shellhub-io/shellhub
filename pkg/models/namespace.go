package models

type Namespace struct {
	Name         string             `json:"name"  validate:"required,hostname_rfc1123"`
	Owner        string             `json:"owner"`
	TenantID     string             `json:"tenant_id" bson:"tenant_id,omitempty"`
	Members      []interface{}      `json:"members" bson:"members"`
	Settings     *NamespaceSettings `json:"settings"`
	Devices      int                `json:"devices" bson:",omitempty"`
	Sessions     int                `json:"sessions" bson:",omitempty"`
	MaxDevices   int                `json:"max_devices" bson:"max_devices"`
	DevicesCount int                `json:"devices_count" bson:"devices_count,omitempty"`
}

type NamespaceSettings struct {
	SessionRecord bool           `json:"session_record" bson:"session_record,omitempty"`
	Webhook       WebhookOptions `json:"webhook" bson:"webhook,omitempty"`
}

type Member struct {
	ID   string `json:"id" bson:"id"`
	Name string `json:"name,omitempty" bson:"-"`
}

type WebhookOptions struct {
	URL    string `json:"url" bson:"url" validate:"required"`
	Active bool   `json:"active" bson:"active,omitempty"`
}
