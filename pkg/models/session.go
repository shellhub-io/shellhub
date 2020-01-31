package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Session struct {
	ID        bson.ObjectId `json:"-" bson:"_id,omitempty"`
	UID       string        `json:"uid"`
	Device    UID           `json:"device"`
	TenantID  string        `json:"tenant_id" bson:"tenant_id"`
	Username  string        `json:"username"`
	IPAddress string        `json:"ip_address" bson:"ip_address"`
	StartedAt time.Time     `json:"started_at" bson:"started_at"`
	LastSeen  time.Time     `json:"last_seen" bson:"last_seen"`
	Active    bool          `json:"active"`
}

type ActiveSession struct {
	ID       bson.ObjectId `json:"-" bson:"_id,omitempty"`
	UID      UID           `json:"uid"`
	LastSeen time.Time     `json:"last_seen" bson:"last_seen"`
}
