package models

type ValidateMFA struct {
	TokenMFA string `json:"token_mfa"`
	Secret   string `json:"secret"`
}

type MFA struct {
	Status   bool `json:"status"`
	Validate bool `json:"validate"`
}

type CreateMFA struct {
	Username string `json:"username"`
	Status   bool   `json:"status"`
}
