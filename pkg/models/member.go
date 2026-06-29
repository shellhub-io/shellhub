package models

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
)

type Member struct {
	ID      string          `json:"id,omitempty"`
	AddedAt time.Time       `json:"added_at"`
	Email   string          `json:"email" validate:"email"`
	Role    authorizer.Role `json:"role" validate:"required,oneof=administrator operator observer"`
}
