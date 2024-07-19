package models

import "github.com/shellhub-io/shellhub/pkg/api/authorizer"

type Member struct {
	ID       string          `json:"id,omitempty" bson:"id,omitempty"`
	Username string          `json:"username,omitempty" bson:"username,omitempty" validate:"username"` // TODO: remove
	Role     authorizer.Role `json:"role" bson:"role" validate:"required,oneof=administrator operator observer"`
}

type MemberChanges struct {
	Role authorizer.Role `bson:"role,omitempty"`
}
