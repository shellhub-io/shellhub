package models

type Type struct {
	Personal bool
	Team     bool
}

func IsTypeTeam (typeNamespace Type) bool {

	return typeNamespace.Team
}

func IsTypePersonal (typeNamespace Type) bool {

	return typeNamespace.Personal
}
