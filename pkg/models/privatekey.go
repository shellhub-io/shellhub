package models

type PrivateKey struct {
	Data        []byte `json:"data"`
	Fingerprint string `json:"fingerprint"`
}
