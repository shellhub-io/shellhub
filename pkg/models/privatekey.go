package models

import "time"

// PrivateKey is a private key the server holds, kept with the fingerprint callers look it up by.
// Data is the key material itself: never log or serialize a PrivateKey into a response.
type PrivateKey struct {
	Data        []byte    `json:"data"`
	Fingerprint string    `json:"fingerprint"`
	CreatedAt   time.Time `json:"created_at"`
}
