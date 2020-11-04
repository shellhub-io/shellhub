package models

type PublicKey struct {
	Name        string `json:"name"`
	Data        []byte `json:"data"`
	Fingerprint string `json:"fingerprint"`
	TenantID    string `json:"tenant_id" bson:"tenant_id"`
}

type PublicKeyAuthRequest struct {
	Fingerprint string `json:"fingerprint"`
	Data        string `json:"data"`
}

type PublicKeyAuthResponse struct {
	Signature string `json:"signature"`
}
