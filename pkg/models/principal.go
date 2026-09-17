package models

// PrincipalKind is what sort of thing acted: a person, or an automation. It is deliberately not
// the membership role. The role says what a principal may do; the kind says what it is, and the
// two were the same field once, which is the mistake this type exists to prevent.
type PrincipalKind string

const (
	// PrincipalUser is a person, a row in users, authorized by their membership role.
	PrincipalUser PrincipalKind = "user"

	// PrincipalAPIKey is an automation, a row in api_keys. It holds no membership and no role
	// over SSH: where it may connect is decided by access policies alone.
	PrincipalAPIKey PrincipalKind = "api-key"

	// PrincipalService is a service account, the synthetic user an automation used to be. It
	// exists only while such rows do, and goes when they are deleted.
	PrincipalService PrincipalKind = "service"
)

// Principal is who a request or a connection is acting as. Carrying the kind beside the id is
// what keeps a caller from having to guess which table the id belongs to.
type Principal struct {
	Kind PrincipalKind
	ID   string
}
