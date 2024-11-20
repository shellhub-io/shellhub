package models

type Type string

const (
	TypePersonal Type = "personal"
	TypeTeam     Type = "team"
)

func NewDefaultType() Type {
	return TypeTeam
}

func IsTypeTeam(typeNamespace string) bool {
	return Type(typeNamespace) == TypeTeam
}

func IsTypePersonal(typeNamespace string) bool {
	return Type(typeNamespace) == TypePersonal
}
