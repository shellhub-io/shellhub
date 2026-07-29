package requests

// WebReauthVerify is the step-up payload for an SSH login's require_reauth gate.
// TenantID/UserID are set server-side from the gateway-injected headers, never
// the body. Community validates Password; the enterprise overlay validates Code
// (TOTP) when the user has MFA.
type WebReauthVerify struct {
	TenantID string `json:"-"`
	UserID   string `json:"-"`
	Password string `json:"password"`
	Code     string `json:"code"`
	// Fingerprint is the SSH identity being re-authenticated. Its freshness is
	// stamped on that identity's last_reauth_at, and it must belong to the caller.
	Fingerprint string `json:"fingerprint" validate:"required"`
	// ApprovalCode is the login waiting on this step-up. Present when a held login
	// should be released by the same call that proves the factor, so verifying and
	// releasing cannot drift apart. Distinct from Code, which is the TOTP.
	ApprovalCode string `json:"approval_code"`
}
