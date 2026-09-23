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
)

// Principal is who a request or a connection is acting as. Carrying the kind beside the id is
// what keeps a caller from having to guess which table the id belongs to.
type Principal struct {
	Kind PrincipalKind `json:"kind"`
	ID   string        `json:"id"`
}

// PrincipalOf is the principal a row carrying both ids acts as: the API key when one is set,
// since a key acts on behalf of the member who created it and is the credential that was
// presented, and otherwise the user. It is nil when neither is set.
func PrincipalOf(userID, apiKeyID string) *Principal {
	switch {
	case apiKeyID != "":
		return &Principal{Kind: PrincipalAPIKey, ID: apiKeyID}
	case userID != "":
		return &Principal{Kind: PrincipalUser, ID: userID}
	default:
		return nil
	}
}
