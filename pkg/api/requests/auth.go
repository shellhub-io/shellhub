package requests

// RenewUserToken is the request to replace a signed-in user's bearer, optionally scoped to one
// namespace.
// With no TenantID the token carries no namespace, which is what a user with no memberships gets.
type RenewUserToken struct {
	UserID   string `identity:"id" json:"-" validate:"required"`
	TenantID string `param:"tenant" json:"-" validate:"omitempty,uuid"`
	// TokenVersion is the version of the bearer being renewed, which the new token carries in
	// place of the user's current one. Nil mints at the user's current version. The handler
	// fills it from the authentication step; no client input reaches it.
	TokenVersion *int `json:"-"`
	// Impersonated mints a token an instance administrator acts as the user with. The admin
	// surface sets it, and every handler that mints a replacement for such a token keeps it; no
	// client input reaches it.
	Impersonated bool `json:"-"`
}
