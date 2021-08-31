package models

import (
	"time"
)

type Namespace struct {
	Name         string             `json:"name"  validate:"required,hostname_rfc1123,excludes=."`
	Owner        string             `json:"owner"`
	TenantID     string             `json:"tenant_id" bson:"tenant_id,omitempty"`
	Members      []Member           `json:"members" bson:"members"`
	Settings     *NamespaceSettings `json:"settings"`
	Devices      int                `json:"devices" bson:",omitempty"`
	Sessions     int                `json:"sessions" bson:",omitempty"`
	MaxDevices   int                `json:"max_devices" bson:"max_devices"`
	DevicesCount int                `json:"devices_count" bson:"devices_count,omitempty"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	Billing      *Billing           `json:"billing" bson:"billing,omitempty"`
}

type NamespaceSettings struct {
	SessionRecord bool `json:"session_record" bson:"session_record,omitempty"`
}

type Member struct {
	ID         string `json:"id" bson:"id"`
	Name       string `json:"name,omitempty" bson:"-"`
	AccessType string `json:"access_type, omitempty" bson:"access_type" validate:"required,oneof=owner admin operator observer`
}
