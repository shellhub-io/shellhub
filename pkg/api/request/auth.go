package request

// AuthTokenGet is the structure for the request data at get auth token endpoint.
type AuthTokenGet struct {
	TenantParam
}

// AuthTokenSwap is the structure for the request data at swap auth token endpoint.
type AuthTokenSwap struct {
	TenantParam
}
