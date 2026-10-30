package authorizer

// UserClaims represents the attributes needed to authenticate a user.
type UserClaims struct {
	ID     string `json:"id"`
	Origin string `json:"origin"`
	// TenantID is the identifier of the tenant to which the claims belongs.
	// It's optional.
	TenantID string `json:"tenant"`
	Role     Role   `json:"-"`
	Username string `json:"name"`
	// MFA indicates whether multi-factor authentication is enabled for the user.
	MFA bool `json:"mfa"`
	// Admin indicates whether the user has administrative privileges.
	Admin        bool `json:"admin"`
	TokenVersion int  `json:"token_version"`
	// Impersonated marks a token an instance administrator was issued to act as the user. A
	// renewal of it stays impersonated.
	Impersonated bool `json:"impersonated,omitempty"`
}

// DeviceClaims represents the attributes needed to authenticate a device.
type DeviceClaims struct {
	UID         string `json:"uid"`
	TenantID    string `json:"tenant"`
	KeyVerified bool   `json:"key_verified,omitempty"`
}
