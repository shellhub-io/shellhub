package models

import (
	"time"
)

type Token struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	TenantID  string    `json:"tenant_id" bson:"tenant_id"`
}
