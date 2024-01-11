package models

type EnableMFA struct {
	TokenMFA string   `json:"token_mfa"`
	Secret   string   `json:"secret"`
	Codes    []string `json:"codes" bson:"codes"`
}

type MFA struct {
	Enable   bool `json:"enable"`
	Validate bool `json:"validate"`
}

type GetCodes struct {
	Codes []string `json:"codes" bson:"codes"`
}

type Code struct {
	Code string `json:"code"`
}
