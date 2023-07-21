package requests

// AuthTokenGet is the structure to represent the request data for get auth token endpoint.
//
//go:generate structsnapshot AuthTokenGet
type AuthTokenGet struct {
	TenantParam
}

// AuthTokenSwap is the structure to represent the request data for swap auth token endpoint.
//
//go:generate structsnapshot AuthTokenSwap
type AuthTokenSwap struct {
	TenantParam
}
