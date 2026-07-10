package entity

import (
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/uptrace/bun"
)

// UserInvitation maps to the user_invitations table.
type UserInvitation struct {
	bun.BaseModel `bun:"table:user_invitations,alias:user_invitations"`

	ID          string    `bun:"id,pk"`
	Email       string    `bun:"email"`
	Status      string    `bun:"status"`
	Invitations int       `bun:"invitations"`
	CreatedAt   time.Time `bun:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at"`
}

func UserInvitationFromModel(model *models.UserInvitation) *UserInvitation {
	return &UserInvitation{
		ID:          model.ID,
		Email:       model.Email,
		Status:      string(model.Status),
		Invitations: model.Invitations,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func UserInvitationToModel(e *UserInvitation) *models.UserInvitation {
	return &models.UserInvitation{
		ID:          e.ID,
		Email:       e.Email,
		Status:      models.UserInvitationStatus(e.Status),
		Invitations: e.Invitations,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}
