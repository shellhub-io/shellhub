package models

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
)

type Member struct {
	ID      string          `json:"id,omitempty" bson:"id,omitempty"`
	AddedAt time.Time       `json:"added_at" bson:"added_at"`
	Email   string          `json:"email" bson:"email,omitempty" validate:"email"`
	Role    authorizer.Role `json:"role" bson:"role" validate:"required,oneof=administrator operator observer"`
}
