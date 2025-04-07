package models

import "time"

type PrivateKey struct {
	Fingerprint string    `json:"fingerprint" bun:"fingerprint,pk"`
	CreatedAt   time.Time `json:"created_at" bun:"created_at"`
	UpdatedAt   time.Time `json:"created_at" bun:"updated_at"`
	Data        []byte    `json:"data" bun:"data,type:bytea"`
}
