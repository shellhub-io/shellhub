package models

// Tenant carries the namespace a request acts on, as the API's gateway reads it out of the
// request context. It is a struct rather than a string so an absent tenant is a nil pointer,
// which is the difference between "no namespace" and "the empty namespace".
type Tenant struct {
	ID string
}
