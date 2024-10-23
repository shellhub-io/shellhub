package models

import env "github.com/shellhub-io/shellhub/pkg/envs"

type Type string

const (
	TypePersonal Type = "personal"
	TypeTeam     Type = "team"
)

func NewDefaultType() Type {
	if env.IsCloud() {
		return TypeTeam
	}

	return TypePersonal
}

func IsTypeTeam(typeNamespace string) bool {
	return Type(typeNamespace) == TypeTeam
}

func IsTypePersonal(typeNamespace string) bool {
	return Type(typeNamespace) == TypePersonal
}
