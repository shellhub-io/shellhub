package models

import "time"

type Slot struct {
	UID       UID       `json:"uid" bson:"uid"`
	Tenant    string    `json:"tenant_id" bson:"tenant_id"`
	Status    string    `json:"status" bson:"status"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
