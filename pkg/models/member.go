package models

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
)

type MembershipStatus string

const (
	MembershipStatusPending  MembershipStatus = "pending"
	MembershipStatusAccepted MembershipStatus = "accepted"
)

type Membership struct {
	UserID      string `json:"user_id" bun:"user_id,pk,type:uuid"`
	NamespaceID string `json:"-" bun:"namespace_id,pk,type:uuid"`

	// CreatedAt represents the timestamp when the membership was created
	CreatedAt time.Time `json:"created_at" bun:"created_at"`
	// UpdatedAt represents the timestamp when the membership was last updated
	UpdatedAt time.Time `json:"updated_at" bun:"updated_at"`

	Status MembershipStatus `json:"status" bun:"status"`
	Role   authorizer.Role  `json:"role" bun:"role"`
}
