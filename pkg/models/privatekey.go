package models

import "time"

type PrivateKey struct {
	Data        []byte    `json:"data"`
	Fingerprint string    `json:"fingerprint"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
}
