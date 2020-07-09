package models

import (
	"time"
)

type Session struct {
	UID           string    `json:"uid"`
	DeviceUID     UID       `json:"device_uid,omitempty" bson:"device_uid"`
	Device        *Device   `json:"device" bson:",omitempty"`
	TenantID      string    `json:"tenant_id" bson:"tenant_id"`
	Username      string    `json:"username"`
	IPAddress     string    `json:"ip_address" bson:"ip_address"`
	StartedAt     time.Time `json:"started_at" bson:"started_at"`
	LastSeen      time.Time `json:"last_seen" bson:"last_seen"`
	Active        bool      `json:"active" bson:",omitempty"`
	Authenticated bool      `json:"authenticated" bson:"authenticated"`
}

type ActiveSession struct {
	UID      UID       `json:"uid"`
	LastSeen time.Time `json:"last_seen" bson:"last_seen"`
}

type RecordedSession struct {
	UID      UID       `json:"uid"`
	Message  string    `json:"message", bson:"message"`
	TenantID string    `json:"tenant_id" bson:"tenant_id,omitempty"`
	Time     time.Time `json:"time" bson:"time,omitempty"`
	Width    int       `json:"width" bson:"width,omitempty"`
	Height   int       `json:"height" bson:"height,omitempty"`
}
