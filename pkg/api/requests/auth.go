package requests

// AuthTokenGet is the structure to represent the request data for get auth token endpoint.
type AuthTokenGet struct {
	UserParam
}

// AuthTokenSwap is the structure to represent the request data for swap auth token endpoint.
type AuthTokenSwap struct {
	TenantParam
}

// CreateUserToken is the request to mint a token for a user, optionally scoped to one namespace.
// With no TenantID the token carries no namespace, which is what a user with no memberships gets.
type CreateUserToken struct {
	UserID   string `param:"id" header:"X-ID" validate:"required"`
	TenantID string `param:"tenant" validate:"omitempty,uuid"`
}
