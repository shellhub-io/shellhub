package requests

// AuthTokenGet is the structure to represent the request data for get auth token endpoint.
type AuthTokenGet struct {
	UserParam
}

// AuthTokenSwap is the structure to represent the request data for swap auth token endpoint.
type AuthTokenSwap struct {
	TenantParam
}

type CreateUserToken struct {
	UserID   string `param:"id" header:"X-ID" validate:"required"`
	TenantID string `param:"tenant" validate:"omitempty,uuid"`
}
