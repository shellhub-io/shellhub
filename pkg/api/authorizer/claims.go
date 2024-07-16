package authorizer

// UserClaims represents the attributes needed to authenticate a user.
type UserClaims struct {
	ID string `json:"id"`
	// TenantID is the identifier of the tenant to which the claims belongs.
	// It's optional.
	TenantID string `json:"tenant"`
	Role     Role   `json:"-"`
	Username string `json:"name"`
	// MFA indicates whether multi-factor authentication is enabled for the user.
	MFA bool `json:"mfa"`
}

// DeviceClaims represents the attributes needed to authenticate a device.
type DeviceClaims struct {
	UID      string `json:"uid"`
	TenantID string `json:"tenant"`
}
