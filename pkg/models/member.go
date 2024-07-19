package models

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
)

type MemberStatus string

const (
	MemberStatusPending  MemberStatus = "pending"
	MemberStatusAccepted MemberStatus = "accepted"
)

type Member struct {
	ID       string          `json:"id,omitempty" bson:"id,omitempty"`
	AddedAt  time.Time       `json:"added_at" bson:"added_at"`
	Username string          `json:"username,omitempty" bson:"username,omitempty" validate:"username"` // TODO: remove
	Role     authorizer.Role `json:"role" bson:"role" validate:"required,oneof=administrator operator observer"`
	Status   MemberStatus    `json:"status" bson:"status"`
}

type MemberChanges struct {
	Role   authorizer.Role `bson:"role,omitempty"`
	Status MemberStatus    `bson:"status,omitempty"`
}
