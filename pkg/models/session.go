package models

import (
	"time"
)

type Session struct {
	UID       string    `json:"uid"`
	Device    UID       `json:"device"`
	TenantID  string    `json:"tenant_id" bson:"tenant_id"`
	Username  string    `json:"username"`
	IPAddress string    `json:"ip_address" bson:"ip_address"`
	StartedAt time.Time `json:"started_at" bson:"started_at"`
	LastSeen  time.Time `json:"last_seen" bson:"last_seen"`
	Active    bool      `json:"active" bson:",omitempty"`
}

type ActiveSession struct {
	UID      UID       `json:"uid"`
	LastSeen time.Time `json:"last_seen" bson:"last_seen"`
}
