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
	ID      string    `json:"id,omitempty" bson:"id,omitempty"`
	AddedAt time.Time `json:"added_at" bson:"added_at"`

	// ExpiresAt specifies the expiration date of the invite. This attribute is only applicable in *Cloud* instances,
	// and it is ignored for members whose status is not 'pending'.
	ExpiresAt time.Time `json:"expires_at" bson:"expires_at"`

	Email  string          `json:"email" bson:"email,omitempty" validate:"email"`
	Role   authorizer.Role `json:"role" bson:"role" validate:"required,oneof=administrator operator observer"`
	Status MemberStatus    `json:"status" bson:"status"`
}
