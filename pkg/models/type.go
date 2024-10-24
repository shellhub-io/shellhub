package models

type Type string

const (
	TypePersonal Type = "personal"
	TypeTeam     Type = "team"
)

func IsTypeTeam(typeNamespace Type) bool {
	return typeNamespace == TypeTeam
}

func IsTypePersonal(typeNamespace Type) bool {
	return typeNamespace == TypePersonal
}
