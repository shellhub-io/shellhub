package entity

import (
	"time"

	"github.com/uptrace/bun"
)

type UserInvitation struct {
	bun.BaseModel `bun:"table:user_invitations"`

	ID          string    `bun:"id,pk,type:uuid"`
	Email       string    `bun:"email"`
	Status      string    `bun:"status"`
	Invitations int       `bun:"invitations"`
	CreatedAt   time.Time `bun:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at"`
}
