package models

import "time"

type PrivateKey struct {
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	Fingerprint string    `json:"fingerprint"`
	Data        []byte    `json:"data"`
}
