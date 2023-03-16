package requests

// AuthTokenGet is the structure to represent the request data for get auth token endpoint.
type AuthTokenGet struct {
	TenantParam
}

// AuthTokenSwap is the structure to represent the request data for swap auth token endpoint.
type AuthTokenSwap struct {
	TenantParam
}
