package models

import "time"

type PublicKeyFields struct {
	Name string `json:"name"`
}

type PublicKey struct {
	Data            []byte    `json:"data"`
	Fingerprint     string    `json:"fingerprint"`
	CreatedAt       time.Time `json:"created_at" bson:"created_at"`
	TenantID        string    `json:"tenant_id" bson:"tenant_id"`
	PublicKeyFields `bson:",inline"`
}

type PublicKeyUpdate struct {
	PublicKeyFields `bson:",inline"`
}

type PublicKeyAuthRequest struct {
	Fingerprint string `json:"fingerprint"`
	Data        string `json:"data"`
}

type PublicKeyAuthResponse struct {
	Signature string `json:"signature"`
}
