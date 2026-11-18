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

// ParseType returns the type raw names, and TypeTeam for anything that names no type, the empty
// string included.
func ParseType(raw string) Type {
	if Type(raw) == TypePersonal {
		return TypePersonal
	}

	return TypeTeam
}
