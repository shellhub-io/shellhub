package models

// Username carries the authenticated user's username as the API's gateway reads it out of a
// request's context, alongside ID and Tenant.
type Username struct {
	ID string
}
