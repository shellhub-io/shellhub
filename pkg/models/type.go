package models

// Type is a namespace's kind. It decides whether the namespace can hold more than one member,
// and cloud billing reads it to pick a plan.
type Type string

const (
	// TypePersonal is a single-owner namespace with no invitations.
	TypePersonal Type = "personal"
	// TypeTeam is a namespace that can hold members, and the type a namespace gets by default.
	TypeTeam Type = "team"
)

// NewDefaultType returns the type a namespace is created with when the caller does not choose one.
func NewDefaultType() Type {
	return TypeTeam
}

// IsTypeTeam reports whether the raw string names the team type. An unrecognized value is not a
// team, so a corrupted or future value degrades to the more restrictive answer.
func IsTypeTeam(typeNamespace string) bool {
	return Type(typeNamespace) == TypeTeam
}

// IsTypePersonal reports whether the raw string names the personal type. An unrecognized value is
// not personal, and neither predicate is the negation of the other.
func IsTypePersonal(typeNamespace string) bool {
	return Type(typeNamespace) == TypePersonal
}
